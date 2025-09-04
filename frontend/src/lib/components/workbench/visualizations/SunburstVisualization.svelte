<script lang="ts">
  /* @ts-nocheck */
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
  $: radius = size / 2 - 10;

  const colorScale = scaleOrdinal(schemeCategory10);

  function getNodeColor(node: any): string {
    const isSelected = node.data.id === selectedNodeId;
    const isInPath = selectedNodePath.some((p) => p.id === node.data.id);
    if (isSelected) return '#3b82f6';
    if (isInPath) return '#6366f1';
    const ancestors = typeof node.ancestors === 'function' ? node.ancestors() : [];
    const top = ancestors.find((n: any) => n.depth === 1) ?? node.parent ?? node;
    const key = top?.data?.id ?? String(node.depth);
    return colorScale(key);
  }

  function handleArcClick(node: any) {
    baseComponent?.handleNodeSelect(node);
  }

  function handleArcHover(node: any | null) {
    baseComponent?.handleNodeHover(node);
  }

  function shouldShowLabel(angle: number, r0: number, r1: number): boolean {
    const thickness = r1 - r0;
    return angle > 0.08 && thickness > 14;
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
  <div style="width: {size}px; height: {size}px;">
    <Chart padding={{ top: 10, right: 10, bottom: 10, left: 10 }}>
      <Svg center>
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

          <Partition hierarchy={summed} size={[2 * Math.PI, radius * radius]}>
            {#snippet children({ nodes })}
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
                  class="cursor-pointer sunburst-arc"
                  onclick={() => handleArcClick(node)}
                  onmouseenter={() => handleArcHover(node)}
                  onmouseleave={() => handleArcHover(null)}
                >
                  <title>{node.data.name} — {node.value ?? 0}</title>
                </Arc>
                {#if shouldShowLabel(end - start, r0, r1)}
                  {@const a = (start + end) / 2}
                  {@const rc = (r0 + r1) / 2}
                  <text
                    x={Math.cos(a - Math.PI / 2) * rc}
                    y={Math.sin(a - Math.PI / 2) * rc}
                    class="pointer-events-none select-none"
                    style="font-size: 10px;"
                    text-anchor="middle"
                    dy="0.35em"
                  >
                    {node.data.name}
                  </text>
                {/if}
              {/each}
            {/snippet}
          </Partition>
        {/if}
      </Svg>
    </Chart>
  </div>
</HierarchyVisualizationBase>

<style>
  :global(.sunburst-arc:hover) {
    filter: brightness(1.1);
  }
</style>