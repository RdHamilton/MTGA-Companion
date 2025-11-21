# E2E Testing in CI Environments

## Current Status

E2E tests with Playwright are **not currently enabled** in the GitHub Actions CI pipeline. This document explains why and provides guidance for future implementation.

## Challenge: Wails Applications in CI

Wails applications are desktop applications that:
1. Require a window manager and display server to run
2. Launch a native window with embedded browser
3. Cannot easily run in standard headless CI environments

This makes traditional CI E2E testing challenging compared to web applications.

## Current Approach

**Local Development Testing**: E2E tests are designed to run locally during development:
```bash
# Start the Wails dev server
wails dev

# In another terminal, run E2E tests
cd frontend
npm run test:e2e
```

## Future CI Integration Options

### Option 1: Xvfb Virtual Display (Linux Only)

Use a virtual X server to run the Wails app headlessly:

```yaml
- name: Setup Xvfb
  run: |
    sudo apt-get update
    sudo apt-get install -y xvfb libgl1-mesa-dev xorg-dev

- name: Run with Xvfb
  run: |
    xvfb-run -a wails dev &
    sleep 10
    cd frontend && npm run test:e2e
```

**Pros**:
- Most similar to real environment
- Tests actual Wails binary

**Cons**:
- Linux only (doesn't test macOS/Windows)
- Complex setup with potential timing issues
- Slower execution

### Option 2: Web-Only Development Mode

Create a special development mode that runs just the frontend:

```yaml
- name: Run frontend dev server
  run: |
    cd frontend
    npm run dev &
    sleep 5

- name: Run E2E tests against web version
  run: |
    cd frontend
    npm run test:e2e
```

**Pros**:
- Simple and fast
- Cross-platform
- Standard web testing approach

**Cons**:
- Doesn't test Wails runtime bindings
- Requires mocking all backend calls
- Not testing the actual shipped application

### Option 3: Dedicated Test Environment

Set up a dedicated test server with proper display support:

**Pros**:
- Real environment testing
- Can test across platforms
- More reliable than CI runners

**Cons**:
- Additional infrastructure cost
- More complex to maintain
- Slower feedback loop

### Option 4: Selective Testing

Run E2E tests only for critical workflows and only on release branches:

```yaml
if: github.ref == 'refs/heads/main' || startsWith(github.ref, 'refs/tags/')
```

**Pros**:
- Balances coverage with CI time
- Focuses on high-value scenarios

**Cons**:
- Delayed feedback for some changes
- Still requires solving the display server problem

## Recommendation

**Short-term**: Continue running E2E tests locally during development and PR reviews. The comprehensive component tests (Vitest) provide good coverage in CI.

**Long-term**: Implement Option 1 (Xvfb) for Linux and run a subset of critical E2E tests on major branches only. This provides some automated E2E coverage without overwhelming CI resources.

## Running E2E Tests Locally

E2E tests should be run before merging significant changes:

1. Start the development server:
   ```bash
   wails dev
   ```

2. Wait for the app to be fully loaded (http://localhost:34115)

3. Run tests:
   ```bash
   cd frontend
   npm run test:e2e          # Headless mode
   npm run test:e2e:ui       # Interactive mode
   npm run test:e2e:debug    # Debug mode
   ```

## CI Test Coverage Strategy

The current CI strategy focuses on:

1. **Backend Tests** (Go): Unit and integration tests for backend logic
2. **Frontend Component Tests** (Vitest): Comprehensive coverage of React components
3. **Linting & Formatting**: Code quality checks
4. **Build Verification**: Ensures the application builds successfully

This provides strong confidence in code quality while avoiding the complexity of desktop app E2E testing in CI.

## Related Documentation

- [E2E Test Suite Documentation](./README.md)
- [GitHub Actions Workflow](../../../.github/workflows/ci.yml)
- [Playwright Configuration](../../playwright.config.ts)
