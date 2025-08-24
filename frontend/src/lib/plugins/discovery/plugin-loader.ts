/**
 * Plugin Loader
 * 
 * Handles discovery and loading of plugins from various sources.
 * Supports local file system plugins and future remote plugin registries.
 */

import { validateManifest } from '../manifest.js';
import { PluginError } from '../api.js';
import type { PluginManifest } from '../api.js';

export interface PluginSource {
  id: string;
  name: string;
  description: string;
  version: string;
  manifest: PluginManifest;
  code: string;
  source: 'local' | 'remote' | 'bundled';
  path?: string;
  url?: string;
}

export interface PluginDiscoveryOptions {
  includeLocalPlugins?: boolean;
  includeBundledPlugins?: boolean;
  pluginDirectories?: string[];
  remoteRegistries?: string[];
}

/**
 * Default plugin discovery options.
 */
export const DEFAULT_DISCOVERY_OPTIONS: PluginDiscoveryOptions = {
  includeLocalPlugins: true,
  includeBundledPlugins: true,
  pluginDirectories: [
    './plugins',
    '~/.archon/plugins',
    '%APPDATA%/Archon/plugins' // Windows
  ],
  remoteRegistries: []
};

/**
 * Plugin loader that discovers and loads plugins from various sources.
 */
export class PluginLoader {
  private options: PluginDiscoveryOptions;
  private discoveredPlugins = new Map<string, PluginSource>();

  constructor(options: PluginDiscoveryOptions = DEFAULT_DISCOVERY_OPTIONS) {
    this.options = { ...DEFAULT_DISCOVERY_OPTIONS, ...options };
  }

  /**
   * Discovers available plugins from all configured sources.
   */
  async discoverPlugins(): Promise<PluginSource[]> {
    const plugins: PluginSource[] = [];

    // Discover bundled plugins
    if (this.options.includeBundledPlugins) {
      plugins.push(...await this.discoverBundledPlugins());
    }

    // Discover local plugins (if in a non-browser environment)
    if (this.options.includeLocalPlugins && typeof window === 'undefined') {
      for (const directory of this.options.pluginDirectories || []) {
        plugins.push(...await this.discoverLocalPlugins(directory));
      }
    }

    // Update discovered plugins cache
    this.discoveredPlugins.clear();
    for (const plugin of plugins) {
      this.discoveredPlugins.set(plugin.id, plugin);
    }

    return plugins;
  }

  /**
   * Gets a discovered plugin by ID.
   */
  getPlugin(pluginId: string): PluginSource | null {
    return this.discoveredPlugins.get(pluginId) || null;
  }

  /**
   * Gets all discovered plugins.
   */
  getDiscoveredPlugins(): PluginSource[] {
    return Array.from(this.discoveredPlugins.values());
  }

  /**
   * Loads a plugin from a source object.
   */
  async loadPlugin(pluginSource: PluginSource): Promise<{ manifest: PluginManifest; code: string }> {
    // Validate manifest
    const validation = validateManifest(pluginSource.manifest);
    if (!validation.valid) {
      throw new PluginError(
        `Invalid plugin manifest for ${pluginSource.id}: ${validation.errors.join(', ')}`,
        'INVALID_MANIFEST',
        pluginSource.id
      );
    }

    return {
      manifest: pluginSource.manifest,
      code: pluginSource.code
    };
  }

