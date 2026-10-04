# dashboard-ci-frontend-build

<!--
目标：让 CI 跑任何 Go job（test / lint-go / drift-check-steering / build /
smoke）之前先 build dashboard frontend（pnpm install + pnpm build），
产 `internal/visualize/static/dist/` 满足 `//go:embed all:static/dist/assets/*`
的产物要求。同时本地开发者跑 `make dashboard-dist` 一键复现。

零 Go 代码 / 零 embed 改动：本 spec 仅动 Makefile + ci.yml + 文档。
dashboard-frontend-react-vite-fsd 已交付 dist 产物的内容（index.html +
assets/*.js + assets/*.css），本 spec 负责把"产出 dist"接到 CI。
-->

## User Stories

- As a CI 维护者 I want `test` / `lint-go` / `drift-check-steering` / `build` / `smoke` 5 个 Go job 之前自动跑 dashboard frontend build so that `//go:embed all:static/dist/assets/*` 不再因 dist 缺失而失败，CI 恢复绿灯。
- As a 本地开发者 I want `make dashboard-dist` 一行命令复现 CI 行为 so that `go build ./cmd/free` 与 `free-kiro unit test` 不需要先手动 cd 到 `internal/visualize/static` 跑 npm。
- As a CI 优化者 I want pnpm 用 `actions/cache@v4` 持久化 `node_modules` so that 重复 PR 跑 `pnpm install --frozen-lockfile` < 30s，避免每次冷启 90s+。
- As a free-kiro 维护者 I want `dist/` 继续在 `.gitignore` 且不入 git so that build artifact 不污染仓库。

## Acceptance Criteria

### Makefile 新增 dashboard 入口

- [AC-1] THE SYSTEM SHALL `Makefile` 提供 `dashboard-dist` target，顺序执行 `pnpm install --frozen-lockfile` 与 `pnpm build`，退出码为 0 时 `internal/visualize/static/dist/index.html` 与 `internal/visualize/static/dist/assets/` 至少 4 个文件（`index-<hash>.js` / `index-<hash>.css` / `react-vendor-<hash>.js` / `query-vendor-<hash>.js`）同时存在。
- [AC-2] WHEN `pnpm` 不在 PATH 上 THE SYSTEM SHALL `make dashboard-dist` 退出码 1 并打印 `corepack enable pnpm` 或 `npm i -g pnpm@9` 的安装命令。
- [AC-3] THE SYSTEM SHALL `Makefile` 提供 `ci` target，顺序执行 `dashboard-dist` + `test` (go test) + `lint-go` (golangci-lint) + `build` (go build)，任意一步非 0 退出立即终止后续步骤。
- [AC-4] WHERE `internal/visualize/static/node_modules` 不存在 THE SYSTEM SHALL `make dashboard-dist` 自动跑 `pnpm install --frozen-lockfile`，存在则跳过 install 步骤，直接跑 `pnpm build`。

### ci.yml 5 个 Go job 前置 frontend build

- [AC-5] WHEN PR 触发 ci.yml 的 `test` job THE SYSTEM SHALL 在 `go vet` 与 `go test` 之前先跑 dashboard frontend build（即 `make dashboard-dist` 等价命令），build 失败时整个 job 退出码 ≥ 1。
- [AC-6] WHEN PR 触发 ci.yml 的 `lint-go` job THE SYSTEM SHALL 在 `golangci-lint-action@v8` 之前先跑 dashboard frontend build。
- [AC-7] WHEN PR 触发 ci.yml 的 `drift-check-steering` job THE SYSTEM SHALL 在 `go install ./cmd/free-kiro` 之前先跑 dashboard frontend build。
- [AC-8] WHEN PR 触发 ci.yml 的 `build` job THE SYSTEM SHALL 在 cross-compile `go build ./cmd/free-kiro` 之前先跑 dashboard frontend build。
- [AC-9] WHEN PR 触发 ci.yml 的 `smoke` job THE SYSTEM SHALL 在 `go build -o /tmp/free-kiro ./cmd/free-kiro` 之前先跑 dashboard frontend build。

### pnpm 缓存与 Node 版本

- [AC-10] THE SYSTEM SHALL ci.yml 用 `actions/setup-node@v4` 设置 `node-version: '20'`（与 `internal/visualize/static/package.json` 的 `engines.node >= 20` 对齐）。
- [AC-11] THE SYSTEM SHALL ci.yml 用 `actions/cache@v4` 缓存 `internal/visualize/static/node_modules`，cache key `${{ runner.os }}-pnpm-${{ hashFiles('internal/visualize/static/pnpm-lock.yaml') }}`，cache path `internal/visualize/static/node_modules`。
- [AC-12] WHEN cache hit 时 THE SYSTEM SHALL `pnpm install --frozen-lockfile` 退出码 0 且耗时 ≤ 30 秒（CI log 实测）；cache miss 时 ≤ 180 秒。

### 文档同步

- [AC-13] THE SYSTEM SHALL `docs/STEERING.md` § "dashboard frontend dev / CI" 段落说明本地用 `make dashboard-dist`、CI 自动跑同一目标。
- [AC-14] THE SYSTEM SHALL `Makefile` 的 `help` target 输出包含 `dashboard-dist` 与 `ci` 两个新 target。

## Out of Scope

- 修复 `pnpm build` TypeError "Cannot read properties of null" runtime bug（属于 dashboard-frontend-components spec 范围）
- 删除 `static/legacy/index.html`（属于 dashboard-frontend-foundation 既有保留路径）
- React 19 升级或 Server Components 探索（属于 dashboard-frontend-react-vite-fsd 锁定的 React 18.x 范围）
- GoReleaser release.yml 集成 frontend build（dashboard frontend 不进 binary 包，仅 //go:embed 进 Go binary）
- axe-core a11y 自动化校验（属于 dashboard-frontend-components）