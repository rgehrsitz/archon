/**
 * Permission System Tests (Vitest Migration)
 * 
 * Tests the core permission management logic including pattern matching,
 * temporal permissions, and risk categorization using Vitest framework.
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { 
  PermissionManager, 
  getPermissionCategory, 
  getPermissionDescription,
  createPermissionRequest,
  PermissionCategory
} from '../runtime/permissions.js'
import type { Permission } from '../api.js'

describe('PermissionManager', () => {
  let manager: PermissionManager

  describe('initialization', () => {
    it('initializes with declared permissions', () => {
      const permissions: Permission[] = ['readRepo', 'writeRepo', 'secrets:jira*']
      manager = new PermissionManager(permissions)
      
      const declared = manager.getDeclaredPermissions()
      expect(declared).toHaveLength(3)
      expect(declared).toContain('readRepo')
      expect(declared).toContain('writeRepo')
      expect(declared).toContain('secrets:jira*')
    })

    it('denies undeclared permissions by default', () => {
      manager = new PermissionManager(['readRepo'])
      
      expect(manager.hasPermission('writeRepo')).toBe(false)
      expect(manager.hasPermission('net')).toBe(false)
    })
  })

  describe('permission granting and revoking', () => {
    beforeEach(() => {
      manager = new PermissionManager(['readRepo', 'writeRepo'])
    })

    it('grants declared permissions', () => {
      // Initially denied
      expect(manager.hasPermission('readRepo')).toBe(false)
      
      // Grant permission
      manager.grantPermission('readRepo')
      expect(manager.hasPermission('readRepo')).toBe(true)
      
      // Other permission still denied
      expect(manager.hasPermission('writeRepo')).toBe(false)
    })

    it('revokes permissions', () => {
      manager.grantPermission('readRepo')
      expect(manager.hasPermission('readRepo')).toBe(true)
      
      manager.revokePermission('readRepo')
      expect(manager.hasPermission('readRepo')).toBe(false)
    })

    it('provides granted permission summary', () => {
      // Initially no grants
      let grants = manager.getGrantedPermissions()
      expect(grants).toHaveLength(0)
      
      // Grant some permissions
      manager.grantPermission('readRepo')
      manager.grantPermission('writeRepo', { temporary: true, duration: 3600000 })
      
      grants = manager.getGrantedPermissions()
      expect(grants).toHaveLength(2)
      
      const readRepoGrant = grants.find(g => g.permission === 'readRepo')
      expect(readRepoGrant?.granted).toBe(true)
      expect(readRepoGrant?.temporary).toBeFalsy()
      
      const writeRepoGrant = grants.find(g => g.permission === 'writeRepo')
      expect(writeRepoGrant?.granted).toBe(true)
      expect(writeRepoGrant?.temporary).toBe(true)
      expect(writeRepoGrant?.expiresAt).toBeInstanceOf(Date)
    })
  })

  describe('temporary permissions', () => {
    beforeEach(() => {
      manager = new PermissionManager(['readRepo'])
    })

    it('handles temporary permissions with expiry', async () => {
      // Grant temporary permission for 50ms
      manager.grantPermission('readRepo', { temporary: true, duration: 50 })
      expect(manager.hasPermission('readRepo')).toBe(true)
      
      // Advance time by 60ms
      vi.advanceTimersByTime(60)
      
      // Should be expired now
      expect(manager.hasPermission('readRepo')).toBe(false)
    })

    it('tracks expiry dates correctly', () => {
      const before = new Date()
      manager.grantPermission('readRepo', { temporary: true, duration: 3600000 })
      const after = new Date(Date.now() + 3600000 + 1000) // Add 1s tolerance
      
      const grants = manager.getGrantedPermissions()
      const grant = grants.find(g => g.permission === 'readRepo')
      
      expect(grant?.expiresAt).toBeInstanceOf(Date)
      expect(grant?.expiresAt!.getTime()).toBeGreaterThan(before.getTime())
      expect(grant?.expiresAt!.getTime()).toBeLessThan(after.getTime())
    })
  })

  describe('pattern matching', () => {
    beforeEach(() => {
      manager = new PermissionManager(['secrets:jira*', 'secrets:github.token'])
    })

    it('matches wildcard patterns for secrets', () => {
      // Grant wildcard permission
      manager.grantPermission('secrets:jira*')
      
      // Should match specific jira secrets
      expect(manager.hasPermission('secrets:jira.token' as Permission)).toBe(true)
      expect(manager.hasPermission('secrets:jira.oauth.refresh' as Permission)).toBe(true)
      
      // Should not match other services
      expect(manager.hasPermission('secrets:github.token' as Permission)).toBe(false)
    })

    it('handles exact matches', () => {
      manager.grantPermission('secrets:github.token')
      expect(manager.hasPermission('secrets:github.token')).toBe(true)
    })

    it('handles edge cases in pattern matching', () => {
      const wildcardManager = new PermissionManager(['secrets:*'])
      wildcardManager.grantPermission('secrets:*' as Permission)
      
      // Wildcard should match everything under secrets
      expect(wildcardManager.hasPermission('secrets:anything' as Permission)).toBe(true)
      expect(wildcardManager.hasPermission('secrets:very.long.nested.key' as Permission)).toBe(true)
      
      // But not non-secrets
      expect(wildcardManager.hasPermission('readRepo')).toBe(false)
    })

    it('handles empty prefix patterns', () => {
      const emptyManager = new PermissionManager(['secrets:' as Permission])
      emptyManager.grantPermission('secrets:' as Permission)
      
      // Should only match exact empty suffix
      expect(emptyManager.hasPermission('secrets:' as Permission)).toBe(true)
      expect(emptyManager.hasPermission('secrets:something' as Permission)).toBe(false)
    })
  })

  describe('permission requests', () => {
    let mockCallback: ReturnType<typeof vi.fn>

    beforeEach(() => {
      manager = new PermissionManager(['readRepo', 'writeRepo'])
      mockCallback = vi.fn()
      manager.onPermissionRequest(mockCallback)
    })

    it('requests permission from callbacks', async () => {
      mockCallback.mockResolvedValue(true)
      
      const result = await manager.requestPermission({
        permission: 'readRepo',
        reason: 'Test reason'
      })
      
      expect(result).toBe(true)
      expect(mockCallback).toHaveBeenCalledWith({
        permission: 'readRepo',
        reason: 'Test reason'
      })
      expect(manager.hasPermission('readRepo')).toBe(true)
    })

    it('rejects undeclared permissions', async () => {
      await expect(manager.requestPermission({
        permission: 'net' // Not in declared permissions
      })).rejects.toThrow('Permission net not declared')
    })

    it('returns true if already granted', async () => {
      manager.grantPermission('readRepo')
      
      const result = await manager.requestPermission({
        permission: 'readRepo'
      })
      
      expect(result).toBe(true)
      expect(mockCallback).not.toHaveBeenCalled()
    })

    it('handles callback rejection', async () => {
      mockCallback.mockResolvedValue(false)
      
      const result = await manager.requestPermission({
        permission: 'readRepo'
      })
      
      expect(result).toBe(false)
      expect(manager.hasPermission('readRepo')).toBe(false)
    })
  })
})

describe('Permission utility functions', () => {
  describe('getPermissionCategory', () => {
    it('categorizes low risk permissions', () => {
      expect(getPermissionCategory('readRepo')).toBe(PermissionCategory.LOW_RISK)
      expect(getPermissionCategory('ui')).toBe(PermissionCategory.LOW_RISK)
    })

    it('categorizes medium risk permissions', () => {
      expect(getPermissionCategory('attachments')).toBe(PermissionCategory.MEDIUM_RISK)
      expect(getPermissionCategory('indexWrite')).toBe(PermissionCategory.MEDIUM_RISK)
    })

    it('categorizes high risk permissions', () => {
      expect(getPermissionCategory('writeRepo')).toBe(PermissionCategory.HIGH_RISK)
      expect(getPermissionCategory('net')).toBe(PermissionCategory.HIGH_RISK)
      expect(getPermissionCategory('secrets:jira*' as Permission)).toBe(PermissionCategory.HIGH_RISK)
    })
  })

  describe('getPermissionDescription', () => {
    it('provides human-readable descriptions', () => {
      const readDesc = getPermissionDescription('readRepo')
      expect(readDesc).toContain('Read project data')
      
      const writeDesc = getPermissionDescription('writeRepo')
      expect(writeDesc).toContain('Create, modify, and delete')
      
      const secretDesc = getPermissionDescription('secrets:jira*' as Permission)
      expect(secretDesc).toContain('Access stored secrets for jira')
    })

    it('handles unknown permissions', () => {
      const unknownDesc = getPermissionDescription('unknown:permission' as Permission)
      expect(unknownDesc).toContain('Unknown permission')
    })
  })

  describe('createPermissionRequest', () => {
    it('creates valid requests with defaults', () => {
      const request = createPermissionRequest('readRepo')
      
      expect(request.permission).toBe('readRepo')
      expect(request.reason).toContain('read your project data')
      expect(request.temporary).toBeUndefined()
      expect(request.duration).toBeUndefined()
    })

    it('accepts overrides', () => {
      const request = createPermissionRequest('writeRepo', {
        reason: 'Custom reason',
        temporary: true,
        duration: 1800000
      })
      
      expect(request.permission).toBe('writeRepo')
      expect(request.reason).toBe('Custom reason')
      expect(request.temporary).toBe(true)
      expect(request.duration).toBe(1800000)
    })

    it('handles secret permissions with fallback reason', () => {
      const request = createPermissionRequest('secrets:unknown' as Permission)
      expect(request.reason).toContain('stored authentication tokens')
    })
  })
})