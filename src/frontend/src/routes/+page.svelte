<!-- +page.svelte -->
<script lang="ts">
  import Layout from '../lib/components/Layout.svelte';
  import ComponentTree from '../lib/components/ComponentTree.svelte';
  
  let components = [
    {
      id: 'root',
      name: 'Root Component',
      type: 'system',
      parentId: null
    },
    {
      id: 'child1',
      name: 'Child Component 1',
      type: 'device',
      parentId: 'root'
    },
    {
      id: 'child2',
      name: 'Child Component 2',
      type: 'device',
      parentId: 'root'
    }
  ];
  
  let selectedId: string | null = null;
</script>

<Layout>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- Component Tree -->
    <div class="lg:col-span-1">
      <ComponentTree {components} bind:selectedId />
    </div>
    
    <!-- Component Details -->
    <div class="lg:col-span-2">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div class="p-4 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-800 dark:text-white">
            {#if selectedId}
              Component Details
            {:else}
              Select a Component
            {/if}
          </h2>
        </div>
        
        <div class="p-4">
          {#if selectedId}
            {#each components as component}
              {#if component.id === selectedId}
                <dl class="grid grid-cols-1 gap-4">
                  <div>
                    <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Name</dt>
                    <dd class="mt-1 text-sm text-gray-900 dark:text-white">{component.name}</dd>
                  </div>
                  <div>
                    <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Type</dt>
                    <dd class="mt-1 text-sm text-gray-900 dark:text-white">{component.type}</dd>
                  </div>
                  <div>
                    <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">ID</dt>
                    <dd class="mt-1 text-sm text-gray-900 dark:text-white">{component.id}</dd>
                  </div>
                </dl>
              {/if}
            {/each}
          {:else}
            <p class="text-gray-500 dark:text-gray-400">Select a component to view its details.</p>
          {/if}
        </div>
      </div>
    </div>
  </div>
</Layout> 