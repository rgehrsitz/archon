/**
 * Plugin Manifest System
 * 
 * Handles loading, parsing, and validating plugin manifests according to ADR-013.
 */

import type { PluginManifest, Permission } from './api.js';

/**
 * Supported plugin manifest versions.
 */
export const SUPPORTED_MANIFEST_VERSIONS = ['1.0'];

/**
 * Valid permission strings.
 */
export const VALID_PERMISSIONS: Permission[] = [
  'readRepo',
  'writeRepo', 
  'attachments',
  'net',
  'indexWrite',
  'ui'
  // Note: secrets permissions are validated separately due to dynamic nature
];

/**
 * Manifest validation error.
 */
export class ManifestError extends Error {
  constructor(
    message: string,
    public code: string,
    public field?: string
  ) {
    super(message);
    this.name = 'ManifestError';
  }
}

/**
 * Raw manifest structure as loaded from manifest.json.
 */
export interface RawManifest {
  id: string;
  name: string;
  version: string;
  type?: string;
  description?: string;
  author?: string;
  manifestVersion?: string;
  permissions: string[];
  entryPoint: string;
  minimumArchonVersion?: string;
  integrity?: string;
  
  // Additional metadata
  homepage?: string;
  repository?: string;
  license?: string;
  keywords?: string[];
}

/**
 * Plugin installation metadata.
 */
export interface PluginInstallation {
  manifest: PluginManifest;
  path: string;
  installedAt: Date;
  enabled: boolean;
  source: 'local' | 'url' | 'registry';
}

/**
 * Loads and parses a plugin manifest from JSON text.
 */
export function parseManifest(manifestText: string): RawManifest {
  let raw: any;
  
  try {
    raw = JSON.parse(manifestText);
  } catch (error) {
    throw new ManifestError(
      'Invalid JSON in manifest',
      'MANIFEST_INVALID_JSON'
    );
  }

  return raw as RawManifest;
}

/**
 * Validation result for manifest validation.
 */
export interface ValidationResult {
  valid: boolean;
  errors: string[];
  manifest?: PluginManifest;
}

/**
 * Validates a raw manifest and returns validation result.
 */
export function validateManifest(raw: RawManifest): ValidationResult {
  const errors: string[] = [];

  // Validate required fields
  if (!raw.id || typeof raw.id !== 'string') {
    errors.push('Missing or invalid "id" field');
  } else if (!isValidPluginId(raw.id)) {
    errors.push('Plugin id must be alphanumeric with dots, dashes, or underscores');
  }

  if (!raw.name || typeof raw.name !== 'string') {
    errors.push('Missing or invalid "name" field');
  } else if (raw.name.length > 200) {
    errors.push('Plugin name is too long (maximum 200 characters)');
  }

  if (!raw.version || typeof raw.version !== 'string') {
    errors.push('Missing or invalid "version" field');
  } else if (!isValidVersion(raw.version)) {
    errors.push('Version must be a valid semantic version (e.g., "1.0.0")');
  }

  if (!raw.entryPoint || typeof raw.entryPoint !== 'string') {
    errors.push('Missing or invalid "entryPoint" field');
  } else if (!isValidEntryPoint(raw.entryPoint)) {
    errors.push('entryPoint must be a valid JavaScript or TypeScript file');
  }

  if (!Array.isArray(raw.permissions)) {
    errors.push('Missing or invalid "permissions" field (must be array)');
  } else if (raw.permissions.length > 50) {
    errors.push('Too many permissions (maximum 50)');
  }

  // Validate optional description length
  if (raw.description && raw.description.length > 1000) {
    errors.push('Plugin description is too long (maximum 1000 characters)');
  }

  // Validate manifest version
  const manifestVersion = raw.manifestVersion || '1.0';
  if (!SUPPORTED_MANIFEST_VERSIONS.includes(manifestVersion)) {
    errors.push(`Unsupported manifest version: ${manifestVersion}`);
  }

  // Validate type field (required)
  if (!raw.type || typeof raw.type !== 'string') {
    errors.push('Missing or invalid "type" field');
  } else if (!isValidPluginType(raw.type)) {
    errors.push(`Invalid plugin type: ${raw.type}`);
  }

  // Validate permissions
  const permissions: Permission[] = [];
  if (Array.isArray(raw.permissions)) {
    for (const perm of raw.permissions) {
      if (typeof perm !== 'string') {
        errors.push(`Invalid permission type: ${typeof perm}`);
        continue;
      }

      // Check for secrets permissions (dynamic pattern)
      if (perm.startsWith('secrets:')) {
        const secretPattern = perm.substring(8);
        if (!secretPattern || secretPattern.length === 0) {
          errors.push(`Invalid secrets permission pattern: ${perm}`);
        } else {
          permissions.push(perm as Permission);
        }
      }
      // Check standard permissions
      else if (VALID_PERMISSIONS.includes(perm as Permission)) {
        permissions.push(perm as Permission);
      } else {
        errors.push(`Unknown permission: ${perm}`);
      }
    }
  }

  // Validate optional fields
  if (raw.minimumArchonVersion && !isValidVersion(raw.minimumArchonVersion)) {
    errors.push('minimumArchonVersion must be a valid semantic version');
  }

  if (raw.integrity && !isValidIntegrity(raw.integrity)) {
    errors.push('integrity must be a valid SHA-256 hash');
  }

  // Validate metadata URLs if present
  if (raw.homepage && !isValidUrl(raw.homepage)) {
    errors.push('Invalid homepage URL');
  }

  if (raw.repository && !isValidUrl(raw.repository)) {
    errors.push('Invalid repository URL');
  }

  const metadata = raw as any;
  if (metadata.metadata) {
    if (metadata.metadata.website !== undefined && !isValidUrl(metadata.metadata.website)) {
      errors.push('Invalid website URL in metadata');
    }
  }

  if (errors.length > 0) {
    return { valid: false, errors };
  }

  const manifest: PluginManifest = {
    id: raw.id,
    name: raw.name,
    version: raw.version,
    type: raw.type as any || 'Importer',
    description: raw.description,
    author: raw.author,
    permissions,
    entryPoint: raw.entryPoint,
    minimumArchonVersion: raw.minimumArchonVersion,
    integrity: raw.integrity
  };

  return { valid: true, errors: [], manifest };
}

