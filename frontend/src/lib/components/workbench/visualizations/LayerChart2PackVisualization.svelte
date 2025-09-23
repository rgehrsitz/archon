<script lang="ts">
  import { Chart, Svg, Pack, Circle, Text } from 'layerchart';
  import { hierarchyDataAdapter } from '$lib/services/HierarchyDataAdapter.js';
  import type { ArchonNode } from '$lib/types/visualization.js';
  import { createEventDispatcher, onMount } from 'svelte';

  // Props
  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 600;
  export let height: number = 600;

  // Event dispatcher
  const dispatch = createEventDispatcher<{
    nodeSelect: { node: ArchonNode; path: ArchonNode[] };
    nodeHover?: { node: ArchonNode | null };
  }>();

  // Internal state
  let hierarchyData: any = null;
  let loading = true;
  let error: string | null = null;

  // Reactive data loading
  let lastProjectId: string = '';
  let hasLoadedOnce = false;
  
  $: if (projectId !== lastProjectId) {
    lastProjectId = projectId;
    if (hasLoadedOnce) {
      loadHierarchyData();
    }
  }

  async function loadHierarchyData() {
    console.log('LayerChart2PackVisualization: Loading hierarchy data for project:', projectId);
    loading = true;
    error = null;
    hierarchyData = null;

    try {
      // Load full hierarchy
      hierarchyData = await hierarchyDataAdapter.buildFullHierarchy();
      console.log('LayerChart2PackVisualization: Hierarchy data loaded successfully', hierarchyData);
    } catch (err) {
      console.error('LayerChart2PackVisualization: Failed to load hierarchy data:', err);
      error = err instanceof Error ? err.message : 'Failed to load hierarchy data';
    } finally {
      loading = false;
      hasLoadedOnce = true;
    }
  }

  onMount(() => {
    loadHierarchyData();
  });

  function getNodeColor(node: any): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (isSelected) return '#3b82f6';
    if (isInPath) return '#6366f1';
    
    // Default color palette
    const colors = [
      '#2563eb', '#dc2626', '#059669', '#d97706', '#7c3aed', '#0891b2',
      '#ea580c', '#be185d', '#0d9488', '#65a30d', '#ca8a04', '#9333ea'
    ];
    const colorIndex = (node.data.id?.charCodeAt(0) || 0) % colors.length;
    return colors[colorIndex];
  }

  function getNodeOpacity(node: any): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some(p => p.id === node.data.id);
    
    if (node.depth === 0) return 0.05;
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

  function handleNodeClick(node: any) {
    if (!node?.data) return;
    
    // Build path to this node
    const path: ArchonNode[] = [];
    let current = node;
    while (current?.parent) {
      path.unshift(current.parent.data);
      current = current.parent;
    }
    path.push(node.data);
    
    dispatch('nodeSelect', { node: node.data, path });
  }

  function handleNodeHover(node: any | null) {
    dispatch('nodeHover', { node: node?.data || null });
  }

  function shouldShowLabel(node: any): boolean {
    return node.r > 20 && node.depth > 0;
  }

  function getFontSize(node: any): number {
    return Math.min(14, Math.max(10, node.r / 3));
  }

  function getTruncatedName(name: string, radius: number): string {
    const maxChars = Math.floor(radius / 2.5);
    return name.length > maxChars ? name.substring(0, maxChars) + '...' : name;
  }
</script>

{#if loading}
  <div class="w-full h-full flex items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800 rounded-lg">
    <div class="text-center">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
      <div class="text-slate-200">Loading hierarchy data...</div>
    </div>
  </div>
{:else if error}
  <div class="w-full h-full flex items-center justify-center bg-gradient-to-br from-red-900 to-red-800 rounded-lg">
    <div class="text-center p-8">
      <div class="text-6xl mb-6 opacity-60">⚠️</div>
      <div class="text-xl font-semibold mb-3 text-red-200">Error Loading Data</div>
      <div class="text-sm text-red-300 max-w-sm">{error}</div>
    </div>
  </div>
{:else if hierarchyData}
  <div class="w-full h-full bg-gradient-to-br from-slate-900 to-slate-800 rounded-lg shadow-sm border border-slate-600">
    <div class="w-full h-full p-6 flex items-center justify-center">
      <div style="width: {width - 48}px; height: {height - 48}px;">
        <Chart data={hierarchyData} padding={{ top: 20, right: 20, bottom: 20, left: 20 }}>
          <Svg>
            <Pack padding={2} size={[width - 88, height - 88]}>
              {#snippet children({ nodes }: { nodes: any[] })}
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
  </div>
{:else}
  <div class="w-full h-full flex items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800 rounded-lg">
    <div class="text-center p-8">
      <div class="text-6xl mb-6 opacity-60">🎯</div>
      <div class="text-xl font-semibold mb-3 text-slate-200">No Data Available</div>
      <div class="text-sm text-slate-400 max-w-sm">
        Load a project to see the circle packing visualization of your hierarchy
      </div>
    </div>
  </div>
{/if}

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
