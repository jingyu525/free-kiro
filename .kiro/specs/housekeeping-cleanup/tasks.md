# housekeeping-cleanup — Tasks

- [x] #1 建 internal/text/text.go，提供 RangeLines(text string) []string，写 test 覆盖空串/末尾无 newline/纯 newline [deps: none]
- [x] #2 lint: ears.go 删本地 rangeLines，import internal/text [deps: #1]
- [x] #3 taskgraph: parse.go 删本地 rangeLines + 自实现 atoi，改用 internal/text + strconv.Atoi [deps: #1]
- [x] #4 lint: baseline.go 删 dead code Empty() 方法 [deps: none]
- [x] #5 hooks: models/hook.go 加 Disabled bool 字段；envelope.go 读 Disabled；dispatch.go 删 timeout=0 歧义分支，改为 if h.Disabled [deps: none]
- [x] #6 spec: engine.go 删除 var _ = json.Marshal 占位符 + encoding/json import；同时把 stderr 直接打印改成 wrap 报出（用 cli/print helper 或新 helper） [deps: none]
- [x] #7 spec: engine.go Complete 中 _ = e.ws.ClearCurrent() 改为 wrap 报出 [deps: none]
- [x] #8 visualize: server.go URL() 改 net.SplitHostPort + IPv6 bracket 处理；补 IPv4/IPv6 测试 [deps: none]
- [x] #9 visualize: server.go godoc 同步（删 "fsnotify future enhancement" + 改 "200 ms" 注释为 shutdownTimeout） [deps: none]
- [x] #10 hooks: dispatch.go 删除 var _ = ferrors.New 占位符 + ferrors import（验证 dispatch 不再需 ferrors） [deps: none]
- [x] #11 verify: 跑既有测试 + 新增 test + go test -race + go vet + free-kiro lint 全绿 [deps: #2,#3,#4,#5,#6,#7,#8,#9,#10]