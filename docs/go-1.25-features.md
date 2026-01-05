# Go 1.25 Features and Codebase Opportunities

This document explores Go 1.25 features and identifies areas of the MTGA-Companion codebase that could benefit from them.

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

## Recommended Follow-up Issues

### High Priority

1. **Migrate concurrent tests to `testing/synctest`**
   - Estimated impact: Faster, more reliable tests
   - Files: ~10 test files with 50+ time.Sleep calls
   - Effort: Medium

### Medium Priority

2. **Adopt `sync.WaitGroup.Go()` pattern**
   - Estimated impact: Cleaner code, less boilerplate
   - Files: 4 files with WaitGroup patterns
   - Effort: Low

3. **Benchmark `encoding/json/v2`**
   - Estimated impact: Potential performance improvement
   - Files: 56 files using encoding/json
   - Effort: Low (just benchmarking)

### Low Priority (Future)

4. **Add Flight Recorder for daemon debugging**
   - Estimated impact: Better debugging capabilities
   - New feature, not a migration
   - Effort: Medium

5. **Benchmark `greenteagc` garbage collector**
   - Estimated impact: Reduced memory overhead
   - Experimental, wait for stability
   - Effort: Low

## Platform Notes

- **macOS 12+ required** - Already targeting this
- **Windows 32-bit ARM deprecated** - Not a target platform
- **Linux container support improved** - Benefits Docker deployments

## References

- [Go 1.25 Release Notes](https://tip.golang.org/doc/go1.25)
- [Go 1.25 Blog Post](https://go.dev/blog/go1.25)
- [Go 1.25 Interactive Tour](https://antonz.org/go-1-25/)