/**
 * Legacy validation function that throws errors (for backward compatibility).
 */
export function validateManifestStrict(raw: RawManifest): PluginManifest {
  const result = validateManifest(raw);
  
  if (!result.valid) {
    throw new ManifestError(
      `Manifest validation failed: ${result.errors.join(', ')}`,
      'MANIFEST_VALIDATION_FAILED'
    );
  }
  
  return result.manifest!;
}

/**
 * Validates a plugin ID format.
 */
export function isValidPluginId(id: string): boolean {
  // Must be reverse domain notation (com.example.plugin)
  // Allow lowercase letters, numbers, dots, and hyphens only
  // Must contain at least one dot (domain separation)
  // Cannot start/end with dots or have consecutive dots
  // No spaces or special characters except dots and hyphens
  if (!id || id.length === 0) return false;
  if (!id.includes('.')) return false; // Must have domain separation
  if (id.startsWith('.') || id.endsWith('.')) return false;
  if (id.includes('..')) return false; // No consecutive dots
  if (id.includes(' ')) return false; // No spaces
  
  // Only lowercase alphanumeric, dots, and hyphens allowed
  // This regex rejects uppercase letters and special chars like @
  return /^[a-z0-9.-]+$/.test(id);
}

/**
 * Validates a semantic version format.
 */
export function isValidVersion(version: string): boolean {
  // Simple semver validation (major.minor.patch with optional pre-release and build metadata)
  return /^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$/.test(version);
}

/**
 * Validates an entry point file path.
 */
export function isValidEntryPoint(entryPoint: string): boolean {
  // Must be a JS or TS file, no path traversal, no URLs
  if (entryPoint.includes('..') || entryPoint.includes('://')) {
    return false;
  }
  return /\.(js|ts)$/i.test(entryPoint);
}

/**
 * Validates a plugin type.
 */
export function isValidPluginType(type: string): boolean {
  const validTypes = [
    'Importer', 'Exporter', 'Transformer', 'Validator', 'Panel', 
    'Provider', 'AttachmentProcessor', 'ConflictResolver', 
    'SearchIndexer', 'UIContrib'
  ];
  return validTypes.includes(type);
}

/**
 * Validates a URL format.
 */
