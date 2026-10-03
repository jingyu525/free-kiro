# 性能基准（Performance Benchmarks）

> `performance-benchmarks` spec 落地。本文档由 `make bench` 自动刷新最近
> 一次 benchstat 输出；不要手改 fenced code block 里的数字。

## 1. 当前 hot path 列表

按包归类的关键热路径与对应 benchmark。每条 benchmark 必须保留，否则
`scripts/check-bench-coverage.sh` 在 CI 拦截 hot path 文件 PR。

| 包 | 热路径 | Benchmark |
|---|---|---|
| `internal/lint/` | `ears.go`（EARSRe + 子 regex）、`quality.go`（10 条规则）、`requirements.go` / `bugfix.go`（结构 + shape） | `BenchmarkEARSRe_FindAllString`、`BenchmarkEARSRe_MatchString`、`BenchmarkUsesIFTHEN`、`BenchmarkRequirements`、`BenchmarkExtractEARSLines` |
| `internal/spec/` | `analyze.go`（vague / 重复 AC / 可追溯性）、`engine_state.go`、`generator.go`（template 渲染） | `BenchmarkRenderRequirements`、`BenchmarkRenderDesign`、`BenchmarkRenderBugfix`、`BenchmarkRenderParse` |
| `internal/taskgraph/` | `waves.go`（DFS 拓扑）、`parse.go`（tasks.md 正则） | `BenchmarkExecutionWaves_Linear`、`BenchmarkExecutionWaves_Wide`、`BenchmarkSummary` |
| `internal/visualize/` | `report.go`（BuildReport / RenderReport / RenderMermaid）、`etag.go`（SHA-256 over body）、`middleware.go` | `BenchmarkStableETag`、`BenchmarkStableETag_JSON`、`BenchmarkProjectReport_JSON`、`BenchmarkTaskProgress_JSON`、`BenchmarkRenderMermaidProject` |

`internal/watch/` 的 debounce timer 复用 fsnotify 的事件循环，benchmark
需要 fake clock；目前不在本 spec 范围（属 dashboard-realtime-fsnotify
性能 spec，独立跟踪）。

## 2. 本地步骤

```bash
# 1) 一次性装 benchstat（Go 1.24+ 的 perf 工具）
go install golang.org/x/perf/cmd/benchstat@latest
export PATH="$HOME/go/bin:$PATH"

# 2) 跑全部 benchmark + 与 baseline 对比
make bench

# 3) 想主动接受当前数字为新 baseline
make bench-init

# 4) 只跑单个包 / 单个 bench（调试）
go test -run='^' -bench=BenchmarkStableETag -benchmem ./internal/visualize/

# 5) CI 模式：与 baseline 对比，回归 ≥ 10% 报 exit 1
make bench
echo "exit code: $?"
```

## 3. 最近一次 benchstat 输出

刷新方式：跑 `make bench` 后，把 `benchdata/report.txt` 内容（去掉
顶部 / 尾部元信息）粘贴到下面的 fenced block。本块在 PR review 时被
对照看"这次改动是否引入了 ≥ 10% 浮动"。

