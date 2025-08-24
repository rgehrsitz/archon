import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()],
  test: {
    // Test environment - jsdom for DOM testing
    environment: 'jsdom',
    
    // Global setup and teardown
    setupFiles: ['./src/tests/setup.ts'],
    
    // Make vitest APIs global (describe, it, expect, vi)
    globals: true,
    
    // Coverage configuration
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov', 'json'],
      reportsDirectory: './coverage',
      exclude: [
        'node_modules/',
        'src/tests/',
        '**/*.test.ts',
        '**/*.spec.ts',
        '**/*.test.svelte.ts',
        'src/app.html',
        'src/app.d.ts',
        'vite.config.ts',
        'vitest.config.ts'
      ],
      
      // Coverage thresholds
      thresholds: {
        global: {
          branches: 75,
          functions: 80,
          lines: 80,
          statements: 80
        },
        // Higher threshold for critical plugin system
        'src/lib/plugins/runtime/': {
          branches: 85,
          functions: 90,
          lines: 90,
          statements: 90
        },
        'src/lib/plugins/manifest.ts': {
          branches: 90,
          functions: 95,
          lines: 95,
          statements: 95
        }
      }
    },
    
    // Test file patterns
    include: [
      'src/**/*.{test,spec}.{js,ts}',
      'src/**/*.test.svelte.ts'
    ],
    
    // Files to exclude
    exclude: [
      'node_modules/',
      '.svelte-kit/',
      'build/',
      'dist/'
    ],
    
    // Timeout configuration
    testTimeout: 10000,
    hookTimeout: 10000,
    
    // Reporter configuration
    reporter: process.env.CI ? ['junit', 'github-actions'] : ['verbose'],
    outputFile: {
      junit: './test-results/junit.xml'
    },
    
    // Watch mode configuration
    watch: {
      exclude: ['node_modules/**', '.svelte-kit/**', 'build/**']
    }
  },
  
  // Resolve configuration for imports
  resolve: {
    alias: {
      '$lib': '/src/lib',
      '$app': '/src/app'
    }
  },
  
  // Define global variables for testing
  define: {
    // Mock browser globals that might not be available in jsdom
    __APP_VERSION__: JSON.stringify('test'),
    __BUILD_TIME__: JSON.stringify('test')
  }
})