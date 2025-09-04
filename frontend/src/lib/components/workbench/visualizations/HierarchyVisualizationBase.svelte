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
  let hierarchyData: HierarchyNode<ArchonNode> | null = null;
  let loading = true;
  let error: string | null = null;

  // Reactive data loading - only trigger when projectId changes
  let lastProjectId: string = '';
  $: if (projectId && projectId !== lastProjectId) {
    lastProjectId = projectId;
    loadHierarchyData();
  }

  async function loadHierarchyData() {
    if (!projectId) {
      console.warn('No projectId provided to HierarchyVisualizationBase');
      loading = false;
      return;
    }
    
    console.log('Loading hierarchy data for project:', projectId);
    loading = true;
    error = null;
    hierarchyData = null;

    try {
      // Load lightweight hierarchy for better performance initially
      // Override this in specific implementations if needed
      console.log('Building lightweight hierarchy...');
      const data = await hierarchyDataAdapter.buildLightweightHierarchy();
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
  export function handleNodeSelect(hierarchyNode: HierarchyNode<ArchonNode>) {
    const node = hierarchyNode.data;
    const path = hierarchyDataAdapter.getNodePath(hierarchyData!, node.id);
    
    dispatch('nodeSelect', { node, path });
  }

  // Handle node hover
  export function handleNodeHover(hierarchyNode: HierarchyNode<ArchonNode> | null) {
    const node = hierarchyNode?.data || null;
    dispatch('nodeHover', { node });
  }

  // Lifecycle
  onMount(() => {
    console.log('HierarchyVisualizationBase mounted with projectId:', projectId);
    if (projectId) {
      loadHierarchyData();
    }
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