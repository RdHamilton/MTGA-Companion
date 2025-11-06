#!/bin/bash
# Development workflow script for MTGA-Companion

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
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

show_help() {
    cat << EOF
Development workflow script for MTGA-Companion

Usage: ./scripts/dev.sh [COMMAND]

Commands:
    fmt         Format all Go code with gofmt
    vet         Run go vet on all packages
    lint        Run golangci-lint (if installed)
    build       Build the application
    check       Run fmt, vet, and lint
    all         Run check and build (default)
    help        Show this help message

Examples:
    ./scripts/dev.sh           # Run check and build
    ./scripts/dev.sh fmt       # Just format code
    ./scripts/dev.sh check     # Run all checks without building
EOF
}

fmt_code() {
    print_step "Formatting code..."
    go fmt ./...
    print_success "Code formatted"
}

vet_code() {
    print_step "Running go vet..."
    go vet ./...
    print_success "go vet passed"
}

lint_code() {
    if command -v golangci-lint &> /dev/null; then
        print_step "Running golangci-lint..."
        golangci-lint run
        print_success "Linting passed"
    else
        echo "golangci-lint not installed, skipping..."
        echo "Install with: brew install golangci-lint"
    fi
}

build_app() {
    print_step "Building application..."
    go build -o bin/mtga-companion ./cmd/mtga-companion
    print_success "Build complete: bin/mtga-companion"
}

run_checks() {
    fmt_code
    vet_code
    lint_code
}

run_all() {
    run_checks
    build_app
}

# Parse command
COMMAND=${1:-all}

case "$COMMAND" in
    fmt)
        fmt_code
        ;;
    vet)
        vet_code
        ;;
    lint)
        lint_code
        ;;
    build)
        build_app
        ;;
    check)
        run_checks
        ;;
    all)
        run_all
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
