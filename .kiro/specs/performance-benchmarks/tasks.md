# performance-benchmarks — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

Wave 划分（按依赖自动计算，free-kiro task list 可视化）：
  - Wave 1：#1-#4   testutil/perf fixture + 3 个 fixture 文件
  - Wave 2：#5-#8   lint / spec analyze / taskgraph / spec templates benchmark
  - Wave 3：#9-#11  visualize (report + etag) + watch (debounce) benchmark
  - Wave 4：#12-#14 Makefile bench + benchstat baseline 初始化 + scripts/check-bench-coverage.sh
  - Wave 5：#15-#17 docs/PERF.md + CI bench-guard job + E2E 跑通
  - Wave 6：#18     commit + free-kiro spec complete
-->

## Wave 1 — fixture + testutil

- [ ] #1 `testdata/perf/small.md`（10 AC）、`medium.md`（100 AC）、`large.md`（1000 AC）三个静态 spec 文本，每行合法 EARS 句式 + 可测量响应；放在 `testdata/` 不入库 go 文件 [no deps]
- [ ] #2 `internal/testutil/perf/perf.go` 新建：`type Size int` + `const (Small/Medium/Large)` + `func Load(tb testing.TB, size Size) []byte` 一次 ReadFile 后 sync.Once 缓存到 package-level `[]byte`；bench 复用同一 buffer（防 IO 污染）[deps: #1]
- [ ] #3 `internal/testutil/perf/perf_test.go` 新建：TestLoad_Returns_AllSizes + TestLoad_Cached 与 sync.Once 检查 [deps: #2]
- [ ] #4 `go test ./internal/testutil/perf/...` + `go vet ./...` 0 error；`free-kiro lint performance-benchmarks` 仍 0 error [deps: #2,#3]

## Wave 2 — lint / spec / taskgraph / templates benchmark

- [ ] #5 `internal/lint/ears_bench_test.go` 新建：`BenchmarkEARSRe_AllTemplates` 用 `b.Run("small"/"medium"/"large", ...)` 子表测 `earsRe.MustCompile.EARSPattern.Match` 在 AC 文本上的 ns/op；package-level `regexp.MustCompile` 防止 compile 成本混入 [no deps]
- [ ] #6 `internal/lint/quality_bench_test.go` 新建：`BenchmarkCheckACQuality` 跑 `quality.go` 10 条 semantic rule；子表 small/medium/large [no deps]
- [ ] #7 `internal/spec/analyze_bench_test.go` 新建：`BenchmarkAnalyzeSpec` 测 `spec.Analyze` + vague + 重复 AC 检测全链路 [no deps]
- [ ] #8 `internal/taskgraph/waves_bench_test.go` 新建：`BenchmarkExecutionWaves` 用 10/100/1000 task 节点测拓扑排序 + wave 划分 [no deps]
- [ ] #9 `internal/spec/templates/render_bench_test.go` 新建：`BenchmarkRenderRequirements` 测 `text/template.Execute` 在不同 size fixture 上的 ns/op [no deps]
- [ ] #10 `go test -run=^$ -bench=. -benchmem -benchtime=100x ./internal/lint/... ./internal/spec/... ./internal/taskgraph/...` ≥ 9 个 BenchmarkXxx 命名，0 FAIL [deps: #5-#9]

## Wave 3 — visualize + watch benchmark

- [ ] #11 `internal/visualize/report_bench_test.go` 新建：`BenchmarkBuildReport` 用 5/20/50 specs 模拟 ProjectReport 构造 + JSON encoding [no deps]
- [ ] #12 `internal/visualize/etag_bench_test.go` 新建：`BenchmarkEtagFor` 测 SHA-256 over response bytes；子表 1KB / 10KB / 100KB [no deps]
- [ ] #13 `internal/watch/debounce_bench_test.go` 新建：`BenchmarkDebounceEvents` 用 100/1K/10K events 模拟 300ms debounce timer 累计（不真等 300ms，用 fake clock）[no deps]
- [ ] #14 `go test -run=^$ -bench=. -benchmem -benchtime=100x ./internal/visualize/... ./internal/watch/...` ≥ 9 个 BenchmarkXxx 命名，0 FAIL；总 BenchXxx 累计 ≥ 18（AC-2 通过）[deps: #11,#12,#13]

## Wave 4 — Makefile + baseline + 覆盖门禁

- [ ] #15 `Makefile` 新增 `bench:` target；首次跑时 cp current.txt → baseline.txt；非首次跑 benchstat delta ≥ 10% 报 exit 1 [no deps]
- [ ] #16 `benchdata/.gitignore` 新建：`current.txt` + `report.txt` 入库忽略；首次跑 `make bench` 生成 baseline.txt 后 `git add benchdata/baseline.txt` [deps: #15]
- [ ] #17 `scripts/check-bench-coverage.sh` 新建 + `chmod +x`：扫 `git diff --name-only origin/main...HEAD`，命中 hot_path 列表（ears.go / quality.go / analyze.go / waves.go / report.go / log.go 等）但同包无 `_bench_test.go` 改动时 exit 1 [deps: —]
- [ ] #18 `make bench` 跑通 → baseline.txt 首次写入 `benchdata/baseline.txt`；`benchstat baseline.txt current.txt` 输出有 delta table（含 ±N% 字样）[deps: #15,#16]

## Wave 5 — docs + CI

- [ ] #19 `docs/PERF.md` 新建：4 节（hot path 列表 / 本地跑步骤 / 最近 benchstat fenced block / baseline 更新流程 PR 模板）；把 Wave 4 跑出的首次 benchstat 输出填进 fenced block [deps: #18]
- [ ] #20 `.github/workflows/ci.yml` 新增 `bench-guard` job：trigger `schedule: cron: '0 3 * * 1'` + `workflow_dispatch`；跑 `make bench`；上传 `benchdata/report.txt` artifact 保留 7 天 [no deps]
- [ ] #21 `scripts/check-bench-coverage.sh` 集成进现有 `lint-go` job 末尾；本 PR 验证脚本在 hot_path 列表文件未改 + 无 _bench_test.go 改动时 exit 0（baseline 状态），改 hot_path 但不改 bench 时 exit 1 [deps: #17]

## Wave 6 — 验证 + commit + complete

- [ ] #22 `free-kiro lint performance-benchmarks` 0 ERROR；`go vet ./...` 0 issue；`gofmt -l .` 无输出；`go test -race ./...` 通过；`go test -run=^$ -bench=. -benchmem ./...` ≥ 18 个 BenchmarkXxx 命名 [deps: #1-#21]
- [ ] #23 `git commit -m "perf(bench): 建立性能基准套件 + benchstat baseline"` + push + `free-kiro spec complete performance-benchmarks` [deps: #22]