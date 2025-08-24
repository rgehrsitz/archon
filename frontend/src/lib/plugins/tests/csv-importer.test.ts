/**
 * CSV Importer Plugin Tests (Vitest Migration)
 * 
 * Tests the example CSV importer plugin to validate parsing logic,
 * data validation, and node creation functionality using Vitest.
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { register as csvImporterRegister } from '../examples/csv-importer.js'
import type { PluginContext, Host, Logger, PluginManifest, Mutation } from '../api.js'

describe('CSV Importer Plugin', () => {
  let mockContext: PluginContext
  let mockHost: Host
  let mockLogger: Logger
  let appliedMutations: Mutation[]

  // Test data
  const sampleCsv = `Name,Age,Email,Department
John Doe,30,john@example.com,Engineering
Jane Smith,25,jane@example.com,Design
Bob Johnson,35,bob@example.com,Marketing`

  const csvWithoutHeader = `John Doe,30,john@example.com,Engineering
Jane Smith,25,jane@example.com,Design
Bob Johnson,35,bob@example.com,Marketing`

  const malformedCsv = `Name,Age,Email,Department
John Doe,30,john@example.com
Jane Smith,25
Bob Johnson,35,bob@example.com,Marketing,Extra`

  const emptyCsv = ''
  const headerOnlyCsv = 'Name,Age,Email,Department'

  beforeEach(() => {
    appliedMutations = []
    
    mockLogger = {
      debug: vi.fn(),
      info: vi.fn(),
      warn: vi.fn(),
      error: vi.fn(),
    }

    mockHost = {
      async getNode(id: string) {
        if (id === 'parent123') {
          return {
            id: 'parent123',
            name: 'Parent Node',
            description: 'Test parent',
            properties: {},
            children: []
          }
        }
        return null
      },
      
      async listChildren() { return [] },
      async query() { return [] },
      
      async apply(edits: Mutation[]) {
        appliedMutations.push(...edits)
      },
      
      async commit() { return 'mock_commit' },
      async snapshot() { return 'mock_snapshot' },
      async readAttachment() { throw new Error('Not implemented') },
      async writeAttachment() { throw new Error('Not implemented') },
      async fetch() { throw new Error('Not implemented') },
      async indexPut() {},
      
      get ui() { throw new Error('Not implemented') },
      get secrets() { throw new Error('Not implemented') }
    }

    const manifest: PluginManifest = {
      id: 'com.example.csv-importer',
      name: 'CSV Importer',
      version: '1.0.0',
      type: 'Importer',
      permissions: ['readRepo', 'writeRepo'],
      entryPoint: 'index.js'
    }

    mockContext = { manifest, host: mockHost, logger: mockLogger }
  })

  describe('plugin registration', () => {
    it('registers correctly with manifest data', () => {
      const importer = csvImporterRegister(mockContext)
      
      expect(importer.id).toBe('com.example.csv-importer')
      expect(importer.name).toBe('CSV Importer')
      expect(importer.type).toBe('Importer')
      expect(typeof importer.import).toBe('function')
      expect(typeof importer.validate).toBe('function')
    })

    it('logs initialization message', () => {
      csvImporterRegister(mockContext)
      
      expect(mockLogger.info).toHaveBeenCalledWith('CSV Importer Plugin 1.0.0 loaded')
    })
  })

  describe('CSV parsing', () => {
    let importer: ReturnType<typeof csvImporterRegister>

    beforeEach(() => {
      importer = csvImporterRegister(mockContext)
    })

    it('parses basic CSV with headers', async () => {
      const result = await importer.import(sampleCsv, 'parent123')
      
      expect(result.success).toBe(true)
      expect(result.imported).toBe(3)
      expect(result.skipped).toBe(0)
      expect(result.errors).toHaveLength(0)
      
      // Check metadata
      expect(result.metadata.headers).toHaveLength(4)
      expect(result.metadata.headers).toEqual(['Name', 'Age', 'Email', 'Department'])
      expect(result.metadata.totalRows).toBe(3)
      expect(result.metadata.delimiter).toBe(',')
      expect(result.metadata.hasHeader).toBe(true)
      
      // Check mutations were created
      expect(appliedMutations).toHaveLength(3)
      
      const firstMutation = appliedMutations[0]
      expect(firstMutation.type).toBe('create')
      expect(firstMutation.parentId).toBe('parent123')
      expect(firstMutation.data?.name).toBe('John Doe')
      expect(firstMutation.data?.properties?.Age).toBe('30')
      expect(firstMutation.data?.properties?.Email).toBe('john@example.com')
    })

    it('handles CSV without headers', async () => {
      const result = await importer.import(csvWithoutHeader, 'parent123', { hasHeader: false })
      
      expect(result.success).toBe(true)
      expect(result.imported).toBe(3)
      
      // Should generate column names
      expect(result.metadata.headers).toEqual(['Column 1', 'Column 2', 'Column 3', 'Column 4'])
      expect(result.metadata.hasHeader).toBe(false)
      
      const firstMutation = appliedMutations[0]
      expect(firstMutation.data?.properties?.['Column 1']).toBe('John Doe')
    })

    it('handles different delimiters', async () => {
      const tsvData = 'Name\tAge\tEmail\nJohn Doe\t30\tjohn@example.com'
      
      const result = await importer.import(tsvData, 'parent123', { delimiter: '\t' })
      
      expect(result.success).toBe(true)
      expect(result.imported).toBe(1)
      expect(result.metadata.delimiter).toBe('\t')
      
      const mutation = appliedMutations[0]
      expect(mutation.data?.name).toBe('John Doe')
      expect(mutation.data?.properties?.Age).toBe('30')
    })

    it('handles empty data gracefully', async () => {
      // Empty CSV
      const emptyResult = await importer.import(emptyCsv, 'parent123')
      expect(emptyResult.success).toBe(true)
      expect(emptyResult.imported).toBe(0)
      expect(emptyResult.metadata.message).toContain('No data')
      
      // Header only
      appliedMutations.length = 0 // Reset mutations
      const headerResult = await importer.import(headerOnlyCsv, 'parent123')
      expect(headerResult.success).toBe(true)
      expect(headerResult.imported).toBe(0)
    })

    it('creates hierarchical structure when requested', async () => {
      const groupedCsv = `Department,Name,Role
Engineering,John Doe,Developer
Engineering,Jane Smith,Senior Developer
Marketing,Bob Johnson,Manager
Marketing,Alice Brown,Coordinator`

      const result = await importer.import(groupedCsv, 'parent123', { createHierarchy: true })
      
      expect(result.success).toBe(true)
      expect(result.imported).toBeGreaterThan(0)
      expect(result.metadata.createHierarchy).toBe(true)
      
      // Should create group nodes + item nodes
      expect(appliedMutations.length).toBeGreaterThan(4)
      
      // Check for group creation
      const groupMutations = appliedMutations.filter(m => 
        m.data?.properties?.['csv.import.type'] === 'group'
      )
      expect(groupMutations.length).toBeGreaterThanOrEqual(2) // Engineering and Marketing groups
    })

    it('adds import metadata to nodes', async () => {
      await importer.import(sampleCsv, 'parent123')
      
      const mutation = appliedMutations[0]
      
      // Check import metadata
      expect(mutation.data?.properties?.['csv.import.timestamp']).toBeDefined()
      expect(mutation.data?.properties?.['csv.import.rowIndex']).toBe(1)
      expect(mutation.data?.description).toContain('Imported from CSV')
    })
  })

  describe('data validation', () => {
    let importer: ReturnType<typeof csvImporterRegister>

    beforeEach(() => {
      importer = csvImporterRegister(mockContext)
    })

    it('validates correct CSV data', async () => {
      const result = await importer.validate(sampleCsv)
      
      expect(result.valid).toBe(true)
      expect(result.errors).toHaveLength(0)
    })

    it('detects malformed CSV', async () => {
      const result = await importer.validate(malformedCsv)
      
      expect(result.valid).toBe(false)
      expect(result.errors.length).toBeGreaterThan(0)
      expect(result.errors[0]).toContain('columns')
    })

    it('handles empty data in validation', async () => {
      const result = await importer.validate(emptyCsv)
      
      expect(result.valid).toBe(false)
      expect(result.errors).toContain('No data found')
    })

    it('validates delimiter consistency', async () => {
      const inconsistentCsv = 'Name,Age\nJohn;30\nJane,25'
      
      const result = await importer.validate(inconsistentCsv, { delimiter: ',' })
      
      expect(result.valid).toBe(false)
      expect(result.errors.length).toBeGreaterThan(0)
    })
  })

  describe('plugin configuration', () => {
    let importer: ReturnType<typeof csvImporterRegister>

    beforeEach(() => {
      importer = csvImporterRegister(mockContext)
    })

    it('returns supported file extensions', () => {
      const extensions = importer.getSupportedExtensions()
      
      expect(extensions).toContain('csv')
      expect(extensions).toContain('tsv')
      expect(extensions).toContain('txt')
    })

    it('provides options schema', () => {
      const schema = importer.getOptionsSchema()
      
      expect(schema.delimiter).toBeDefined()
      expect(schema.hasHeader).toBeDefined()
      expect(schema.createHierarchy).toBeDefined()
      
      expect(schema.delimiter.default).toBe(',')
      expect(schema.hasHeader.default).toBe(true)
      expect(schema.createHierarchy.default).toBe(false)
    })

    it('includes option descriptions and types', () => {
      const schema = importer.getOptionsSchema()
      
      expect(schema.delimiter.type).toBe('string')
      expect(schema.delimiter.description).toBeDefined()
      expect(schema.delimiter.enum).toContain(',')
      expect(schema.delimiter.enum).toContain('\t')
      
      expect(schema.hasHeader.type).toBe('boolean')
      expect(schema.createHierarchy.type).toBe('boolean')
    })
  })

  describe('error handling', () => {
    let importer: ReturnType<typeof csvImporterRegister>

    beforeEach(() => {
      importer = csvImporterRegister(mockContext)
    })

    it('handles parsing errors gracefully', async () => {
      // Mock host to return null for parent node
      mockHost.getNode = vi.fn().mockResolvedValue(null)
      
      const result = await importer.import(sampleCsv, 'nonexistent_parent')
      
      expect(result.success).toBe(false)
      expect(result.imported).toBe(0)
      expect(result.errors.length).toBeGreaterThan(0)
      expect(result.errors[0]).toContain('nonexistent_parent not found')
    })

    it('handles host service failures', async () => {
      // Mock apply to throw error
      mockHost.apply = vi.fn().mockRejectedValue(new Error('Database error'))
      
      const result = await importer.import(sampleCsv, 'parent123')
      
      expect(result.success).toBe(false)
      expect(result.errors.length).toBeGreaterThan(0)
    })

    it('logs operations correctly', async () => {
      await importer.import(sampleCsv, 'parent123')
      
      // Check that logging occurred
      expect(mockLogger.info).toHaveBeenCalledWith('Starting CSV import')
      expect(mockLogger.info).toHaveBeenCalledWith(expect.stringContaining('Applied'))
    })

    it('handles malformed rows gracefully', async () => {
      const partiallyMalformedCsv = `Name,Age,Email
John Doe,30,john@example.com
,, 
Jane Smith,25,jane@example.com`

      const result = await importer.import(partiallyMalformedCsv, 'parent123')
      
      expect(result.success).toBe(true)
      expect(result.imported).toBe(2) // Should skip empty row
      expect(result.skipped).toBe(1)
    })
  })

  describe('edge cases', () => {
    let importer: ReturnType<typeof csvImporterRegister>

    beforeEach(() => {
      importer = csvImporterRegister(mockContext)
    })

    it('handles special characters in data', async () => {
      const specialCsv = `Name,Description
"Smith, John","Description with commas"
O'Brien,Name with apostrophe`

      const result = await importer.import(specialCsv, 'parent123')
      
      expect(result.success).toBe(true)
      expect(result.imported).toBeGreaterThanOrEqual(1)
    })

    it('handles large datasets efficiently', async () => {
      // Create a large CSV (100 rows)
      const headers = 'Name,Age,Email'
      const rows: string[] = [headers]
      
      for (let i = 1; i <= 100; i++) {
        rows.push(`User${i},${20 + (i % 30)},user${i}@example.com`)
      }
      
      const largeCsv = rows.join('\n')
      
      const result = await importer.import(largeCsv, 'parent123')
      
      expect(result.success).toBe(true)
      expect(result.imported).toBe(100)
      expect(result.metadata.totalRows).toBe(100)
      expect(appliedMutations).toHaveLength(100)
    })

    it('handles Unicode and international characters', async () => {
      const unicodeCsv = `Name,City,Country
José García,São Paulo,Brasil
李明,北京,中国
محمد أحمد,القاهرة,مصر`

      const result = await importer.import(unicodeCsv, 'parent123')
      
      expect(result.success).toBe(true)
      expect(result.imported).toBe(3)
      
      const firstMutation = appliedMutations[0]
      expect(firstMutation.data?.name).toBe('José García')
    })

    it('handles different line endings', async () => {
      const csvWithCRLF = sampleCsv.replace(/\n/g, '\r\n')
      const csvWithCR = sampleCsv.replace(/\n/g, '\r')
      
      // Test CRLF
      appliedMutations.length = 0 // Reset
      let result = await importer.import(csvWithCRLF, 'parent123')
      expect(result.success).toBe(true)
      expect(result.imported).toBe(3)
      
      // Test CR 
      appliedMutations.length = 0 // Reset
      result = await importer.import(csvWithCR, 'parent123')
      expect(result.success).toBe(true)
      expect(result.imported).toBe(3)
    })
  })
})