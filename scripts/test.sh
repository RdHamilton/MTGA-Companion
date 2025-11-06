#!/bin/bash
# Testing workflow script for MTGA-Companion

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

show_help() {
    cat << EOF
Testing workflow script for MTGA-Companion

Usage: ./scripts/test.sh [COMMAND] [OPTIONS]

Commands:
    unit        Run all unit tests
    coverage    Run tests with coverage report
    race        Run tests with race detection
    verbose     Run tests with verbose output
    bench       Run benchmarks
    specific    Run specific test (requires -name flag)
    all         Run unit tests with race detection (default)
    help        Show this help message

Options:
    -name       Test name pattern (for specific command)
    -pkg        Package path (default: ./...)

Examples:
    ./scripts/test.sh                      # Run all tests with race detection
    ./scripts/test.sh unit                 # Run unit tests
    ./scripts/test.sh coverage             # Run tests with coverage
    ./scripts/test.sh specific -name TestParseLog -pkg ./internal/mtga
    ./scripts/test.sh bench                # Run benchmarks
EOF
}

run_unit_tests() {
    local pkg=${1:-./...}
    print_step "Running unit tests..."
    go test $pkg
    print_success "Unit tests passed"
}

run_with_coverage() {
    local pkg=${1:-./...}
    print_step "Running tests with coverage..."
    go test -coverprofile=coverage.out $pkg

    if [ -f coverage.out ]; then
        print_step "Generating coverage report..."
        go tool cover -func=coverage.out
        echo ""
        print_success "Coverage report generated: coverage.out"
        print_warning "View HTML report with: go tool cover -html=coverage.out"
    fi
}

run_race_tests() {
    local pkg=${1:-./...}
    print_step "Running tests with race detection..."
    go test -race $pkg
    print_success "Race detection tests passed"
}

run_verbose_tests() {
    local pkg=${1:-./...}
    print_step "Running tests with verbose output..."
    go test -v $pkg
}

run_benchmarks() {
    local pkg=${1:-./...}
    print_step "Running benchmarks..."
    go test -bench=. -benchmem $pkg
    print_success "Benchmarks complete"
}

run_specific_test() {
    local test_name=$1
    local pkg=${2:-./...}

    if [ -z "$test_name" ]; then
        print_error "Test name required. Use -name flag"
        echo ""
        show_help
        exit 1
    fi

    print_step "Running test: $test_name in $pkg"
    go test -v -run "^${test_name}$" $pkg
    print_success "Test complete"
}

run_all() {
    local pkg=${1:-./...}
    run_race_tests $pkg
}

# Parse arguments
COMMAND=${1:-all}
shift || true

PKG="./..."
TEST_NAME=""

while [ $# -gt 0 ]; do
    case "$1" in
        -name)
            TEST_NAME="$2"
            shift 2
            ;;
        -pkg)
            PKG="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

case "$COMMAND" in
    unit)
        run_unit_tests $PKG
        ;;
    coverage)
        run_with_coverage $PKG
        ;;
    race)
        run_race_tests $PKG
        ;;
    verbose)
        run_verbose_tests $PKG
        ;;
    bench)
        run_benchmarks $PKG
        ;;
    specific)
        run_specific_test "$TEST_NAME" $PKG
        ;;
    all)
        run_all $PKG
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        echo ""
        show_help
        exit 1
        ;;
esac
