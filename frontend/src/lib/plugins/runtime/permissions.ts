/**
 * Permission Management System
 * 
 * Manages plugin permissions and runtime enforcement as specified in ADR-013.
 */

import type { Permission } from '../api.js';
import { PluginError } from '../api.js';

/**
 * Permission grant status.
 */
export interface PermissionGrant {
  permission: Permission;
  granted: boolean;
  grantedAt?: Date;
  temporary?: boolean;
  expiresAt?: Date;
}

/**
 * Permission request context.
 */
export interface PermissionRequest {
  permission: Permission;
  reason?: string;
  temporary?: boolean;
  duration?: number; // milliseconds
}

/**
 * Permission manager that enforces access control.
 */
export class PermissionManager {
  private grants = new Map<Permission, PermissionGrant>();
  private requestCallbacks = new Set<PermissionRequestCallback>();

  constructor(private declaredPermissions: Permission[]) {
    // Initialize grants for declared permissions
    for (const permission of declaredPermissions) {
      this.grants.set(permission, {
        permission,
        granted: false // Requires explicit grant
      });
    }
  }

  /**
   * Checks if a permission is available and granted.
   */
  hasPermission(permission: Permission): boolean {
    // Check for exact match first
    if (this.grants.has(permission)) {
      const grant = this.grants.get(permission)!;
      
      // Check if temporary permission has expired
      if (grant.temporary && grant.expiresAt && grant.expiresAt < new Date()) {
        grant.granted = false;
        grant.temporary = false;
        grant.expiresAt = undefined;
      }
      
      return grant.granted;
    }

    // Check for pattern matches (e.g., secrets:*)
    for (const [grantedPerm, grant] of this.grants) {
      if (this.matchesPattern(grantedPerm, permission)) {
        if (grant.temporary && grant.expiresAt && grant.expiresAt < new Date()) {
          grant.granted = false;
          grant.temporary = false;
          grant.expiresAt = undefined;
          continue;
        }
        return grant.granted;
      }
    }

    return false;
  }

  /**
   * Requests permission from the user (triggers UI consent dialog).
   */
  async requestPermission(request: PermissionRequest): Promise<boolean> {
    // Check if permission is declared in manifest
    const isDeclared = this.declaredPermissions.some(declared => 
      declared === request.permission || this.matchesPattern(declared, request.permission)
    );

    if (!isDeclared) {
      throw new PluginError(
        `Permission ${request.permission} not declared in plugin manifest`,
        'PERMISSION_NOT_DECLARED',
        'permissions'
      );
    }

    // If already granted, return true
    if (this.hasPermission(request.permission)) {
      return true;
    }

    // Trigger permission request callbacks (UI will handle the actual prompt)
    const results = await Promise.all(
      Array.from(this.requestCallbacks).map(callback => callback(request))
    );

    // All callbacks must approve
    const approved = results.every(result => result === true);

    if (approved) {
      this.grantPermission(request.permission, {
        temporary: request.temporary,
        duration: request.duration
      });
    }

    return approved;
  }

  /**
   * Grants a permission (typically called after user consent).
   */
  grantPermission(
    permission: Permission, 
    options: { temporary?: boolean; duration?: number } = {}
  ): void {
    const grant: PermissionGrant = {
      permission,
      granted: true,
      grantedAt: new Date(),
      temporary: options.temporary,
      expiresAt: options.duration 
        ? new Date(Date.now() + options.duration)
        : undefined
    };

    this.grants.set(permission, grant);
  }

  /**
   * Revokes a permission.
   */
  revokePermission(permission: Permission): void {
    const grant = this.grants.get(permission);
    if (grant) {
      grant.granted = false;
      grant.temporary = false;
      grant.expiresAt = undefined;
    }
  }

  /**
   * Gets all granted permissions.
   */
  getGrantedPermissions(): PermissionGrant[] {
    return Array.from(this.grants.values()).filter(grant => grant.granted);
  }

