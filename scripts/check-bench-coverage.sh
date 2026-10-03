#!/usr/bin/env bash
# scripts/check-bench-coverage.sh
#
# performance-benchmarks spec 落地：关键热路径改动必须新增/更新对应
# BenchmarkXxx。在 CI `lint-go` job 末尾跑；命中 hot path 但同包
# 没有 _bench_test.go 改动时 exit 1。
#
# Usage:
#   ./scripts/check-bench-coverage.sh [<base-ref>]
#
# Default base-ref = origin/main（与 .github/workflows/ci.yml 的 diff base 一致）。
# 本地跑可指定其他 base：./scripts/check-bench-coverage.sh HEAD~5

set -euo pipefail

BASE_REF="${1:-origin/main}"

# hot_path → bench_package 映射。任何文件改动应触发同包 _bench_test.go 改动。
# 维护成本低：加 hot path 同时加 BenchmarkXxx；删 hot path 同步删 bench。
declare -A HOT_PATHS=(
    ["internal/lint/ears.go"]="internal/lint"
    ["internal/lint/quality.go"]="internal/lint"
    ["internal/lint/requirements.go"]="internal/lint"
    ["internal/lint/bugfix.go"]="internal/lint"
    ["internal/spec/analyze.go"]="internal/spec"
    ["internal/spec/engine_state.go"]="internal/spec"
    ["internal/spec/generator.go"]="internal/spec"
    ["internal/taskgraph/waves.go"]="internal/taskgraph"
    ["internal/taskgraph/parse.go"]="internal/taskgraph"
    ["internal/visualize/report.go"]="internal/visualize"
    ["internal/visualize/etag.go"]="internal/visualize"
    ["internal/visualize/middleware.go"]="internal/visualize"
)

if ! git rev-parse --verify --quiet "$BASE_REF" >/dev/null; then
    # 本地可能没有 origin/main（fresh clone, no fetch）；fallback 到 HEAD~5
    # 让脚本仍能在本地跑、CI 仍能跑。
    if git rev-parse --verify --quiet "HEAD~5" >/dev/null; then
        BASE_REF="HEAD~5"
        echo "::notice::$BASE_REF not found, falling back to HEAD~5"
    else
        echo "::notice::no base ref available, skipping check"
        exit 0
    fi
fi

changed=$(git diff --name-only "$BASE_REF"...HEAD 2>/dev/null || git diff --name-only "$BASE_REF" HEAD)

missing=0
for path in "${!HOT_PATHS[@]}"; do
    if echo "$changed" | grep -qx "$path"; then
        pkg="${HOT_PATHS[$path]}"
        # 同包是否有任何 _bench_test.go 改动？
        if ! echo "$changed" | grep -q "^${pkg}/.*_bench_test.go$"; then
            echo "::error::$path changed but no _bench_test.go in $pkg"
            missing=$((missing + 1))
        fi
    fi
done

if [ $missing -gt 0 ]; then
    echo
    echo "performance-benchmarks spec requires BenchmarkXxx coverage for hot paths."
    echo "Run 'make bench' locally to verify, then add/extend the relevant bench."
    exit 1
fi

echo "✓ bench coverage OK (no hot path changed without _bench_test.go)"