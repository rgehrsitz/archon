<!-- plugins/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';

  let newPluginUrl = '';
  let plugins: Array<{id: string, name: string, version: string, enabled: boolean, description: string}> = [];
  let loading = true;
  let error = '';

  onMount(async () => {
    // Mock data for now
    plugins = [
      { id: '1', name: 'Sample Plugin', version: '1.0.0', enabled: true, description: 'A sample plugin for demonstration' },
      { id: '2', name: 'Another Plugin', version: '2.1.0', enabled: false, description: 'Another example plugin' },
    ];
    loading = false;
  });

  async function handleInstallPlugin() {
    if (!newPluginUrl) return;
    try {
      // Mock installation logic
      const newPlugin = {
        id: Date.now().toString(),
        name: 'New Plugin',
        version: '1.0.0',
        enabled: false,
        description: 'Newly installed plugin'
      };
      plugins = [...plugins, newPlugin];
      newPluginUrl = '';
    } catch (error) {
      console.error('Failed to install plugin:', error);
    }
  }
  async function handleTogglePlugin(id: string, enabled: boolean) {
    try {
      // Mock toggle logic
      plugins = plugins.map(plugin => 
        plugin.id === id ? { ...plugin, enabled } : plugin
      );
    } catch (error) {
      console.error('Failed to toggle plugin:', error);
    }
  }

  async function handleUninstallPlugin(id: string) {
    if (!confirm('Are you sure you want to uninstall this plugin?')) return;
    try {
      // Mock uninstall logic
      plugins = plugins.filter(plugin => plugin.id !== id);
    } catch (error) {
      console.error('Failed to uninstall plugin:', error);
    }
  }
  async function handleUpdateConfig(id: string, config: Record<string, any>) {
    try {
      // Mock config update logic
      plugins = plugins.map(plugin => 
        plugin.id === id ? { ...plugin, config } : plugin
      );
    } catch (error) {
      console.error('Failed to update plugin config:', error);
    }
  }
</script>

<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-6">Plugins</h1>

  <!-- Install Plugin Form -->
  <div class="bg-white rounded-lg shadow p-6 mb-6">
    <h2 class="text-lg font-semibold mb-4">Install New Plugin</h2>
    <form onsubmit={handleInstallPlugin} class="space-y-4">
      <div>
        <label for="pluginUrl" class="block text-sm font-medium text-slate-600">Plugin URL</label>
        <input
          type="url"
          id="pluginUrl"
          bind:value={newPluginUrl}
          class="mt-1 block w-full rounded-md border-slate-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          placeholder="https://example.com/plugin"
          required
        />
      </div>
      <button
        type="submit"
        class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        Install Plugin
      </button>
    </form>
  </div>

  <!-- Plugins List -->
  <div class="bg-white rounded-lg shadow">
    <div class="px-6 py-4 border-b border-slate-200">
      <h2 class="text-lg font-semibold">Installed Plugins</h2>
    </div>
    {#if loading}
      <div class="p-4 text-slate-500">Loading plugins...</div>
    {:else if error}
      <div class="p-4 text-red-500">Error: {error}</div>
    {:else if plugins.length === 0}
      <div class="p-4 text-slate-500">No plugins installed</div>
    {:else}
      <div class="divide-y divide-slate-200">
        {#each plugins as plugin (plugin.id)}
          <div class="p-4 hover:bg-slate-50">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-lg font-medium text-slate-900">{plugin.name}</h3>
                <p class="text-sm text-slate-500">{plugin.description}</p>
                <div class="mt-1 text-xs text-slate-400">
                  Version {plugin.version}
                </div>
              </div>
              <div class="flex items-center space-x-4">
                <label for="pluginEnabled-{plugin.id}" class="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    id="pluginEnabled-{plugin.id}"
                    class="sr-only peer"
                    checked={plugin.enabled}
                    onchange={(e) => handleTogglePlugin(plugin.id, e.currentTarget.checked)}
                  />
                  <div class="w-11 h-6 bg-slate-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                  <span class="ml-3 text-sm font-medium text-slate-900">Enabled</span>
                </label>
                <button
                  onclick={() => handleUninstallPlugin(plugin.id)}
                  class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                >
                  Uninstall
                </button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>