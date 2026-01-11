#!/bin/bash
# GC Benchmark Comparison Script
#
# Compares default GC vs greenteagc (Go 1.25+ experimental GC)
#
# Prerequisites:
#   - Go 1.25 or later
#   - benchstat: go install golang.org/x/perf/cmd/benchstat@latest
#
# Usage:
#   ./run_gc_comparison.sh [count]
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

echo -e "${BLUE}=== GC Benchmark Comparison ===${NC}"
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

# Run benchmarks with default GC
echo -e "${GREEN}Running benchmarks with default GC...${NC}"
go test -bench=. -benchmem -count="$COUNT" ./benchmarks/... > "$RESULTS_DIR/default_gc_$TIMESTAMP.txt" 2>&1
echo "Results saved to: $RESULTS_DIR/default_gc_$TIMESTAMP.txt"
echo ""

# Run benchmarks with greenteagc
echo -e "${GREEN}Running benchmarks with greenteagc...${NC}"
GOEXPERIMENT=greenteagc go test -bench=. -benchmem -count="$COUNT" ./benchmarks/... > "$RESULTS_DIR/greenteagc_$TIMESTAMP.txt" 2>&1
echo "Results saved to: $RESULTS_DIR/greenteagc_$TIMESTAMP.txt"
echo ""

# Compare results
echo -e "${BLUE}=== Comparison Results ===${NC}"
echo ""
benchstat "$RESULTS_DIR/default_gc_$TIMESTAMP.txt" "$RESULTS_DIR/greenteagc_$TIMESTAMP.txt"

# Save comparison
benchstat "$RESULTS_DIR/default_gc_$TIMESTAMP.txt" "$RESULTS_DIR/greenteagc_$TIMESTAMP.txt" > "$RESULTS_DIR/comparison_$TIMESTAMP.txt"
echo ""
echo "Comparison saved to: $RESULTS_DIR/comparison_$TIMESTAMP.txt"
echo ""

echo -e "${GREEN}Done!${NC}"
echo ""
echo "Key metrics to watch:"
echo "  - time/op: Lower is better (execution time)"
echo "  - allocs/op: Lower is better (number of allocations)"
echo "  - B/op: Lower is better (bytes allocated)"
echo ""
echo "A negative percentage means greenteagc is faster/better."
echo "A positive percentage means default GC is faster/better."
