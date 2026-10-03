# unify-steering-docs-and-code-reality — Design

## Architecture

3 个文档作为独立子任务（product.md / tech.md / structure.md），按文件分散，无相互依赖：

| 模块 | 职责 | 文件 |
|---|---|---|
| product.md 修订 | L26 状态机、L27 EARS 数量、L33 advisory 数量、L36 4 widget、L38-41 前置依赖 / watch / 自升级 / fsnotify fallback（AC-1 ~ AC-9） | `.kiro/steering/product.md` |
| tech.md 修订 | 移除 `~/.prev` 段落、降级覆盖率阈值、降级性能预算、改 GoReleaser 平台数、补 fsnotify fallback 说明（AC-10 ~ AC-14） | `.kiro/steering/tech.md` |
| structure.md 升级 | 顶层布局保留 + `internal/cli` 命令文件抽象 + `internal/lint` 按职责分组 + `internal/visualize/static` 补完整结构 + 根目录增补 + 3 处文档漂移修正（AC-15） | `.kiro/steering/structure.md` |

不引入新依赖（free-kiro 精简哲学：Go stdlib + 3 个直接依赖保持不变）。

## Components

### 1. product.md 字段级修订（AC-1 ~ AC-9）

**AC-1 状态机描述**：把 L26 线性箭头改为"主流路径为线性；planning 阶段（requirements / design / tasks）可互转；APPROVED → TASKS 是合法回填；完整合法迁移表见 `internal/models/phase.go` 的 `allowed` 切片"。

**AC-2 EARS 数量**：把 L27 "5 种 EARS 句式 + 无条件基线" 改为 "5 种 EARS 触发句式（when / while / where / unless / if-then）+ 1 个 ubiquitous 基线，共 6 类正则分支"，并附"`internal/lint/ears.go` 的 `EARSRe` 含 6 个 alternation"作可追溯引用。

**AC-3 advisory 数量**：L33 改为 4 类 finding 代码（`vague-language` / `duplicate-acceptance-criteria` / `uncovered-acceptance-criteria` / `tasks-without-requirements`），附 "`internal/spec/analyze.go:53-115` 是真实实现" 作可追溯引用。

**AC-4/5 前置依赖**：L38 issue→spec 行加 "前置依赖 `gh` CLI（`brew install gh` + `gh auth login`）"；L39 PRD→spec 行区分 `--from-prd`（仅需网络）与 `--from-browser`（前置 `bsk` CLI + Chromium 登录态）。

**AC-6 watch presets**：L40 列出全部 5 个 `--preset`（`default` / `lint` / `status` / `reactive` / `full`），标注 `reactive = lint + status`。

**AC-7 4 widget**：L36 明确指代 `stat-card` / `sparkline-cell` / `phase-badge` / `waves-progress`，注 "另有 9 个辅助 widget 见 `internal/visualize/static/src/widgets/`"。

**AC-8 Windows re-exec**：L41 加 "Windows 下不自动 re-exec，升级后需手动重跑 `free-kiro upgrade` 或重启进程"。

**AC-9 fsnotify fallback**：L35 加 "fsnotify 在容器 / 网络 FS 不可用时降级为 2s mtime 轮询（见 `internal/visualize/server_sse.go:154-170` `watchChangesFallback`）"。

### 2. tech.md 字段级修订（AC-10 ~ AC-14）

**AC-10 自升级回滚**：L69 删除"失败回滚路径：保留上一份二进制到 `~/.local/share/free-kiro/.prev`"，改为 "升级失败行为：失败时不替换旧二进制，进程继续以旧版本运行；用户可重新运行 `free-kiro upgrade --check` 排查"。

**AC-11 覆盖率阈值**：L38 改为 "建议目标：核心 lint 引擎 ≥ 80%、其他 ≥ 60%；CI 当前仅打印总数 `go test -cover`，不强制阈值"。

**AC-12 性能预算**：L73-76 改为 "目标预算（待基准验证）：CLI 冷启动 < 100ms / `free-kiro lint`（10 spec）< 500ms / SSE 推送 < 200ms / dashboard 首屏 < 500ms；当前未在 `internal/cli/` `internal/lint/` `internal/visualize/` 中定义命名常量，也无对应 `go test -bench` 测试"。