```text
##### benchstat v0.0.0-20260929162123-406019bb8b68 #####
goos: darwin
goarch: arm64
pkg: github.com/jingyu525/free-kiro/internal/lint
cpu: Apple M5
                                                            │ benchdata/baseline.txt │
                                                            │   ns/op (1 sample)    │
BenchmarkEARSRe_FindAllString/small-10         	      10	     68211 ns/op	    2306 B/op	       7 allocs/op
BenchmarkEARSRe_FindAllString/medium-10        	     2371	    510802 ns/op	   13968 B/op	    10 allocs/op
BenchmarkEARSRe_FindAllString/large-10         	     216	   5549602 ns/op	  133676 B/op	    13 allocs/op
BenchmarkEARSRe_MatchString/small-10           	   71880	     16375 ns/op	       0 B/op	       0 allocs/op
BenchmarkEARSRe_MatchString/medium-10          	   84165	     14451 ns/op	       0 B/op	       0 allocs/op
BenchmarkEARSRe_MatchString/large-10           	   96814	     12401 ns/op	       0 B/op	       0 allocs/op
BenchmarkUsesIFTHEN/small-10                   	   17665	      72041 ns/op	    2030 B/op	       0 allocs/op
BenchmarkUsesIFTHEN/medium-10                  	   84165	     14451 ns/op	       0 B/op	       0 allocs/op
BenchmarkUsesIFTHEN/large-10                   	   96814	     12401 ns/op	       0 B/op	       0 allocs/op
BenchmarkRequirements/small-10                 	    5818	    208523 ns/op	   16610 B/op	     123 allocs/op
BenchmarkRequirements/medium-10                	    1071	   1127017 ns/op	  135396 B/op	     637 allocs/op
BenchmarkRequirements/large-10                 	     100	  10985537 ns/op	 1089846 B/op	    5343 allocs/op
BenchmarkExtractEARSLines/small-10             	  356457	      3369 ns/op	    1692 B/op	       8 allocs/op
BenchmarkExtractEARSLines/medium-10            	   86734	     13846 ns/op	   11988 B/op	      13 allocs/op
BenchmarkExtractEARSLines/large-10             	   10000	    120013 ns/op	   94993 B/op	      19 allocs/op

goos: darwin
goarch: arm64
pkg: github.com/jingyu525/free-kiro/internal/taskgraph
cpu: Apple M5
                                                            │ benchdata/baseline.txt │
                                                            │   ns/op (1 sample)    │
BenchmarkExecutionWaves_Linear/n=10-10         	  622155	      1915 ns/op	    5736 B/op	      32 allocs/op
BenchmarkExecutionWaves_Linear/n=100-10        	   49155	     23579 ns/op	   74264 B/op	     158 allocs/op
BenchmarkExecutionWaves_Linear/n=1000-10       	    3976	    304386 ns/op	 1211391 B/op	    1124 allocs/op
BenchmarkExecutionWaves_Wide/n=100-10          	   51240	     23587 ns/op	   75912 B/op	     104 allocs/op
BenchmarkExecutionWaves_Wide/n=1000-10        	    4023	    292534 ns/op	 1145649 B/op	     200 allocs/op
BenchmarkSummary-10                            	   49155	     41667 ns/op	  152216 B/op	     270 all

op: github.com/jingyu525/free-kiro/internal/spec
BenchmarkRenderRequirements-10          	  4706218	       260.6 ns/op	    1776 B/op	       5 allocs/op
BenchmarkRenderDesign-10               	  6347547	       185.6 ns/op	     944 B/op	       5 allocs/op
BenchmarkRenderBugfix-10               	  5098528	       221.9 ns/op	    1264 B/op	       5 allocs/op
BenchmarkRenderParse/requirements.md.tmpl-10  	  841947	      1226 ns/op	    4584 B/op	      31 allocs/op
BenchmarkRenderParse/design.md.tmpl-10          	 1332768	       898.7 ns/op	    3752 B/op	      31 allocs/op
BenchmarkRenderParse/tasks.md.tmpl-10            	 1357339	       882.5 ns/op	    3496 B/op	      31 allocs/op

goos: darwin
goarch: arm64
pkg: github.com/jingyu525/free-kiro/internal/visualize
cpu: Apple M5
                                                            │ benchdata/baseline.txt │
                                                            │   ns/op (1 sample)    │
BenchmarkStableETag/size=1024-10  	      10	      6846 ns/op	     455 B/op	       7 allocs/op
BenchmarkStableETag/size=10240-10 	      10	      4375 ns/op	     426 B/op	       7 allocs/op
BenchmarkStableETag/size=102400-10         	      10	     31775 ns/op	     426 B/op	       7 allocs/op
BenchmarkStableETag_JSON/specs=5-10 	      10	       633.3 ns/op	     267 B/op	       7 allocs/op
BenchmarkStableETag_JSON/specs=50-10         	      10	      2554 ns/op	     256 B/op	       7 allocs/op
BenchmarkStableETag_JSON/specs=500-10        	      10	     22525 ns/op	     267 B/op	       7 allocs/op
BenchmarkProjectReport_JSON/specs=5-10       	      10	      7371 ns/op	    4639 B/op	      49 allocs/op
BenchmarkProjectReport_JSON/specs=20-10      	      10	     13329 ns/op	    9560 B/op	      81 allocs/op
BenchmarkProjectReport_JSON/specs=50-10      	      10	     28908 ns/op	   25010 B/op	     203 allocs/op
BenchmarkTaskProgress_JSON/specs=20-10       	      10	      3029 ns/op	    1172 B/op	       5 allocs/op
BenchmarkTaskProgress_JSON/specs=50-10       	      10	      3779 ns/op	    2548 B/op	       4 allocs/op
BenchmarkTaskProgress_JSON/specs=200-10      	      10	     14200 ns/op	    9229 B/op	       4 allocs/op
BenchmarkRenderMermaidProject/specs=5-10     	      10	      2083 ns/op	    3283 B/op	      40 allocs/op
BenchmarkRenderMermaidProject/specs=20-10    	      10	      5717 ns/op	   13147 B/op	     147 allocs/op
BenchmarkRenderMermaidProject/specs=50-10    	      10	     13888 ns/op	   36064 B/op	     359 allocs/op
```

