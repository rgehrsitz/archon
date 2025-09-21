<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import { Chart, Svg, Tree, Circle, Line, Text } from 'layerchart';
  // Temporary type aliases for d3-hierarchy (workaround for import issues)
  type HierarchyPointNode<T> = any;
  type HierarchyPointLink<T> = any;
  import type { ArchonNode } from '$lib/types/visualization.js';
  import { scaleOrdinal } from 'd3-scale';
  import { schemeCategory10 } from 'd3-scale-chromatic';

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 800;
  export let height: number = 600;

  // Base component reference
  let baseComponent: HierarchyVisualizationBase;

  // Tree orientation
  export let orientation: 'horizontal' | 'vertical' = 'horizontal';

  // Color scale
  const colorScale = scaleOrdinal(schemeCategory10);

  function getNodeColor(node: HierarchyPointNode<ArchonNode>): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return '#3b82f6';
    if (isInPath) return '#6366f1';
    
    // Color by depth
    return colorScale(node.depth.toString());
  }

  function getNodeSize(node: HierarchyPointNode<ArchonNode>): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return 8;
    if (isInPath) return 6;
    
    // Size based on whether it has children and depth
    const baseSize = node.children ? 6 : 4;
    return Math.max(3, baseSize - node.depth * 0.5);
  }

  function getLinkColor(link: HierarchyPointLink<ArchonNode>): string {
    const isInSelectedPath = selectedNodePath.some(p => p.id === link.source.data.id || p.id === link.target.data.id);
    return isInSelectedPath ? '#6366f1' : '#94a3b8';
  }

  function getLinkWidth(link: HierarchyPointLink<ArchonNode>): number {
    const isInSelectedPath = selectedNodePath.some(p => p.id === link.source.data.id || p.id === link.target.data.id);
    return isInSelectedPath ? 2 : 1;
  }

  function handleNodeClick(node: HierarchyPointNode<ArchonNode>) {
    if (baseComponent) {
      baseComponent.handleNodeSelect(node);
    }
  }

  function handleNodeHover(node: HierarchyPointNode<ArchonNode> | null) {
    if (baseComponent) {
      baseComponent.handleNodeHover(node);
    }
  }

  function shouldShowLabel(node: HierarchyPointNode<ArchonNode>): boolean {
    // Show labels for root, selected node, nodes in path, and nodes with fewer than 3 siblings
    return node.depth === 0 || 
           node.data.id === selectedNodeId ||
           selectedNodePath.some(p => p.id === node.data.id) ||
           (node.parent?.children?.length || 0) < 4;
  }

  function getLabelPosition(node: HierarchyPointNode<ArchonNode>): { x: number; y: number } {
    if (orientation === 'horizontal') {
      return {
        x: node.y + 12,
        y: node.x
      };
    } else {
      return {
        x: node.x,
        y: node.y + 15
      };
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
  <div style="width: {width}px; height: {height}px;">
    <Chart padding={{ top: 20, right: 20, bottom: 20, left: 20 }}>
      <Svg>
        <Tree
          {...{ data: hierarchyData } as any}
          {orientation}
          nodeSize={orientation === 'horizontal' ? [30, 200] : [120, 30]}
        >
        {#snippet children({ nodes, links }: { nodes: any[], links: any[] })}
          <!-- Draw links first (behind nodes) -->
          {#each links as link}
            <Line
              x1={orientation === 'horizontal' ? link.source.y : link.source.x}
              y1={orientation === 'horizontal' ? link.source.x : link.source.y}
              x2={orientation === 'horizontal' ? link.target.y : link.target.x}
              y2={orientation === 'horizontal' ? link.target.x : link.target.y}
              stroke={getLinkColor(link)}
              strokeWidth={getLinkWidth(link)}
              opacity={0.6}
              class="transition-all duration-200"
            />
          {/each}

          <!-- Draw nodes -->
          {#each nodes as node}
            <g
              role="button"
              tabindex="0"
              onclick={() => handleNodeClick(node)}
              onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleNodeClick(node)}
              onmouseenter={() => handleNodeHover(node)}
              onmouseleave={() => handleNodeHover(null)}
            >
              <Circle
                cx={orientation === 'horizontal' ? node.y : node.x}
                cy={orientation === 'horizontal' ? node.x : node.y}
                r={getNodeSize(node)}
                fill={getNodeColor(node)}
                stroke="white"
                strokeWidth={2}
                opacity={0.9}
                class="cursor-pointer hover:stroke-4 hover:scale-110 transition-all duration-200"
              />
              
              <!-- Node labels -->
              {#if shouldShowLabel(node)}
                {@const labelPos = getLabelPosition(node)}
                <Text
                  x={labelPos.x}
                  y={labelPos.y}
                  textAnchor={orientation === 'horizontal' ? 'start' : 'middle'}
                  dy="0.35em"
                  fill="#334155"
                  class="pointer-events-none select-none font-medium"
                  style="font-size: 10px;"
                >
                  {node.data.name.length > 15 ? node.data.name.substring(0, 15) + '...' : node.data.name}
                </Text>
                
                <!-- Type label for selected or root nodes -->
                {#if (node.data.id === selectedNodeId || node.depth === 0) && node.data.type}
                  <Text
                    x={labelPos.x}
                    y={labelPos.y + 12}
                    textAnchor={orientation === 'horizontal' ? 'start' : 'middle'}
                    dy="0.35em"
                    fill="#64748b"
                    class="pointer-events-none select-none"
                    style="font-size: 8px;"
                  >
                    {node.data.type}
                  </Text>
                {/if}
              {/if}
            </g>
          {/each}
        {/snippet}
        </Tree>
      </Svg>
    </Chart>
  </div>
</HierarchyVisualizationBase>

<!-- Controls for orientation -->
<div class="absolute top-2 right-2 flex gap-1">
  <button
    class="px-2 py-1 text-xs rounded bg-background border border-border hover:bg-accent {orientation === 'horizontal' ? 'bg-accent' : ''}"
    onclick={() => orientation = 'horizontal'}
  >
    Horizontal
  </button>
  <button
    class="px-2 py-1 text-xs rounded bg-background border border-border hover:bg-accent {orientation === 'vertical' ? 'bg-accent' : ''}"
    onclick={() => orientation = 'vertical'}
  >
    Vertical
  </button>
</div>

<style>
  :global(.tree-node:hover) {
    transform: scale(1.1);
    stroke-width: 4px;
  }
</style>