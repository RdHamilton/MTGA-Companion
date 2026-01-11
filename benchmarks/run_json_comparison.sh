#!/bin/bash
# JSON v1 vs v2 Benchmark Comparison Script
#
# Compares encoding/json (v1) vs encoding/json/v2 performance
# Note: json/v2 is experimental and requires GOEXPERIMENT=jsonv2
#
# Prerequisites:
#   - Go 1.25 or later
#   - benchstat: go install golang.org/x/perf/cmd/benchstat@latest
#
# Usage:
#   ./run_json_comparison.sh [count]
#
# Arguments:
#   count: Number of benchmark runs (default: 5)

set -e

COUNT=${1:-5}
RESULTS_DIR="benchmark_results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== JSON v1 vs v2 Benchmark Comparison ===${NC}"
echo "Go version: $(go version)"
echo "Benchmark runs: $COUNT"
echo ""

# Create results directory
mkdir -p "$RESULTS_DIR"

# Check for benchstat
if ! command -v benchstat &> /dev/null; then
    echo -e "${YELLOW}Installing benchstat...${NC}"
    go install golang.org/x/perf/cmd/benchstat@latest
fi

# Note: Both benchmarks run with jsonv2 experiment enabled
# because the benchmark file requires it (has build tag)
# V1 functions (json.Marshal) still use v1 implementation

echo -e "${GREEN}Running JSON benchmarks...${NC}"
echo "(Note: Both V1 and V2 benchmarks run in same binary with GOEXPERIMENT=jsonv2)"
echo ""

# Run all JSON benchmarks
GOEXPERIMENT=jsonv2 go test -bench="BenchmarkJSON" -benchmem -count="$COUNT" ./benchmarks/... > "$RESULTS_DIR/json_benchmark_$TIMESTAMP.txt" 2>&1
echo "Results saved to: $RESULTS_DIR/json_benchmark_$TIMESTAMP.txt"
echo ""

# Display results
echo -e "${BLUE}=== Results ===${NC}"
echo ""
cat "$RESULTS_DIR/json_benchmark_$TIMESTAMP.txt"
echo ""

# Extract and compare V1 vs V2
echo -e "${BLUE}=== V1 vs V2 Comparison ===${NC}"
echo ""

# Create separate files for V1 and V2
grep "V1" "$RESULTS_DIR/json_benchmark_$TIMESTAMP.txt" | sed 's/V1//' > "$RESULTS_DIR/json_v1_$TIMESTAMP.txt"
grep "V2" "$RESULTS_DIR/json_benchmark_$TIMESTAMP.txt" | sed 's/V2//' > "$RESULTS_DIR/json_v2_$TIMESTAMP.txt"

if [ -s "$RESULTS_DIR/json_v1_$TIMESTAMP.txt" ] && [ -s "$RESULTS_DIR/json_v2_$TIMESTAMP.txt" ]; then
    benchstat "$RESULTS_DIR/json_v1_$TIMESTAMP.txt" "$RESULTS_DIR/json_v2_$TIMESTAMP.txt"
    benchstat "$RESULTS_DIR/json_v1_$TIMESTAMP.txt" "$RESULTS_DIR/json_v2_$TIMESTAMP.txt" > "$RESULTS_DIR/json_comparison_$TIMESTAMP.txt"
    echo ""
    echo "Comparison saved to: $RESULTS_DIR/json_comparison_$TIMESTAMP.txt"
else
    echo "Could not extract V1 and V2 results for comparison"
fi

echo ""
echo -e "${GREEN}Done!${NC}"
echo ""
echo "Key metrics to watch:"
echo "  - time/op: Lower is better (execution time)"
echo "  - allocs/op: Lower is better (number of allocations)"
echo "  - B/op: Lower is better (bytes allocated)"
echo ""
echo "Benchstat compares V1 (baseline) vs V2 (new):"
echo "  - Negative delta: V2 is faster than V1 (improvement)"
echo "  - Positive delta: V2 is slower than V1 (regression)"
