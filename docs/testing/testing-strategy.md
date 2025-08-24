# Archon Testing Strategy

**Version:** 1.0  
**Date:** August 24, 2025  
**Status:** Implementation in Progress

## Overview

This document defines the comprehensive testing strategy for the Archon project, a multi-language desktop application built with Go (backend), TypeScript (frontend logic), and Svelte (UI components). The strategy ensures code quality, reliability, and maintainability across the entire stack.

## Architecture & Technology Stack

```
┌─────────────────┬──────────────────┬─────────────────┐
│   Go Backend    │  TypeScript/JS   │  Svelte UI      │
│   (Wails API)   │    Frontend      │   Components    │
├─────────────────┼──────────────────┼─────────────────┤
│ Unit Tests      │ Unit Tests       │ Component Tests │
│ • go test       │ • Vitest         │ • @testing-lib  │
│ • testify/assert│ • Custom mocks   │ • User events   │
│                 │                  │                 │
│ Integration     │ Integration      │ E2E Tests       │
│ • Wails bridge  │ • Plugin system  │ • Playwright    │
│ • Database      │ • API clients    │ • Real workflows│
│ • File system   │ • Host services  │ • Cross-browser │
└─────────────────┴──────────────────┴─────────────────┘
```

## Testing Layers

### 1. Unit Tests

#### Go Backend (`internal/*/`)
- **Framework:** Standard `go test` with `testify/assert`
- **Coverage Target:** 85%+
- **Scope:** Individual functions, business logic, data structures
- **Location:** `*_test.go` files alongside source code

```go
// Example: internal/store/nodes_test.go
func TestNodeStore_Create(t *testing.T) {
    store := NewNodeStore(":memory:")
    node := &Node{Name: "Test Node"}
    
    err := store.Create(node)
    assert.NoError(t, err)
    assert.NotEmpty(t, node.ID)
}
```

#### TypeScript Frontend (`frontend/src/lib/`)
- **Framework:** Vitest (Vite-native, fast, modern)
- **Coverage Target:** 80%+ (plugin system: 90%+)
- **Scope:** Business logic, API clients, plugin system
- **Location:** `*.test.ts` files in `tests/` subdirectories

```typescript
// Example: src/lib/plugins/tests/permissions.test.ts
import { describe, it, expect, vi } from 'vitest'
import { PermissionManager } from '../runtime/permissions.js'

describe('PermissionManager', () => {
  it('grants declared permissions', () => {
    const manager = new PermissionManager(['readRepo'])
    manager.grantPermission('readRepo')
    expect(manager.hasPermission('readRepo')).toBe(true)
  })
})
```

### 2. Component Tests

#### Svelte Components (`frontend/src/lib/components/`)
- **Framework:** `@testing-library/svelte` with Vitest
- **Coverage Target:** 75%+
- **Scope:** Component behavior, user interactions, props/events
- **Location:** `*.test.svelte.ts` files alongside components

```typescript
// Example: src/lib/components/ui/plugin-registry.test.svelte.ts
import { render, screen, fireEvent } from '@testing-library/svelte'
import { vi } from 'vitest'
import PluginRegistry from './plugin-registry.svelte'

describe('PluginRegistry', () => {
  it('displays available plugins', async () => {
    const mockManager = createMockPluginManager()
    render(PluginRegistry, { props: { pluginManager: mockManager } })
    
    expect(screen.getByText('Plugin Registry')).toBeInTheDocument()
    expect(screen.getByText('CSV Importer')).toBeInTheDocument()
  })
})
```

### 3. Integration Tests

#### Go-TypeScript Bridge
- **Framework:** Go test with Wails context mocking
- **Scope:** Wails API calls, data serialization, error handling
- **Location:** `frontend/tests/integration/`

#### Plugin System Integration  
- **Framework:** Vitest with realistic mocks
- **Scope:** Plugin loading, permission flow, host service integration
- **Location:** `frontend/src/lib/plugins/tests/integration/`

