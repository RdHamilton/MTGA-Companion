# Go 1.25 Features in MTGA-Companion

This document describes Go 1.25 features implemented in MTGA-Companion and identifies areas for future adoption.

## Implementation Status (v1.4.1)

| Feature | Status | PR | Notes |
|---------|--------|----|----|
| Flight Recorder | ✅ Implemented | #873 | Low-overhead execution tracing |
| GC Benchmarks | ✅ Implemented | #874 | Compare default vs greenteagc |
| JSON v2 Benchmarks | ✅ Implemented | #875 | Compare v1 vs v2 performance |
| sync.WaitGroup.Go() | ✅ Implemented | #791 | Cleaner goroutine management |
| testing/synctest | 🔄 Planned | - | Future test improvement |

## Go 1.25 Key Features

### 1. `testing/synctest` Package (GA)
The new `testing/synctest` package provides support for testing concurrent code with virtualized time.

**Benefits:**
- Tests run in isolated "bubbles" with fake clocks
- Time advances instantaneously when all goroutines block
- No more `time.Sleep()` in tests for synchronization

**Codebase Impact:** HIGH
- `internal/mtga/logreader/poller_test.go` - 25+ `time.Sleep` calls
- `internal/mtga/logreader/notifier_test.go` - 4 `time.Sleep` calls
- `internal/mtga/logreader/poller_manager_test.go` - 5 `time.Sleep` calls
- `internal/storage/scheduler_test.go` - 7 `time.Sleep` calls
- `internal/api/websocket/hub_test.go` - 5 `time.Sleep` calls
- `internal/ipc/client_test.go` - 3 `time.Sleep` calls

### 2. `sync.WaitGroup.Go()` Method
Simplifies the common pattern of `wg.Add(1); go func() { defer wg.Done(); ... }()`.

**Before:**
```go
s.wg.Add(1)
go s.refreshWorker(ctx)
```

**After:**
```go
s.wg.Go(func() { s.refreshWorker(ctx) })
```

**Codebase Impact:** MEDIUM
- `internal/mtga/cards/refresh/scheduler.go:99-104` - 2 instances
- `internal/mtga/logreader/poller_manager.go:138` - 1 instance
- `internal/gui/collection_facade_autofetch_test.go:465` - 1 instance

### 3. `runtime/trace.FlightRecorder` API
Lightweight runtime trace capture into an in-memory ring buffer.

**Codebase Impact:** LOW (new feature opportunity)
- Could add to daemon for debugging rare issues
- Useful for capturing traces when log parsing errors occur

### 4. Container-Aware GOMAXPROCS
Automatically considers cgroup CPU bandwidth limits on Linux.

**Codebase Impact:** NONE (automatic benefit)
- No manual GOMAXPROCS settings in codebase
- Will automatically benefit when running in containers

### 5. `encoding/json/v2` (Experimental)
Major revision of JSON handling with performance improvements.

**Codebase Impact:** MEDIUM (56 files use `encoding/json`)
- Experimental, requires `GOEXPERIMENT=jsonv2`
- Consider benchmarking after it graduates to stable

### 6. Experimental Garbage Collector (`greenteagc`)
10-40% reduction in GC overhead for real-world programs.

**Codebase Impact:** LOW (experimental)
- Worth benchmarking for memory-intensive operations
- Draft rating calculations and large collection handling

## Deprecations Check

**No deprecated APIs found in codebase:**
- `go/ast.FilterPackage()` - Not used
- `go/ast.PackageExports()` - Not used
- `go/ast.MergePackageFiles()` - Not used
- `go/parser.ParseDir()` - Not used

## Implemented Features

### Flight Recorder (#794, PR #873)

**Location**: `internal/daemon/flight_recorder.go`

Uses `runtime/trace.FlightRecorder` for low-overhead execution tracing:

```go
// Create and start flight recorder
config := DefaultFlightRecorderConfig()
fr := NewFlightRecorder(config)
fr.Start()

// Capture trace on error
tracePath, err := fr.CaptureTrace("error-reason")

// Stop when done
fr.Stop()
```

**Configuration**:
- `MinAge`: Minimum age of events to keep (default: 10s)
- `MaxBytes`: Maximum buffer size (default: 10MB)
- `OutputDir`: Trace file directory (default: temp)
- `MaxTraceFiles`: Maximum retained files (default: 5)

### GC Benchmarks (#795, PR #874)

**Location**: `benchmarks/gc_bench_test.go`

Compare default GC vs experimental `greenteagc`:

```bash
# Run comparison
./benchmarks/run_gc_comparison.sh

# Or manually
go test -bench=. -benchmem ./benchmarks/...
GOEXPERIMENT=greenteagc go test -bench=. -benchmem ./benchmarks/...
```

Available benchmarks:
- `BenchmarkCollectionAllocation` - Card collection loading
- `BenchmarkDraftSessionAllocation` - Draft pick processing
- `BenchmarkMatchHistoryAllocation` - Match history loading
- `BenchmarkJSONMarshal/Unmarshal` - JSON serialization
- `BenchmarkMapOperations` - Card lookup
- `BenchmarkSliceGrowth` - Dynamic arrays
- `BenchmarkConcurrentAllocation` - Parallel allocation

### JSON v2 Benchmarks (#793, PR #875)

**Location**: `benchmarks/json_bench_test.go`

Compare `encoding/json` (v1) vs experimental `encoding/json/v2`:

```bash
# Run comparison
./benchmarks/run_json_comparison.sh
```

Requires `GOEXPERIMENT=jsonv2` build tag.

### sync.WaitGroup.Go() (#791)

Adopted cleaner goroutine pattern:

```go
// Before
s.wg.Add(1)
go func() {
    defer s.wg.Done()
    s.refreshWorker(ctx)
}()

// After
s.wg.Go(func() { s.refreshWorker(ctx) })
```

## Remaining Opportunities

### High Priority

1. **Migrate concurrent tests to `testing/synctest`**
   - Estimated impact: Faster, more reliable tests
   - Files: ~10 test files with 50+ time.Sleep calls
   - Effort: Medium

### Medium Priority

2. **Production JSON v2 adoption**
   - Once `encoding/json/v2` graduates to stable
   - 56 files currently using encoding/json
   - Would provide performance improvements

### Low Priority

3. **Production greenteagc adoption**
   - Once garbage collector is stable
   - Monitor benchmarks for improvements
   - No code changes required

## Platform Notes

- **macOS 12+ required** - Already targeting this
- **Windows 32-bit ARM deprecated** - Not a target platform
- **Linux container support improved** - Benefits Docker deployments

## References

- [Go 1.25 Release Notes](https://tip.golang.org/doc/go1.25)
- [Go 1.25 Blog Post](https://go.dev/blog/go1.25)
- [Go 1.25 Interactive Tour](https://antonz.org/go-1-25/)
