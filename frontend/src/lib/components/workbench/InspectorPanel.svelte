<script lang="ts">
  import { Input } from '../ui/input/index.js';
  import { Label } from '../ui/label/index.js';
  import { Textarea } from '../ui/textarea/index.js';
  import { Button } from '../ui/button/index.js';
  import { Badge } from '../ui/badge/index.js';
  import { Separator } from '../ui/separator/index.js';
  import { updateNode } from '../../api/nodes.js';
  
  export let selectedNode: any = null;
  let className = '';
  export { className as class };
  
  // Form state for editing
  let editedName = '';
  let editedDescription = '';
  let editedProperties: Record<string, any> = {};
  let isDirty = false;
  
  // Watch for node changes and reset form
  $: if (selectedNode) {
    resetForm();
  }
  
  function resetForm() {
    if (!selectedNode) return;
    
    editedName = selectedNode.name || '';
    editedDescription = selectedNode.description || '';
    editedProperties = { ...(selectedNode.properties || {}) };
    isDirty = false;
  }
  
  function markDirty() {
    isDirty = true;
  }
  
  async function handleSave() {
    if (!selectedNode || !isDirty) return;
    
    try {
      const updatedNode = await updateNode({
        id: selectedNode.id,
        name: editedName,
        description: editedDescription,
        properties: editedProperties
      });
      
      // Update the selectedNode with the response
      selectedNode = updatedNode;
      isDirty = false;
      
      console.log('Node saved successfully:', updatedNode);
    } catch (error) {
      console.error('Failed to save node:', error);
      // TODO: Show error toast/notification to user
    }
  }
  
  function handleCancel() {
    resetForm();
  }
  
  function addProperty() {
    const key = prompt('Property name:');
    if (key && !editedProperties[key]) {
      editedProperties[key] = { value: '', typeHint: 'string' };
      editedProperties = { ...editedProperties };
      markDirty();
    }
  }
  
  function removeProperty(key: string) {
    delete editedProperties[key];
    editedProperties = { ...editedProperties };
    markDirty();
  }
  
  function getPropertyTypeIcon(typeHint: string): string {
    switch (typeHint) {
      case 'string': return '📝';
      case 'number': return '🔢';
      case 'boolean': return '☑️';
      case 'date': return '📅';
      default: return '❓';
    }
  }
</script>

<aside class="h-full overflow-hidden flex flex-col bg-muted/20 {className}">
  {#if selectedNode}
    <!-- Header -->
    <div class="p-4 border-b">
      <div class="flex items-start gap-3">
        <span class="text-2xl">
          {selectedNode.hasChildren ? '📁' : '📄'}
        </span>
        <div class="flex-1 min-w-0">
          <h2 class="font-semibold truncate" title={selectedNode.name}>
            {selectedNode.name}
          </h2>
          <p class="text-xs text-muted-foreground">
            {selectedNode.hasChildren ? 'Folder' : 'Item'}
          </p>
        </div>
      </div>
    </div>
    
    <!-- Form Content -->
    <div class="flex-1 overflow-y-auto">
      <div class="p-4 space-y-6">
        <!-- Basic Info -->
        <div class="space-y-4">
          <div>
            <Label for="node-name">Name</Label>
            <Input
              id="node-name"
              bind:value={editedName}
              placeholder="Node name"
              oninput={markDirty}
              class="mt-1"
            />
          </div>
          
          <div>
            <Label for="node-description">Description</Label>
            <Textarea
              id="node-description"
              bind:value={editedDescription}
              placeholder="Describe this node..."
              oninput={markDirty}
              class="mt-1 min-h-20"
            />
          </div>
        </div>
        
        <Separator />
        
        <!-- Properties -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <Label>Properties</Label>
            <Button
              size="sm"
              variant="outline"
              onclick={addProperty}
            >
              {#snippet children()}
                + Add
              {/snippet}
            </Button>
          </div>
          
          {#if Object.keys(editedProperties).length > 0}
            <div class="space-y-3">
              {#each Object.entries(editedProperties) as [key, prop]}
                <div class="p-3 rounded border bg-background">
                  <div class="flex items-center gap-2 mb-2">
                    <span class="text-sm">
                      {getPropertyTypeIcon(prop.typeHint || 'string')}
                    </span>
                    <code class="text-sm font-mono bg-muted px-1 rounded">
                      {key}
                    </code>
                    <Badge variant="secondary" class="text-xs">
                      {prop.typeHint || 'string'}
                    </Badge>
                    <button
                      class="ml-auto text-xs text-destructive hover:text-destructive/80"
                      onclick={() => removeProperty(key)}
                    >
                      ×
                    </button>
                  </div>
                  
                  <Input
                    bind:value={prop.value}
                    placeholder="Value..."
                    oninput={markDirty}
                    class="text-sm"
                  />
                </div>
              {/each}
            </div>
          {:else}
            <div class="text-center py-8 text-muted-foreground">
              <div class="text-2xl mb-2">🏷️</div>
              <div class="text-sm">No properties</div>
              <div class="text-xs">Click "Add" to create one</div>
            </div>
          {/if}
        </div>
        
        <Separator />
        
        <!-- Metadata -->
        <div class="space-y-2 text-xs text-muted-foreground">
          <div>
            <span class="font-medium">ID:</span> 
            <code class="bg-muted px-1 rounded">{selectedNode.id}</code>
          </div>
          {#if selectedNode.createdAt}
            <div>
              <span class="font-medium">Created:</span>
              {new Date(selectedNode.createdAt).toLocaleString()}
            </div>
          {/if}
          {#if selectedNode.updatedAt}
            <div>
              <span class="font-medium">Modified:</span>
              {new Date(selectedNode.updatedAt).toLocaleString()}
            </div>
          {/if}
        </div>
      </div>
    </div>
    
    <!-- Footer Actions -->
    {#if isDirty}
      <div class="p-4 border-t bg-muted/10">
        <div class="flex gap-2">
          <Button
            size="sm"
            onclick={handleSave}
            class="flex-1"
          >
            {#snippet children()}
              Save Changes
            {/snippet}
          </Button>
          <Button
            size="sm"
            variant="outline"
            onclick={handleCancel}
          >
            {#snippet children()}
              Cancel
            {/snippet}
          </Button>
        </div>
      </div>
    {/if}
    
  {:else}
    <!-- Empty State -->
    <div class="h-full flex flex-col items-center justify-center text-center p-8">
      <div class="text-4xl mb-4 opacity-40">📋</div>
      <div class="text-sm font-medium mb-2">No Selection</div>
      <div class="text-xs text-muted-foreground max-w-48">
        Select a node from the Miller columns to view and edit its properties
      </div>
    </div>
  {/if}
</aside>