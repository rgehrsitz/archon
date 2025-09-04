<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import type { HierarchyNode, HierarchyRectangularNode } from 'd3-hierarchy';
  import type { ArchonNode } from '$lib/types/visualization.js';
  
  // Try to import LayerChart components
  let Chart: any, Svg: any, Treemap: any, Rect: any, Text: any;
  
  // Dynamic import to check if LayerChart is available
  let layerChartError: string | null = null;
  
  import('layerchart').then((lc) => {
    Chart = lc.Chart;
    Svg = lc.Svg;
    Treemap = lc.Treemap;
    Rect = lc.Rect;
    Text = lc.Text;
    console.log('LayerChart imported successfully');
  }).catch((err) => {
    layerChartError = `Failed to import LayerChart: ${err.message}`;
    console.error('LayerChart import failed:', err);
  });

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 800;
  export let height: number = 600;

  // Base component reference for accessing shared functionality
  let baseComponent: HierarchyVisualizationBase;

  function getNodeColor(node: any, depth: number): string {
    const colors = ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6', '#06b6d4'];
    const isSelected = node.id === selectedNodeId;
    
    if (isSelected) {
      return '#1e40af'; // Darker blue for selected
    }
    
    return colors[depth % colors.length];
  }

  function handleNodeClick(node: any) {
    console.log('TreemapVisualization: Node clicked:', node);
    if (baseComponent) {
      baseComponent.handleNodeSelect(node);
    }
  }
</script>

<HierarchyVisualizationBase 
  bind:this={baseComponent}
  {projectId} 
  {selectedNodeId} 
  {selectedNodePath} 
  {width} 
  {height}
  let:hierarchyData
>
  <div class="w-full h-full bg-slate-100">
    {#if layerChartError}
      <div class="flex items-center justify-center h-full">
        <div class="text-center p-6">
          <div class="text-4xl mb-4">⚠️</div>
          <div class="text-lg font-medium mb-2">LayerChart Error</div>
          <div class="text-sm text-muted-foreground mb-4">{layerChartError}</div>
          <div class="text-xs text-muted-foreground">
            This visualization requires LayerChart library components
          </div>
        </div>
      </div>
    {:else if !Chart || !Svg || !Treemap}
      <div class="flex items-center justify-center h-full">
        <div class="text-center p-6">
          <div class="text-4xl mb-4">⏳</div>
          <div class="text-lg font-medium mb-2">Loading LayerChart...</div>
          <div class="text-sm text-muted-foreground">
            Initializing visualization components
          </div>
        </div>
      </div>
    {:else if hierarchyData}
      <!-- Simple CSS-based treemap fallback -->
      <div class="p-4 h-full overflow-auto">
        <h3 class="text-lg font-bold mb-4">Treemap Visualization (Fallback)</h3>
        <div class="flex flex-wrap gap-2">
          {#each hierarchyData.descendants().slice(1) as node, i}
            {@const size = Math.max(60, Math.min(200, 60 + (node.value || 1) * 30))}
            <div 
              class="text-white p-3 rounded cursor-pointer hover:scale-105 transition-transform flex items-center justify-center text-center text-sm font-medium shadow-lg"
              style="width: {size}px; height: {size}px; background-color: {getNodeColor(node.data, node.depth)};"
              onclick={() => handleNodeClick(node)}
            >
              <div>
                <div class="font-bold">{node.data.name}</div>
                <div class="text-xs opacity-80">D:{node.depth} V:{node.value}</div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</HierarchyVisualizationBase>

<style>
  :global(.treemap-rect:hover) {
    stroke-width: 2px;
    filter: brightness(1.1);
  }
</style>