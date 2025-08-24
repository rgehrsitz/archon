/**
 * Plugin Manifest Validation Tests (Vitest Migration)
 * 
 * Tests the plugin manifest parsing, validation, and security checks using Vitest.
 * Critical for plugin system security.
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { validateManifest, isValidVersion, compareVersions } from '../manifest.js'
import type { PluginManifest } from '../api.js'

describe('Plugin Manifest Validation', () => {
  let validManifest: PluginManifest

  beforeEach(() => {
    validManifest = {
      id: 'com.example.test-plugin',
      name: 'Test Plugin',
      version: '1.0.0',
      type: 'Importer',
      description: 'A test plugin for validation',
      author: 'Test Author',
      license: 'MIT',
      permissions: ['readRepo', 'writeRepo'],
      entryPoint: 'index.js',
      archonVersion: '^1.0.0'
    }
  })

  describe('valid manifest', () => {
    it('accepts valid manifest', () => {
      const result = validateManifest(validManifest)
      
      expect(result.valid).toBe(true)
      expect(result.errors).toHaveLength(0)
    })
  })

  describe('required fields validation', () => {
    it('rejects manifest without required fields', () => {
      const invalidManifest = {
        name: 'Test Plugin',
        version: '1.0.0'
        // Missing required fields: id, type, permissions, entryPoint
      } as PluginManifest
      
      const result = validateManifest(invalidManifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.length).toBeGreaterThan(0)
      expect(result.errors.some(error => error.includes('id'))).toBe(true)
      expect(result.errors.some(error => error.includes('type'))).toBe(true)
      expect(result.errors.some(error => error.includes('permissions'))).toBe(true)
      expect(result.errors.some(error => error.includes('entryPoint'))).toBe(true)
    })

    it('accepts manifest with minimal required fields', () => {
      const minimalManifest: PluginManifest = {
        id: 'com.example.minimal',
        name: 'Minimal Plugin',
        version: '1.0.0',
        type: 'Importer',
        permissions: [],
        entryPoint: 'index.js'
      }
      
      const result = validateManifest(minimalManifest)
      expect(result.valid).toBe(true)
    })
  })

  describe('plugin ID validation', () => {
    const invalidIds = [
      { id: '', description: 'empty string' },
      { id: 'test', description: 'no domain' },
      { id: 'com.', description: 'empty name' },
      { id: 'com..test', description: 'double dot' },
      { id: 'com.example.', description: 'trailing dot' },
      { id: '.com.example.test', description: 'leading dot' },
      { id: 'com.example.test.', description: 'trailing dot' },
      { id: 'COM.EXAMPLE.TEST', description: 'uppercase not allowed' },
      { id: 'com.example.test plugin', description: 'spaces' },
      { id: 'com.example.test@1.0', description: 'special chars' }
    ]

    it.each(invalidIds)('rejects invalid ID: $description', ({ id }) => {
      const manifest = { ...validManifest, id }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => error.includes('id'))).toBe(true)
    })

    const validIds = [
      'com.example.test',
      'org.archon.csv-importer',
      'dev.local.my-plugin',
      'io.github.user.plugin-name'
    ]

    it.each(validIds)('accepts valid ID: %s', (id) => {
      const manifest = { ...validManifest, id }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(true)
    })
  })

  describe('plugin type validation', () => {
    const validTypes = [
      'Importer', 'Exporter', 'Transformer', 'Validator', 'Panel', 
      'Provider', 'AttachmentProcessor', 'ConflictResolver', 
      'SearchIndexer', 'UIContrib'
    ]

    it.each(validTypes)('accepts valid type: %s', (type) => {
      const manifest = { ...validManifest, type: type as any }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(true)
    })

    const invalidTypes = [
      { type: 'InvalidType', description: 'unknown type' },
      { type: 'importer', description: 'lowercase' },
      { type: 'IMPORTER', description: 'uppercase' },
      { type: '', description: 'empty string' },
      { type: undefined, description: 'undefined' }
    ]

    it.each(invalidTypes)('rejects invalid type: $description', ({ type }) => {
      const manifest = { ...validManifest, type: type as any }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => error.includes('type'))).toBe(true)
    })
  })

  describe('permissions validation', () => {
    it('accepts valid permissions', () => {
      const manifest = { ...validManifest, permissions: ['readRepo', 'writeRepo', 'ui'] }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(true)
    })

    it('accepts empty permissions array', () => {
      const manifest = { ...validManifest, permissions: [] }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(true)
    })

    it('rejects invalid permission types', () => {
      const manifest = { ...validManifest, permissions: ['invalidPerm'] as any }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => error.includes('permission'))).toBe(true)
    })

    it('rejects non-array permissions', () => {
      const manifest = { ...validManifest, permissions: 'readRepo' as any }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => error.includes('permissions'))).toBe(true)
    })

    it('accepts valid secret permissions', () => {
      const validSecrets = ['secrets:jira*', 'secrets:github.token', 'secrets:test']
      const manifest = { ...validManifest, permissions: validSecrets as any }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(true)
    })

    it('rejects invalid secret format', () => {
      const manifest = { ...validManifest, permissions: ['secrets'] as any } // No colon
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => error.includes('permission'))).toBe(true)
    })

    it('limits permissions array size', () => {
      const tooManyPermissions = Array(1000).fill('readRepo') as any
      const manifest = { ...validManifest, permissions: tooManyPermissions }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => 
        error.includes('permissions') && error.includes('many')
      )).toBe(true)
    })
  })

  describe('entry point validation', () => {
    const validEntryPoints = ['index.js', 'plugin.js', 'main.ts', 'dist/bundle.js']

    it.each(validEntryPoints)('accepts valid entry point: %s', (entryPoint) => {
      const manifest = { ...validManifest, entryPoint }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(true)
    })

    const invalidEntryPoints = [
      { entryPoint: '', description: 'empty string' },
      { entryPoint: 'index', description: 'no extension' },
      { entryPoint: 'plugin.txt', description: 'wrong extension' },
      { entryPoint: '../malicious.js', description: 'path traversal' },
      { entryPoint: 'http://evil.com/script.js', description: 'URL' }
    ]

    it.each(invalidEntryPoints)('rejects invalid entry point: $description', ({ entryPoint }) => {
      const manifest = { ...validManifest, entryPoint }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => error.includes('entryPoint'))).toBe(true)
    })
  })

  describe('security validation', () => {
    const maliciousIds = [
      '../../../etc/passwd',
      '/etc/shadow',
      'C:\\Windows\\System32',
      'file:///etc/passwd',
      'javascript:alert(1)',
      '<script>alert(1)</script>',
      'eval(maliciousCode)',
      'require("child_process")'
    ]

    it.each(maliciousIds)('prevents malicious plugin ID: %s', (id) => {
      const manifest = { ...validManifest, id }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
    })

    it('prevents excessively long name', () => {
      const longString = 'x'.repeat(10000)
      const manifest = { ...validManifest, name: longString }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => 
        error.includes('name') && error.includes('long')
      )).toBe(true)
    })

    it('prevents excessively long description', () => {
      const longString = 'x'.repeat(10000)
      const manifest = { ...validManifest, description: longString }
      const result = validateManifest(manifest)
      
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => 
        error.includes('description') && error.includes('long')
      )).toBe(true)
    })
  })

  describe('metadata validation', () => {
    it('accepts valid metadata', () => {
      const manifestWithMetadata = {
        ...validManifest,
        metadata: {
          category: 'Data Import',
          tags: ['csv', 'import', 'data'],
          website: 'https://example.com',
          repository: 'https://github.com/user/repo'
        }
      }
      
      const result = validateManifest(manifestWithMetadata)
      expect(result.valid).toBe(true)
    })

    const invalidUrls = ['javascript:alert(1)', 'file:///etc/passwd', 'not-a-url', '']

    it.each(invalidUrls)('rejects invalid website URL: %s', (url) => {
      const manifest = {
        ...validManifest,
        metadata: { website: url }
      }
      
      const result = validateManifest(manifest)
      expect(result.valid).toBe(false)
      expect(result.errors.some(error => error.includes('website'))).toBe(true)
    })

    const validUrls = ['https://example.com', 'http://localhost:3000', 'https://github.com/user/repo']

    it.each(validUrls)('accepts valid website URL: %s', (url) => {
      const manifest = {
        ...validManifest,
        metadata: { website: url }
      }
      
      const result = validateManifest(manifest)
      expect(result.valid).toBe(true)
    })
  })
})

describe('Version utilities', () => {
  describe('isValidVersion', () => {
    const validVersions = [
      '1.0.0', '0.1.0', '10.2.5', '1.0.0-alpha', 
      '1.0.0-beta.1', '2.1.0-rc.1+build.123'
    ]

    it.each(validVersions)('validates semantic version: %s', (version) => {
      expect(isValidVersion(version)).toBe(true)
    })

    const invalidVersions = [
      '1.0', '1', '1.0.0.0', 'v1.0.0', '1.0.0-', '1.0.0+', '', 'latest'
    ]

    it.each(invalidVersions)('rejects invalid version: %s', (version) => {
      expect(isValidVersion(version)).toBe(false)
    })
  })

  describe('compareVersions', () => {
    it('compares equal versions', () => {
      expect(compareVersions('1.0.0', '1.0.0')).toBe(0)
      expect(compareVersions('2.1.3', '2.1.3')).toBe(0)
    })

    it('identifies greater versions', () => {
      expect(compareVersions('2.0.0', '1.9.9')).toBeGreaterThan(0)
      expect(compareVersions('1.1.0', '1.0.9')).toBeGreaterThan(0)
      expect(compareVersions('1.0.1', '1.0.0')).toBeGreaterThan(0)
    })

    it('identifies lesser versions', () => {
      expect(compareVersions('1.0.0', '2.0.0')).toBeLessThan(0)
      expect(compareVersions('1.0.0', '1.1.0')).toBeLessThan(0)
      expect(compareVersions('1.0.0', '1.0.1')).toBeLessThan(0)
    })

    it('handles pre-release versions', () => {
      // Pre-release is less than release
      expect(compareVersions('1.0.0-alpha', '1.0.0')).toBeLessThan(0)
      expect(compareVersions('2.1.0-beta', '2.1.0')).toBeLessThan(0)
      
      // Compare pre-release versions
      expect(compareVersions('1.0.0-alpha', '1.0.0-beta')).toBeLessThan(0)
      expect(compareVersions('1.0.0-beta.1', '1.0.0-beta.2')).toBeLessThan(0)
      expect(compareVersions('1.0.0-rc.1', '1.0.0-alpha.1')).toBeGreaterThan(0)
    })
  })
})