<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import { Chart, Svg, Pack, Circle, Text } from 'layerchart';
  import type { HierarchyCircularNode } from 'd3-hierarchy';
  import type { ArchonNode } from '$lib/types/visualization.js';
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

  // Ensure square aspect ratio for circle packing
  $: size = Math.min(width, height);

  // Professional color palette
  const colorPalette = [
    '#2563eb', '#dc2626', '#059669', '#d97706', '#7c3aed', '#0891b2',
    '#ea580c', '#be185d', '#0d9488', '#65a30d', '#ca8a04', '#9333ea'
  ];

  function getNodeColor(node: any): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return '#1e40af';
    if (isInPath) return '#6366f1';
    
    // Use sophisticated color assignment
    const colorIndex = (node.data.id?.charCodeAt(0) || 0) % colorPalette.length;
    return colorPalette[colorIndex];
  }

  function getNodeOpacity(node: any): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (node.depth === 0) return 0.05; // Root circle is barely visible
    if (isSelected) return 0.95;
    if (isInPath) return 0.85;
    return 0.75;
  }

  function getStrokeColor(node: any): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return '#1e40af';
    if (isInPath) return '#6366f1';
    return '#ffffff';
  }

  function getStrokeWidth(node: any): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return 3;
    if (isInPath) return 2;
    return 1.5;
  }

  function handleCircleClick(node: any) {
    if (baseComponent) {
      baseComponent.handleNodeSelect(node);
    }
  }

  function handleCircleHover(node: any | null) {
    if (baseComponent) {
      baseComponent.handleNodeHover(node);
    }
  }

  function shouldShowLabel(node: any): boolean {
    // Show labels for circles with radius > 20 and depth > 0
    return node.r > 20 && node.depth > 0;
  }

  function getFontSize(node: any): number {
    // Dynamic font size based on circle radius
    return Math.min(14, Math.max(10, node.r / 3));
  }

  function getTruncatedName(name: string, radius: number): string {
    const maxChars = Math.floor(radius / 2.5);
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
  <div class="w-full h-full bg-gradient-to-br from-slate-900 to-slate-800 rounded-lg shadow-sm border border-slate-600">
    {#if hierarchyData}
      <!-- Debug info -->
      <div class="absolute top-2 left-2 bg-slate-800 text-slate-200 text-xs p-2 rounded z-10 border border-slate-600">
        <div>Circle Pack Data: {hierarchyData ? 'loaded' : 'null'}</div>
        <div>Type: {typeof hierarchyData}</div>
        {#if hierarchyData}
          <div>Has children: {hierarchyData.children ? hierarchyData.children.length : 'none'}</div>
          <div>Data: {JSON.stringify(hierarchyData.data).substring(0, 50)}...</div>
        {/if}
      </div>
      <div class="w-full h-full p-6 flex items-center justify-center">
        <div style="width: {size - 48}px; height: {size - 48}px;">
          <Chart data={hierarchyData} padding={{ top: 20, right: 20, bottom: 20, left: 20 }}>
            <Svg>
              <Pack padding={2} size={[size - 88, size - 88]}>
                {#snippet children({ nodes }: { nodes: any[] })}
                  {#each nodes as node}
                    <g
                      role="button"
                      tabindex="0"
                      onclick={() => handleCircleClick(node)}
                      onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleCircleClick(node)}
                      onmouseenter={() => handleCircleHover(node)}
                      onmouseleave={() => handleCircleHover(null)}
                    >
                      <Circle
                        cx={node.x}
                        cy={node.y}
                        r={node.r}
                        fill={getNodeColor(node)}
                        opacity={getNodeOpacity(node)}
                        stroke={getStrokeColor(node)}
                        strokeWidth={getStrokeWidth(node)}
                        class="circle-pack-node cursor-pointer transition-all duration-300 ease-out"
                      />
                        {#if shouldShowLabel(node)}
                          <Text
                            x={node.x}
                            y={node.y}
                            textAnchor="middle"
                            dy="0.35em"
                            fill={node.depth === 1 ? '#ffffff' : '#1e293b'}
                            class="pointer-events-none select-none font-semibold drop-shadow-sm"
                            style={`font-size: ${getFontSize(node)}px;`}
                          >
                            {getTruncatedName(node.data?.name || node.data?.id || 'Unknown', node.r)}
                          </Text>
                        {#if node.r > 30 && (node.data.type || (node.children && node.children.length > 0))}
                          <Text
                            x={node.x}
                            y={node.y + getFontSize(node) + 3}
                            textAnchor="middle"
                            dy="0.35em"
                            fill={node.depth === 1 ? '#e2e8f0' : '#64748b'}
                            class="pointer-events-none select-none font-medium"
                            style={`font-size: ${Math.max(8, getFontSize(node) - 2)}px;`}
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
        </div>
      </div>
    {:else}
      <div class="flex items-center justify-center h-full">
        <div class="text-center p-8">
          <div class="text-6xl mb-6 opacity-60">🎯</div>
          <div class="text-xl font-semibold mb-3 text-slate-200">No Data Available</div>
          <div class="text-sm text-slate-400 max-w-sm">
            Load a project to see the circle packing visualization of your hierarchy
          </div>
        </div>
      </div>
    {/if}
  </div>
</HierarchyVisualizationBase>

<style>
  :global(.circle-pack-node:hover) {
    filter: brightness(1.15) drop-shadow(0 6px 12px rgba(0, 0, 0, 0.2));
    transform: scale(1.08);
    transform-origin: center;
    stroke-width: 3px;
  }

  :global(.circle-pack-node) {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
</style>