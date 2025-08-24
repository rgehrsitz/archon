/**
 * Vitest Global Test Setup
 * 
 * Configures the testing environment with necessary polyfills,
 * mocks, and global utilities for all tests.
 */

import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Web Worker for plugin sandbox tests
class MockWorker {
  url: string
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((error: ErrorEvent) => void) | null = null

  constructor(url: string) {
    this.url = url
  }

  postMessage(data: any) {
    // Mock worker message handling
    setTimeout(() => {
      if (this.onmessage) {
        this.onmessage({ data: { type: 'ready' } } as MessageEvent)
      }
    }, 0)
  }

  terminate() {
    // Mock cleanup
  }

  addEventListener(type: string, listener: EventListener) {
    if (type === 'message') {
      this.onmessage = listener as (event: MessageEvent) => void
    } else if (type === 'error') {
      this.onerror = listener as (error: ErrorEvent) => void
    }
  }

  removeEventListener(type: string, listener: EventListener) {
    if (type === 'message') {
      this.onmessage = null
    } else if (type === 'error') {
      this.onerror = null
    }
  }
}

// Mock Worker globally
global.Worker = MockWorker as any

// Mock URL.createObjectURL and revokeObjectURL
global.URL = global.URL || {}
global.URL.createObjectURL = vi.fn(() => 'mock-blob-url')
global.URL.revokeObjectURL = vi.fn()

// Mock Blob for plugin code
global.Blob = vi.fn().mockImplementation((content, options) => ({
  content,
  options,
  size: content[0]?.length || 0,
  type: options?.type || 'text/plain'
})) as any

// Mock fetch for network requests
global.fetch = vi.fn()

// Mock ResizeObserver (often needed for UI components)
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock matchMedia for responsive design tests
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock localStorage and sessionStorage
const mockStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
  length: 0,
  key: vi.fn(),
}

Object.defineProperty(window, 'localStorage', {
  value: mockStorage
})

Object.defineProperty(window, 'sessionStorage', {
  value: mockStorage
})

// Mock console methods in test environment to reduce noise
// but keep them available for debugging
const originalConsole = { ...console }

global.console = {
  ...console,
  // Suppress info logs in tests unless DEBUG env is set
  info: process.env.DEBUG ? originalConsole.info : vi.fn(),
  debug: process.env.DEBUG ? originalConsole.debug : vi.fn(),
  // Keep warn and error visible
  warn: originalConsole.warn,
  error: originalConsole.error,
  log: process.env.DEBUG ? originalConsole.log : vi.fn(),
}

// Mock Wails runtime for testing
global.window = global.window || {}
;(global.window as any).go = {
  api: {
    NodeService: {
      GetNode: vi.fn(),
      CreateNode: vi.fn(),
      UpdateNode: vi.fn(),
      DeleteNode: vi.fn(),
      MoveNode: vi.fn(),
      ReorderChildren: vi.fn(),
      ListChildren: vi.fn(),
    },
    SearchService: {
      SearchNodes: vi.fn(),
      SearchProperties: vi.fn(),
    },
    GitService: {
      Status: vi.fn(),
    },
    SnapshotService: {
      Create: vi.fn(),
      List: vi.fn(),
    }
  }
}

// Global test utilities
declare global {
  var testUtils: {
    mockPluginManifest: (overrides?: any) => any
    mockPluginContext: (overrides?: any) => any
    mockHost: (overrides?: any) => any
    mockLogger: () => any
  }
}

// Test utility factory functions
global.testUtils = {
  mockPluginManifest: (overrides = {}) => ({
    id: 'com.test.plugin',
    name: 'Test Plugin',
    version: '1.0.0',
    type: 'Importer',
    permissions: ['readRepo'],
    entryPoint: 'index.js',
    ...overrides
  }),

  mockPluginContext: (overrides = {}) => ({
    manifest: global.testUtils.mockPluginManifest(),
    host: global.testUtils.mockHost(),
    logger: global.testUtils.mockLogger(),
    ...overrides
  }),

  mockHost: (overrides = {}) => ({
    getNode: vi.fn().mockResolvedValue(null),
    listChildren: vi.fn().mockResolvedValue([]),
    query: vi.fn().mockResolvedValue([]),
    apply: vi.fn().mockResolvedValue(undefined),
    commit: vi.fn().mockResolvedValue('mock-commit-hash'),
    snapshot: vi.fn().mockResolvedValue('mock-snapshot-hash'),
    readAttachment: vi.fn().mockRejectedValue(new Error('Not implemented')),
    writeAttachment: vi.fn().mockRejectedValue(new Error('Not implemented')),
    fetch: vi.fn().mockRejectedValue(new Error('Not implemented')),
    indexPut: vi.fn().mockResolvedValue(undefined),
    ui: {
      registerCommand: vi.fn(),
      showPanel: vi.fn(),
      showModal: vi.fn().mockResolvedValue(undefined),
      notify: vi.fn(),
    },
    secrets: {
      get: vi.fn().mockResolvedValue(null),
      set: vi.fn().mockResolvedValue(undefined),
      delete: vi.fn().mockResolvedValue(undefined),
    },
    ...overrides
  }),

  mockLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  })
}

// Use fake timers by default for all tests
vi.useFakeTimers()

// Setup and teardown hooks
beforeEach(() => {
  // Clear all mocks before each test
  vi.clearAllMocks()
  
  // Reset DOM
  document.body.innerHTML = ''
  
  // Reset local storage
  localStorage.clear()
  sessionStorage.clear()
})

afterEach(() => {
  // Clean up any pending timers
  vi.runOnlyPendingTimers()
  
  // Reset timers but keep fake timers active
  vi.clearAllTimers()
})

// Global error handler for uncaught promises
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason)
  // Don't exit process in tests, just log the error
})