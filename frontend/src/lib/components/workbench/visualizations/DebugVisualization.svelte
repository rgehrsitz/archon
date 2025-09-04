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
    console.log('Debug: Node clicked:', node);
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
  <div class="p-4 bg-slate-100 h-full overflow-auto">
    <h3 class="text-lg font-bold mb-4">Debug Visualization</h3>
    
    {#if hierarchyData}
      <div class="mb-4">
        <strong>Root Node:</strong> {hierarchyData.data.name}
      </div>
      
      <div class="mb-4">
        <strong>Total Descendants:</strong> {hierarchyData.descendants().length}
      </div>
      
      <div class="mb-4">
        <strong>Max Depth:</strong> {hierarchyData.height}
      </div>
      
      <div class="space-y-2">
        <h4 class="font-semibold">Hierarchy Structure:</h4>
        <div class="text-sm font-mono bg-white p-3 rounded border max-h-96 overflow-auto">
          {#each hierarchyData.descendants() as node}
            <div 
              class="cursor-pointer hover:bg-blue-50 px-1 rounded {selectedNodeId === node.data.id ? 'bg-blue-100 font-bold' : ''}"
              style="margin-left: {node.depth * 20}px"
              onclick={() => handleNodeClick(node)}
            >
              {''.repeat(node.depth)}└ {node.data.name} 
              <span class="text-gray-500">
                (depth: {node.depth}, 
                children: {node.children?.length || 0},
                value: {node.value})
              </span>
            </div>
          {/each}
        </div>
      </div>
      
      <div class="mt-4">
        <strong>Raw Data:</strong>
        <pre class="text-xs bg-white p-2 rounded border mt-2 max-h-48 overflow-auto">
{JSON.stringify(hierarchyData.data, null, 2)}
        </pre>
      </div>
    {:else}
      <div class="text-red-600">No hierarchy data available</div>
    {/if}
  </div>
</HierarchyVisualizationBase>