# Testing Guide

This document provides comprehensive guidance on testing for the MTGA Companion project.

## Table of Contents

- [Testing Strategy](#testing-strategy)
- [Test Types](#test-types)
  - [Backend Tests (Go)](#backend-tests-go)
  - [Frontend Component Tests (Vitest)](#frontend-component-tests-vitest)
  - [End-to-End Tests (Playwright)](#end-to-end-tests-playwright)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Test Utilities and Helpers](#test-utilities-and-helpers)
- [Mocking Wails Bindings](#mocking-wails-bindings)
- [Best Practices](#best-practices)
- [Coverage Targets](#coverage-targets)
- [Troubleshooting](#troubleshooting)

## Testing Strategy

MTGA Companion employs a multi-layered testing strategy:

1. **Backend Unit Tests**: Test Go business logic, data processing, and storage
2. **Frontend Component Tests**: Test React components in isolation with mocked dependencies
3. **E2E Tests**: Test complete user workflows from start to finish

### When to Use Each Test Type

| Scenario | Test Type | Reason |
|----------|-----------|--------|
| Data processing logic | Backend Unit Test | Tests pure business logic |
| Database queries | Backend Integration Test | Tests data persistence |
| React component rendering | Component Test | Fast, isolated UI testing |
| User interactions | Component Test | Tests clicks, inputs, events |
| Complete workflows | E2E Test | Tests full user journey |
| Cross-component behavior | E2E Test | Tests integration points |

## Test Types

### Backend Tests (Go)

Backend tests use Go's built-in testing framework.

**Location**: `*_test.go` files alongside source code

**Running**:
```bash
# Run all backend tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/storage

# Run with race detection
go test -race ./...
```

**Example**:
```go
func TestStatisticsCalculation(t *testing.T) {
    matches := []models.Match{
        {Result: "win"},
        {Result: "loss"},
        {Result: "win"},
    }

    stats := CalculateStatistics(matches)

    if stats.WinRate != 0.666 {
        t.Errorf("Expected win rate 0.666, got %f", stats.WinRate)
    }
}
```

### Frontend Component Tests (Vitest)

Component tests use Vitest and React Testing Library.

**Location**: `frontend/src/**/*.test.tsx`

**Running**:
```bash
cd frontend

# Run all component tests
npm run test

# Run in watch mode (development)
npm run test:ui

# Run once (CI)
npm run test:run

# Run with coverage
npm run test:coverage
```

**Example**:
```typescript
import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import { render } from '../test/utils/testUtils';
import Footer from './Footer';
import { mockWailsApp } from '../test/mocks/wailsApp';

describe('Footer Component', () => {
  it('should display total matches count', async () => {
    mockWailsApp.GetStats.mockResolvedValue({
      TotalMatches: 100,
      WinRate: 0.6,
    });

    render(<Footer />);

    await waitFor(() => {
      expect(screen.getByText('100')).toBeInTheDocument();
    });
  });
});
```

### End-to-End Tests (Playwright)

E2E tests use Playwright to test the full application.

**Location**: `frontend/tests/e2e/**/*.spec.ts`

**Running**:
```bash
# First, start the Wails dev server
wails dev

# In another terminal:
cd frontend

# Run E2E tests (headless)
npm run test:e2e

# Run with UI (for debugging)
npm run test:e2e:ui

# Run in debug mode
npm run test:e2e:debug

# View last test report
npx playwright show-report
```

**Example**:
```typescript
import { test, expect } from '@playwright/test';

test('should navigate to draft view', async ({ page }) => {
  await page.goto('/');

  // Wait for app to load
  await expect(page.locator('.app-container')).toBeVisible();

  // Click draft tab
  await page.getByText('Draft').click();

  // Verify navigation
  await expect(page).toHaveURL(/\/draft/);
});
```

## Writing Tests

### Component Test Structure

Follow the Arrange-Act-Assert pattern:

```typescript
describe('ComponentName', () => {
  beforeEach(() => {
    // Reset mocks before each test
    vi.clearAllMocks();
  });

  it('should do something when condition', async () => {
    // Arrange: Set up test data and mocks
    mockWailsApp.GetData.mockResolvedValue(testData);

    // Act: Render component and interact
    render(<ComponentName />);
    await userEvent.click(screen.getByText('Button'));

    // Assert: Verify expected outcome
    await waitFor(() => {
      expect(screen.getByText('Result')).toBeInTheDocument();
    });
  });
});
```

### E2E Test Structure

Focus on user workflows:

```typescript
test.describe('Workflow Name', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.app-container')).toBeVisible();
  });

  test('should complete workflow step', async ({ page }) => {
    // Navigate
    await page.getByText('Tab Name').click();

    // Interact
    await page.locator('input').fill('data');
    await page.getByText('Submit').click();

    // Verify
    await expect(page.getByText('Success')).toBeVisible();
  });
});
```

## Test Utilities and Helpers

### Custom Render Function

Use the custom render function for components that need routing:

```typescript
import { render } from '../test/utils/testUtils';

render(<Component />, { initialRoute: '/draft' });
```

The custom render wraps components with:
- React Router's `MemoryRouter`
- Any necessary providers

### Mock Data Generators

Create test data using model constructors:

```typescript
import { models } from '../../wailsjs/go/models';

function createMockMatch(overrides = {}) {
  return new models.Match({
    ID: 'test-match-1',
    Result: 'win',
    Format: 'Standard',
    Timestamp: new Date(),
    ...overrides,
  });
}

// Use in tests
const testMatch = createMockMatch({ Result: 'loss' });
```

## Mocking Wails Bindings

### Setup

Wails bindings are mocked in `frontend/src/test/mocks/`:

- `wailsApp.ts`: Mocks Go backend functions
- `wailsRuntime.ts`: Mocks Wails runtime functions (EventsOn, EventsEmit, etc.)

### Using Mocks

```typescript
import { mockWailsApp } from '../test/mocks/wailsApp';
import { mockEventEmitter } from '../test/mocks/wailsRuntime';

// Mock function return values
mockWailsApp.GetStats.mockResolvedValue({
  TotalMatches: 100,
  WinRate: 0.6,
});

// Verify function calls
expect(mockWailsApp.GetStats).toHaveBeenCalled();

// Emit events
mockEventEmitter.emit('stats:updated', { matches: 5 });

// Clear event listeners
mockEventEmitter.clear();
```

### Adding New Mocks

When adding new Wails bindings:

1. Add the mock function to `wailsApp.ts`:
```typescript
export const mockWailsApp = {
  // ...existing mocks
  NewFunction: vi.fn(() => Promise.resolve([] as any[])),
};
```

2. The mock is automatically available via `window.go.main.App`:
```typescript
// No additional setup needed, mocks are global
```

## Best Practices

### General

1. **Test behavior, not implementation**: Focus on what the component does, not how it does it
2. **Write descriptive test names**: Use "should [expected behavior] when [condition]"
3. **One assertion per test** (when possible): Makes failures easier to diagnose
4. **Use data-testid sparingly**: Prefer semantic queries (getByText, getByRole)
5. **Avoid testing third-party libraries**: Trust that React, Recharts, etc. work correctly

### Component Tests

1. **Mock external dependencies**: Don't make real API calls or access real storage
2. **Test user interactions**: Click buttons, fill forms, navigate
3. **Use waitFor for async updates**: Don't use arbitrary timeouts
4. **Test loading and error states**: Not just the happy path
5. **Keep tests isolated**: Each test should be independent

```typescript
// Good: Tests behavior
it('should show error message when fetch fails', async () => {
  mockWailsApp.GetData.mockRejectedValue(new Error('Failed'));
  render(<Component />);
  await waitFor(() => {
    expect(screen.getByText(/error/i)).toBeInTheDocument();
  });
});

// Bad: Tests implementation details
it('should set loading state to false', async () => {
  const { rerender } = render(<Component />);
  expect(component.state.loading).toBe(false); // Don't access state
});
```

### E2E Tests

1. **Test complete workflows**: Full user journeys from start to finish
2. **Handle timing gracefully**: Use waitFor, not fixed sleeps
3. **Test both data and no-data states**: Empty states are important
4. **Keep tests idempotent**: Should work regardless of existing data
5. **Use meaningful selectors**: Prefer text content over CSS selectors

```typescript
// Good: Waits for element to be visible
await expect(page.locator('.results')).toBeVisible();

// Bad: Arbitrary timeout
await page.waitForTimeout(5000);
```

## Coverage Targets

### Backend (Go)

- **Target**: 70% overall coverage
- **Critical paths**: 90%+ (data processing, calculations)
- **UI code**: Can be lower (Fyne widgets)

### Frontend (React)

- **Target**: 80% overall coverage
- **Components**: 75%+ (all major UI components)
- **Utilities**: 90%+ (helper functions, calculations)
- **E2E**: Focus on critical workflows, not coverage percentage

### Checking Coverage

```bash
# Backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Frontend
cd frontend
npm run test:coverage
# Open frontend/coverage/index.html
```

## Troubleshooting

### Common Issues

#### "Cannot find module '@wailsio/runtime'"

**Cause**: Wails runtime not properly mocked

**Solution**: Ensure test setup imports the mock:
```typescript
import { mockEventEmitter } from '../test/mocks/wailsRuntime';
```

#### "TypeError: Cannot read property 'then' of undefined"

**Cause**: Mock function not returning a Promise

**Solution**: Ensure mocks return Promises:
```typescript
mockWailsApp.GetData.mockResolvedValue(data); // Good
mockWailsApp.GetData.mockReturnValue(data);   // Bad - missing Promise
```

#### "Test times out waiting for element"

**Cause**: Element never appears or wrong selector

**Solutions**:
1. Check if mock data is set correctly
2. Verify the element exists with screen.debug()
3. Increase timeout: `await waitFor(() => {...}, { timeout: 10000 })`

#### "E2E tests fail with connection refused"

**Cause**: Wails dev server not running

**Solution**: Start the dev server first:
```bash
wails dev
# Wait for server to start
# Then run E2E tests
```

#### "Tests pass locally but fail in CI"

**Possible causes**:
1. Timing differences (use waitFor, not timeouts)
2. Different data in test database
3. Environment-specific dependencies

**Solution**: Check CI logs, run with same environment variables

### Getting Help

1. Check existing tests for examples
2. Review error messages carefully - they usually point to the issue
3. Use `screen.debug()` to see what's actually rendered
4. Check the [Vitest docs](https://vitest.dev/)
5. Check the [Testing Library docs](https://testing-library.com/react)
6. Check the [Playwright docs](https://playwright.dev/)

## Related Documentation

- [Component Test Examples](../frontend/src/components/)
- [E2E Test Suite](../frontend/tests/e2e/)
- [CI/CD Testing](../frontend/tests/e2e/CI_README.md)
- [Mocking Guide](../frontend/src/test/README.md) (if exists)

## Contributing

When adding new features:

1. **Write tests first** (TDD) or alongside the feature
2. **Include both happy path and error cases**
3. **Update this documentation** if adding new patterns
4. **Ensure tests pass** before creating PR:
   ```bash
   # Backend
   go test ./...

   # Frontend
   cd frontend
   npm run test:run

   # E2E (locally)
   npm run test:e2e
   ```

Questions? Open an issue or discussion!
