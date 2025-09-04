<script lang="ts">
  import { onMount } from 'svelte';
  import { GetRootNode, ListChildren } from '../../../../../wailsjs/go/api/NodeService.js';
  import type { ArchonNode } from '$lib/types/visualization.js';

  // Props (same as other visualizations)
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 800;
  export let height: number = 600;

  // State
  let loading = true;
  let error: string | null = null;
  let rootNode: any = null;
  let children: any[] = [];

  async function loadData() {
    console.log('DirectAPITest: Starting to load data for project:', projectId);
    loading = true;
    error = null;

    try {
      console.log('DirectAPITest: Calling GetRootNode...');
      rootNode = await GetRootNode();
      console.log('DirectAPITest: Root node received:', rootNode);

      console.log('DirectAPITest: Calling ListChildren...');
      children = await ListChildren(rootNode.id);
      console.log('DirectAPITest: Children received:', children);

      loading = false;
    } catch (err) {
      console.error('DirectAPITest: Error occurred:', err);
      error = err instanceof Error ? err.message : 'Unknown error';
      loading = false;
    }
  }

  onMount(() => {
    console.log('DirectAPITest: Component mounted with projectId:', projectId);
    console.log('DirectAPITest: typeof projectId:', typeof projectId);
    console.log('DirectAPITest: projectId === undefined:', projectId === undefined);
    console.log('DirectAPITest: projectId === null:', projectId === null);
    console.log('DirectAPITest: projectId length:', projectId?.length);
    
    if (projectId) {
      loadData();
    } else {
      console.error('DirectAPITest: No projectId provided, cannot load data');
      error = 'No project ID provided';
      loading = false;
    }
  });

  // Also reactive check
  $: {
    console.log('DirectAPITest: Reactive - projectId changed to:', projectId);
    if (projectId && !loading && !error && !rootNode) {
      console.log('DirectAPITest: Reactive trigger - loading data');
      loadData();
    }
  }

  function handleNodeClick(node: any) {
    console.log('DirectAPITest: Node clicked:', node);
    // Could emit events here if needed
  }
</script>

<div class="w-full h-full bg-gray-50 p-4">
  <h3 class="text-lg font-bold mb-4">Direct API Test</h3>
  
  {#if loading}
    <div class="flex items-center justify-center h-64">
      <div class="text-center">
        <div class="w-8 h-8 border-3 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <div class="text-sm text-gray-600">Loading via direct API calls...</div>
      </div>
    </div>
  {:else if error}
    <div class="bg-red-50 border border-red-200 rounded p-4">
      <div class="text-red-800 font-medium">Error</div>
      <div class="text-red-600 text-sm mt-1">{error}</div>
    </div>
  {:else}
    <div class="space-y-4">
      <div class="bg-white rounded border p-4">
        <h4 class="font-semibold mb-2">Root Node</h4>
        <div class="text-sm">
          <div><strong>ID:</strong> {rootNode?.id || 'N/A'}</div>
          <div><strong>Name:</strong> {rootNode?.name || 'N/A'}</div>
        </div>
      </div>

      <div class="bg-white rounded border p-4">
        <h4 class="font-semibold mb-2">Children ({children.length})</h4>
        <div class="grid gap-2">
          {#each children as child, i}
            <button
              class="text-left p-2 bg-blue-50 hover:bg-blue-100 rounded border cursor-pointer transition-colors"
              onclick={() => handleNodeClick(child)}
            >
              <div class="font-medium">{child.name}</div>
              <div class="text-xs text-gray-600">ID: {child.id}</div>
              <div class="text-xs text-gray-600">
                Children: {child.children ? child.children.length : 0}
              </div>
            </button>
          {/each}
        </div>
      </div>

      <div class="bg-white rounded border p-4">
        <h4 class="font-semibold mb-2">Raw Data</h4>
        <pre class="text-xs bg-gray-100 p-2 rounded overflow-auto max-h-32">
{JSON.stringify({ rootNode, children }, null, 2)}
        </pre>
      </div>
    </div>
  {/if}
</div>