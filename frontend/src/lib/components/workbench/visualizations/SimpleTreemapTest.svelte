<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import type { ArchonNode } from '$lib/types/visualization.js';

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 800;
  export let height: number = 600;

  // Base component reference
  let baseComponent: HierarchyVisualizationBase;

  function handleNodeClick(node: any) {
    console.log('SimpleTreemap: Node clicked:', node);
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
  <div class="w-full h-full bg-slate-200 p-4">
    <h3 class="text-lg font-bold mb-4">Simple Treemap Test (No LayerChart)</h3>
    
    {#if hierarchyData}
      <div class="grid grid-cols-3 gap-2">
        {#each hierarchyData.descendants().slice(1) as node, i}
          {@const size = Math.max(50, Math.min(200, 50 + i * 20))}
          <div 
            class="bg-blue-{(i % 9 + 1) * 100} text-white p-2 rounded cursor-pointer hover:scale-105 transition-transform flex items-center justify-center text-center text-sm font-medium"
            style="width: {size}px; height: {size}px;"
            onclick={() => handleNodeClick(node)}
          >
            {node.data.name}
            <br>
            <span class="text-xs opacity-80">D:{node.depth}</span>
          </div>
        {/each}
      </div>
    {:else}
      <div class="text-red-600">No hierarchy data available</div>
    {/if}
  </div>
</HierarchyVisualizationBase>