export function isValidUrl(url: string): boolean {
  // Only allow http/https URLs, reject javascript:, file:, etc.
  if (!url || url.trim().length === 0) {
    return false;
  }
  
  try {
    const parsed = new URL(url);
    return parsed.protocol === 'http:' || parsed.protocol === 'https:';
  } catch {
    return false;
  }
}

/**
 * Validates an integrity hash format.
 */
export function isValidIntegrity(hash: string): boolean {
  // SHA-256 hex hash (64 characters)
  return /^[a-fA-F0-9]{64}$/.test(hash);
}

/**
 * Compares two semantic versions.
 * Returns -1 if a < b, 0 if a === b, 1 if a > b
 */
export function compareVersions(a: string, b: string): number {
  const parseVersion = (version: string) => {
    const [main, pre] = version.split('-');
    const [major, minor, patch] = main.split('.').map(Number);
    return { major, minor, patch, prerelease: pre };
  };

  const aVer = parseVersion(a);
  const bVer = parseVersion(b);

  // Compare major.minor.patch
  if (aVer.major !== bVer.major) return aVer.major - bVer.major;
  if (aVer.minor !== bVer.minor) return aVer.minor - bVer.minor;
  if (aVer.patch !== bVer.patch) return aVer.patch - bVer.patch;

  // Handle prerelease
  if (aVer.prerelease && !bVer.prerelease) return -1;
  if (!aVer.prerelease && bVer.prerelease) return 1;
  if (aVer.prerelease && bVer.prerelease) {
    return aVer.prerelease.localeCompare(bVer.prerelease);
  }

  return 0;
}

/**
 * Checks if a plugin manifest is compatible with the current Archon version.
 */
export function isCompatible(manifest: PluginManifest, archonVersion: string): boolean {
  if (!manifest.minimumArchonVersion) {
    return true; // No minimum version specified
  }

  return compareVersions(archonVersion, manifest.minimumArchonVersion) >= 0;
}

/**
 * Plugin discovery result.
 */
export interface DiscoveryResult {
  valid: PluginInstallation[];
  invalid: Array<{
    path: string;
    error: ManifestError;
  }>;
}

/**
 * Default manifest template for creating new plugins.
 */
export const DEFAULT_MANIFEST_TEMPLATE: Partial<RawManifest> = {
  manifestVersion: '1.0',
  version: '1.0.0',
  permissions: ['readRepo'],
  entryPoint: 'index.js'
};

/**
 * Creates a manifest template with the given plugin metadata.
 */
export function createManifestTemplate(
  id: string,
  name: string,
  options: Partial<RawManifest> = {}
): RawManifest {
  return {
    ...DEFAULT_MANIFEST_TEMPLATE,
    id,
    name,
    ...options
  } as RawManifest;
}

/**
 * Serializes a manifest to JSON with pretty formatting.
 */
export function serializeManifest(manifest: RawManifest): string {
  return JSON.stringify(manifest, null, 2);
}

/**
 * Permission display names for UI.
 */
export const PERMISSION_DESCRIPTIONS: Record<Permission | string, string> = {
  'readRepo': 'Read project data (nodes, properties, structure)',
  'writeRepo': 'Modify project data (create, update, delete nodes)',
  'attachments': 'Access file attachments (read and write)',
  'net': 'Access network resources (HTTP requests)',
  'indexWrite': 'Write to search index',
  'ui': 'Contribute to user interface (commands, panels, dialogs)',
  'secrets:*': 'Access stored secrets for external services'
};

/**
 * Gets a human-readable description for a permission.
 */
export function getPermissionDescription(permission: Permission): string {
  if (permission.startsWith('secrets:')) {
    const scope = permission.substring(8);
    return `Access secrets for ${scope.replace('*', 'any')} services`;
  }

  return PERMISSION_DESCRIPTIONS[permission] || `Unknown permission: ${permission}`;
}

/**
 * Groups permissions by risk level for UI display.
 */
export function categorizePermissions(permissions: Permission[]): {
  low: Permission[];
  medium: Permission[];
  high: Permission[];
} {
  const low: Permission[] = [];
  const medium: Permission[] = [];
  const high: Permission[] = [];

  for (const permission of permissions) {
    if (permission === 'readRepo' || permission === 'ui') {
      low.push(permission);
    } else if (permission === 'attachments' || permission === 'indexWrite') {
      medium.push(permission);
    } else {
      // writeRepo, net, secrets:* are high risk
      high.push(permission);
    }
  }

  return { low, medium, high };
}