# perf-small

<!--
性能基准 fixture：10 条 EARS AC，覆盖 6 种 EARS 句式 + 可测量响应。
不依赖任何 go:embed；bench 通过 internal/testutil/perf.Load 一次性
ReadFile + sync.Once 缓存到 []byte，b.N 期间复用同一 buffer 防 IO 污染。
-->

## User Stories

- As a 性能基准测试者 I want 一份 10 AC 的最小 spec 文本 so that benchmark
  能测出小规模下的 lint / analyze / wave 性能基线。

## Acceptance Criteria

[AC-1] WHEN the bench harness starts THE SYSTEM SHALL load this file within 5 ms.
[AC-2] WHILE the benchmark loop is running THE SYSTEM SHALL reuse the same in-memory buffer so that IO does not pollute the measurement.
[AC-3] WHERE the fixture size is Small THE SYSTEM SHALL contain exactly 10 AC lines.
[AC-4] UNLESS the AC list is empty THE SYSTEM SHALL CONTINUE TO report at least 1 ns per MatchString call.
[AC-5] IF the regex engine is warm THEN the matcher hits THE SYSTEM SHALL return within 200 ns.
[AC-6] THE SYSTEM SHALL contain at least one AC line per EARS template.
[AC-7] WHEN the file is parsed THE SYSTEM SHALL complete within 1 ms on a 2020-era CPU.
[AC-8] WHILE the engine is idle THE SYSTEM SHALL NOT allocate more than 2 KiB per call.
[AC-9] WHERE a benchmark runs at b.N = 1000 iterations THE SYSTEM SHALL report a stable ns/op within 5 percent variance.
[AC-10] THE SYSTEM SHALL remain under 2 KiB after gzip compression.

## Out of Scope

- 不依赖任何运行时配置
- 不引用外部 spec 文件
- 不包含图片 / 表格等非纯文本元素