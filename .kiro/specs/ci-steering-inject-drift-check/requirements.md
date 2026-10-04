# ci-steering-inject-drift-check

新增 pre-commit hook + Makefile target + CI step 三层防御，自动跑
`free-kiro steering inject` 让项目根 4 个 IDE 指令文件（CLAUDE.md /
AGENTS.md / .cursorrules / .cursor/rules/free-kiro.md）始终与
`.kiro/steering/*.md` 中 `mode: always` 文档保持同步，避免开发者改
steering 后忘记跑 inject 导致 IDE 上下文过期。

## User Stories

- As a free-kiro 贡献者 I want `git commit` 前自动跑 `free-kiro steering inject` so that 我改了 `.kiro/steering/*.md` 之后 IDE 指令文件不会 drift，PR 也不会被 CI 拒收。

## Acceptance Criteria

- [AC-1] WHEN `.githooks/pre-commit` 在 git commit 阶段触发且 staged diff 含 `.kiro/steering/*.md` 或 `internal/ide/templates/*.md` THE SYSTEM SHALL 跑 `free-kiro steering inject` 并在 inject 退出码为 0 时自动 `git add` 4 个项目根 IDE 指令文件（CLAUDE.md / AGENTS.md / .cursorrules / .cursor/rules/free-kiro.md），共 4 个目标文件被 add 进 commit。
- [AC-2] WHEN pre-commit hook 跑 inject 后获得退出码 ≥ 1 THE SYSTEM SHALL 让 git commit 终止并保留退出码 1，并在 stderr 打印 1 行 `free-kiro steering inject failed; commit aborted`。
- [AC-3] WHEN `.githooks/pre-commit` 运行时 `.kiro/steering/*.md` 与 `internal/ide/templates/*.md` 在 staged diff 中均无改动 THE SYSTEM SHALL 跳过 inject 直接放行 commit，以避免每次 commit 都跑 ~50ms 的 inject。
- [AC-4] WHEN 开发者跑 `make install-hooks` THE SYSTEM SHALL 执行 `git config core.hooksPath .githooks` 并打印 1 行成功消息，开发者下一次 `git commit` 即自动跑 pre-commit hook。
- [AC-5] WHEN 开发者跑 `make drift-check` THE SYSTEM SHALL 跑 `free-kiro steering inject` 后用 `git diff --exit-code CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md` 检查 4 个文件是否有未提交改动，drift 时进程以退出码 1 结束并打印 1 行 `drift: <files>` 提示，clean 时进程以退出码 0 结束。
- [AC-6] WHEN `.github/workflows/ci.yml` 在 PR / push 阶段触发 THE SYSTEM SHALL 新增 1 个名为 `Drift check (steering inject)` 的 step 调用 `make drift-check`，让 CI job 在 drift 时 exit 1 并显示 4 个目标文件的 drift 列表。
- [AC-7] WHEN drift 检测发现 1 个文件改动时 THE SYSTEM SHALL 在 stderr 打印 1 行 `drift: <relpath>`（每文件 1 行，文件列表按字母序排列，最多 4 行），drift 检测本身 exit 1。
- [AC-8] THE SYSTEM SHALL 让 `.githooks/pre-commit` 脚本可执行（`chmod +x`），`make install-hooks` 在设置 hooksPath 后同时把 `.githooks/pre-commit` chmod 到 0755（兜底跨平台兼容性）。
- [AC-9] THE SYSTEM SHALL 让 pre-commit hook 仅依赖 PATH 上的 `free-kiro` 与 `git` 命令，不引入 bash 之外的 shell（保持 POSIX 兼容），单文件大小 ≤ 30 行。

## Out of Scope

- 不引入 lefthook / pre-commit framework / husky 等第三方 hook 管理工具——本 spec 走 `core.hooksPath` 单文件方案，保持工具链极简。
- 不在 pre-commit hook 里跑 `make drift-check`——`make` 在某些 CI 容器里可能未安装；hook 直接调 `free-kiro` 命令即可，CI 才走 Makefile（CI 容器有 make）。
- 不动 `.github/workflows/lint.yml` / `release.yml` / `spec-lint.yml.example`——只改 `ci.yml`。
- 不写 hook 跨平台适配（PowerShell / Windows .git 路径差异）——本仓库主开发平台是 macOS / Linux，Windows 贡献者跑 `bash .githooks/pre-commit` 即可。
- 不把 `.kiro/AGENTS.md`（workspace steering store 用）纳入 drift 检查——它由 steering store 读，不进 IDE agent loader；不属于 inject 的目标文件。
