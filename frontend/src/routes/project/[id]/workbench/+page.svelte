<script lang="ts">
  import { location } from 'svelte-spa-router';
  import MillerColumns from '$lib/components/workbench/MillerColumns.svelte';
  import TreeView from '$lib/components/workbench/TreeView.svelte';
  import InspectorPanel from '$lib/components/workbench/InspectorPanel.svelte';
  import CommandBar from '$lib/components/workbench/CommandBar.svelte';
  
  // Get project ID from route params (svelte-spa-router style)
  let projectId: string = '';
  $: {
    const match = $location.match(/^\/project\/([^\/]+)\/workbench/);
    projectId = match ? match[1] : '';
  }
  
  // Layout toggle state
  let viewMode: 'miller' | 'tree' = 'miller';
  let selectedNode: any = null;
  let selectedNodePath: any[] = [];
  
  // Derived selected node ID for passing to child components
  $: selectedNodeId = selectedNode?.id || null;
  
  // Handle node selection from both views
  function handleNodeSelect(event: CustomEvent) {
    selectedNode = event.detail.node;
    selectedNodePath = event.detail.path || [];
  }
  
  function handleViewModeChange(event: CustomEvent) {
    viewMode = event.detail.mode;
  }
</script>

<svelte:head>
  <title>Workbench - Archon</title>
</svelte:head>

<div class="h-full flex flex-col bg-background">
  <!-- Command Bar -->
  <CommandBar 
    {projectId}
    {viewMode}
    on:viewModeChange={handleViewModeChange}
  />
  
  <!-- Main Workbench Area -->
  <div class="flex-1 flex overflow-hidden">
    <!-- Miller Columns / Tree View -->
    <div class="flex-1 min-w-0">
      {#if viewMode === 'miller'}
        <MillerColumns 
          {projectId}
          {selectedNodeId}
          {selectedNodePath}
          on:nodeSelect={handleNodeSelect}
        />
      {:else}
        <TreeView 
          {projectId}
          {selectedNodeId}
          {selectedNodePath}
          on:nodeSelect={handleNodeSelect}
        />
      {/if}
    </div>
    
    <!-- Inspector Panel -->
    <InspectorPanel 
      {selectedNode}
      class="w-80 border-l"
    />
  </div>
</div>