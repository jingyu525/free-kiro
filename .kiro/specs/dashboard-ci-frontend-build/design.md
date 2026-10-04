# dashboard-ci-frontend-build — Design

<!--
本 spec 是"devops glue"，不改 Go 代码也不改 frontend 代码。
目标是把 dashboard-frontend-react-vite-fsd 已经交付的 dist 产物接到
CI 5 个 Go job 之前，避免 `//go:embed all:static/dist/assets/*` 因 dist
缺失而失败。

30 天后回看要点：
- ci.yml 5 个 job 各加 1 个 "Build dashboard frontend" step（共享相同 setup）
- Makefile 加 `dashboard-dist` + `ci` 两个 target，本地 `make ci` 等价于 CI
- pnpm 用 actions/cache 持久化 node_modules，cache hit < 30s
-->

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          CI / 本地 CI 入口                            │
│                                                                     │
│   GitHub Actions workflow (.github/workflows/ci.yml)                 │
│   ┌────────────────────────────────────────────────────────────┐     │
│   │ 5 个 Go job（test / lint-go / drift-check-steering /       │     │
│   │ build / smoke）每个 job 内部顺序：                          │     │
│   │   1. setup-node@v4 (node 20)                               │     │
│   │   2. actions/cache@v4  缓存 node_modules                   │     │
│   │   3. corepack enable pnpm                                 │     │
│   │   4. make dashboard-dist   ← 新增                          │     │
│   │   5. (各 job 原 Go 命令)                                   │     │
│   └────────────────────────────────────────────────────────────┘     │
│                                                                     │
│   Makefile (本地入口)                                               │
│   ┌────────────────────────────────────────────────────────────┐     │
│   │ dashboard-dist   → pnpm install --frozen-lockfile +        │     │
│   │                    pnpm build                              │     │
│   │ ci               → dashboard-dist + test + lint-go +       │     │
│   │                    build                                   │     │
│   └────────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────┘
                                ↓
                                ↓
        internal/visualize/static/dist/{index.html,assets/*}  ← 产物
                                ↓
        //go:embed all:static/dist/assets/* 满足
```

**关键决策**：

| 决策 | 选项 | 选择 | 理由 |
|---|---|---|---|
| frontend build 触发位置 | ci.yml 5 job 内联 vs composite action 抽取 | **composite action 抽取**（`action.yaml` in `.github/actions/build-dashboard/`） | 5 job 共享同一段 4-step setup，避免重复定义；后续若加新 Go job 直接复用 |
| pnpm 安装方式 | corepack enable vs npm i -g pnpm | **corepack enable pnpm** | 与 `package.json` 的 `packageManager: pnpm@9.x` 自动对齐；不污染 runner 镜像 |
| cache key | hash pnpm-lock.yaml vs hash package.json | **hash pnpm-lock.yaml** | 仅当依赖图变化才失效；package.json 其他字段（如 scripts）改了不重 install |
| cache path | `~/.local/share/pnpm/store` vs `internal/visualize/static/node_modules` | **`internal/visualize/static/node_modules`** | 与 `pnpm install` 实际写入路径一致；size ≤ 200MB 可控 |
| Makefile dashboard-dist | 调用 `make` 嵌套 vs 直接 shell 命令 | **直接 shell 调用 pnpm** | 减少一层 make 嵌套依赖；target 注释清晰说明步骤 |

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `.github/actions/build-dashboard/action.yaml` | composite action：setup-node@v4 + cache + corepack + make dashboard-dist | 单一 `runs.using: composite` + 5 个 `steps` |
| `Makefile` `dashboard-dist` target | 本地一键复现：pnpm install --frozen-lockfile + pnpm build | `make dashboard-dist` |
| `Makefile` `ci` target | 本地 CI 入口：dashboard-dist + test + lint-go + build | `make ci` |
| `.github/workflows/ci.yml` 5 job 改造 | 每个 job 加 1 步 `uses: ./.github/actions/build-dashboard` | `test` / `lint-go` / `drift-check-steering` / `build` / `smoke` |

## Data Model

无新数据结构。本 spec 不动 Go 代码，不动 embed 指令。

## Error Handling

| 失败点 | 表现 | 修复建议 |
|---|---|---|
| `pnpm` 不在 PATH | Makefile dashboard-dist exit 1 + 打印 `corepack enable pnpm` 或 `npm i -g pnpm@9` | 按提示安装 |
| `pnpm install` lockfile drift | 退出码 ≥ 1 | 重新跑 `pnpm install` 在非锁文件里，确认后 `pnpm-lock.yaml` 入库 |
| `pnpm build` 失败（TypeError / import 错） | 退出码 ≥ 1，dist 不更新 | 看 vite log；dashboard-frontend-components 跟 TypeError |
| CI cache 失效（pnpm-lock.yaml 改） | install 重跑 < 180s | 接受；首次 PR cold |
| `actions/cache@v4` post step 失败 | 仅 warn，不阻断 job | 接受；runner 偶发网络 |

## Testing Strategy

| 层 | 工具 | 覆盖什么 |
|---|---|---|
| 本地 smoke | `make dashboard-dist && ls internal/visualize/static/dist/` | Makefile target 正确产出 |
| 本地 CI 等价 | `make ci` | 与 GitHub Actions 5 job 等价 |
| GitHub Actions | ci.yml 5 job 全绿 | 5 个 Go job 不再因 embed 失败 |
| regression | `git diff origin/main -- .github/workflows/ci.yml Makefile` | 本 spec 改动 ≤ 2 个文件 + 1 个新 composite action |

## Migration / Rollout

**展开**：

1. **PR #1（本次）**：本 spec 三件套 + 实施代码合并
   - 新增 `.github/actions/build-dashboard/action.yaml`
   - 修改 `.github/workflows/ci.yml` 5 个 job
   - 修改 `Makefile` 加 `dashboard-dist` + `ci` target
   - 修改 `docs/STEERING.md` 同步新流程
2. **PR 合入 main 后**：下次 push main 触发 ci.yml 全绿，验证 dashboard spec #33 留的"Makefile dashboard-dist target 未建"债务清除

**回滚**：

- `git revert <commit>` 即可回滚全部改动（composite action + ci.yml + Makefile + docs）
- 无数据库迁移、无 schema 变更

**特性开关**：不需要（CI 直接生效）

**分阶段**：不需要（单一 PR 即可，dashboard frontend 在 main 已经是必须产物）