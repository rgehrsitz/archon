<script lang="ts">
  import { onMount } from 'svelte';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs/index.js';
  import { Switch } from '$lib/components/ui/switch/index.js';
  import { Separator } from '$lib/components/ui/separator/index.js';
  import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
  import { 
    getPlugins, 
    getEnabledPlugins, 
    enablePlugin, 
    disablePlugin, 
    getPluginPermissions,
    grantPermission,
    revokePermission,
    type PluginInstallation,
    type PluginPermissionGrant 
  } from '$lib/api/plugins.js';
  
  let plugins: PluginInstallation[] = [];
  let enabledPlugins: PluginInstallation[] = [];
  let loading = true;
  let error: string | null = null;
  
  onMount(async () => {
    await loadPlugins();
  });
  
  async function loadPlugins() {
    try {
      loading = true;
      error = null;
      plugins = await getPlugins();
      enabledPlugins = await getEnabledPlugins();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load plugins';
      console.error('Error loading plugins:', err);
    } finally {
      loading = false;
    }
  }
  
  async function togglePlugin(plugin: PluginInstallation) {
    try {
      if (plugin.enabled) {
        await disablePlugin(plugin.manifest.id);
      } else {
        await enablePlugin(plugin.manifest.id);
      }
      await loadPlugins(); // Reload to get updated state
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to toggle plugin';
      console.error('Error toggling plugin:', err);
    }
  }
  
  function getPluginTypeIcon(type: string): string {
    switch (type) {
      case 'Importer': return '📥';
      case 'Exporter': return '📤';
      case 'Transformer': return '🔄';
      case 'Validator': return '✅';
      case 'Panel': return '📋';
      case 'Provider': return '🔌';
      case 'AttachmentProcessor': return '📎';
      case 'ConflictResolver': return '🤝';
      case 'SearchIndexer': return '🔍';
      case 'UIContrib': return '🎨';
      default: return '🔧';
    }
  }
  
  function formatDate(dateString: string): string {
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return dateString;
    }
  }
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="max-w-6xl mx-auto p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Plugin Manager</h1>
      <p class="text-muted-foreground">Manage your Archon plugins and permissions</p>
    </div>
    
    {#if error}
      <Alert>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    {/if}
    
    <Tabs value="installed" class="w-full">
      <TabsList>
        <TabsTrigger value="installed">Installed ({plugins.length})</TabsTrigger>
        <TabsTrigger value="permissions">Permissions</TabsTrigger>
        <TabsTrigger value="discover" disabled>Discover</TabsTrigger>
      </TabsList>
      
      <TabsContent value="installed" class="space-y-4">
        {#if loading}
          <div class="flex items-center justify-center py-8">
            <div class="text-center">
              <div class="text-2xl mb-2">⏳</div>
              <div class="text-sm text-muted-foreground">Loading plugins...</div>
            </div>
          </div>
        {:else if plugins.length === 0}
          <div class="flex items-center justify-center py-8">
            <div class="text-center">
              <div class="text-4xl mb-4">🔌</div>
              <div class="text-lg font-medium mb-2">No Plugins Installed</div>
              <div class="text-sm text-muted-foreground mb-4">
                Install plugins to extend Archon's functionality
              </div>
              <Button disabled>
                Install Plugin
              </Button>
            </div>
          </div>
        {:else}
          <div class="grid gap-4">
            {#each plugins as plugin}
              <Card>
                <CardHeader class="pb-3">
                  <div class="flex items-start justify-between">
                    <div class="flex items-start gap-3">
                      <span class="text-2xl">{getPluginTypeIcon(plugin.manifest.type)}</span>
                      <div class="flex-1 min-w-0">
                        <CardTitle class="text-lg">{plugin.manifest.name}</CardTitle>
                        <CardDescription class="mt-1">
                          {plugin.manifest.description || 'No description available'}
                        </CardDescription>
                        <div class="flex items-center gap-2 mt-2">
                          <Badge variant="secondary">{plugin.manifest.type}</Badge>
                          <Badge variant="outline">v{plugin.manifest.version}</Badge>
                          {#if plugin.manifest.author}
                            <Badge variant="outline">by {plugin.manifest.author}</Badge>
                          {/if}
                        </div>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <Switch 
                        checked={plugin.enabled}
                        onCheckedChange={() => togglePlugin(plugin)}
                      />
                      <span class="text-sm text-muted-foreground">
                        {plugin.enabled ? 'Enabled' : 'Disabled'}
                      </span>
                    </div>
                  </div>
                </CardHeader>
                <CardContent class="pt-0">
                  <div class="space-y-2 text-sm text-muted-foreground">
                    <div class="flex items-center gap-2">
                      <span class="font-medium">ID:</span>
                      <code class="bg-muted px-1 rounded text-xs">{plugin.manifest.id}</code>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="font-medium">Installed:</span>
                      <span>{formatDate(plugin.installedAt)}</span>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="font-medium">Source:</span>
                      <Badge variant="outline" class="text-xs">{plugin.source}</Badge>
                    </div>
                    {#if plugin.manifest.permissions.length > 0}
                      <div class="flex items-center gap-2">
                        <span class="font-medium">Permissions:</span>
                        <div class="flex gap-1 flex-wrap">
                          {#each plugin.manifest.permissions as permission}
                            <Badge variant="outline" class="text-xs">{permission}</Badge>
                          {/each}
                        </div>
                      </div>
                    {/if}
                  </div>
                </CardContent>
              </Card>
            {/each}
          </div>
        {/if}
      </TabsContent>
      
      <TabsContent value="permissions" class="space-y-4">
        {#if loading}
          <div class="flex items-center justify-center py-8">
            <div class="text-center">
              <div class="text-2xl mb-2">⏳</div>
              <div class="text-sm text-muted-foreground">Loading permissions...</div>
            </div>
          </div>
        {:else if plugins.length === 0}
          <div class="flex items-center justify-center py-8">
            <div class="text-center">
              <div class="text-4xl mb-4">🔐</div>
              <div class="text-lg font-medium mb-2">No Plugins to Manage</div>
              <div class="text-sm text-muted-foreground">
                Install plugins to manage their permissions
              </div>
            </div>
          </div>
        {:else}
          <div class="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Plugin Permissions</CardTitle>
                <CardDescription>Manage permissions for installed plugins</CardDescription>
              </CardHeader>
              <CardContent>
                <div class="space-y-4">
                  {#each plugins as plugin}
                    <div class="border rounded-lg p-4">
                      <div class="flex items-start justify-between mb-3">
                        <div class="flex items-center gap-2">
                          <span class="text-lg">{getPluginTypeIcon(plugin.manifest.type)}</span>
                          <div>
                            <div class="font-medium">{plugin.manifest.name}</div>
                            <div class="text-sm text-muted-foreground">{plugin.manifest.id}</div>
                          </div>
                        </div>
                        <Badge variant={plugin.enabled ? 'default' : 'secondary'}>
                          {plugin.enabled ? 'Enabled' : 'Disabled'}
                        </Badge>
                      </div>
                      
                      {#if plugin.manifest.permissions.length > 0}
                        <div class="space-y-2">
                          <div class="text-sm font-medium">Required Permissions:</div>
                          <div class="flex gap-2 flex-wrap">
                            {#each plugin.manifest.permissions as permission}
                              <Badge variant="outline" class="text-xs">{permission}</Badge>
                            {/each}
                          </div>
                        </div>
                      {:else}
                        <div class="text-sm text-muted-foreground">No permissions required</div>
                      {/if}
                    </div>
                  {/each}
                </div>
              </CardContent>
            </Card>
          </div>
        {/if}
      </TabsContent>
      
      <TabsContent value="discover" class="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Discover Plugins</CardTitle>
            <CardDescription>Browse and install plugins from the registry</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="text-center py-8">
              <div class="text-4xl mb-4">🔍</div>
              <div class="text-lg font-medium mb-2">Plugin Discovery</div>
              <div class="text-sm text-muted-foreground mb-4">
                Browse and install plugins from the Archon plugin registry
              </div>
              <Button disabled>
                Browse Registry
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</div>