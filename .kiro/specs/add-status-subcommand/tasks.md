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

- [x] #1 新增 internal/cli/status.go：定义 statusCmd，Cobra flag 包含 --json；Run 调 workspace.ListSpecs + ActiveSpec + models.Summary 聚合数据，默认走 text/tabwriter 表格输出
- [x] #2 statusCmd 加 --json 输出分支：复用 internal/models.Summary 的 JSON 结构（与 serve /api/summary 同源），输出到 stdout [deps: #1]
- [x] #3 新增 internal/cli/status_test.go：TestStatusHumanOutput + TestStatusJSONOutput + TestStatusNoWorkspace + TestStatusNoSpecs + TestStatusPointerMissing + TestStatusIsReadOnly 六个用例（用 t.TempDir 建隔离 .kiro/） [deps: #1, #2]
- [x] #4 在 internal/cli/root.go init() 注册 statusCmd：rootCmd.AddCommand(statusCmdFactory())，并把 statusCmd 加进根 help 文案 [deps: #1]
- [x] #5 跑 go test ./internal/cli/... 全绿 + go vet ./... 无警告 + free-kiro lint cli-version-flag + free-kiro lint add-status-subcommand 都 OK [deps: #3, #4]
- [x] #6 跑 free-kiro watch --preset reactive --debounce 300ms --verbose，touch 一个 spec 文件触发事件，断言 status --human 不再报 "unknown command"，watch 流程全部 exit 0 [deps: #5]