  /**
   * Gets all declared permissions (from manifest).
   */
  getDeclaredPermissions(): Permission[] {
    return [...this.declaredPermissions];
  }

  /**
   * Registers a callback for permission requests.
   */
  onPermissionRequest(callback: PermissionRequestCallback): void {
    this.requestCallbacks.add(callback);
  }

  /**
   * Removes a permission request callback.
   */
  offPermissionRequest(callback: PermissionRequestCallback): void {
    this.requestCallbacks.delete(callback);
  }

  /**
   * Checks if a permission pattern matches a specific permission.
   */
  private matchesPattern(pattern: Permission, permission: Permission): boolean {
    if (pattern === permission) {
      return true;
    }

    // Handle secrets:* patterns
    if (pattern.startsWith('secrets:') && permission.startsWith('secrets:')) {
      const patternSuffix = pattern.substring(8);
      const permissionSuffix = permission.substring(8);

      // Handle wildcard patterns
      if (patternSuffix.endsWith('*')) {
        const prefix = patternSuffix.slice(0, -1);
        return permissionSuffix.startsWith(prefix);
      }
    }

    return false;
  }
}

/**
 * Permission request callback type.
 */
export type PermissionRequestCallback = (request: PermissionRequest) => Promise<boolean>;

/**
 * Permission category for UI grouping.
 */
export enum PermissionCategory {
  LOW_RISK = 'low',
  MEDIUM_RISK = 'medium',
  HIGH_RISK = 'high'
}

/**
 * Gets the risk category for a permission.
 */
export function getPermissionCategory(permission: Permission): PermissionCategory {
  switch (permission) {
    case 'readRepo':
    case 'ui':
      return PermissionCategory.LOW_RISK;
    
    case 'attachments':
    case 'indexWrite':
      return PermissionCategory.MEDIUM_RISK;
    
    case 'writeRepo':
    case 'net':
    default:
      if (permission.startsWith('secrets:')) {
        return PermissionCategory.HIGH_RISK;
      }
      return PermissionCategory.HIGH_RISK;
  }
}

/**
 * Gets a human-readable description of what a permission allows.
 */
export function getPermissionDescription(permission: Permission): string {
  switch (permission) {
    case 'readRepo':
      return 'Read project data, including nodes, properties, and structure';
    case 'writeRepo':
      return 'Create, modify, and delete project data';
    case 'attachments':
      return 'Access and modify file attachments';
    case 'net':
      return 'Make network requests to external services';
    case 'indexWrite':
      return 'Add entries to the search index';
    case 'ui':
      return 'Add commands, panels, and other UI elements';
    default:
      if (permission.startsWith('secrets:')) {
        const scope = permission.substring(8);
        return `Access stored secrets for ${scope.replace('*', 'any')} services`;
      }
      return `Unknown permission: ${permission}`;
  }
}

/**
 * Permission risk level indicators for UI.
 */
export const PERMISSION_RISK_ICONS = {
  [PermissionCategory.LOW_RISK]: '🟢',
  [PermissionCategory.MEDIUM_RISK]: '🟡',
  [PermissionCategory.HIGH_RISK]: '🔴'
};

/**
 * Default permission request reasons.
 */
export const PERMISSION_REASONS = {
  'readRepo': 'This plugin needs to read your project data to function',
  'writeRepo': 'This plugin needs to modify your project data',
  'attachments': 'This plugin needs to work with file attachments',
  'net': 'This plugin needs to connect to external services',
  'indexWrite': 'This plugin needs to add content to the search index',
  'ui': 'This plugin wants to add interface elements',
  'secrets:*': 'This plugin needs to access stored authentication tokens'
};

/**
 * Creates a standard permission request with default reason.
 */
export function createPermissionRequest(
  permission: Permission,
  overrides: Partial<PermissionRequest> = {}
): PermissionRequest {
  const defaultReason = PERMISSION_REASONS[permission as keyof typeof PERMISSION_REASONS] 
    || PERMISSION_REASONS['secrets:*'];
  
  return {
    permission,
    reason: defaultReason,
    ...overrides
  };
}