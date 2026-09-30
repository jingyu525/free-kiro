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

- [x] #1 创建 `examples/todo-app/` 目录骨架：用 `free-kiro spec quick add-task-priority --prompt "todo app priority sorting"` 生成三份文档作为起点；手动精修 requirements.md（去掉 TODO 占位符、改为 ≥4 条 EARS AC）+ design.md（描述 priority 枚举 / JSON store / 排序算法）+ tasks.md（5 个任务、显式 2 个 wave 拓扑）
- [x] #2 在 `examples/todo-app/.kiro/specs/add-task-priority/tasks.md` 落地 `#1 data model + JSON store`、`#2 CLI arg parser`（Wave 1）以及 `#3 sort logic`、`#4 list renderer`、`#5 e2e test`（Wave 2），[deps: #1]
- [x] #3 跑 `free-kiro lint examples/todo-app` 验证样例 0 ERROR；同时跑 `free-kiro task list add-task-priority` 验证输出 ≥2 wave；把验证命令 + 输出快照写入 `docs/DEMO_VERIFICATION.md`（NEW，初始创建空文件 + 第 1 段），[deps: #1, #2]
- [x] #4 重写 README.md 第 2 节从"30 秒上手"为"5 分钟 demo"：新增 `cd examples/todo-app && ../../dist/free-kiro_*/free-kiro serve` 起步指令 + 顶部段落注入 "5 分钟端到端 demo" 字面量；保留原"30 秒上手"内容快照到 `docs/DEMO_VERIFICATION.md` 末尾（迁移审计），[deps: #3]
- [x] #5 扩 `internal/ide/ide.go`：增 `Cursor ID = "cursor"` + `Continue ID = "continue"` + `OpenCode ID = "opencode"` 三个常量；扩 `configPathFor` 覆盖 `~/.cursor/settings.json` / `~/.continue/config.json` / `~/.opencode/settings.json`；扩 `Parse` 接受 5 种字符串（case-insensitive）+ 未知值返回 `UsageError` 列出全部 5 个名字；同步 `All()` 返回 5 个 ID [deps: #1]
- [x] #6 扩 `internal/ide/ide_test.go` ≥15 case：5 IDE × 3 路径（Parse 合法字符串 / configPathFor 返回正确路径 / DetectAll DirExists 判定） + 3 case 验证 Parse 未知值返回 UsageError；跑 `go test -race ./internal/ide/...` 全绿，[deps: #5]
- [x] #7 新增 `internal/cli/demo.go`：实现 `runDemo(cmd, args)` 包含 5 步 onboarding 打印 + cwd 必须在 repo root 的检查（exit 3 + stderr 中文引导）+ `--no-color` 抑制 ANSI + `--ide <name>` 打印对应 hook 片段 + 30 秒 marker 去重；不引入新 Go 依赖 [deps: #5]
- [x] #8 新增 `internal/cli/demo_test.go` ≥6 case：cwd 正确（exit 0）/ cwd 错（exit 3）/ marker 写成功 / 30 秒去重生效 / `--ide claude-code` 打印 hook 片段 / `--no-color` 输出 0 ANSI 转义；跑 `go test -race ./internal/cli/...` 全绿，[deps: #7]
- [x] #9 在 `cmd/free-kiro/root.go`（或 cli 入口）注册 `demoCmd` 到根命令的 `AddCommand` 链；跑 `free-kiro --help` 看到 `demo` 子命令在列表中；同时把 `init --ide` flag 帮助文本更新为 "auto | claude-code | codebuddy | cursor | continue | opencode | none"，[deps: #7]
- [x] #10 新增 `.github/workflows/spec-lint.yml.example`：jobs 包含 `setup-free-kiro@v0` + `free-kiro lint` + `free-kiro doctor`；trigger 为 `pull_request` 到 `main`；用 `actionlint`（本地或 `npx actionlint`）验证 yaml 合法；[deps: #7]
- [x] #11 新增 `docs/CI_INTEGRATION.md`：5 步接入指南（复制 yaml 模板 / 重命名为 spec-lint.yml / commit / 查看 PR 检查 / 可选 status badge）；[deps: #10]
- [x] #12 跑全量门禁：`go vet ./...` 零警告、`go build ./...` 成功、`go test ./...` 全绿、`free-kiro lint top1-demo-onboarding` 零 ERROR、`free-kiro doctor` 7 项全过；手工冒烟 demo 完整流程并把命令 + 退出码 + 输出填入 `docs/DEMO_VERIFICATION.md`；同步更新 `README.md` 的命令速查表加入 `free-kiro demo`；[deps: #3, #4, #6, #8, #9, #11]