# E2E Testing in CI Environments

## Current Status

E2E tests with Playwright are **enabled** in CI using the REST API mode. This allows E2E tests to run without the Wails runtime.

## REST API Testing Architecture

The REST API mode enables E2E testing by running:
1. **Go REST API server** (`cmd/apiserver`) on port 8080
2. **Vite dev server** on port 5173 with `VITE_USE_REST_API=true`

This architecture bypasses the Wails runtime completely, making tests:
- Fast and reliable
- Cross-platform compatible
- Easy to run in CI environments

## CI Integration

### GitHub Actions Example

```yaml
e2e-tests:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.23'

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '20'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json

    - name: Install frontend dependencies
      run: cd frontend && npm ci

    - name: Install Playwright browsers
      run: cd frontend && npx playwright install --with-deps chromium

    - name: Run E2E tests
      run: cd frontend && npm run test:e2e
```

### What Happens During Test Execution

1. Playwright automatically starts the Go API server (`go run ./cmd/apiserver`)
2. Playwright starts the Vite dev server with REST API mode
3. Tests run against the frontend, which makes API calls to the Go server
4. Both servers are shut down after tests complete

## Local Development

Run E2E tests locally:

```bash
cd frontend
npm run test:e2e          # Headless mode
npm run test:e2e:ui       # Interactive mode
npm run test:e2e:debug    # Debug mode
```

The same commands work on macOS, Windows, and Linux.

## CI Test Coverage Strategy

The CI pipeline includes:

1. **Backend Tests** (Go): Unit and integration tests for backend logic
2. **Frontend Unit Tests** (Vitest): Component and hook tests
3. **E2E Tests** (Playwright): Full user workflow tests via REST API
4. **Linting & Formatting**: Code quality checks
5. **Build Verification**: Ensures the application builds successfully

## Comparison: REST API vs Wails Mode

| Feature | REST API Mode | Wails Mode |
|---------|---------------|------------|
| CI Support | Yes | Difficult (needs Xvfb) |
| Cross-platform | Yes | Platform-specific |
| Speed | Fast | Slower |
| Wails bindings | Not tested | Tested |
| Setup complexity | Low | High |

The REST API mode tests the same frontend code and API handlers, but skips testing Wails-specific bindings (window management, dialogs, etc.). For most tests, this is sufficient.

## Troubleshooting

### API server fails to start in CI

Check:
- Go version matches the project requirements
- All Go dependencies are available
- Port 8080 is not in use

### Vite dev server fails to start

Check:
- Node.js version matches requirements
- `npm ci` completed successfully
- Port 5173 is not in use

### Tests timeout waiting for servers

Increase the timeout in `playwright.config.ts`:
```typescript
webServer: [
  {
    command: 'cd .. && go run ./cmd/apiserver',
    url: 'http://localhost:8080/health',
    timeout: 180 * 1000,  // Increase if needed
  },
  // ...
]
```

## Related Documentation

- [E2E Test Suite Documentation](./README.md)
- [Playwright Configuration](../../playwright.config.ts)
- [REST API Server](../../../cmd/apiserver/main.go)
