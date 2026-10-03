# perf-medium

<!--
性能基准 fixture：100 条 EARS AC，循环生成 6 种 EARS 句式 × 数字响应。
bench 通过 internal/testutil/perf.Load 一次性 ReadFile 缓存。
-->

## User Stories

- As a 性能基准测试者 I want 一份 100 AC 的中等 spec 文本 so that benchmark
  能测出中等规模下的 lint / analyze / wave 性能基线。

## Acceptance Criteria

[AC-1] WHEN the bench harness starts THE SYSTEM SHALL load this file within 10 ms.
[AC-2] WHILE the benchmark loop is running THE SYSTEM SHALL reuse the same in-memory buffer.
[AC-3] WHERE the fixture size is Medium THE SYSTEM SHALL contain exactly 100 AC lines.
[AC-4] UNLESS the AC list is empty THE SYSTEM SHALL CONTINUE TO report at least 1 ns per call.
[AC-5] IF the regex engine is warm THEN the matcher hits THE SYSTEM SHALL return within 200 ns.
[AC-6] THE SYSTEM SHALL contain at least 10 AC lines per EARS template.
[AC-7] WHEN the file is parsed THE SYSTEM SHALL complete within 5 ms.
[AC-8] WHILE the engine is idle THE SYSTEM SHALL NOT allocate more than 2 KiB per call.
[AC-9] WHERE a benchmark runs at b.N = 1000 iterations THE SYSTEM SHALL report stable ns/op.
[AC-10] THE SYSTEM SHALL remain under 16 KiB after gzip compression.
[AC-11] WHEN the bench harness loads medium.md THE SYSTEM SHALL detect all 100 AC lines within 2 ms.
[AC-12] WHILE the matcher is hot THE SYSTEM SHALL sustain at least 1 M matches per second.
[AC-13] WHERE the engine is single-threaded THE SYSTEM SHALL NOT use more than 1 CPU core.
[AC-14] UNLESS the input is empty THE SYSTEM SHALL CONTINUE TO report the line count within 5 percent.
[AC-15] IF the file size exceeds 100 KiB THEN THE SYSTEM SHALL warn the caller via stderr.
[AC-16] THE SYSTEM SHALL parse 100 AC lines within 20 ms total.
[AC-17] WHEN a benchmark reports ns/op THE SYSTEM SHALL report a sub-microsecond median.
[AC-18] WHILE the test is in flight THE SYSTEM SHALL NOT exceed 50 MiB resident memory.
[AC-19] WHERE profiling is enabled THE SYSTEM SHALL attribute at least 80 percent of CPU to the matcher.
[AC-20] UNLESS the user opts out THE SYSTEM SHALL CONTINUE TO enable -benchmem by default.
[AC-21] IF the test fails THEN THE SYSTEM SHALL print the failing AC line within 100 ms.
[AC-22] THE SYSTEM SHALL complete the full benchmark suite within 30 s.
[AC-23] WHEN a Go file contains 1000 lines THE SYSTEM SHALL compile in 1 s.
[AC-24] WHILE the cache is warm THE SYSTEM SHALL hit at least 95 percent of calls.
[AC-25] WHERE the platform is linux AMD64 THE SYSTEM SHALL use the standard regexp package.
[AC-26] UNLESS the environment is WSL Windows THE SYSTEM SHALL CONTINUE TO use POSIX file paths.
[AC-27] IF the CPU has AVX2 THEN the matcher THE SYSTEM SHALL use vectorised scan.
[AC-28] THE SYSTEM SHALL expose exactly one public Load function.
[AC-29] WHEN the bench reports allocations THE SYSTEM SHALL report zero allocations per iteration after warmup.
[AC-30] WHILE the file watcher is silent THE SYSTEM SHALL NOT invoke the callback.
[AC-31] WHERE multiple goroutines call Load concurrently THE SYSTEM SHALL return the same buffer.
[AC-32] UNLESS the caller overrides b.N THE SYSTEM SHALL CONTINUE TO honour the default -benchtime.
[AC-33] IF the benchmark is run with -count=3 THEN THE SYSTEM SHALL report all 3 trials.
[AC-34] THE SYSTEM SHALL aggregate results into a single ns/op line per case.
[AC-35] WHEN a benchmark exceeds 10 s wall time THE SYSTEM SHALL emit a warning.
[AC-36] WHILE the engine is in benchmark THE SYSTEM SHALL skip the regexp 10-sample cost.
[AC-37] WHERE the test environment is containerised THE SYSTEM SHALL use the host CPU count.
[AC-38] UNLESS the test file is corrupt THE SYSTEM SHALL CONTINUE TO parse without panic.
[AC-39] IF the input contains Windows line endings THEN THE SYSTEM SHALL normalise to LF.
[AC-40] THE SYSTEM SHALL complete the lint suite inside 5 seconds.
[AC-41] WHEN a benchmark measures JSON encoding THE SYSTEM SHALL encode 100 KiB within 2 ms.
[AC-42] WHILE the cache is full THE SYSTEM SHALL evict at least 10 percent of entries.
[AC-43] WHERE the workload is dominated by IO THE SYSTEM SHALL switch to mmap.
[AC-44] UNLESS the caller disables race detection THE SYSTEM SHALL CONTINUE TO run -race.
[AC-45] IF the matcher backtracks THEN THE SYSTEM SHALL warn about catastrophic backtracking.
[AC-46] THE SYSTEM SHALL expose at most 100 KiB resident memory at peak.
[AC-47] WHEN a benchmark re-runs THE SYSTEM SHALL reuse the previous result within 1 ms.
[AC-48] WHILE the fsnotify watcher is hot THE SYSTEM SHALL coalesce events within 300 ms.
[AC-49] WHERE the file system is BTRFS THE SYSTEM SHALL fall back to polling within 2 s.
[AC-50] UNLESS the kernel supports inotify THE SYSTEM SHALL CONTINUE TO use polling.
[AC-51] IF the bench reports regression > 1 ms THEN THE SYSTEM SHALL exit with code 1.
[AC-52] THE SYSTEM SHALL reject fixtures with more than 10 percent whitespace.
[AC-53] WHEN the bench harness starts THE SYSTEM SHALL pin GOMAXPROCS to the host count.
[AC-54] WHILE the matcher runs THE SYSTEM SHALL keep the cache hot.
[AC-55] WHERE the test is run in CI THE SYSTEM SHALL use the test cache directory.
[AC-56] UNLESS the test is skipped THE SYSTEM SHALL CONTINUE TO emit a benchmark line.
[AC-57] IF the test panics THEN THE SYSTEM SHALL mark the benchmark as FAIL.
[AC-58] THE SYSTEM SHALL produce a stable benchstat delta within 0.1 percent variance.
[AC-59] WHEN the bench reports allocations THE SYSTEM SHALL keep B/op below 1024.
[AC-60] WHILE the loop is steady THE SYSTEM SHALL keep allocs/op at zero.
[AC-61] WHERE the engine sees concurrent calls THE SYSTEM SHALL serialise via sync.Once.
[AC-62] UNLESS the caller bypasses Load THE SYSTEM SHALL CONTINUE TO use the package cache.
[AC-63] IF the cache is cold THEN the loader THE SYSTEM SHALL re-read from disk.
[AC-64] THE SYSTEM SHALL consume at most 64 MiB of virtual memory.
[AC-65] WHEN a benchmark measures parsing THE SYSTEM SHALL parse 100 lines within 1 ms.
[AC-66] WHILE the parser is busy THE SYSTEM SHALL NOT block other goroutines.
[AC-67] WHERE the fixture is loaded from disk THE SYSTEM SHALL parse within 10 ms.
[AC-68] UNLESS the file size is zero THE SYSTEM SHALL CONTINUE TO allocate a buffer.
[AC-69] IF the parser sees a malformed line THEN THE SYSTEM SHALL skip the line.
[AC-70] THE SYSTEM SHALL output exactly 100 acceptance criteria lines.
[AC-71] WHEN the bench harness starts THE SYSTEM SHALL validate the fixture checksum.
[AC-72] WHILE the cache is locked THE SYSTEM SHALL NOT reallocate.
[AC-73] WHERE the platform is darwin ARM64 THE SYSTEM SHALL use hardware SHA.
[AC-74] UNLESS the user disables vector instructions THE SYSTEM SHALL CONTINUE TO use SIMD.
[AC-75] IF the matcher sees 1000 matches per second THEN the engine THE SYSTEM SHALL keep CPU under 50 percent.
[AC-76] THE SYSTEM SHALL generate exactly one benchstat report per run.
[AC-77] WHEN a benchmark reruns THE SYSTEM SHALL include -benchtime in the output filename.
[AC-78] WHILE the loop is active THE SYSTEM SHALL not yield to the scheduler.
[AC-79] WHERE the caller passes b.ResetTimer THE SYSTEM SHALL drop warmup samples.
[AC-80] UNLESS the loop is interrupted THE SYSTEM SHALL CONTINUE TO report b.N.
[AC-81] IF the test times out THEN THE SYSTEM SHALL mark the benchmark as DEADLINE_EXCEEDED.
[AC-82] THE SYSTEM SHALL keep the test binary under 50 MiB.
[AC-83] WHEN the harness reports results THE SYSTEM SHALL sort by ns/op ascending.
[AC-84] WHILE the engine emits events THE SYSTEM SHALL respect the configured debounce.
[AC-85] WHERE the workload is bursty THE SYSTEM SHALL apply backpressure within 50 ms.
[AC-86] UNLESS the bench harness is in verbose mode THE SYSTEM SHALL CONTINUE TO silent stdout.
[AC-87] IF the input contains tabs THEN THE SYSTEM SHALL convert to spaces.
[AC-88] THE SYSTEM SHALL reject fixtures containing fewer than 3 sections.
[AC-89] WHEN the matcher compiles THE SYSTEM SHALL complete within 5 ms.
[AC-90] WHILE the cache is warm THE SYSTEM SHALL hit at least 99 percent of calls.
[AC-91] WHERE the bench runs in parallel THE SYSTEM SHALL use b.RunParallel.
[AC-92] UNLESS the caller requests serial THE SYSTEM SHALL CONTINUE TO parallelise.
[AC-93] IF the matcher returns no match THEN the engine THE SYSTEM SHALL report 0 ns.
[AC-94] THE SYSTEM SHALL pass all 10 quality rules without warnings.
[AC-95] WHEN the bench completes THE SYSTEM SHALL print the delta within 1 second.
[AC-96] WHILE the loop is active THE SYSTEM SHALL keep the working set under 8 MiB.
[AC-97] WHERE the platform is macos AMD64 THE SYSTEM SHALL fall back to software SHA.
[AC-98] UNLESS the test is excluded THE SYSTEM SHALL CONTINUE TO run all 100 AC.
[AC-99] IF the bench sees no allocations THEN the engine THE SYSTEM SHALL report 0 allocs/op.
[AC-100] THE SYSTEM SHALL complete the 100 AC benchmark suite within 10 s.

## Out of Scope

- 不依赖任何运行时配置
- 不引用外部 spec 文件
- 不包含图片 / 表格等非纯文本元素
- 不写 `#[AC-101]` 之后的 ID（避免 lint ears-few-ac 误判）