<!-- ComponentTree.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
    let expandedNodes = new Set<string>();
  let components: Array<{id: string, name: string, type: string, parentId: string | null}> = [];
  let loading = true;
  let error = '';
  
  onMount(async () => {
    // Mock data for now
    components = [
      { id: '1', name: 'App', type: 'Component', parentId: null },
      { id: '2', name: 'Header', type: 'Component', parentId: '1' },
      { id: '3', name: 'Content', type: 'Component', parentId: '1' },
    ];
    loading = false;
  });
  
  function toggleNode(id: string) {
    if (expandedNodes.has(id)) {
      expandedNodes.delete(id);
    } else {
      expandedNodes.add(id);
    }
    expandedNodes = expandedNodes; // Trigger reactivity
  }
  
  function getChildComponents(parentId: string | null) {
    return components.filter(c => c.parentId === parentId);
  }
  
  function handleDragStart(event: DragEvent, componentId: string) {
    if (event.dataTransfer) {
      event.dataTransfer.setData('text/plain', componentId);
    }
  }
  
  function handleDrop(event: DragEvent, targetId: string) {
    event.preventDefault();
    const sourceId = event.dataTransfer?.getData('text/plain');
    if (sourceId && sourceId !== targetId) {
      // Handle drop logic here
      console.log('Dropped', sourceId, 'onto', targetId);
    }
  }
  
  function handleDragOver(event: DragEvent) {
    event.preventDefault();
  }
</script>

<div class="bg-white rounded-lg shadow">
  <div class="p-4 border-b border-slate-200">
    <h2 class="text-lg font-semibold text-slate-800">Component Hierarchy</h2>
  </div>
  
  <div class="p-4">
    {#if loading}
      <div class="text-slate-500">Loading components...</div>
    {:else if error}
      <div class="text-red-500">Error: {error}</div>
    {:else}
      <div class="space-y-2">
        {#each getChildComponents(null) as component (component.id)}
          <div class="component-tree">
            <div class="flex items-center py-1">
              <button
                class="p-1 rounded hover:bg-slate-100"
                on:click={() => toggleNode(component.id)}
                aria-label="Toggle component {component.name}"
              >
                <svg
                  class="w-4 h-4 transform transition-transform"
                  class:rotate-90={expandedNodes.has(component.id)}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </button>
              
              <button
                class="flex-1 px-2 py-1 text-left rounded hover:bg-slate-100"
                draggable="true"
                on:dragstart={(e) => handleDragStart(e, component.id)}
                on:drop={(e) => handleDrop(e, component.id)}
                on:dragover={handleDragOver}
                aria-label="Select component {component.name}"
              >
                <span class="font-medium">{component.name}</span>
                <span class="ml-2 text-sm text-slate-500">({component.type})</span>
              </button>
            </div>
            
            {#if expandedNodes.has(component.id)}
              <div class="ml-6">
                {#each getChildComponents(component.id) as child (child.id)}
                  <div class="flex items-center py-1">
                    <button
                      class="p-1 rounded hover:bg-slate-100"
                      on:click={() => toggleNode(child.id)}
                      aria-label="Toggle component {child.name}"
                    >
                      <svg
                        class="w-4 h-4 transform transition-transform"
                        class:rotate-90={expandedNodes.has(child.id)}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </button>
                    
                    <button
                      class="flex-1 px-2 py-1 text-left rounded hover:bg-slate-100"
                      draggable="true"
                      on:dragstart={(e) => handleDragStart(e, child.id)}
                      on:drop={(e) => handleDrop(e, child.id)}
                      on:dragover={handleDragOver}
                      aria-label="Select component {child.name}"
                    >
                      <span class="font-medium">{child.name}</span>
                      <span class="ml-2 text-sm text-slate-500">({child.type})</span>
                    </button>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .component-tree {
    color: #64748b;
  }
</style> 