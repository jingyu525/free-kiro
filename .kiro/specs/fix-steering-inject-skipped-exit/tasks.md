# fix-steering-inject-skipped-exit — Tasks

- [x] #1 修改 `internal/cli/steering_inject.go` 的 `runInject` 函数：删除 `if n := len(res.Skipped); n > 0 { ... os.Exit(n) }` 分支；删除常量 `injectExitMaxSkipped`；加注释说明 skipped 不影响 exit code，真错入口仍是函数开头的 DocCount==0 / glob 非法
- [x] #2 补单测：在 `internal/cli/steering_inject_test.go` 加 1 个 case 跑 `steeringInjectCmd()`，fixture 5 目标 4 个缺 marker，断言 cmd.Execute() 退出码为 0；如该测试文件不存在则新建 [deps: #1]
- [x] #3 修改 `.kiro/specs/steering-inject-to-ide/requirements.md` AC-3 文字："最终进程退出码等于被跳过的文件数（最多 5）" 改为 "skipped 仅 stderr warning 不影响退出码；仅 DocCount==0（exit 3）/ glob 非法（exit 4）才让进程 exit 非 0" [deps: #1]
- [x] #4 跑 `go test ./internal/cli/ ./internal/steering/` 确认现有测试不破坏 + 新增 case 通过；跑 `free-kiro spec sync steering-inject-to-ide` 捕获新 baseline；跑 `free-kiro spec complete fix-steering-inject-skipped-exit` 标记完成；最后 `git commit` 单独提交（commit message 标 `fix(steering)`） [deps: #1,#2,#3]