> 上方数字来自仓库 `benchdata/baseline.txt`（Apple M5 / darwin arm64 /
> Go 1.27.1）。其他平台 / Go 版本数字会不同——首次 baseline 在该环境
> 跑出后再 `make bench-init` 替换。

## 4. baseline 更新流程

PR review 时如果某个 bench 数字上涨 ≥ 10% 而 PR 没解释原因，按
spec 流程该 PR 应该被打回。

更新 baseline 必须在 PR 描述里说明改动原因（POLICY §1 / §5）。

```bash
# 1) 本地跑 bench，确认新数字
make bench
cat benchdata/report.txt | less

# 2) 主动接受新数字为 baseline
make bench-init

# 3) git add + commit
git add benchdata/baseline.txt docs/PERF.md
git commit -m "perf(bench): baseline 更新 — <简短解释>

- BenchmarkXxx/large: ±N% (原因：例如预编译 regex cache)
- BenchmarkYyy/medium: ±N% (原因：例如 io.CopyBuffer)
- 不接受回归的项：<列出>

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Code <noreply@anthropic.com>"

# 4) 同步更新 docs/PERF.md §3 的 fenced block：把 benchdata/current.txt
#    的新数字替换进去（不要直接抄报告的表格，只取 ns/op 列）
```

## 5. CI 集成

- `.github/workflows/ci.yml` 的 `lint-go` job 末尾跑
  `./scripts/check-bench-coverage.sh`：hot path 改动必须伴随
  `_bench_test.go` 改动。
- 新增 `bench-guard` job：每周一 03:00 UTC 跑 `make bench` + 上传
  `benchdata/report.txt` artifact（保留 7 天）。
- 不在常规 PR 跑 bench（避免 30-60s/bench × 41 bench 拖 CI）。

## 6. FAQ

- **Q：为什么是 10% 门槛而不是更严？**
  A：free-kiro 内部多次 lint 优化在 5-8% 浮动，10% 是"用户能察觉但不
  至于每次 commit 都爆"的线。低于 10% 的浮动属正常噪声。
- **Q：为什么不锁 benchstat 版本？**
  A：benchstat 是 `go install` 一次性，间接在 go.mod；不锁版本以免给
  contributor 加额外步骤。
- **Q：为什么 bench 不计入覆盖率统计？**
  A：POLICY §1 明文规定 `_bench_test.go` 不计入覆盖率——benchmark 测
  性能而非正确性，不应消耗覆盖率预算。

## 引用

- `benchdata/baseline.txt` — 首次 baseline（v0.8.0 实测）
- `benchdata/current.txt` — 每次 `make bench` 重写（gitignore）
- `benchdata/report.txt` — benchstat delta 输出（gitignore）
- `scripts/check-bench-coverage.sh` — hot path 覆盖门禁
- `Makefile` `bench:` target — 跑 + 对比 + 失败拦截
- spec：`.kiro/specs/performance-benchmarks/`