**AC-13 GoReleaser 平台数**：L57 改为 "跨 5 个平台：darwin / linux × amd64+arm64；windows × amd64（windows+arm64 在 `.goreleaser.yaml` 的 `ignore` 段中当前被跳过）"。

**AC-14 fsnotify fallback**：L75 加 "fsnotify 不可用时降级为 2s mtime 轮询，端到端延迟可达 2s+"。

### 3. structure.md β 方案升级（AC-15）

按"按职责分组 + 代表性文件"原则重写：

**(a) 顶层布局保留**：原 ASCII tree 框图作为骨架保留，仅替换内部文本。

**(b) `internal/cli/` 抽象**：改为"~22 个 cobra 命令文件，每个子命令一文件；命令清单见 `cmd/free-kiro/main.go` 的 `rootCmd.AddCommand(...)` 注册"。

**(c) `internal/lint/` 抽象**：按职责分组列出代表性文件：
- `ears.go` — EARS 正则定义
- `bugfix.go` — bugfix 三段式校验
- `requirements.go` & `tasks.go` — 结构 lint
- `linter.go` — lint 总入口
- `baseline.go` — 基线规则
- `quality.go` — 质量指标（semantic gates）

**(d) `internal/visualize/static/` 补完整结构**：明确"完整的 Vite + FSD 工程，含 `src/` `dist/` `scripts/` `node_modules/`"。

**(e) 根目录增补**：在顶层布局段列出 `LICENSE` `CONTRIBUTING.md` `Makefile` `AGENTS.md` `CLAUDE.md` `examples/`。

**(f) 3 处文档漂移修正**：
- `contrib/` 当前不含 Homebrew formula（`brew` 分发走 release.yml，不在 `contrib/`）
- `internal/skill/` 与 `contrib/skills/` 无软链 / 拷贝关系（skill bundle 是独立脚本化分发）
- `ide/templates/` 实际是 `agents_*.md` / `instructions_*.md`，不是 `.tmpl`

## Data Model

无新数据结构。改动全部在 markdown 文档（`.kiro/steering/*.md`）内。

不动 `.baseline.json` schema、不动 linter Issue 结构、不动 `internal/visualize/` 的任何数据结构。

## Error Handling

- **grep 验证失败**：直接 `grep` 退出码 ≥ 1 → CI 退出 1；输出即"哪条 AC 在哪一行失败"的自然证据
- **文档 lint 漂移**：steering 文档不在 `free-kiro lint` 扫描范围（当前行为），但 `go build ./...` + `go test ./...` 仍会跑（确保没有任何意外的 .go 改动）
- **回滚路径**：3 个文档改动互不耦合，`git revert <commit>` 即可恢复任意一个

## Testing Strategy

| 层 | 覆盖 | 文件 / 命令 |
|---|---|---|
| 单元（lint 引擎） | 不涉及（无 .go 改动） | N/A |
| 端到端 grep 校验 | AC-1 ~ AC-15 的 grep 模式逐条断言 | `bash .kiro/specs/unify-steering-docs-and-code-reality/scripts/verify-docs.sh`（新建脚本，Wave 1 task #4） |
| 端到端构建 | 跑 `go build ./...` + `go test ./...` 确认无回归 | shell 验证 |
| 端到端 lint | 跑 `free-kiro lint unify-steering-docs-and-code-reality` 确认 spec 自身 OK | shell 验证 |

**覆盖率门槛**：无 .go 改动，无需 `go test -cover` 检查。

## Migration / Rollout

**默认上线**：合并到 main 即生效。文档改动让所有 `free-kiro spec generate` 注入 steering 后的 IDE 指令文件受益。

**回滚**：3 个文档改动互不耦合，回滚任何一个文件即可恢复。零数据迁移。

**兼容影响**：
- product.md 字段修订 → 下次 `free-kiro steering inject` 自动同步到 `CLAUDE.md` 等 IDE 指令文件
- tech.md 字段修订 → 同上
- structure.md 结构升级 → 仅影响人类阅读，不影响任何自动注入逻辑（steering inject 只取 `mode: always` 的 product.md / structure.md / tech.md body 全文）
- 零兼容性问题，零 schema 变更，零 CLI 行为变更