### 4. End-to-End Tests

#### Full Application Workflows
- **Framework:** Playwright (cross-browser, Wails-friendly)
- **Scope:** User journeys, plugin installation, data import/export
- **Location:** `e2e/` directory at project root

```typescript
// Example: e2e/plugin-workflow.spec.ts
import { test, expect } from '@playwright/test'

test('install and use CSV importer plugin', async ({ page }) => {
  await page.goto('/')
  
  // Navigate to plugin registry
  await page.click('[data-testid="plugin-registry"]')
  
  // Install CSV importer
  await page.click('[data-testid="install-csv-importer"]')
  await expect(page.locator('[data-testid="plugin-installed"]')).toBeVisible()
  
  // Use plugin to import data
  await page.setInputFiles('[data-testid="file-upload"]', 'test-data.csv')
  await page.click('[data-testid="import-button"]')
  
  // Verify import success
  await expect(page.locator('[data-testid="import-success"]')).toBeVisible()
})
```

## Test Configuration

### Vitest Configuration (`frontend/vitest.config.ts`)

```typescript
import { defineConfig } from 'vitest/config'
import { sveltekit } from '@sveltejs/kit/vite'

export default defineConfig({
  plugins: [sveltekit()],
  test: {
    // Test environment
    environment: 'jsdom',
    
    // Global setup
    setupFiles: ['./src/tests/setup.ts'],
    
    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'src/tests/',
        '**/*.test.ts',
        '**/*.spec.ts'
      ],
      thresholds: {
        global: {
          branches: 75,
          functions: 80,
          lines: 80,
          statements: 80
        },
        // Higher threshold for critical plugin system
        'src/lib/plugins/': {
          branches: 85,
          functions: 90,
          lines: 90,
          statements: 90
        }
      }
    },
    
    // Test file patterns
    include: [
      'src/**/*.{test,spec}.{js,ts}',
      'src/**/*.test.svelte.ts'
    ],
    
    // Mock configuration
    globals: true,
    
    // Timeout configuration
    testTimeout: 10000,
    hookTimeout: 10000
  }
})
```

### Playwright Configuration (`e2e/playwright.config.ts`)

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  
  reporter: [
    ['html'],
    ['junit', { outputFile: 'test-results/results.xml' }]
  ],
  
  use: {
    baseURL: 'http://localhost:5173', // Vite dev server
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
  },
})
```

## Package.json Scripts

```json
{
  "scripts": {
    "test": "vitest run",
    "test:watch": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest run --coverage",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:all": "npm run test:coverage && npm run test:e2e"
  }
}
```

## Dependencies

### Frontend Testing Dependencies

```json
{
  "devDependencies": {
    // Core testing framework
    "vitest": "^2.1.8",
    "@vitest/ui": "^2.1.8",
    
    // DOM testing environment
    "jsdom": "^25.0.1",
    "@testing-library/jest-dom": "^6.6.3",
    
    // Svelte component testing
    "@testing-library/svelte": "^5.2.3",
    "@testing-library/user-event": "^14.5.2",
    
    // Mocking and utilities
    "@vitest/spy": "^2.1.8",
    
    // E2E testing
    "@playwright/test": "^1.49.1",
    
    // Coverage reporting
    "@vitest/coverage-v8": "^2.1.8"
  }
}
```

## Migration Path

### Current State (August 2025)
- ✅ Go backend tests using standard `go test`
- ✅ Homegrown TypeScript plugin system tests (functional but non-standard)
- ❌ No Svelte component tests
- ❌ No E2E tests

### Phase 1: Vitest Migration (Current Priority)
1. **Install Vitest and dependencies**
2. **Configure vitest.config.ts**
3. **Migrate existing plugin system tests** from homegrown to Vitest
4. **Add test:* scripts to package.json**
5. **Validate test coverage meets thresholds**

### Phase 2: Component Testing
1. **Add @testing-library/svelte**
2. **Create component tests for plugin UI**
3. **Test permission dialogs and user interactions**
4. **Mock plugin system dependencies**

### Phase 3: E2E Testing  
1. **Install and configure Playwright**
2. **Create critical user journey tests**
3. **Test plugin installation and usage workflows**
4. **Add visual regression testing**

## Test Data Management

### Test Fixtures (`frontend/src/tests/fixtures/`)
```
fixtures/
├── manifests/
│   ├── valid-manifest.json
│   ├── invalid-manifest.json
│   └── malicious-manifest.json
├── csv-data/
│   ├── sample.csv
│   ├── malformed.csv
│   └── large-dataset.csv
└── plugins/
    ├── sample-importer.js
    └── test-validator.js
