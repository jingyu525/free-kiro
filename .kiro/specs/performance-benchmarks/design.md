# performance-benchmarks — Design

<!--
决策记录（30 天后回看"为什么这么做"）：
  - 用 Go 标准 testing.B + benchstat（已 golang.org/x/perf 在 go.mod）
    而非第三方框架：与现有 unit test 同栈，零新依赖
  - benchmark fixture 放 testdata/perf/，避免 b.N 跑出来被 IO 污染
  - "≥ 10% 浮动" 阈值：free-kiro 内部多次 lint 优化在 5-8% 浮动，
    10% 是可察觉但不敏感线
  - 不在常规 PR 跑 bench：3-30s 跑 18+ benchmark 拖 CI；改 cron + 夜间 main
  - baseline.txt 提交进仓库而非 CI 动态生成：让"对比"有真值锚点
-->

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      benchmark 套件拓扑                       │
│                                                              │
│  testdata/perf/                                               │
│    ├─ small.md   (10 AC)                                      │
│    ├─ medium.md  (100 AC)                                     │
│    └─ large.md   (1000 AC)                                    │
│         ▲                                                     │
│         │ 一处 helper (benchmark fixture loader)              │
│         │                                                     │
│  ┌───────┼───────┬───────────┬─────────────┬──────────────┐  │
│  │ lint/ │ spec/ │ taskgraph/│ visualize/ │ watch/  │ tmpl │  │
│  │ (3)   │ (3)   │ (3)       │ (3+3)      │ (3)     │ (3)  │  │
│  │       │       │           │ report+etag│ debounce│      │  │
│  └───────┴───────┴───────────┴─────────────┴──────────────┘  │
│         │                                                     │
│         ▼                                                     │
│   go test -run=^$ -bench=. -benchmem -benchtime=1s           │
│         │                                                     │
│         ▼                                                     │
│   benchdata/current.txt                                       │
│         │                                                     │
│         ▼ benchstat baseline.txt current.txt                  │
│   docs/PERF.md (最近一次报告 fenced code block)              │
└──────────────────────────────────────────────────────────────┘
```

## Components

### 1. benchmark fixture（testdata/perf/）

3 个静态文件，**不依赖**任何 go:embed，让 benchmark 启动期 IO 一次性完成：

| 文件 | 行数 | AC 行数 | 用途 |
|---|---|---|---|
| `small.md` | ~30 | 10 | 模拟"草稿 spec"场景 |
| `medium.md` | ~250 | 100 | 模拟"典型 medium spec"（taskgraph wave 测试） |
| `large.md` | ~2400 | 1000 | 模拟"完整 enterprise spec"（lint 引擎压力） |

每行是合法 `WHEN/WHILE/WHERE/UNLESS/IF-THEN/ubiquitous` EARS 句式 + 数字响应（避免触发 `ears-response-immeasurable`）。

### 2. 6 个 benchmark 落点

| 包 | 函数 | 测什么 | 子表维度 |
|---|---|---|---|
| `internal/lint/ears_bench_test.go` | `BenchmarkEARSRe_AllTemplates` | regexp.MustCompile + MustMatch 10K 次 | small/medium/large |
| `internal/lint/quality_bench_test.go` | `BenchmarkCheckACQuality` | 10 条 semantic quality 规则扫一遍 AC | small/medium/large |
| `internal/spec/analyze_bench_test.go` | `BenchmarkAnalyzeSpec` | spec.Analyze + vague word 扫描 + 重复 AC 检测 | small/medium/large |
| `internal/taskgraph/waves_bench_test.go` | `BenchmarkExecutionWaves` | 任务依赖图拓扑排序 + wave 划分 | N=10/100/1000 tasks |
| `internal/visualize/report_bench_test.go` | `BenchmarkBuildReport` | ProjectReport + JSON encoding | 5/20/50 specs |
| `internal/visualize/etag_bench_test.go` | `BenchmarkEtagFor` | SHA-256 over response bytes | 1KB/10KB/100KB |
| `internal/watch/debounce_bench_test.go` | `BenchmarkDebounceEvents` | 300ms debounce timer 累计 + 触发 | 100/1K/10K events |
| `internal/spec/templates/render_bench_test.go` | `BenchmarkRenderRequirements` | text/template Execute + embed.FS 读 | small/medium/large |

共 **8 个** BenchmarkXxx，每个 3 子表 = **24 个子 benchmark**，超过 AC-2 的 ≥ 18 要求。

### 3. fixture loader（testdata/perf 单一 helper）

放在 `internal/testutil/perf/perf.go`，签名：

```go
// Load 返回 small/medium/large spec markdown 字节切片（已 ReadFile，
// benchmark 内复用同一 []byte；调用方需复制再修改）。
func Load(tb testing.TB, size Size) []byte
```

testing.TB 兼容 `*testing.B` 与 `*testing.T`，避免重复 helper。

### 5. benchstat baseline + 报告生成

```makefile
# Makefile 新增 target
.PHONY: bench
bench:
	go test -run=^$ -bench=. -benchmem -benchtime=1s ./... > benchdata/current.txt 2>&1
	@if [ ! -f benchdata/baseline.txt ]; then \
		cp benchdata/current.txt benchdata/baseline.txt; \
		echo "baseline.txt initialized"; \
		exit 0; \
	fi
	go install golang.org/x/perf/cmd/benchstat@latest
	benchstat benchdata/baseline.txt benchdata/current.txt | tee benchdata/report.txt
	# 退出码非零 = 出现 ≥ 10% 浮动
	benchstat -delta-test=bloater -alpha=0.10 benchdata/baseline.txt benchdata/current.txt > /dev/null
