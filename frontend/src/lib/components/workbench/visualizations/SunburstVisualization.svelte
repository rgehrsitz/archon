<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import { Chart, Svg, Partition, Arc, Text } from 'layerchart';
  import type { HierarchyNode, HierarchyRectangularNode } from 'd3-hierarchy';
  import type { ArchonNode } from '$lib/types/visualization.js';
  import { arc as d3Arc } from 'd3-shape';
  import { scaleOrdinal } from 'd3-scale';
  import { schemeCategory10 } from 'd3-scale-chromatic';

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 600;
  export let height: number = 600;

  // Base component reference
  let baseComponent: HierarchyVisualizationBase;

  // Ensure square aspect ratio for sunburst
  $: size = Math.min(width, height);
  $: radius = size / 2 - 10;

  // Color scale
  const colorScale = scaleOrdinal(schemeCategory10);

  // Arc generator for path calculations
  const arcGenerator = d3Arc<HierarchyRectangularNode<ArchonNode>>()
    .startAngle(d => d.x0)
    .endAngle(d => d.x1)
    .innerRadius(d => Math.sqrt(d.y0) * radius / Math.sqrt(d.height))
    .outerRadius(d => Math.sqrt(d.y1) * radius / Math.sqrt(d.height));

  function getNodeColor(node: HierarchyRectangularNode<ArchonNode>): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return '#3b82f6';
    if (isInPath) return '#6366f1';
    
    // Color by parent or depth for consistent grouping
    const colorKey = node.parent ? node.parent.data.id : node.depth.toString();
    return colorScale(colorKey);
  }

  function getNodeOpacity(node: HierarchyRectangularNode<ArchonNode>): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return 1;
    if (isInPath) return 0.9;
    return 0.7;
  }

  function handleArcClick(node: HierarchyRectangularNode<ArchonNode>) {
    if (baseComponent) {
      baseComponent.handleNodeSelect(node);
    }
  }

  function handleArcHover(node: HierarchyRectangularNode<ArchonNode> | null) {
    if (baseComponent) {
      baseComponent.handleNodeHover(node);
    }
  }

  function shouldShowLabel(node: HierarchyRectangularNode<ArchonNode>): boolean {
    const angle = node.x1 - node.x0;
    const radiusRange = Math.sqrt(node.y1) - Math.sqrt(node.y0);
    return angle > 0.1 && radiusRange * radius > 20; // Only show label if arc is large enough
  }

  function getLabelPosition(node: HierarchyRectangularNode<ArchonNode>): { x: number; y: number; rotation: number } {
    const angle = (node.x0 + node.x1) / 2;
    const labelRadius = Math.sqrt((node.y0 + node.y1) / 2) * radius / Math.sqrt(node.height);
    
    return {
      x: Math.cos(angle - Math.PI / 2) * labelRadius,
      y: Math.sin(angle - Math.PI / 2) * labelRadius,
      rotation: angle > Math.PI ? (angle * 180 / Math.PI) - 90 : (angle * 180 / Math.PI) + 90
    };
  }
</script>

<HierarchyVisualizationBase 
  bind:this={baseComponent}
  {projectId} 
  {selectedNodeId} 
  {selectedNodePath} 
  width={size} 
  height={size}
  let:hierarchyData
>
  <!-- Simple fallback sunburst without LayerChart for now -->
  <div class="w-full h-full p-4 bg-slate-50">
    <h3 class="text-lg font-bold mb-4">Sunburst Visualization</h3>
    <div class="relative w-full h-96 flex items-center justify-center bg-slate-50 rounded-lg">
      <!-- Center circle for root -->
      <div 
        class="absolute w-16 h-16 bg-indigo-600 rounded-full flex items-center justify-center text-white font-bold text-sm cursor-pointer hover:bg-indigo-500 transition-colors z-10"
        onclick={() => handleArcClick({ data: hierarchyData.data, depth: 0 })}
      >
        ROOT
      </div>
      
      <!-- Concentric rings for different depths -->
      <div class="relative w-96 h-96">
        {#each hierarchyData.descendants().slice(1) as node, i}
          {@const angle = (i / hierarchyData.descendants().slice(1).length) * 360}
          {@const radius = 60 + (node.depth * 40)}
          {@const x = Math.cos((angle - 90) * Math.PI / 180) * radius}
          {@const y = Math.sin((angle - 90) * Math.PI / 180) * radius}
          {@const colorIndex = node.depth % 6}
          {@const colors = ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6', '#06b6d4']}
          {@const isSelected = node.data.id === selectedNodeId}
          {@const isInPath = selectedNodePath.some(p => p.id === node.data.id)}
          
          <div 
            class="absolute w-8 h-8 rounded-full cursor-pointer hover:scale-110 transition-all duration-200 flex items-center justify-center text-xs font-medium shadow-md border-2 border-white"
            style="
              left: calc(50% + {x}px - 16px);
              top: calc(50% + {y}px - 16px);
              background-color: {isSelected ? '#1e40af' : isInPath ? '#3730a3' : colors[colorIndex]};
              opacity: {isSelected ? '1' : isInPath ? '0.9' : '0.8'};
            "
            onclick={() => handleArcClick(node)}
            title={node.data.name}
          >
            <div class="text-white text-center leading-tight">
              {node.data.name.substring(0, 2)}
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
</HierarchyVisualizationBase>

<style>
  :global(.sunburst-arc:hover) {
    filter: brightness(1.1);
  }
</style>