```

### Mock Factories (`frontend/src/tests/mocks/`)
```typescript
// mocks/plugin-manager.ts
export function createMockPluginManager(overrides = {}) {
  return {
    getLoadedPlugins: vi.fn(() => []),
    loadPlugin: vi.fn(),
    activatePlugin: vi.fn(),
    getStatistics: vi.fn(() => ({ totalLoaded: 0, totalActive: 0 })),
    ...overrides
  }
}
```

## CI/CD Integration

### GitHub Actions Workflow (`.github/workflows/test.yml`)
```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  go-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -html=coverage.out -o coverage.html

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: npm ci
      - run: npm run test:coverage
      - uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage/

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: npm ci
      - run: npx playwright install
      - run: npm run test:e2e
```

## Quality Gates

### Code Coverage Requirements
- **Go Backend:** 85%+ line coverage
- **TypeScript Core Logic:** 80%+ line coverage  
- **Plugin System:** 90%+ line coverage (critical security component)
- **Svelte Components:** 75%+ line coverage

### Performance Requirements
- **Unit tests:** Must complete in <30 seconds
- **Component tests:** Must complete in <60 seconds
- **E2E tests:** Must complete in <5 minutes
- **Full test suite:** Must complete in <10 minutes

### Test Maintenance
- **Flaky tests:** Fix or remove within 1 week
- **Broken tests:** Block deployment until fixed
- **Coverage regressions:** Block PR merge
- **Test documentation:** Update with API changes

## Special Considerations

### Plugin System Testing
- **Security-first:** Test malicious input handling, permission escalation
- **Sandbox isolation:** Verify Web Worker boundaries are secure
- **Performance:** Test with large plugin loads and complex permissions
- **Error handling:** Test graceful failure and cleanup

### Wails Integration Testing
- **API boundaries:** Test Go-TypeScript data serialization
- **Error propagation:** Verify errors surface correctly in UI
- **Resource management:** Test file system access and cleanup
- **Performance:** Test responsiveness with large data sets

### Cross-Platform Considerations
- **File paths:** Test Windows, macOS, and Linux path handling
- **Permissions:** Test file system permission variations
- **Performance:** Test on different hardware configurations
- **UI rendering:** Test across different screen sizes and DPI settings

## Troubleshooting

### Common Issues
1. **Vitest import errors:** Ensure proper TypeScript path mapping
2. **Svelte component test failures:** Check for proper jsdom setup
3. **Playwright timeouts:** Increase timeout for slow CI environments
4. **Mock leakage:** Use `vi.restoreAllMocks()` in test cleanup

### Debug Commands
```bash
# Run specific test file
npm test permissions.test.ts

# Debug with browser devtools
npm run test:ui

# Run E2E tests in headed mode
npx playwright test --headed

# Generate coverage report
npm run test:coverage
```

---

## Implementation Status

### ✅ Completed
- Go backend testing infrastructure
- Homegrown TypeScript plugin tests (temporary)
- Test strategy documentation

### 🚧 In Progress  
- Vitest integration and migration
- Test configuration setup

### 📅 Planned
- Svelte component testing
- E2E test implementation
- CI/CD integration

---

**Next Action:** Implement Vitest integration following Phase 1 migration plan.