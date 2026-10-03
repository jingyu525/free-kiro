# instructions-template-inject-context — Tasks

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

- [x] #1 在 `internal/ide/templates/instructions_zh.md` 模板的"## 编码规范（SessionStart 必须先 Read）"段之前插入"## 项目上下文（自动注入的 steering）"段，含 2 个 marker 链接（`<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->`）+ always-mode 文档列表（product.md / structure.md / tech.md）+ 1 句编辑指引 + 1 个 `docs/STEERING.md` §引用
- [x] #2 在 `internal/ide/templates/instructions_en.md` 模板的 "## Coding standards (SessionStart must Read first)" 段之前插入英文版 "## Project context (auto-injected steering)" 段，结构对齐 zh 版（4 段：标题 + marker 包裹说明 + 总是模式列表 + 编辑指引），保留 `§"自动注入到 IDE 指令文件"` 中文引用指针（与 zh 版一致）
- [x] #3 跑 `free-kiro init --ide auto --overwrite-instructions` 验证 round-trip（diff 项目根 CLAUDE.md 应含 "## 项目上下文" 段）、跑 `go test ./internal/ide/...` 确认现有测试不破坏、跑 `free-kiro steering inject` 灌回 always 内容、最后 `git commit` 单独提交（commit message 标 `docs(template)`） [deps: #1,#2]
