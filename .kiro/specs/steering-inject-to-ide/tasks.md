# steering-inject-to-ide — Tasks

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

- [x] #1 实现 `internal/steering/inject.go` 的 Store.InjectAll 核心逻辑（LoadAll→过滤 always→排序→拼装 block→按 marker 替换→写回），含 InjectResult / InjectSkip / DefaultInjectTargets / marker 常量定义，覆盖单元测试 [internal/steering/inject_test.go]
- [x] #2 在 `internal/ide/templates/instructions_zh.md` 与 `instructions_en.md` 模板的"## 编码规范"段后追加 `<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->` 包裹的空块（含 `InjectBlockHeader` 注释行），并修改 `internal/cli/init.go` 让 init 写出这些 marker
- [x] #3 在 `internal/cli/steering.go` 的 `steeringCmdFactory()` 注册 `steeringInjectCmd()` 子命令，支持 `--dry-run` / `--only <glob>` 两个 flag，按 InjectResult 决定退出码 [deps: #1]
- [x] #4 实现 `steeringInjectCmd()` RunE：解析 flag、调 Store.InjectAll、stdout 打印 dry-run 内容或成功消息、stderr 打印 skipped 文件警告、按 Skipped 数返回退出码（最多 5），含单测 [internal/cli/steering_inject_test.go] [deps: #1,#3]
- [x] #5 在 `testdata/inject/{before,expected}/` 准备 5 个目标文件的 fixture（每对：before 含 marker 块与历史内容；expected 含正确替换结果）+ 1 个缺 marker 的 fixture，覆盖集成测试 [deps: #1,#3]
- [x] #6 重跑 `free-kiro init --ide auto --overwrite-instructions` 让本仓库 5 个 IDE 指令文件获得 marker 块，再跑 `free-kiro steering inject` 把当前 always 内容（product/structure/tech）灌进去；最后 `git commit` 单独提交 [deps: #2,#4]
- [x] #7 更新 `docs/STEERING.md`（§"在 agent 里集成" 后追加"## 自动注入到 IDE 指令文件"小节）、`docs/HOOKS.md`（§"Agent instructions per IDE" 加一条 marker 说明）、`CLAUDE.md`（顶部加一段指向 steering 内容）；单独 commit [deps: #6]
