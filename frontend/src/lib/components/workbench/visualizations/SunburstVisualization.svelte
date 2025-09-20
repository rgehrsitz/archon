<script lang="ts">
  import HierarchyVisualizationBase from './HierarchyVisualizationBase.svelte';
  import { Chart, Svg, Partition, Arc } from 'layerchart';
  import type { ArchonNode } from '$lib/types/visualization.js';
  import { scaleOrdinal } from 'd3-scale';
  import { schemeCategory10 } from 'd3-scale-chromatic';

  export let projectId: string;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: ArchonNode[] = [];
  export let width: number = 600;
  export let height: number = 600;

  let baseComponent: HierarchyVisualizationBase;

  $: size = Math.min(width, height);
  $: radius = size / 2 - 20;

  // Professional color palette
  const colorPalette = [
    '#2563eb', '#dc2626', '#059669', '#d97706', '#7c3aed', '#0891b2',
    '#ea580c', '#be185d', '#0d9488', '#65a30d', '#ca8a04', '#9333ea'
  ];

  function getNodeColor(node: any): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some((p) => p.id === node.data.id);
    
    if (isSelected) return '#1e40af';
    if (isInPath) return '#6366f1';
    
    // Use sophisticated color assignment
    const colorIndex = (node.data.id?.charCodeAt(0) || 0) % colorPalette.length;
    return colorPalette[colorIndex];
  }

  function getNodeOpacity(node: any): number {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some((p) => p.id === node.data.id);
    
    if (isSelected) return 1.0;
    if (isInPath) return 0.9;
    return 0.8;
  }

  function handleArcClick(node: any) {
    baseComponent?.handleNodeSelect(node);
  }

  function handleArcHover(node: any | null) {
    baseComponent?.handleNodeHover(node);
  }

  function shouldShowLabel(angle: number, r0: number, r1: number): boolean {
    const thickness = r1 - r0;
    return angle > 0.1 && thickness > 16;
  }

  function getFontSize(r0: number, r1: number): number {
    const thickness = r1 - r0;
    return Math.min(12, Math.max(8, thickness / 2));
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
      {@const summed = (() => {
        const h: any = hierarchyData as any;
        const copy = typeof h.copy === 'function' ? h.copy() : h;
        copy.sum((d: ArchonNode) => {
          const q = (d as any)?.metadata?.quantity;
          return typeof q === 'number' && isFinite(q) && q > 0 ? q : 1;
        });
        return copy;
      })()}
      <div class="w-full h-full p-6 flex items-center justify-center">
        <div style="width: {size - 48}px; height: {size - 48}px;">
            <Chart data={summed} padding={{ top: 20, right: 20, bottom: 20, left: 20 }}>
              <Svg center>
                <Partition size={[2 * Math.PI, radius * radius]}>
                  {#snippet children({ nodes }: { nodes: any[] })}
                  {#each nodes.slice(1) as node}
                    {@const start = node.x0}
                    {@const end = node.x1}
                    {@const r0 = Math.sqrt(node.y0)}
                    {@const r1 = Math.sqrt(node.y1)}
                    <Arc
                      value={node.value}
                      startAngle={start}
                      endAngle={end}
                      innerRadius={r0}
                      outerRadius={r1}
                      fill={getNodeColor(node)}
                      opacity={getNodeOpacity(node)}
                      stroke="#ffffff"
                      strokeWidth="1"
                      class="sunburst-arc cursor-pointer transition-all duration-300 ease-out"
                      onclick={() => handleArcClick(node)}
                      onmouseenter={() => handleArcHover(node)}
                      onmouseleave={() => handleArcHover(null)}
                    >
                      <title>{node.data?.name || node.data?.id || 'Unknown'} — {node.value ?? 0}</title>
                    </Arc>
                    {#if shouldShowLabel(end - start, r0, r1)}
                      {@const a = (start + end) / 2}
                      {@const rc = (r0 + r1) / 2}
                        <text
                          x={Math.cos(a - Math.PI / 2) * rc}
                          y={Math.sin(a - Math.PI / 2) * rc}
                          class="pointer-events-none select-none font-semibold drop-shadow-sm"
                          style={`font-size: ${getFontSize(r0, r1)}px;`}
                          text-anchor="middle"
                          dy="0.35em"
                          fill="#ffffff"
                        >
                          {node.data?.name || node.data?.id || 'Unknown'}
                        </text>
                    {/if}
                  {/each}
                {/snippet}
              </Partition>
            </Svg>
          </Chart>
        </div>
      </div>
    {:else}
      <div class="flex items-center justify-center h-full">
        <div class="text-center p-8">
          <div class="text-6xl mb-6 opacity-60">☀️</div>
          <div class="text-xl font-semibold mb-3 text-slate-200">No Data Available</div>
          <div class="text-sm text-slate-400 max-w-sm">
            Load a project to see the sunburst visualization of your hierarchy
          </div>
        </div>
      </div>
    {/if}
  </div>
</HierarchyVisualizationBase>

<style>
  :global(.sunburst-arc:hover) {
    filter: brightness(1.15) drop-shadow(0 4px 8px rgba(0, 0, 0, 0.2));
    stroke-width: 2px;
    stroke: #1e40af;
  }

  :global(.sunburst-arc) {
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }
</style>