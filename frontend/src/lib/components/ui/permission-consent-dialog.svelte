<!--
Permission Consent Dialog Component

Shows permission request details and allows user to grant/deny permissions.
Integrates with the plugin permission system defined in ADR-013.
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './button.svelte';
  import * as Dialog from './dialog.js';
  import { Badge } from './badge.js';
  import { Separator } from './separator.js';
  import { 
    type Permission, 
    type PermissionRequest,
    getPermissionCategory,
    getPermissionDescription,
    PERMISSION_RISK_ICONS,
    PermissionCategory
  } from '../../plugins/runtime/permissions.js';

  interface Props {
    open: boolean;
    request: PermissionRequest;
    pluginName: string;
    pluginId: string;
  }

  let { open = $bindable(), request, pluginName, pluginId }: Props = $props();

  const dispatch = createEventDispatcher<{
    grant: { temporary: boolean; duration?: number };
    deny: void;
  }>();

  const category = $derived(getPermissionCategory(request.permission));
  const description = $derived(getPermissionDescription(request.permission));
  const riskIcon = $derived(PERMISSION_RISK_ICONS[category]);
  
  let temporaryGrant = $state(false);
  let duration = $state(3600000); // 1 hour default

  const riskColors = {
    [PermissionCategory.LOW_RISK]: 'bg-green-50 border-green-200 text-green-800',
    [PermissionCategory.MEDIUM_RISK]: 'bg-yellow-50 border-yellow-200 text-yellow-800',
    [PermissionCategory.HIGH_RISK]: 'bg-red-50 border-red-200 text-red-800'
  };

  function handleGrant() {
    dispatch('grant', { 
      temporary: temporaryGrant,
      duration: temporaryGrant ? duration : undefined
    });
    open = false;
  }

  function handleDeny() {
    dispatch('deny');
    open = false;
  }

  function formatDuration(ms: number): string {
    const minutes = Math.floor(ms / 60000);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours} hour${hours !== 1 ? 's' : ''}`;
    }
    return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-2">
        🔌 Permission Request
      </Dialog.Title>
      <Dialog.Description>
        The plugin "{pluginName}" is requesting additional permissions.
      </Dialog.Description>
    </Dialog.Header>

    <div class="space-y-4">
      <!-- Plugin Info -->
      <div class="rounded-md border border-muted bg-muted/50 p-3">
        <div class="text-sm font-medium">{pluginName}</div>
        <div class="text-xs text-muted-foreground">{pluginId}</div>
      </div>

      <!-- Permission Details -->
      <div class="space-y-3">
        <h4 class="text-sm font-semibold">Requested Permission</h4>
        
        <div class="rounded-md border p-3 {riskColors[category]}">
          <div class="flex items-start gap-3">
            <span class="text-lg" title="{category} risk">{riskIcon}</span>
            <div class="flex-1">
              <div class="font-medium text-sm">
                <code class="bg-black/10 px-1 py-0.5 rounded text-xs">{request.permission}</code>
              </div>
              <div class="text-sm mt-1">{description}</div>
            </div>
          </div>
        </div>

        <!-- Risk Level Badge -->
        <div class="flex items-center gap-2">
          <span class="text-xs text-muted-foreground">Risk Level:</span>
          <Badge variant={category === PermissionCategory.HIGH_RISK ? 'destructive' : 
                         category === PermissionCategory.MEDIUM_RISK ? 'secondary' : 'outline'}>
            {category.toUpperCase()} RISK
          </Badge>
        </div>

        <!-- Custom Reason -->
        {#if request.reason}
          <div>
            <h5 class="text-xs font-medium text-muted-foreground mb-1">Reason</h5>
            <div class="text-sm bg-muted/50 rounded-md p-2 border">
              {request.reason}
            </div>
          </div>
        {/if}
      </div>

      <Separator />

      <!-- Grant Options -->
      <div class="space-y-3">
        <h4 class="text-sm font-semibold">Grant Options</h4>
        
        <div class="space-y-2">
          <label class="flex items-start gap-2 text-sm cursor-pointer">
            <input 
              type="radio" 
              bind:group={temporaryGrant} 
              value={false} 
              class="mt-0.5"
            />
            <div>
              <div class="font-medium">Permanent Access</div>
              <div class="text-xs text-muted-foreground">
                Grant until manually revoked
              </div>
            </div>
          </label>
          
          <label class="flex items-start gap-2 text-sm cursor-pointer">
            <input 
              type="radio" 
              bind:group={temporaryGrant} 
              value={true} 
              class="mt-0.5"
            />
            <div class="flex-1">
              <div class="font-medium">Temporary Access</div>
              <div class="text-xs text-muted-foreground mb-2">
                Grant for a limited time only
              </div>
              
              {#if temporaryGrant}
                <div class="ml-6 space-y-1">
                  <label class="text-xs text-muted-foreground">Duration:</label>
                  <select bind:value={duration} class="text-xs border rounded px-2 py-1 bg-background">
                    <option value={900000}>15 minutes</option>
                    <option value={3600000}>1 hour</option>
                    <option value={14400000}>4 hours</option>
                    <option value={86400000}>24 hours</option>
                    <option value={604800000}>1 week</option>
                  </select>
                  <div class="text-xs text-muted-foreground">
                    ({formatDuration(duration)})
                  </div>
                </div>
              {/if}
            </div>
          </label>
        </div>
      </div>
    </div>

    <Dialog.Footer class="gap-2">
      <Button variant="outline" onclick={handleDeny}>
        Deny
      </Button>
      <Button 
        onclick={handleGrant}
        variant={category === PermissionCategory.HIGH_RISK ? 'destructive' : 'default'}
      >
        {#if category === PermissionCategory.HIGH_RISK}
          ⚠️ Grant Anyway
        {:else}
          Grant Permission
        {/if}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>