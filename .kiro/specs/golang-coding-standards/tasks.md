# golang-coding-standards — Tasks

- [ ] #1 起草 .kiro/steering/coding-style.md 骨架与命名约定章节（含 ✅/❌ 示例）
- [ ] #2 补充 .kiro/steering/coding-style.md 错误处理与并发章节 [deps: #1]
- [ ] #3 补充 .kiro/steering/coding-style.md 接口设计与测试章节 [deps: #2]
- [ ] #4 补充 .kiro/steering/coding-style.md 注释/文档、依赖管理、AI 协作章节 [deps: #3]
- [ ] #5 编写 .golangci.yml 启用 govet / staticcheck / errcheck / gofmt / goimports / revive [deps: #1]
- [ ] #6 在 Makefile 新增 lint-go target 调用 golangci-lint run ./... [deps: #5]
- [ ] #7 在 .github/workflows/ci.yml 新增 lint-go job，失败阻断 merge [deps: #6]
- [ ] #8 更新 CONTRIBUTING.md 增加"编码规范"小节与 .kiro/steering/coding-style.md 链接 [deps: #4]
- [ ] #9 更新 .kiro/AGENTS.md SessionStart 引导附加"先读 .kiro/steering/coding-style.md" [deps: #4]
- [ ] #10 Dogfooding 验证：本 spec 落地 PR 自身必须 make lint-go 通过 [deps: #5,#6,#7]
