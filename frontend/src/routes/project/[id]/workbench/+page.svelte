<script lang="ts">
  import { location } from 'svelte-spa-router';
  import { onMount } from 'svelte';
  import InspectorPanel from '$lib/components/workbench/InspectorPanel.svelte';
  import CommandBar from '$lib/components/workbench/CommandBar.svelte';
  import ProjectHeader from '$lib/components/project/ProjectHeader.svelte';
  import ProjectSettingsDialog from '$lib/components/project/ProjectSettingsDialog.svelte';
  import RenameProjectDialog from '$lib/components/project/RenameProjectDialog.svelte';
  import { visualizationRegistry } from '$lib/services/visualizationRegistry.js';
  import type { VisualizationId } from '$lib/types/visualization.js';
  
  // Get project ID from route params (svelte-spa-router style)
  let projectId: string = '';
  $: {
    const match = $location.match(/^\/project\/([^\/]+)\/workbench/);
    projectId = match ? match[1] : '';
  }
  
  // Visualization selection
  let viewMode: VisualizationId = 'miller';
  let selectedNode: any = null;
  let selectedNodePath: any[] = [];
  
  // Dynamic component reference
  let currentVisualization: any;
  
  // Derived selected node ID for passing to child components
  $: selectedNodeId = selectedNode?.id || null;
  
  // Dialog states
  let showProjectSettings = false;
  let showRenameDialog = false;
  
  // Handle node selection from both views
  function handleNodeSelect(event: CustomEvent) {
    selectedNode = event.detail.node;
    selectedNodePath = event.detail.path || [];
  }
  
  function handleViewModeChange(event: CustomEvent) {
    viewMode = event.detail.mode;
  }

  // Get current visualization component
  $: currentVisualizationData = visualizationRegistry.get(viewMode);
  $: currentVisualizationComponent = currentVisualizationData?.component;
  
  // Debug logging
  $: {
    console.log('Workbench: Current viewMode:', viewMode);
    console.log('Workbench: Registry has visualization:', !!currentVisualizationData);
    console.log('Workbench: Component available:', !!currentVisualizationComponent);
    console.log('Workbench: Available visualizations:', visualizationRegistry.getIds());
    console.log('Workbench: projectId:', projectId);
    console.log('Workbench: selectedNodeId:', selectedNodeId);
    console.log('Workbench: selectedNodePath length:', selectedNodePath?.length);
  }

  onMount(() => {
    // Ensure registry is initialized
    if (visualizationRegistry.getAll().length === 0) {
      console.warn('Visualization registry is empty. Re-importing...');
      import('$lib/services/visualizationRegistry.js');
    }
  });

  function handleBreadcrumbNavigate(event: CustomEvent) {
    const idx: number = event.detail.index;
    if (idx < 0) {
      selectedNodePath = [];
      selectedNode = null;
      return;
    }
    const newPath = selectedNodePath.slice(0, idx + 1);
    selectedNodePath = newPath;
    selectedNode = newPath[newPath.length - 1] || null;
  }
  
  // Project action handlers
  function handleOpenSettings() {
    showProjectSettings = true;
  }
  
  function handleRenameProject() {
    showRenameDialog = true;
  }
  
  function handleCloseProject() {
    // TODO: Implement project close functionality
    console.log('Close project requested');
  }
  
  function handleProjectSettingsSaved(event: CustomEvent) {
    console.log('Project settings saved:', event.detail);
    showProjectSettings = false;
  }
  
  function handleProjectRenamed(event: CustomEvent) {
    console.log('Project renamed:', event.detail);
    showRenameDialog = false;
  }
</script>

<svelte:head>
  <title>Workbench - Archon</title>
</svelte:head>

<div class="h-full flex flex-col bg-background">
  <!-- Project Header -->
  <ProjectHeader 
    {projectId}
    on:openSettings={handleOpenSettings}
    on:renameProject={handleRenameProject}
    on:closeProject={handleCloseProject}
  />
  
  <!-- Command Bar -->
  <CommandBar 
    {projectId}
    {viewMode}
    nodePath={selectedNodePath}
    on:viewModeChange={handleViewModeChange}
    on:breadcrumbNavigate={handleBreadcrumbNavigate}
  />
  
  <!-- Main Workbench Area -->
  <div class="flex-1 flex overflow-hidden">
    <!-- Dynamic Visualization Component -->
    <div class="flex-1 min-w-0 relative">
      {#if currentVisualizationComponent}
        <svelte:component 
          this={currentVisualizationComponent}
          {projectId}
          {selectedNodeId}
          {selectedNodePath}
          on:nodeSelect={handleNodeSelect}
          bind:this={currentVisualization}
        />
        
        <!-- Debug info -->
        <div class="absolute top-2 left-2 bg-black bg-opacity-75 text-white text-xs p-2 rounded z-50">
          <div>projectId: "{projectId}"</div>
          <div>viewMode: {viewMode}</div>
          <div>component: {currentVisualizationComponent ? 'loaded' : 'none'}</div>
        </div>
      {:else}
        <!-- Fallback if visualization not found -->
        <div class="flex items-center justify-center h-full text-center">
          <div>
            <div class="text-4xl mb-4">❌</div>
            <div class="text-lg font-medium mb-2">Visualization Not Found</div>
            <div class="text-sm text-muted-foreground">
              The '{viewMode}' visualization is not available.
            </div>
            <button 
              class="mt-4 px-3 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90"
              onclick={() => viewMode = 'miller'}
            >
              Switch to Miller Columns
            </button>
          </div>
        </div>
      {/if}
    </div>
    
    <!-- Inspector Panel -->
    <InspectorPanel 
      {selectedNode}
      class="w-80 border-l"
    />
  </div>
  
  <!-- Project Dialogs -->
  <ProjectSettingsDialog 
    bind:open={showProjectSettings}
    on:close={() => showProjectSettings = false}
    on:saved={handleProjectSettingsSaved}
  />
  
  <RenameProjectDialog 
    bind:open={showRenameDialog}
    on:close={() => showRenameDialog = false}
    on:renamed={handleProjectRenamed}
  />
</div>