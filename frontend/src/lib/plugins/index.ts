/**
 * Plugin System Index
 * 
 * Main entry point for the Archon plugin system.
 * Exports all components needed for plugin development and integration.
 */

// Core API types and interfaces
export * from './api.js';
export { PluginError } from './api.js';

// Manifest system
export * from './manifest.js';

// Runtime system
export { PluginSandbox, type SandboxConfig, DEFAULT_SANDBOX_CONFIG } from './runtime/sandbox.js';
export { HostService } from './runtime/host-services.js';
export { PermissionManager, type PermissionRequest, type PermissionGrant } from './runtime/permissions.js';
export { UIPermissionManager, type PermissionConsentState } from './runtime/ui-permission-manager.js';
export { UIService } from './runtime/ui-service.js';
export { SecretService, COMMON_SECRET_PATTERNS, SecretConfigHelper } from './runtime/secret-service.js';
export { PluginManager, type LoadedPlugin, type PluginManagerConfig } from './runtime/plugin-manager.js';

// Discovery and loading
export { PluginLoader, type PluginSource, DEFAULT_DISCOVERY_OPTIONS } from './discovery/plugin-loader.js';

// Example plugins for reference
export { register as csvImporterRegister } from './examples/csv-importer.js';
export { register as dataValidatorRegister } from './examples/data-validator.js';

/**
 * Plugin System Factory
 * 
 * Convenience factory for creating a complete plugin system instance
 * with all components wired together.
 */
export class ArchonPluginSystem {
  public readonly pluginManager: PluginManager;
  public readonly pluginLoader: PluginLoader;

  constructor(config: {
    pluginManager?: PluginManagerConfig;
    discoveryOptions?: Parameters<typeof PluginLoader>[0];
  } = {}) {
    this.pluginManager = new PluginManager(config.pluginManager);
    this.pluginLoader = new PluginLoader(config.discoveryOptions);
  }

  /**
   * Initialize the plugin system by discovering and loading available plugins.
   */
  async initialize(): Promise<void> {
    await this.pluginLoader.discoverPlugins();
  }

  /**
   * Install a plugin from a discovered source.
   */
  async installPlugin(pluginId: string): Promise<void> {
    const source = this.pluginLoader.getPlugin(pluginId);
    if (!source) {
      throw new Error(`Plugin ${pluginId} not found in discovered plugins`);
    }

    const { manifest, code } = await this.pluginLoader.loadPlugin(source);
    await this.pluginManager.loadPlugin(manifest, code);
  }

  /**
   * Get system-wide statistics.
   */
  getSystemStats() {
    const discovered = this.pluginLoader.getDiscoveredPlugins();
    const managerStats = this.pluginManager.getStatistics();

    return {
      discovered: discovered.length,
      ...managerStats
    };
  }

  /**
   * Shutdown the plugin system.
   */
  async shutdown(): Promise<void> {
    await this.pluginManager.shutdown();
  }
}

/**
 * Default plugin system instance for convenience.
 */
export const defaultPluginSystem = new ArchonPluginSystem();