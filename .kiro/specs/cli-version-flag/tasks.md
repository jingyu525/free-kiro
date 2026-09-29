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

- [x] #1 在 internal/cli/root.go 给 rootCmd 设置 Version 字段（= buildVersion）并在 init() 里调用 SetVersionTemplate 输出 `free-kiro version <v> (commit <sha>, built <date>)`，commit/date 为空时退化为 "unknown"
- [x] #2 新增 internal/cli/root_test.go 单元测试：断言 rootCmd.Version 非空、`--version` 触发后 stdout 单行输出含 "free-kiro version"、exit 0、dev build 含 "dev" 字面量 [deps: #1]
- [x] #3 跑 `go test ./internal/cli/...` 全绿；`go vet ./...` 无警告；`free-kiro lint cli-version-flag` 仍然 OK [deps: #2]
- [x] #4 跑 `goreleaser build --snapshot --clean` 出 snapshot 二进制，手工执行 `dist/free-kiro_*/free-kiro --version` 确认 ldflags 注入的 version/commit/date 都正确显示 [deps: #3]
