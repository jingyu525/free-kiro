# agents-template-coding-standards — Tasks

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

- [x] #1 在 `internal/ide/templates/agents_zh.md` 模板的"## 完成实现时"段之前插入"## 编码规范（SessionStart 必须先 Read）"段，内容对齐 `.kiro/AGENTS.md` 当前手写版本（含 .kiro/steering/coding-style.md / AGENT_RULES.md / POLICY.md 必读链接 + 3 个文档说明 + AGENT_RULES.md §1 的 4 条编号硬约束），链接路径用 `../../docs/`
- [x] #2 在 `internal/ide/templates/agents_en.md` 模板的 "When you finish implementing a spec" 段之前插入英文版 "## Coding Standards (SessionStart must Read first)" 段，结构对齐 zh 版（5 段：标题 + 必读链接 + 3 个文档说明 + hard-constraints header + 4 条编号约束），链接路径用 `../../docs/`
- [x] #3 跑 `free-kiro init --ide auto --overwrite-instructions` 验证 round-trip（diff 项目根 AGENTS.md 应含 "## 编码规范" 段），跑 `go test ./internal/ide/...` 确认现有测试不破坏，再跑 `free-kiro steering inject` 把 always 内容灌回 4 个项目根 IDE 指令文件（init 重写会清掉 inject 块）；最后 `git commit` 单独提交（commit message 标 `docs(template)`） [deps: #1,#2]