  /**
   * Loads a plugin directly from manifest and code strings.
   */
  async loadPluginFromStrings(
    manifestJson: string, 
    code: string, 
    source: 'local' | 'remote' | 'bundled' = 'local'
  ): Promise<{ manifest: PluginManifest; code: string }> {
    try {
      const manifest = JSON.parse(manifestJson) as PluginManifest;
      
      // Validate manifest
      const validation = validateManifest(manifest);
      if (!validation.valid) {
        throw new PluginError(
          `Invalid plugin manifest: ${validation.errors.join(', ')}`,
          'INVALID_MANIFEST',
          manifest.id || 'unknown'
        );
      }

      return { manifest, code };
    } catch (error) {
      if (error instanceof PluginError) {
        throw error;
      }
      throw new PluginError(
        `Failed to parse plugin manifest: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'MANIFEST_PARSE_ERROR',
        'unknown'
      );
    }
  }

  /**
   * Discovers bundled plugins that ship with Archon.
   */
  private async discoverBundledPlugins(): Promise<PluginSource[]> {
    // For now, return a sample bundled plugin
    // In production, this would read from a bundled plugins directory
    return [
      {
        id: 'archon.sample-importer',
        name: 'Sample Data Importer',
        description: 'A sample plugin that demonstrates data import capabilities',
        version: '1.0.0',
        source: 'bundled' as const,
        manifest: {
          id: 'archon.sample-importer',
          name: 'Sample Data Importer',
          version: '1.0.0',
          description: 'A sample plugin that demonstrates data import capabilities',
          author: 'Archon Team',
          license: 'MIT',
          type: 'Importer',
          permissions: ['readRepo', 'writeRepo'],
          entryPoint: 'index.js',
          archonVersion: '^1.0.0',
          metadata: {
            category: 'Data Import',
            tags: ['sample', 'demo', 'csv'],
            website: 'https://archon.dev'
          }
        },
        code: `
// Sample Importer Plugin
function register(context) {
  const { host, logger, manifest } = context;
  
  logger.info('Sample Importer Plugin loaded');
  
  // Register the import functionality
  return {
    id: manifest.id,
    name: manifest.name,
    
    // Import function that creates sample nodes
    async import(data, parentId) {
      logger.info('Importing sample data');
      
      const mutations = [];
      const lines = data.split('\\n');
      
      for (const line of lines) {
        if (line.trim()) {
          mutations.push({
            type: 'create',
            parentId,
            data: {
              name: line.trim(),
              description: 'Imported from sample data',
              properties: {
                'import.source': 'sample-importer',
                'import.timestamp': new Date().toISOString()
              }
            }
          });
        }
      }
      
      if (mutations.length > 0) {
        await host.apply(mutations);
        logger.info(\`Imported \${mutations.length} items\`);
      }
      
      return { imported: mutations.length };
    },
    
    // Supported file types
    supportedTypes: ['txt', 'csv'],
    
    // UI metadata
    displayName: 'Sample Text Importer',
    description: 'Import line-separated text as nodes'
  };
}

// Export for plugin system
if (typeof module !== 'undefined') {
  module.exports = { register };
}
`
      }
    ];
  }

  /**
   * Discovers plugins in a local directory.
   * Note: This only works in non-browser environments.
   */
  private async discoverLocalPlugins(directory: string): Promise<PluginSource[]> {
    // Browser environment - return empty array
    if (typeof window !== 'undefined') {
      return [];
    }

    try {
      // This would use Node.js filesystem APIs
      // For now, return empty array as this is frontend code
      console.warn(`Local plugin discovery not implemented for directory: ${directory}`);
      return [];
    } catch (error) {
      console.warn(`Failed to discover plugins in ${directory}:`, error);
      return [];
    }
  }

  /**
   * Installs a plugin from a URL or file path.
   */
  async installPlugin(source: string): Promise<PluginSource> {
    if (source.startsWith('http://') || source.startsWith('https://')) {
      return this.installRemotePlugin(source);
    } else {
      return this.installLocalPlugin(source);
    }
  }

  /**
   * Installs a plugin from a remote URL.
   */
  private async installRemotePlugin(url: string): Promise<PluginSource> {
    throw new PluginError(
      'Remote plugin installation not yet implemented',
      'NOT_IMPLEMENTED',
      'unknown'
    );
  }

  /**
   * Installs a plugin from a local file.
   */
  private async installLocalPlugin(path: string): Promise<PluginSource> {
    throw new PluginError(
      'Local plugin installation not yet implemented',
      'NOT_IMPLEMENTED',
      'unknown'
    );
  }

  /**
   * Validates a plugin package structure.
   */
  private validatePluginPackage(manifest: PluginManifest, code: string): boolean {
    // Basic validation
    if (!manifest.id || !manifest.name || !manifest.version) {
      return false;
    }

    // Check if code contains a register function
    if (!code.includes('function register') && !code.includes('register:') && !code.includes('register =')) {
      return false;
    }

    return true;
  }
}