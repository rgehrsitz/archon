<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import { Chart, Svg, Treemap, Rect, Text, Group, Bounds, ChartClipPath, RectClipPath } from 'layerchart';
  import type { HierarchyNode, HierarchyRectangularNode } from 'd3-hierarchy';
  import type { ArchonNode } from '$lib/types/visualization.js';

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 800;
  export let height: number = 600;

  // Base component reference for accessing shared functionality
  let baseComponent: HierarchyVisualizationBase;
  
  // Selected node for zoom functionality
  let selectedNode: any = null;

  // Professional color palette
  const colorPalette = [
    '#2563eb', '#dc2626', '#059669', '#d97706', '#7c3aed', '#0891b2',
    '#ea580c', '#be185d', '#0d9488', '#65a30d', '#ca8a04', '#9333ea'
  ];


  function getNodeOpacity(node: any): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return 1.0;
    if (isInPath) return 0.9;
    return 0.8;
  }

  function getStrokeColor(node: any): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return '#1e40af'; // Dark blue for selected
    if (isInPath) return '#6366f1'; // Purple for path nodes
    return '#ffffff'; // White for normal nodes
  }

  function getStrokeWidth(node: any): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return 3;
    if (isInPath) return 2;
    return 1;
  }

  function handleNodeClick(node: any) {
    if (baseComponent) {
      baseComponent.handleNodeSelect(node);
    } else {
      console.error('TreemapVisualization: No base component available');
    }
  }

  function handleNodeHover(node: any | null, event: any) {
    // Add hover effects
    if (node && event && event.type === 'mouseenter') {
      // Could add tooltip or highlight effects here
    }
  }

  function shouldShowLabel(node: any): boolean {
    // Only show labels for nodes that are large enough
    const area = (node.x1 - node.x0) * (node.y1 - node.y0);
    return area > 1000; // Minimum area threshold
  }

  function getFontSize(node: any): number {
    const area = (node.x1 - node.x0) * (node.y1 - node.y0);
    const baseSize = Math.sqrt(area) / 20;
    return Math.min(14, Math.max(8, baseSize));
  }

  function isNodeVisible(node: any, selected: any): boolean {
    if (!selected) return true;
    return node === selected || selected.ancestors().includes(node);
  }

  function getNodeColor(node: any, colorBy: string = 'children'): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) {
      return '#1e40af'; // Darker blue for selected
    }
    
    if (isInPath) {
      return '#6366f1'; // Purple for path nodes
    }
    
    // Use a more sophisticated color assignment based on node properties
    const colorIndex = (node.data.id?.charCodeAt(0) || 0) % colorPalette.length;
    return colorPalette[colorIndex];
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
  <div class="w-full h-full bg-gradient-to-br from-slate-900 to-slate-800 rounded-lg shadow-sm border border-slate-600">
    {#if hierarchyData}
      <!-- Debug info -->
      <div class="absolute top-2 left-2 bg-slate-800 text-slate-200 text-xs p-2 rounded z-10 border border-slate-600">
        <div>Treemap Data: {hierarchyData ? 'loaded' : 'null'}</div>
        <div>Type: {typeof hierarchyData}</div>
        {#if hierarchyData}
          <div>Has children: {hierarchyData.children ? hierarchyData.children.length : 'none'}</div>
          <div>Data: {JSON.stringify(hierarchyData.data).substring(0, 50)}...</div>
          <div>Root name: {hierarchyData.data?.name || 'no name'}</div>
          <div>Value: {hierarchyData.value}</div>
          <div>Depth: {hierarchyData.depth}</div>
          {#if hierarchyData.children}
            <div>First child: {hierarchyData.children[0]?.data?.name}</div>
            <div>Child count: {hierarchyData.children.length}</div>
          {/if}
        {/if}
      </div>
      <!-- Interactive Treemap with LayerChart -->
      <div class="w-full h-full p-4">
        <Chart 
          data={hierarchyData}
          width={width - 32} 
          height={height - 32}
          padding={{ top: 10, right: 10, bottom: 10, left: 10 }}
        >
          <Svg>
            <Treemap 
              let:nodes 
              bind:selected={selectedNode}
            >
              <!-- Debug: Show nodes count -->
              <div class="absolute top-16 left-2 bg-red-800 text-red-200 text-xs p-2 rounded z-20 border border-red-600">
                <div>Nodes count: {nodes ? nodes.length : 'undefined'}</div>
                {#if nodes && nodes.length > 0}
                  <div>First node: {nodes[0]?.data?.name}</div>
                  <div>Node value: {nodes[0]?.value}</div>
                {/if}
              </div>
              {#each nodes as node}
                <Group
                  x={node.x0}
                  y={node.y0}
                  onclick={() => {
                    console.log('Treemap node clicked:', node.data?.name);
                    if (node.children) {
                      selectedNode = node;
                    }
                    handleNodeClick(node);
                  }}
                >
                  {@const nodeWidth = node.x1 - node.x0}
                  {@const nodeHeight = node.y1 - node.y0}
                  <RectClipPath width={nodeWidth} height={nodeHeight}>
                    {@const nodeColor = getNodeColor(node, 'children')}
                    {#if isNodeVisible(node, selectedNode)}
                      <g>
                        <Rect
                          width={nodeWidth}
                          height={nodeHeight}
                          stroke={getStrokeColor(node)}
                          stroke-opacity={0.2}
                          fill={nodeColor}
                          rx={5}
                        />
                        <Text
                          value="{node.data?.name || node.data?.id || 'Unknown'} ({node.children?.length ?? 0})"
                          class="text-[10px] font-medium fill-white"
                          verticalAnchor="start"
                          x={4}
                          y={2}
                        />
                        <Text
                          value={node.value ?? 0}
                          class="text-[8px] font-extralight fill-white"
                          verticalAnchor="start"
                          x={4}
                          y={16}
                        />
                      </g>
                    {/if}
                  </RectClipPath>
                </Group>
              {/each}
            </Treemap>
          </Svg>
        </Chart>
      </div>
      
      <!-- Native SVG Test -->
      <div class="absolute top-20 left-4 bg-yellow-100 border border-yellow-300 rounded p-2 z-20">
        <h4 class="text-sm font-medium mb-2">Native SVG Test</h4>
        <svg width="200" height="100" viewBox="0 0 200 100">
          <rect 
            x="10" 
            y="10" 
            width="80" 
            height="40" 
            fill="#3b82f6"
            stroke="#ffffff"
            stroke-width="2"
            onclick={() => alert('Native SVG clicked!')}
            class="cursor-pointer"
          />
          <text x="50" y="35" text-anchor="middle" fill="white" font-size="12">Click Me</text>
        </svg>
      </div>
    {:else}
      <div class="flex items-center justify-center h-full">
        <div class="text-center p-8">
          <div class="text-6xl mb-6 opacity-60">📊</div>
          <div class="text-xl font-semibold mb-3 text-slate-200">No Data Available</div>
          <div class="text-sm text-slate-400 max-w-sm">
            Load a project to see the treemap visualization of your hierarchy
          </div>
        </div>
      </div>
    {/if}
  </div>
</HierarchyVisualizationBase>

<style>
  :global(.treemap-rect:hover) {
    stroke-width: 2px;
    stroke: #1e40af;
    filter: brightness(1.1) drop-shadow(0 4px 8px rgba(0, 0, 0, 0.15));
    transform: scale(1.02);
    transform-origin: center;
  }

  :global(.treemap-text) {
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  }
</style>