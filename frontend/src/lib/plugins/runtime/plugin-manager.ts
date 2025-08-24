/**
 * Plugin Manager
 * 
 * Orchestrates plugin lifecycle, permission management, and UI integration.
 * Central coordinator for the plugin system.
 */

import type { PluginManifest, Plugin } from '../api.js';
import { PluginSandbox, type SandboxConfig } from './sandbox.js';
import { validateManifest } from '../manifest.js';
import { PluginError } from '../api.js';
import type { UIPermissionManager } from './ui-permission-manager.js';

export interface LoadedPlugin {
  manifest: PluginManifest;
  sandbox: PluginSandbox;
  permissionManager: UIPermissionManager;
  active: boolean;
  loadedAt: Date;
}

export interface PluginManagerConfig {
  sandboxConfig?: Partial<SandboxConfig>;
  autoActivate?: boolean;
  developmentMode?: boolean;
}

/**
 * Manages plugin lifecycle and provides integration points for the UI.
 */
export class PluginManager {
  private plugins = new Map<string, LoadedPlugin>();
  private config: PluginManagerConfig;

  constructor(config: PluginManagerConfig = {}) {
    this.config = {
      autoActivate: true,
      developmentMode: false,
      ...config
    };
  }

  /**
   * Loads a plugin from manifest and code.
   */
  async loadPlugin(manifest: PluginManifest, code: string): Promise<string> {
    // Validate manifest
    const validation = validateManifest(manifest);
    if (!validation.valid) {
      throw new PluginError(
        `Invalid plugin manifest: ${validation.errors.join(', ')}`,
        'INVALID_MANIFEST',
        manifest.id
      );
    }

    // Check if plugin is already loaded
    if (this.plugins.has(manifest.id)) {
      throw new PluginError(
        `Plugin ${manifest.id} is already loaded`,
        'PLUGIN_ALREADY_LOADED',
        manifest.id
      );
    }

    // Create sandbox with config
    const sandboxConfig = {
      timeoutMs: 60000,
      memoryLimitMB: 256,
      allowedOrigins: [],
      ...this.config.sandboxConfig
    };

    const sandbox = new PluginSandbox(manifest, sandboxConfig);

    try {
      // Initialize sandbox
      await sandbox.initialize(code);

      // Get permission manager from host service
      const permissionManager = sandbox.getHostService().getPermissionManager();

      // Create plugin entry
      const plugin: LoadedPlugin = {
        manifest,
        sandbox,
        permissionManager,
        active: false,
        loadedAt: new Date()
      };

      this.plugins.set(manifest.id, plugin);

      // Auto-activate if configured
      if (this.config.autoActivate) {
        await this.activatePlugin(manifest.id);
      }

      return manifest.id;
    } catch (error) {
      // Cleanup on failure
      await sandbox.terminate();
      throw error;
    }
  }

  /**
   * Activates a loaded plugin by calling its register function.
   */
  async activatePlugin(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      throw new PluginError(`Plugin ${pluginId} not found`, 'PLUGIN_NOT_FOUND', pluginId);
    }

    if (plugin.active) {
      return; // Already active
    }

    try {
      // Execute the plugin's register function
      const result = await plugin.sandbox.execute('register');
      
      if (!result.success) {
        throw result.error || new PluginError(
          'Plugin activation failed',
          'ACTIVATION_FAILED',
          pluginId
        );
      }

      plugin.active = true;
    } catch (error) {
      throw new PluginError(
        `Failed to activate plugin: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'ACTIVATION_FAILED',
        pluginId
      );
    }
  }

  /**
   * Deactivates a plugin and cleans up resources.
   */
  async deactivatePlugin(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      return; // Not loaded
    }

    plugin.active = false;
    
    // Close any open permission dialogs
    plugin.permissionManager.closeConsentDialog();
    
    // TODO: Call plugin deactivation hook if exists
  }

  /**
   * Unloads a plugin completely.
   */
  async unloadPlugin(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      return; // Not loaded
    }

    // Deactivate first
    await this.deactivatePlugin(pluginId);

    // Terminate sandbox
    await plugin.sandbox.terminate();

    // Remove from loaded plugins
    this.plugins.delete(pluginId);
  }

  /**
   * Gets a loaded plugin by ID.
   */
  getPlugin(pluginId: string): LoadedPlugin | null {
    return this.plugins.get(pluginId) || null;
  }

  /**
   * Lists all loaded plugins.
   */
  getLoadedPlugins(): LoadedPlugin[] {
    return Array.from(this.plugins.values());
  }

  /**
   * Lists active plugins.
   */
  getActivePlugins(): LoadedPlugin[] {
    return Array.from(this.plugins.values()).filter(p => p.active);
  }

  /**
   * Gets plugin permission manager for UI integration.
   */
  getPluginPermissionManager(pluginId: string): UIPermissionManager | null {
    const plugin = this.plugins.get(pluginId);
    return plugin?.permissionManager || null;
  }

  /**
   * Executes a plugin method (for testing/debugging).
   */
  async executePluginMethod(pluginId: string, method: string, ...args: any[]): Promise<any> {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      throw new PluginError(`Plugin ${pluginId} not found`, 'PLUGIN_NOT_FOUND', pluginId);
    }

    if (!plugin.active) {
      throw new PluginError(`Plugin ${pluginId} is not active`, 'PLUGIN_NOT_ACTIVE', pluginId);
    }

    const result = await plugin.sandbox.execute(method, ...args);
    
    if (!result.success) {
      throw result.error || new PluginError(
        'Plugin method execution failed',
        'EXECUTION_FAILED',
        pluginId
      );
    }

    return result.result;
  }

  /**
   * Grants development permissions to all loaded plugins (for testing).
   */
  grantDevelopmentPermissions(): void {
    if (!this.config.developmentMode) {
      console.warn('Development permissions can only be granted in development mode');
      return;
    }

    for (const plugin of this.plugins.values()) {
      plugin.permissionManager.preGrantPermissions(plugin.manifest.permissions);
    }
  }

  /**
   * Revokes all permissions for all plugins.
   */
  revokeAllPermissions(): void {
    for (const plugin of this.plugins.values()) {
      plugin.permissionManager.resetAllPermissions();
    }
  }

  /**
   * Gets plugin system statistics.
   */
  getStatistics(): {
    totalLoaded: number;
    totalActive: number;
    totalPermissionsGranted: number;
    memoryUsage: number; // Approximation
  } {
    const plugins = this.getLoadedPlugins();
    let totalPermissionsGranted = 0;

    for (const plugin of plugins) {
      totalPermissionsGranted += plugin.permissionManager.getPermissionSummary().granted;
    }

    return {
      totalLoaded: plugins.length,
      totalActive: this.getActivePlugins().length,
      totalPermissionsGranted,
      memoryUsage: plugins.length * 50 // Rough estimate: 50MB per plugin
    };
  }

  /**
   * Shuts down all plugins and cleans up resources.
   */
  async shutdown(): Promise<void> {
    const pluginIds = Array.from(this.plugins.keys());
    
    await Promise.all(
      pluginIds.map(id => this.unloadPlugin(id))
    );
  }
}