```

### 6. CI 集成（cron + 夜间 main）

`.github/workflows/ci.yml` 新增 `bench-guard` job：
- trigger: `schedule: cron: '0 3 * * 1'`（周一凌晨 3 点）+ `workflow_dispatch`
- 运行 `make bench`
- 上传 `benchdata/report.txt` 作为 artifact，保留 7 天
- **不**在常规 PR 必跑

### 7. benchmark 覆盖门禁脚本

`scripts/check-bench-coverage.sh`：

```bash
#!/usr/bin/env bash
# 关键热路径改动后必须新增/更新对应 BenchmarkXxx
set -euo pipefail

declare -A hot_paths=(
    ["internal/lint/ears.go"]="internal/lint"
    ["internal/lint/quality.go"]="internal/lint"
    ["internal/spec/analyze.go"]="internal/spec"
    ["internal/taskgraph/waves.go"]="internal/taskgraph"
    ["internal/visualize/report.go"]="internal/visualize"
    ["internal/visualize/etag.go"]="internal/visualize"
)

changed=$(git diff --name-only origin/main...HEAD)
missing=0
for path in "${!hot_paths[@]}"; do
    if echo "$changed" | grep -q "^$path\$"; then
        pkg="${hot_paths[$path]}"
        bench_files=$(echo "$changed" | grep "^$pkg/.*_bench_test.go" || true)
        if [ -z "$bench_files" ]; then
            echo "::error::$path changed but no benchmark in $pkg"
            missing=$((missing + 1))
        fi
    fi
done

exit $missing
```

CI `lint-go` job 末尾调 `./scripts/check-bench-coverage.sh`。

## Data Model

### 新增 Go 类型（testutil/perf/perf.go）

```go
package perf

type Size int

const (
    Small Size = iota
    Medium
    Large
)

