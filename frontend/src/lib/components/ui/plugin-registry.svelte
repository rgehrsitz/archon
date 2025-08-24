<!--
Plugin Registry Component

Provides a comprehensive interface for managing plugins - discovery, installation,
activation, deactivation, and permission management.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import Button from './button.svelte';
  import { Badge } from './badge.js';
  import { Separator } from './separator.js';
  import PermissionConsentDialog from './permission-consent-dialog.svelte';
  import PluginPermissionPanel from './plugin-permission-panel.svelte';
  import { PluginLoader, type PluginSource } from '../../plugins/discovery/plugin-loader.js';
  import { PluginManager, type LoadedPlugin } from '../../plugins/runtime/plugin-manager.js';
  import type { PermissionConsentState } from '../../plugins/runtime/ui-permission-manager.js';

  interface Props {
    pluginManager: PluginManager;
    developmentMode?: boolean;
  }

  let { pluginManager, developmentMode = false }: Props = $props();

  let pluginLoader = new PluginLoader();
  let availablePlugins = $state<PluginSource[]>([]);
  let loadedPlugins = $state<LoadedPlugin[]>([]);
  let selectedPlugin = $state<LoadedPlugin | null>(null);
  let showPermissionPanel = $state(false);
  let isLoading = $state(false);
  let error = $state<string | null>(null);

  // Permission consent dialog state
  let consentState = $state<PermissionConsentState>({
    isOpen: false,
    request: null,
    pluginName: '',
    pluginId: '',
    resolve: null
  });

  async function refreshAvailablePlugins() {
    try {
      isLoading = true;
      error = null;
      availablePlugins = await pluginLoader.discoverPlugins();
    } catch (err) {
      error = `Failed to discover plugins: ${err instanceof Error ? err.message : 'Unknown error'}`;
    } finally {
      isLoading = false;
    }
  }

  function refreshLoadedPlugins() {
    loadedPlugins = pluginManager.getLoadedPlugins();
  }

  async function handleInstallPlugin(pluginSource: PluginSource) {
    try {
      isLoading = true;
      error = null;
      
      const { manifest, code } = await pluginLoader.loadPlugin(pluginSource);
      await pluginManager.loadPlugin(manifest, code);
      
      refreshLoadedPlugins();
    } catch (err) {
      error = `Failed to install plugin: ${err instanceof Error ? err.message : 'Unknown error'}`;
    } finally {
      isLoading = false;
    }
  }

  async function handleUninstallPlugin(pluginId: string) {
    try {
      await pluginManager.unloadPlugin(pluginId);
      refreshLoadedPlugins();
      
      // Close permission panel if it was for this plugin
      if (selectedPlugin?.manifest.id === pluginId) {
        selectedPlugin = null;
        showPermissionPanel = false;
      }
    } catch (err) {
      error = `Failed to uninstall plugin: ${err instanceof Error ? err.message : 'Unknown error'}`;
    }
  }

  async function handleActivatePlugin(pluginId: string) {
    try {
      await pluginManager.activatePlugin(pluginId);
      refreshLoadedPlugins();
    } catch (err) {
      error = `Failed to activate plugin: ${err instanceof Error ? err.message : 'Unknown error'}`;
    }
  }

  async function handleDeactivatePlugin(pluginId: string) {
    try {
      await pluginManager.deactivatePlugin(pluginId);
      refreshLoadedPlugins();
    } catch (err) {
      error = `Failed to deactivate plugin: ${err instanceof Error ? err.message : 'Unknown error'}`;
    }
  }

  function handleShowPermissions(plugin: LoadedPlugin) {
    selectedPlugin = plugin;
    showPermissionPanel = true;
  }

  function handleConsentResult(granted: boolean, options: { temporary?: boolean; duration?: number } = {}) {
    selectedPlugin?.permissionManager.handleConsentResult(granted, options);
  }

  function handleDevelopmentPermissions() {
    if (developmentMode) {
      pluginManager.grantDevelopmentPermissions();
      refreshLoadedPlugins();
    }
  }

  function getPluginStatusColor(plugin: LoadedPlugin): string {
    if (!plugin.active) return 'bg-gray-100 text-gray-700';
    
    const summary = plugin.permissionManager.getPermissionSummary();
    if (summary.granted === summary.total) return 'bg-green-100 text-green-700';
    if (summary.granted > 0) return 'bg-yellow-100 text-yellow-700';
    return 'bg-red-100 text-red-700';
  }

  function getPluginStatusText(plugin: LoadedPlugin): string {
    if (!plugin.active) return 'Inactive';
    
    const summary = plugin.permissionManager.getPermissionSummary();
    if (summary.granted === summary.total) return 'Active';
    if (summary.granted > 0) return 'Partial Permissions';
    return 'No Permissions';
  }

  function isPluginInstalled(pluginId: string): boolean {
    return loadedPlugins.some(p => p.manifest.id === pluginId);
  }

  onMount(() => {
    refreshAvailablePlugins();
    refreshLoadedPlugins();

    // Set up permission consent dialog for any loaded plugins
    for (const plugin of loadedPlugins) {
      plugin.permissionManager.getConsentState().subscribe(state => {
        consentState = state;
      });
    }
  });
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-2xl font-bold">Plugin Registry</h2>
      <p class="text-muted-foreground">Manage your Archon plugins</p>
    </div>
    <div class="flex items-center gap-2">
      {#if developmentMode}
        <Button variant="outline" onclick={handleDevelopmentPermissions}>
          Grant Dev Permissions
        </Button>
      {/if}
      <Button onclick={refreshAvailablePlugins} disabled={isLoading}>
        {isLoading ? 'Refreshing...' : 'Refresh'}
      </Button>
    </div>
  </div>

  <!-- Error Display -->
  {#if error}
    <div class="rounded-md border border-red-200 bg-red-50 p-3 text-red-700">
      <strong>Error:</strong> {error}
      <Button variant="ghost" size="sm" onclick={() => error = null} class="ml-2">
        Dismiss
      </Button>
    </div>
  {/if}

  <!-- Statistics -->
  {#snippet statsGrid()}
    {@const stats = pluginManager.getStatistics()}
    <div class="grid grid-cols-4 gap-4">
      <div class="rounded-md border bg-card p-3 text-center">
        <div class="text-2xl font-bold">{availablePlugins.length}</div>
        <div class="text-sm text-muted-foreground">Available</div>
      </div>
      <div class="rounded-md border bg-card p-3 text-center">
        <div class="text-2xl font-bold">{stats.totalLoaded}</div>
        <div class="text-sm text-muted-foreground">Installed</div>
      </div>
      <div class="rounded-md border bg-card p-3 text-center">
        <div class="text-2xl font-bold">{stats.totalActive}</div>
        <div class="text-sm text-muted-foreground">Active</div>
      </div>
      <div class="rounded-md border bg-card p-3 text-center">
        <div class="text-2xl font-bold">{stats.totalPermissionsGranted}</div>
        <div class="text-sm text-muted-foreground">Permissions</div>
      </div>
    </div>
  {/snippet}
  
  {@render statsGrid()}

  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- Available Plugins -->
    <div class="space-y-4">
      <h3 class="text-lg font-semibold">Available Plugins</h3>
      
      {#if availablePlugins.length === 0 && !isLoading}
        <div class="text-center text-muted-foreground text-sm py-8 border rounded-md">
          No plugins found. Check your plugin directories or refresh to discover plugins.
        </div>
      {:else}
        <div class="space-y-2">
          {#each availablePlugins as plugin (plugin.id)}
            <div class="rounded-md border p-4">
              <div class="flex items-start justify-between">
                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-2">
                    <h4 class="font-medium">{plugin.name}</h4>
                    <Badge variant="outline">{plugin.version}</Badge>
                    <Badge variant="secondary">{plugin.source}</Badge>
                    <Badge variant="outline">{plugin.manifest.type}</Badge>
                  </div>
                  <p class="text-sm text-muted-foreground mb-2">{plugin.description}</p>
                  <div class="text-xs text-muted-foreground">
                    {plugin.manifest.permissions.length} permission{plugin.manifest.permissions.length !== 1 ? 's' : ''}
                  </div>
                </div>
                <div class="flex items-center gap-2">
                  {#if isPluginInstalled(plugin.id)}
                    <Badge variant="outline">Installed</Badge>
                  {:else}
                    <Button 
                      size="sm" 
                      onclick={() => handleInstallPlugin(plugin)}
                      disabled={isLoading}
                    >
                      Install
                    </Button>
                  {/if}
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Installed Plugins -->
    <div class="space-y-4">
      <h3 class="text-lg font-semibold">Installed Plugins</h3>
      
      {#if loadedPlugins.length === 0}
        <div class="text-center text-muted-foreground text-sm py-8 border rounded-md">
          No plugins installed. Install plugins from the available list to get started.
        </div>
      {:else}
        <div class="space-y-2">
          {#each loadedPlugins as plugin (plugin.manifest.id)}
            {@const statusColor = getPluginStatusColor(plugin)}
            {@const statusText = getPluginStatusText(plugin)}
            {@const permissionSummary = plugin.permissionManager.getPermissionSummary()}
            
            <div class="rounded-md border p-4">
              <div class="flex items-start justify-between mb-3">
                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-1">
                    <h4 class="font-medium">{plugin.manifest.name}</h4>
                    <Badge variant="outline">{plugin.manifest.version}</Badge>
                    <Badge variant="outline">{plugin.manifest.type}</Badge>
                  </div>
                  <p class="text-sm text-muted-foreground mb-2">{plugin.manifest.description}</p>
                  <div class="text-xs {statusColor} rounded px-2 py-1 inline-block">
                    {statusText}
                  </div>
                </div>
              </div>
              
              <Separator class="my-3" />
              
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-4 text-sm">
                  <div>
                    Permissions: <strong>{permissionSummary.granted}/{permissionSummary.total}</strong>
                  </div>
                  <div class="text-muted-foreground">
                    Loaded {new Date(plugin.loadedAt).toLocaleTimeString()}
                  </div>
                </div>
                
                <div class="flex items-center gap-2">
                  <Button 
                    variant="outline" 
                    size="sm"
                    onclick={() => handleShowPermissions(plugin)}
                  >
                    Permissions
                  </Button>
                  
                  {#if plugin.active}
                    <Button 
                      variant="outline" 
                      size="sm"
                      onclick={() => handleDeactivatePlugin(plugin.manifest.id)}
                    >
                      Deactivate
                    </Button>
                  {:else}
                    <Button 
                      size="sm"
                      onclick={() => handleActivatePlugin(plugin.manifest.id)}
                    >
                      Activate
                    </Button>
                  {/if}
                  
                  <Button 
                    variant="destructive" 
                    size="sm"
                    onclick={() => handleUninstallPlugin(plugin.manifest.id)}
                  >
                    Uninstall
                  </Button>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  {#if developmentMode}
    <div class="text-xs text-muted-foreground bg-yellow-50 border border-yellow-200 rounded-md p-3">
      <strong>Development Mode:</strong> Additional debugging features are available. 
      Use "Grant Dev Permissions" to automatically grant all permissions for testing.
    </div>
  {/if}
</div>

<!-- Permission Consent Dialog -->
{#if consentState.isOpen && consentState.request}
  <PermissionConsentDialog
    bind:open={consentState.isOpen}
    request={consentState.request}
    pluginName={consentState.pluginName}
    pluginId={consentState.pluginId}
    on:grant={(e) => handleConsentResult(true, e.detail)}
    on:deny={() => handleConsentResult(false)}
  />
{/if}

<!-- Permission Panel Dialog -->
{#if showPermissionPanel && selectedPlugin}
  <div class="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center">
    <div class="bg-background rounded-lg border shadow-lg w-full max-w-2xl max-h-[80vh] overflow-y-auto">
      <div class="p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold">Plugin Permissions</h3>
          <Button variant="ghost" onclick={() => showPermissionPanel = false}>×</Button>
        </div>
        
        <PluginPermissionPanel
          permissionManager={selectedPlugin.permissionManager}
          pluginName={selectedPlugin.manifest.name}
          pluginId={selectedPlugin.manifest.id}
        />
      </div>
    </div>
  </div>
{/if}