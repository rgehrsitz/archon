<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import { Chart, Svg, Pack, Circle, Text } from 'layerchart';
  import type { HierarchyNode, HierarchyCircularNode } from 'd3-hierarchy';
  import type { ArchonNode } from '$lib/types/visualization.js';
  import { scaleOrdinal } from 'd3-scale';
  import { schemeSet3 } from 'd3-scale-chromatic';

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 600;
  export let height: number = 600;

  // Base component reference
  let baseComponent: HierarchyVisualizationBase;

  // Ensure square aspect ratio for circle packing
  $: size = Math.min(width, height);

  // Color scale - using Set3 for better circle distinction
  const colorScale = scaleOrdinal(schemeSet3);

  function getNodeColor(node: HierarchyCircularNode<ArchonNode>): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return '#3b82f6';
    if (isInPath) return '#6366f1';
    
    // Color by depth with some parent-based variation
    const colorKey = `${node.depth}-${node.parent?.data.id || 'root'}`;
    return colorScale(colorKey);
  }

  function getNodeOpacity(node: HierarchyCircularNode<ArchonNode>): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (node.depth === 0) return 0.1; // Root circle is mostly transparent
    if (isSelected) return 0.9;
    if (isInPath) return 0.8;
    return 0.6;
  }

  function getStrokeWidth(node: HierarchyCircularNode<ArchonNode>): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return 3;
    if (isInPath) return 2;
    return 1;
  }

  function handleCircleClick(node: HierarchyCircularNode<ArchonNode>) {
    if (baseComponent) {
      baseComponent.handleNodeSelect(node);
    }
  }

  function handleCircleHover(node: HierarchyCircularNode<ArchonNode> | null) {
    if (baseComponent) {
      baseComponent.handleNodeHover(node);
    }
  }

  function shouldShowLabel(node: HierarchyCircularNode<ArchonNode>): boolean {
    // Show labels for circles with radius > 15 and depth > 0
    return node.r > 15 && node.depth > 0;
  }

  function getFontSize(node: HierarchyCircularNode<ArchonNode>): number {
    // Dynamic font size based on circle radius
    return Math.min(12, Math.max(8, node.r / 4));
  }

  function getTruncatedName(name: string, radius: number): string {
    const maxChars = Math.floor(radius / 3);
    return name.length > maxChars ? name.substring(0, maxChars) + '...' : name;
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
  <Chart width={size} height={size} padding={{ top: 10, right: 10, bottom: 10, left: 10 }}>
    <Svg>
      <Pack 
        hierarchy={hierarchyData}
        padding={3}
        size={[size - 20, size - 20]}
      >
        {#snippet children({ nodes })}
          {#each nodes as node}
            <g transform="translate({10}, {10})">
              <Circle
                cx={node.x}
                cy={node.y}
                r={node.r}
                fill={getNodeColor(node)}
                opacity={getNodeOpacity(node)}
                stroke={node.depth === 0 ? '#e2e8f0' : 'white'}
                strokeWidth={getStrokeWidth(node)}
                class="cursor-pointer hover:brightness-110 transition-all duration-200 {node.depth > 0 ? 'hover:scale-105' : ''}"
                onclick={() => handleCircleClick(node)}
                onmouseenter={() => handleCircleHover(node)}
                onmouseleave={() => handleCircleHover(null)}
              />
              
              <!-- Text labels for larger circles -->
              {#if shouldShowLabel(node)}
                <Text
                  x={node.x}
                  y={node.y}
                  textAnchor="middle"
                  dy="0.35em"
                  fontSize={getFontSize(node)}
                  fill={node.depth === 1 ? 'white' : '#1e293b'}
                  class="pointer-events-none select-none font-medium drop-shadow-sm"
                >
                  {getTruncatedName(node.data.name, node.r)}
                </Text>
                
                <!-- Smaller secondary label for node type or child count -->
                {#if node.r > 25 && (node.data.type || (node.children && node.children.length > 0))}
                  <Text
                    x={node.x}
                    y={node.y + getFontSize(node) + 2}
                    textAnchor="middle"
                    dy="0.35em"
                    fontSize={Math.max(6, getFontSize(node) - 2)}
                    fill={node.depth === 1 ? '#e2e8f0' : '#64748b'}
                    class="pointer-events-none select-none"
                  >
                    {node.data.type || `${node.children?.length || 0} items`}
                  </Text>
                {/if}
              {/if}
            </g>
          {/each}
        {/snippet}
      </Pack>
    </Svg>
  </Chart>
</HierarchyVisualizationBase>

<style>
  :global(.circle-pack-node:hover) {
    filter: brightness(1.1);
    transform: scale(1.05);
    transform-origin: center;
  }
</style>