/**
 * Plugin System Test Suite
 * 
 * Runs all plugin system tests and provides a unified test runner.
 */

import { runPermissionTests } from './permissions.test.js';
import { runManifestTests } from './manifest.test.js';
import { runCSVImporterTests } from './csv-importer.test.js';

/**
 * Runs all plugin system tests and returns results.
 */
export async function runAllPluginTests() {
  console.log('🚀 Starting Plugin System Test Suite');
  console.log('=====================================\n');

  const results = {
    permissions: { passed: 0, failed: 0 },
    manifest: { passed: 0, failed: 0 },
    csvImporter: { passed: 0, failed: 0 },
    total: { passed: 0, failed: 0 }
  };

  let hasFailures = false;

  // Run Permission Tests
  try {
    results.permissions = await runPermissionTests();
    console.log('');
  } catch (error) {
    console.error('❌ Permission tests failed:', error);
    hasFailures = true;
  }

  // Run Manifest Tests  
  try {
    results.manifest = await runManifestTests();
    console.log('');
  } catch (error) {
    console.error('❌ Manifest tests failed:', error);
    hasFailures = true;
  }

  // Run CSV Importer Tests
  try {
    results.csvImporter = await runCSVImporterTests();
    console.log('');
  } catch (error) {
    console.error('❌ CSV Importer tests failed:', error);
    hasFailures = true;
  }

  // Calculate totals
  results.total.passed = results.permissions.passed + results.manifest.passed + results.csvImporter.passed;
  results.total.failed = results.permissions.failed + results.manifest.failed + results.csvImporter.failed;

  // Print summary
  console.log('🏁 Test Suite Complete');
  console.log('======================');
  console.log(`📊 Permission System:  ${results.permissions.passed} passed, ${results.permissions.failed} failed`);
  console.log(`📋 Manifest Validation: ${results.manifest.passed} passed, ${results.manifest.failed} failed`);
  console.log(`📈 CSV Importer Plugin: ${results.csvImporter.passed} passed, ${results.csvImporter.failed} failed`);
  console.log(`\n🎯 TOTAL: ${results.total.passed} passed, ${results.total.failed} failed`);

  if (results.total.failed === 0) {
    console.log('\n✨ All tests passed! Plugin system is ready.');
  } else {
    console.log(`\n💥 ${results.total.failed} test(s) failed. Please fix before proceeding.`);
    hasFailures = true;
  }

  if (hasFailures) {
    throw new Error(`${results.total.failed} test(s) failed`);
  }

  return results;
}

/**
 * Test runner for command line usage.
 */
export async function runTestsFromCLI() {
  try {
    await runAllPluginTests();
    console.log('\n🎉 Test suite completed successfully!');
    process.exit(0);
  } catch (error) {
    console.error('\n💀 Test suite failed:', error);
    process.exit(1);
  }
}

// Auto-run in Node.js environment
if (typeof process !== 'undefined' && process.argv?.[1]?.endsWith('index.ts')) {
  runTestsFromCLI();
}

/**
 * Individual test runners for selective testing.
 */
export {
  runPermissionTests,
  runManifestTests,
  runCSVImporterTests
};

/**
 * Quick smoke test for CI/development.
 */
export async function runSmokeTests() {
  console.log('🔥 Running Plugin System Smoke Tests...\n');

  // Basic permission functionality
  const { PermissionManager } = await import('../runtime/permissions.js');
  const manager = new PermissionManager(['readRepo']);
  manager.grantPermission('readRepo');
  
  if (!manager.hasPermission('readRepo')) {
    throw new Error('Smoke test failed: Permission not granted');
  }

  // Basic manifest validation
  const { validateManifest } = await import('../manifest.js');
  const validManifest = {
    id: 'com.test.plugin',
    name: 'Test Plugin',
    version: '1.0.0',
    type: 'Importer' as const,
    permissions: ['readRepo' as const],
    entryPoint: 'index.js'
  };
  
  const result = validateManifest(validManifest);
  if (!result.valid) {
    throw new Error('Smoke test failed: Valid manifest rejected');
  }

  // Basic plugin creation
  const { register } = await import('../examples/csv-importer.js');
  const mockContext = {
    manifest: validManifest,
    host: {
      async getNode() { return null; },
      async apply() {},
    } as any,
    logger: {
      info: () => {},
      debug: () => {},
      warn: () => {},
      error: () => {}
    }
  };
  
  const plugin = register(mockContext);
  if (!plugin || plugin.type !== 'Importer') {
    throw new Error('Smoke test failed: Plugin creation failed');
  }

  console.log('✅ Permission management working');
  console.log('✅ Manifest validation working');
  console.log('✅ Plugin creation working');
  console.log('\n🔥 Smoke tests passed! Core functionality verified.\n');
  
  return true;
}

/**
 * Performance benchmark for plugin operations.
 */
export async function runPerformanceBenchmarks() {
  console.log('⚡ Running Plugin System Performance Benchmarks...\n');

  const { PermissionManager } = await import('../runtime/permissions.js');
  const { validateManifest } = await import('../manifest.js');

  // Benchmark permission checks
  const manager = new PermissionManager(['readRepo', 'writeRepo', 'secrets:*']);
  manager.grantPermission('readRepo');
  manager.grantPermission('secrets:*' as any);

  const permissionStart = Date.now();
  for (let i = 0; i < 10000; i++) {
    manager.hasPermission('readRepo');
    manager.hasPermission('secrets:jira.token' as any);
  }
  const permissionTime = Date.now() - permissionStart;

  // Benchmark manifest validation
  const manifest = {
    id: 'com.test.plugin',
    name: 'Test Plugin',
    version: '1.0.0',
    type: 'Importer' as const,
    permissions: ['readRepo' as const, 'writeRepo' as const],
    entryPoint: 'index.js'
  };

  const manifestStart = Date.now();
  for (let i = 0; i < 1000; i++) {
    validateManifest(manifest);
  }
  const manifestTime = Date.now() - manifestStart;

  console.log(`📈 Permission checks: 10,000 operations in ${permissionTime}ms (${(10000 / permissionTime * 1000).toFixed(0)} ops/sec)`);
  console.log(`📋 Manifest validation: 1,000 operations in ${manifestTime}ms (${(1000 / manifestTime * 1000).toFixed(0)} ops/sec)`);
  console.log('\n⚡ Performance benchmarks completed.\n');

  return {
    permissionOpsPerSec: Math.round(10000 / permissionTime * 1000),
    manifestOpsPerSec: Math.round(1000 / manifestTime * 1000)
  };
}