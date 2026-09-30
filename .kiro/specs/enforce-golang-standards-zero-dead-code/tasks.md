# enforce-golang-standards-zero-dead-code — Tasks

依赖不能自引用、不能指向不存在任务、不能成环。

## Wave 1 — 配置先行（无依赖，可立即开工）

- [x] #1 在 `.golangci.yml` 的 `linters.enable:` 数组下追加 `- unused`（独立
  linter，不是 staticcheck 子检查）；保留既有 `staticcheck` / `errcheck` /
  `revive` / `govet` 启用项不变 [deps: ]

- [x] #2 在 `docs/CODING_STYLE.md` 第 8 章追加"零死代码（U1000
  zero-tolerance）"段落，与既有"零 `//nolint` 豁免"条款并列 [deps: ]

## Wave 2 — 自动修复（依赖 Wave 1）

- [x] #3 跑 `gofmt -s -w .` + `goimports -local
  github.com/jingyu525/free-kiro -w .` 一次性格式化 13 个文件
  （internal/lint/baseline.go 等） [deps: #1]

- [x] #4 修 `internal/ide/ide.go:426` 的 `defer f.Close()` → `defer func() { _ = f.Close() }()`
  闭包延迟吞错（保持 scanner 能读到 EOF），加注释说明吞错原因
  [deps: #1]

- [x] #5 修 `internal/hooks/hooks_test.go:400` 未使用参数 `p` → rename 为
  `_` [deps: #1]

- [x] #6 修 `internal/lint/quality.go:171` 参数 `min` → `minAC`，消除
  builtin `min` 遮蔽；同步更新函数体（line 180 `strconv.Itoa(min)`）+
  `internal/spec/analyze.go` 无 caller 变更 [deps: #1]

## Wave 3 — 删死代码（依赖 Wave 2；按 4 个包并行）

- [x] #7 删 `internal/cli/stubs.go` 的 3 个 U1000 函数：
  `attachPhaseFlags` / `stubCmd` / `stubFor` [deps: #3,#4,#5,#6]

- [x] #8 删 `internal/spec/generator.go` 的 6 个 U1000 函数：
  `writeIfMissing` / `ensureMeta` / `listSpecDocs` / `firstDocExists` /
  `fmtMissing` / `phaseForDoc` [deps: #3,#4,#5,#6]

- [x] #9 删 `internal/steering/{frontmatter.go:115, store.go:175}` 的
  `splitLines` + `deriveName`（已由 `internal/text/text.SplitLines` 等
  替代） [deps: #3,#4,#5,#6]

- [x] #10 删 `internal/upgrade/upgrade.go:33` 常量 `httpTimeout` + 删
  `internal/visualize/mermaid.go:122` 函数 `taskWaves`（已由
  `internal/visualize/taskgraph.go` 接管） [deps: #3,#4,#5,#6]

- [x] #11 删 `internal/visualize/server.go:45,180` 的 `osDirEntry` 类型
  别名 + `(*Server).loadWorkspace` 方法 [deps: #3,#4,#5,#6]

- [x] #12 删 `internal/visualize/server_static.go` 整文件（4 个 U1000 函数
  `atoiLocal` / `detectBrowserAutoOpen` / `openBrowser` /
  `detectBrowserOpen` 全部 dead code，无 build tag 隔离） [deps: #3,#4,#5,#6]

## Wave 4 — 复核（依赖 Wave 3）

- [x] #13 跑 `go test ./...` 确认全绿；任一 failure 必须回溯到 Wave 2 / 3
  排查 [deps: #7,#8,#9,#10,#11,#12]

- [x] #14 跑 `golangci-lint run ./... --timeout=5m` 确认退出码 0 + 0
  issue；不通过则回到 Wave 2 / 3 继续修 [deps: #13]

- [x] #15 跑 `staticcheck ./...` 确认退出码 0；与 #14 互为冗余校验
  [deps: #13]

- [x] #16 跑 `free-kiro lint` 确认所有 spec（含本 spec）通过；
  通过后调用 `free-kiro spec complete
  enforce-golang-standards-zero-dead-code` 把 spec 标记为 DONE
  [deps: #14,#15]
