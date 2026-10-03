# ci-steering-inject-drift-check — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

依赖不能：
  - 自引用（#1 不能依赖 #1）
  - 指向不存在的任务
  - 形成循环

完成后把 [ ] 改成 [x]；wave 视图用 `free-kiro task list <spec>`。
-->

- [x] #1 新建 `.githooks/pre-commit` 脚本（bash，≤ 30 行，可执行 0755）：仅在 staged diff 含 `.kiro/steering/*.md` 或 `internal/ide/templates/*.md` 时跑 `free-kiro steering inject`；inject 退出码 0 时自动 `git add CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md`；PATH 上无 free-kiro 或 inject 失败时 abort commit 并打印 1 行 stderr 提示
- [x] #2 修改 `Makefile`：在 `precommit` target 之前追加 `install-hooks` 与 `drift-check` 2 个 target，并把 `drift-check` 加进现有 `precommit` 依赖链（让本地 `make precommit` 也跑 drift 检查）[deps: #1]
- [x] #3 修改 `.github/workflows/ci.yml`：在现有 `lint-go` step 之后追加 1 个名为 `Drift check (steering inject)` 的 step，调用 `make drift-check`，drift 时让 job 失败 [deps: #2]
- [x] #4 跑 `make install-hooks` 配置本地 hooksPath；手动改 `.kiro/steering/tech.md` + `git commit` 验证 pre-commit 自动 inject + 自动 add；跑 `make drift-check` 验证干净状态 exit 0；最后 `git commit` 单独提交（commit message 标 `chore(ci)`） [deps: #1,#2,#3]
