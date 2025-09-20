<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { hierarchyDataAdapter } from '$lib/services/HierarchyDataAdapter.js';
  import type { HierarchyVisualizationProps, HierarchyVisualizationEvents, ArchonNode } from '$lib/types/visualization.js';
  import type { HierarchyNode } from 'd3-hierarchy';

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 800;
  export let height: number = 600;

  // Event dispatcher
  const dispatch = createEventDispatcher<HierarchyVisualizationEvents>();

  // Internal state
  let hierarchyData: any | null = null;
  let loading = true;
  let error: string | null = null;

  // Reactive data loading - trigger when projectId changes or on mount
  let lastProjectId: string = '';
  let hasLoadedOnce = false;
  
  $: if (projectId !== lastProjectId) {
    lastProjectId = projectId;
    if (hasLoadedOnce) {
      loadHierarchyData();
    }
  }

  async function loadHierarchyData() {
    console.log('Loading hierarchy data...', projectId ? `for project: ${projectId}` : 'without specific project');
    loading = true;
    error = null;
    hierarchyData = null;

    try {
      // Load full hierarchy for spatial visualizations
      console.log('Building full hierarchy...');
      const data = await hierarchyDataAdapter.buildFullHierarchy();
      console.log('Hierarchy data loaded successfully:', data);
      
      if (!data) {
        throw new Error('No hierarchy data returned');
      }
      
      hierarchyData = data;
      loading = false;
    } catch (err) {
      console.error('Failed to load hierarchy data:', err);
      console.error('Error type:', typeof err);
      console.error('Error instance:', err instanceof Error);
      error = err instanceof Error ? err.message : 'Failed to load data';
      hierarchyData = null;
      loading = false;
    }
    
    console.log('Loading complete. Loading:', loading, 'Error:', error, 'Has data:', !!hierarchyData);
  }

  // Handle node selection - converts d3-hierarchy node back to Archon format
  export function handleNodeSelect(hierarchyNode: any) {
    console.log('HierarchyVisualizationBase: handleNodeSelect called with:', hierarchyNode);
    const node = hierarchyNode.data;
    console.log('HierarchyVisualizationBase: Extracted node:', node);
    const path = hierarchyDataAdapter.getNodePath(hierarchyData!, node.id);
    console.log('HierarchyVisualizationBase: Node path:', path);
    
    console.log('HierarchyVisualizationBase: Dispatching nodeSelect event...');
    dispatch('nodeSelect', { node, path });
  }

  // Handle node hover
  export function handleNodeHover(hierarchyNode: any | null) {
    const node = hierarchyNode?.data || null;
    dispatch('nodeHover', { node });
  }

  // Lifecycle
  onMount(() => {
    console.log('HierarchyVisualizationBase mounted with projectId:', projectId);
    hasLoadedOnce = true;
    loadHierarchyData();
  });

  onDestroy(() => {
    hierarchyDataAdapter.clearCache();
  });

  // Expose data for child components
  export { hierarchyData, loading, error };
</script>

<div class="hierarchy-visualization-base w-full h-full relative bg-background">
  {#if loading}
    <!-- Loading State -->
    <div class="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm">
      <div class="flex flex-col items-center gap-3">
        <div class="w-8 h-8 border-3 border-primary border-t-transparent rounded-full animate-spin"></div>
        <div class="text-sm text-muted-foreground">Loading visualization...</div>
      </div>
    </div>
  {:else if error}
    <!-- Error State -->
    <div class="absolute inset-0 flex items-center justify-center bg-background">
      <div class="flex flex-col items-center gap-3 text-center max-w-md">
        <div class="text-4xl opacity-60">⚠️</div>
        <div class="text-lg font-medium text-foreground">Visualization Error</div>
        <div class="text-sm text-muted-foreground">{error}</div>
        <button 
          class="mt-2 px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90"
          onclick={loadHierarchyData}
        >
          Retry
        </button>
      </div>
    </div>
  {:else if !hierarchyData}
    <!-- No Data State -->
    <div class="absolute inset-0 flex items-center justify-center bg-background">
      <div class="flex flex-col items-center gap-3 text-center">
        <div class="text-4xl opacity-60">📊</div>
        <div class="text-lg font-medium text-foreground">No Data</div>
        <div class="text-sm text-muted-foreground">No hierarchy data available for visualization</div>
      </div>
    </div>
  {:else}
    <!-- Visualization Content - to be overridden by child components -->
    <slot {hierarchyData} {width} {height} {selectedNodeId} {selectedNodePath} />
  {/if}
</div>

<style>
  .hierarchy-visualization-base {
    container-type: size;
  }
</style>