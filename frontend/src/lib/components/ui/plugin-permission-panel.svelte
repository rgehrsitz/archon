<!--
Plugin Permission Management Panel

Provides a comprehensive view of plugin permissions with management controls.
Used in plugin settings and management interfaces.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import Button from './button.svelte';
  import { Badge } from './badge.js';
  import { Separator } from './separator.js';
  import {
    type Permission,
    getPermissionCategory,
    getPermissionDescription,
    PERMISSION_RISK_ICONS,
    PermissionCategory
  } from '../../plugins/runtime/permissions.js';
  import type { UIPermissionManager } from '../../plugins/runtime/ui-permission-manager.js';

  interface Props {
    permissionManager: UIPermissionManager;
    pluginName: string;
    pluginId: string;
    readonly?: boolean;
  }

  let { permissionManager, pluginName, pluginId, readonly = false }: Props = $props();

  let summary = $state<{
    total: number;
    granted: number;
    pending: number;
    denied: number;
    details: Array<{
      permission: Permission;
      granted: boolean;
      temporary: boolean;
      expiresAt?: Date;
    }>;
  }>({ total: 0, granted: 0, pending: 0, denied: 0, details: [] });

  function refreshSummary() {
    summary = permissionManager.getPermissionSummary();
  }

  function handleRevokePermission(permission: Permission) {
    permissionManager.revokePermission(permission);
    refreshSummary();
  }

  function handleGrantPermission(permission: Permission, temporary = false) {
    permissionManager.grantPermission(permission, { 
      temporary,
      duration: temporary ? 3600000 : undefined // 1 hour
    });
    refreshSummary();
  }

  function handleResetAll() {
    permissionManager.resetAllPermissions();
    refreshSummary();
  }

  function formatTimeUntilExpiry(expiresAt: Date): string {
    const now = new Date();
    const msUntilExpiry = expiresAt.getTime() - now.getTime();
    
    if (msUntilExpiry <= 0) {
      return 'Expired';
    }
    
    const minutes = Math.floor(msUntilExpiry / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);
    
    if (days > 0) {
      return `${days}d ${hours % 24}h`;
    } else if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    } else {
      return `${minutes}m`;
    }
  }

  onMount(() => {
    refreshSummary();
    
    // Refresh every 30 seconds to update expiry times
    const interval = setInterval(refreshSummary, 30000);
    return () => clearInterval(interval);
  });
</script>

<div class="space-y-4">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h3 class="text-lg font-semibold">Plugin Permissions</h3>
      <p class="text-sm text-muted-foreground">{pluginName}</p>
    </div>
    <div class="flex items-center gap-2">
      <Badge variant="outline">
        {summary.granted}/{summary.total} granted
      </Badge>
      {#if !readonly && summary.granted > 0}
        <Button variant="outline" size="sm" onclick={handleResetAll}>
          Revoke All
        </Button>
      {/if}
    </div>
  </div>

  <Separator />

  <!-- Permission Summary -->
  <div class="grid grid-cols-3 gap-4 text-center">
    <div class="rounded-md border border-green-200 bg-green-50 p-3">
      <div class="text-2xl font-bold text-green-700">{summary.granted}</div>
      <div class="text-xs text-green-600">Granted</div>
    </div>
    <div class="rounded-md border border-yellow-200 bg-yellow-50 p-3">
      <div class="text-2xl font-bold text-yellow-700">{summary.pending}</div>
      <div class="text-xs text-yellow-600">Pending</div>
    </div>
    <div class="rounded-md border border-red-200 bg-red-50 p-3">
      <div class="text-2xl font-bold text-red-700">{summary.denied}</div>
      <div class="text-xs text-red-600">Denied</div>
    </div>
  </div>

  <!-- Detailed Permission List -->
  <div class="space-y-2">
    <h4 class="text-sm font-semibold">Declared Permissions</h4>
    
    {#if summary.details.length === 0}
      <div class="text-center text-muted-foreground text-sm py-8">
        No permissions declared by this plugin.
      </div>
    {:else}
      <div class="space-y-2">
        {#each summary.details as detail (detail.permission)}
          {@const category = getPermissionCategory(detail.permission)}
          {@const description = getPermissionDescription(detail.permission)}
          {@const riskIcon = PERMISSION_RISK_ICONS[category]}
          
          <div class="rounded-md border p-3 {detail.granted ? 'bg-green-50 border-green-200' : 'bg-gray-50'}">
            <div class="flex items-start justify-between gap-3">
              <div class="flex items-start gap-3 flex-1">
                <span class="text-lg" title="{category} risk">{riskIcon}</span>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 mb-1">
                    <code class="text-xs bg-black/10 px-1.5 py-0.5 rounded font-mono">
                      {detail.permission}
                    </code>
                    <Badge 
                      variant={category === PermissionCategory.HIGH_RISK ? 'destructive' : 
                               category === PermissionCategory.MEDIUM_RISK ? 'secondary' : 'outline'}
                      class="text-xs"
                    >
                      {category}
                    </Badge>
                    {#if detail.temporary}
                      <Badge variant="outline" class="text-xs">
                        🕐 TEMP
                      </Badge>
                    {/if}
                  </div>
                  <div class="text-sm text-muted-foreground">
                    {description}
                  </div>
                  {#if detail.granted && detail.temporary && detail.expiresAt}
                    <div class="text-xs text-orange-600 mt-1">
                      Expires in {formatTimeUntilExpiry(detail.expiresAt)}
                    </div>
                  {/if}
                </div>
              </div>
              
              {#if !readonly}
                <div class="flex items-center gap-1">
                  {#if detail.granted}
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onclick={() => handleRevokePermission(detail.permission)}
                    >
                      Revoke
                    </Button>
                  {:else}
                    <div class="flex gap-1">
                      <Button 
                        variant="outline" 
                        size="sm"
                        onclick={() => handleGrantPermission(detail.permission, false)}
                      >
                        Grant
                      </Button>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onclick={() => handleGrantPermission(detail.permission, true)}
                        title="Grant temporarily for 1 hour"
                      >
                        🕐 1h
                      </Button>
                    </div>
                  {/if}
                </div>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
  
  {#if !readonly}
    <div class="text-xs text-muted-foreground">
      💡 <strong>Tip:</strong> Permissions can be granted permanently or temporarily. 
      Temporary permissions automatically expire and are safer for testing.
    </div>
  {/if}
</div>