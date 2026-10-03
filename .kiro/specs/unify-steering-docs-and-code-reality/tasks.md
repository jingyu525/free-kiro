# unify-steering-docs-and-code-reality — Tasks

每完成一条把 `[ ]` 改成 `[x]`；wave 视图用 `free-kiro task list unify-steering-docs-and-code-reality`。

依赖图：
- **Wave 1**：4 个文档 / 脚本改动互不依赖，全部并行（#1 product.md / #2 tech.md / #3 structure.md / #4 verify 脚本）
- **Wave 2**：端到端验证依赖 Wave 1 全部 4 个任务（#5）
- **Wave 3**：spec 状态机推进依赖 Wave 2（#6）

## Wave 1 — 文档修订（全部可并行）

- [ ] #1 改 `.kiro/steering/product.md`：(a) L26 状态机改为"主流路径为线性；planning 阶段可互转；APPROVED → TASKS 是合法回填"；(b) L27 EARS 改为"5 种触发句式 + 1 个 ubiquitous 基线，共 6 类正则分支"；(c) L33 advisory 改为 4 类 finding 代码；(d) L38-39 issue→spec / PRD→spec 行加 `gh` / `bsk` 前置依赖声明；(e) L40 watch 列出全部 5 个 preset；(f) L36 "4 widget" 指明 `stat-card` / `sparkline-cell` / `phase-badge` / `waves-progress`；(g) L41 加 Windows re-exec 限制说明；(h) L35 加 fsnotify fallback 说明 [no deps]

- [ ] #2 改 `.kiro/steering/tech.md`：(a) L69 删除 `~/.prev` 段落，改为"升级失败行为：失败时不替换旧二进制，进程继续以旧版本运行"；(b) L38 覆盖率阈值改为"建议目标；CI 当前仅打印总数，不强制"；(c) L73-76 性能预算改为"目标预算（待基准验证）；当前无命名常量、无基准测试"；(d) L57 GoReleaser 平台数改为"5 个"并标注 windows+arm64 忽略；(e) L75 加 fsnotify fallback 2s+ 说明 [no deps]

- [ ] #3 改 `.kiro/steering/structure.md`：走 β 方案（按职责分组 + 代表性文件）。(a) 顶层布局保留；(b) `internal/cli/` 改为"~22 个 cobra 命令文件，命令清单见 `cmd/free-kiro/main.go` 的 `rootCmd.AddCommand(...)`"；(c) `internal/lint/` 按职责列出 6 个代表性文件（`ears.go` / `bugfix.go` / `requirements.go` / `tasks.go` / `linter.go` / `baseline.go` / `quality.go`）；(d) `internal/visualize/static/` 补"完整的 Vite + FSD 工程，含 src/ dist/ scripts/ node_modules/"；(e) 顶层布局段增补 ≥ 6 个根目录文件（`LICENSE` / `CONTRIBUTING.md` / `Makefile` / `AGENTS.md` / `CLAUDE.md` / `examples/`）；(f) 3 处文档漂移修正（`contrib/` 无 Homebrew formula；`internal/skill/` 与 `contrib/skills/` 无软链；`ide/templates/` 是 `.md` 不是 `.tmpl`）[no deps]

- [ ] #4 新建 `.kiro/specs/unify-steering-docs-and-code-reality/scripts/verify-docs.sh`：bash 脚本实现 AC-1 ~ AC-15 的 grep 端到端校验（每条 AC 一个 if-else，匹配 grep 返回 0 时打印 "PASS: AC-N"，否则打印 "FAIL: AC-N" 并 exit 1）。脚本要可被 `bash verify-docs.sh` 直接调用，零外部依赖（只用 `grep` / `sed` / `cat`）。[no deps]

## Wave 2 — 端到端验证（依赖 Wave 1）

- [ ] #5 跑以下 4 项端到端校验：(a) `bash .kiro/specs/unify-steering-docs-and-code-reality/scripts/verify-docs.sh` 全部 AC PASS；(b) `go build ./...` 退出 0；(c) `go test ./...` 退出 0；(d) `free-kiro lint unify-steering-docs-and-code-reality` 退出 0。任一项失败则修复并重跑。期望 4 项全绿，缺一不可。[deps: #1, #2, #3, #4]

## Wave 3 — 审批上线（依赖 Wave 2）

- [ ] #6 推进 spec 状态机：`free-kiro spec approve unify-steering-docs-and-code-reality` → `free-kiro spec start unify-steering-docs-and-code-reality` → `git add .kiro/steering/` + `git commit -m "docs(steering): 对齐 product.md / tech.md / structure.md 与代码实现"` + `git push` → `free-kiro spec complete unify-steering-docs-and-code-reality`。[deps: #5]
