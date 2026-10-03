# performance-benchmarks

<!--
目标：为 free-kiro CLI 关键热路径建立系统化 Go benchmark 套件，作为
后续性能回归基线 + CI 门槛。重点覆盖：
  - EARS lint 引擎（EARSRe 匹配 + 质量 10 条 lint 规则 + Hint 解析）
  - spec wave 计算（任务依赖图拓扑排序）
  - visualize ETag SHA-256
  - dashboard ProjectReport JSON encoding
  - fsnotify 事件 debounce
  - requirements.md 模板渲染

不在本 spec 范围：
  - Go runtime / GC 调优（属 runtime 层）
  - HTTP server 吞吐（属可视化压测，独立 spec）
  - 第三方依赖升级带来的性能变化（属 deps upgrade spec）
-->

## User Stories

- As a free-kiro 维护者 I want 关键热路径都有 `BenchmarkXxx` 函数 so that
  我能在改 lint 引擎 / wave 算法 / ETag 实现时立刻知道性能回归了多少。
- As a free-kiro CI 维护者 I want `make bench` 自动跑 benchstat 与 baseline
  对比 so that 性能回归 ≥ 10% 时 PR 自动被拦下，不需要人工比对历史数字。
- As a 性能分析师 I want `docs/PERF.md` 收录最近一次 benchstat 输出 so that
  我能在 PR review 里直接对比某次改动的 ns/op 浮动而不必本地重跑。

## Acceptance Criteria

### Benchmark 函数覆盖

- [AC-1] THE SYSTEM SHALL 在以下 6 个包各提供 ≥ 1 个 `BenchmarkXxx(b *testing.B)`
  函数，使用 `b.Run` 子表覆盖小/中/大三种 spec 规模（10 / 100 / 1000 AC 行）：
  `internal/lint/`、`internal/spec/`（analyze 路径）、`internal/taskgraph/`、
  `internal/visualize/`（report + etag）、`internal/watch/`（debounce）、
  `internal/spec/templates/`（requirements.md 渲染）。

- [AC-2] WHEN `go test -run=^$ -bench=. -benchmem ./...` 在仓库根目录执行
  THE SYSTEM SHALL 输出 ≥ 18 个 `BenchmarkXxx` 命名（6 包 × ≥ 3 子表），每个
  benchmark 至少跑 `-benchtime=1s` 且报告 `ns/op` / `B/op` / `allocs/op` 三列。

- [AC-3] THE SYSTEM SHALL benchmark fixture 文件使用 `testdata/perf/`
  下的 `small.md`（10 AC）/ `medium.md`（100 AC）/ `large.md`（1000 AC）三个
  静态 spec 文本，由 `init()` 或 `helper` 一次读取 + `[]byte` 复用，避免
  benchmark 自身被 IO 噪声污染。

### benchstat 报告与 baseline

- [AC-4] WHEN `make bench` 在仓库根目录执行 THE SYSTEM SHALL 跑完所有
  benchmark 后调用 `benchstat`（golang.org/x/perf/cmd/benchstat 已有依赖或
  stdlib）生成 `benchdata/baseline.txt` 与 `benchdata/current.txt` 对比报告，
  写入 `docs/PERF.md` 的 fenced code block；任何 ns/op 单项浮动 ≥ 10%
  THE SYSTEM SHALL 退出码非零（性能回归 fail）。

- [AC-5] THE SYSTEM SHALL `benchdata/baseline.txt` 首次提交包含本仓库
  v0.8.0 实测数字（commit `9412e8c` 前后），格式为 benchstat 兼容
  `name  old time/op  new time/op  delta` 表格；后续修改 baseline 必须
  在 PR 描述里说明改动原因（"X% 提升来自改键 Y"）。

### CI 与回归门槛

- [AC-6] WHERE CI 在 `.github/workflows/ci.yml` 新增 `bench-guard` job
  THE SYSTEM SHALL 该 job 跑 `make bench`，**不**在常规 PR 必跑（避免
  benchtime 拖慢 CI），改为每周 cron + main 分支夜间跑；输出报告 artifact
  上传 7 天保留。

- [AC-7] IF 关键热路径（如 `internal/lint/ears.go` 的 `EARSRe` 编译、
  `internal/spec/spec.go` 的 `BuildReport`）新增了 ≥ 10 行代码
  THEN THE SYSTEM SHALL lint-helper 脚本 `scripts/check-bench-coverage.sh`
  报错并提示"该文件缺 BenchmarkXxx"；该脚本由 CI 在 `lint-go` job 末尾跑。

### Documentation

- [AC-8] THE SYSTEM SHALL `docs/PERF.md` 包含 4 节：(1) 当前 hot path 列表
  与对应 benchmark、(2) 跑 benchmark 的本地步骤、(3) 最近一次 benchstat
  输出（fenced code block）、(4) baseline 更新流程（PR 模板 + benchstat
  命令示例）。

## Out of Scope

- Go runtime / GC / GOGC / GOMAXPROCS 调优（属 runtime 范畴）
- HTTP server 端到端压测（属 dashboard 压测 spec，独立）
- pprof / trace 可视化（属 profiling spec，独立）
- 跨平台性能数字对比（只记录 Linux/macOS amd64，不收 Windows arm）
- 第三方依赖升级带来的性能变化（属 deps upgrade spec）

## Acceptance Notes

- "≥ 10% 浮动" 来自 free-kiro 项目惯例：HTTP API / lint 引擎属于"用户能等
  但不应该突然变慢"的范畴；10% 是用户能察觉但不至于每次 commit 都爆的阈值。
- benchmark **不计入** 覆盖率统计（POLICY §1 明文规定 `_bench_test.go`
  不计入覆盖率），但新增 benchmark 仍走正常 review 流程。