func Load(tb testing.TB, size Size) []byte
```

### benchdata/ 新增

- `benchdata/baseline.txt`：首次跑 bench 时从 current.txt 复制；benchstat 格式
- `benchdata/current.txt`：每次 `make bench` 重写
- `benchdata/report.txt`：benchstat delta 输出
- `benchdata/.gitignore`：`current.txt` / `report.txt` 入库，`*.txt.bak` 不入库

## Testing Strategy

### Lint gate

```bash
free-kiro lint performance-benchmarks   # 0 ERROR
go vet ./...                            # 0 issue
gofmt -l .                              # 无输出
```

### Benchmark 自身验证

```bash
go test -run=^$ -bench=. -benchmem -benchtime=100x ./internal/lint/... ./internal/spec/... \
    ./internal/taskgraph/... ./internal/visualize/... ./internal/watch/... \
    ./internal/spec/templates/...
# 期望：≥ 18 个 BenchmarkXxx 重命名（PASS），无 FAIL
```

### baseline 初始化

```bash
make bench      # 第一次跑会创建 baseline.txt
git add benchdata/baseline.txt
git commit -m "perf(bench): 初始化 v0.8.0 baseline"
```

### Smoke

```bash
# 跑一次 `make bench` 看 docs/PERF.md 是否更新 fenced block
make bench
git diff docs/PERF.md      # 期望只动 fenced block 内的数字
```

## Compatibility / Rollout

### 兼容性边界
- 新增 `_bench_test.go` 文件不影响 `go build ./...`（go test 才会编译）
- `benchdata/current.txt` / `report.txt` 不入库仓库（gitignore）
- `benchdata/baseline.txt` 入库，所有现有 76+ spec 文档不变
- `Makefile` 新增 target 不动现有 target
- `docs/PERF.md` 是新文件（不存在冲突）

### 风险与回滚

- **风险 A**：benchstat 在 PR 流程引入额外 30-60s，CI 卡顿。**对策**：
  仅 cron + 夜间 main 跑，不进 PR lint-go（PM/owner 在 PR 描述里 @ 报告链接）。
- **风险 B**：fixture 文本被反复 ReadFile。**对策**：testutil/perf.Load
  内部 sync.Once 缓存到 `[]byte`，b.N 复用同一 buffer；benchmark 仅做
  copy 进 loop body 防优化掉。
- **风险 C**：lint 引擎 regex 在大文件下首次 compile 慢（cold start）。**对策**：
  BenchmarkEARSRe 内部用 `var compiled = regexp.MustCompile(...)` package-level，
  只测 MustMatch 调用成本而非 compile 成本（与生产 EARS 用户路径一致）。
- **风险 D**：Makefile 引入 golang.org/x/perf benchstat 新下载。**对策**：
  benchstat 是 `go install` 一次性，已在 go.mod indirect；不锁版本。
- **风险 E**：perf.Fixture 命名冲突（testdata/perf vs 现有 testdata）。
  **对策**：`testdata/` 是 Go 约定保留目录，新文件不会与 unit test fixture 混。

### 落地步骤

```
1. testdata/perf/{small,medium,large}.md 新建 3 fixture 文件
2. internal/testutil/perf/perf.go 新建 + perf_test.go 验证 Load 正确性
3. 8 个 _bench_test.go 文件落地（每个包一个，避免跨包 import）
4. Makefile 新增 bench target
5. benchdata/.gitignore + baseline.txt（首次 commit）
6. scripts/check-bench-coverage.sh 新建 + chmod +x
7. .github/workflows/ci.yml 新增 bench-guard job
8. docs/PERF.md 新建（4 节 + 首次 benchstat 输出）
9. free-kiro lint performance-benchmarks → 0 ERROR
10. go test -run=^$ -bench=. -benchmem 跑通
11. make bench 跑通 → baseline.txt 初始化
12. commit + free-kiro spec complete performance-benchmarks
```

## 引用

- `docs/POLICY.md` §1 覆盖率（_bench_test.go 不计入覆盖率）
- `docs/AGENT_RULES.md` §1（先 spec 后代码，> 50 行强制 spec）
- `docs/EARS.md` "## 五种模板 + 无条件基线"（fixture 文本需符合 EARS 句式）
- `Makefile` 现有 target 索引
- `golang.org/x/perf/cmd/benchstat`（仓库 indirect 已有）