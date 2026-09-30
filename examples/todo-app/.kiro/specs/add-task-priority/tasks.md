<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B
-->

- [ ] #1 在 `examples/todo-app/store.go` 实现 `Task` 结构体 + `Priority` 类型 + `ParsePriority` + `TodoStore` 结构 + `Load`/`Save`/`Append` 方法；使用 `t.TempDir()` 写测试覆盖 缺文件 / 原子写 / round-trip 三个 case
- [ ] #2 在 `examples/todo-app/main.go` 用 cobra 搭 CLI 框架：`rootCmd` + `addCmd`（读 `--priority` flag，调用 `ParsePriority`，调用 `store.Append`）+ `listCmd`（调用 `store.Load` + 排序后打印）；不引入第三方依赖
- [ ] #3 在 `examples/todo-app/sort.go` 实现 `SortTasks([]Task)` 按 priority asc 然后 created_at asc 稳定排序；写 ≥4 个 case 的表驱动测试 [deps: #1]
- [ ] #4 在 `examples/todo-app/main_test.go` 用 `cmd.SetArgs` + `cmd.Execute()` 做 e2e：`add` 默认 P2 / `add --priority X` 拒绝 / `list` 排序输出 / `list` 空文件输出 [deps: #1, #2, #3]
- [ ] #5 跑 `cd examples/todo-app && go build ./...` + `go test ./...` 全绿；手工冒烟：`go run . add "buy milk" --priority P0` + `go run . list` 看到 P0 优先 [deps: #4]