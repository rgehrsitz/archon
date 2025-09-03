<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { GetRootNode, ListChildren } from '../../../../wailsjs/go/api/NodeService.js';
  
  export const projectId: string = undefined!;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: any[] = [];
  
  const dispatch = createEventDispatcher<{
    nodeSelect: { node: any, path: any[] };
  }>();
  
  // Column management
  interface Column {
    id: string;
    parentId: string | null;
    nodes: any[];
    loading: boolean;
    selectedNodeId: string | null;
    title: string;
  }
  
  let columns: Column[] = [];
  let containerRef: HTMLDivElement;
  let containerWidth = 0;
  let isReconstructingFromPath = false;
  let lastProcessedPath: string = '';
  let columnStyles: string[] = [];
  
  // Hover state management
  let hoveredColumnIndex = -1;
  let hoverTimeout: NodeJS.Timeout | null = null;
  let isMouseOverColumns = false;
  // Keep containerWidth reactive to layout changes
  onMount(() => {
    const measure = () => {
      if (containerRef) {
        const newWidth = containerRef.clientWidth;
        if (newWidth !== containerWidth) {
          containerWidth = newWidth;
          console.log('Container width updated:', containerWidth);
        }
      }
    };
    measure();
    const ro = new ResizeObserver(() => {
      console.log('ResizeObserver callback fired');
      measure();
    });
    if (containerRef) {
      ro.observe(containerRef);
    }
    const onWinResize = () => {
      console.log('Window resize event');
      measure();
    };
    window.addEventListener('resize', onWinResize);
    return () => { ro.disconnect(); window.removeEventListener('resize', onWinResize); };
  });

  
  // Layout constants
  const COLUMN_WIDTH = 280;
  const COLUMN_OVERLAP = 40; // How much columns overlap
  const HIDDEN_COLUMN_WIDTH = 16; // Width when partially hidden
  const REVEAL_WIDTH = 180; // Width to reveal when hovering a hidden column
  const REVEAL_GAP = 12; // Small gap between revealed hidden column and visible stack
  const TRANSITION_DURATION = 300;
  const HOVER_DELAY = 0; // No delay for immediate response
  
  // Helper function to transform backend nodes to frontend format
  function transformNode(node: any) {
    return {
      id: node.id,
      name: node.name,
      hasChildren: node.children && node.children.length > 0,
      type: node.type,
      metadata: node.metadata
    };
  }
  
  onMount(async () => {
    await loadRootColumn();
  });
  
  // React to selectedNodeId changes from parent
  $: if (selectedNodeId && columns.length > 0) {
    // Find and highlight the selected node in the appropriate column
    columns.forEach((column, columnIndex) => {
      const node = column.nodes.find(n => n.id === selectedNodeId);
      if (node) {
        column.selectedNodeId = selectedNodeId;
      }
    });
  }
  
  // Reconstruct columns when selectedNodePath changes
  $: {
    const currentPathKey = selectedNodePath.map(n => n.id).join('->');
    if (selectedNodePath.length > 0 && 
        columns.length <= 1 && 
        !isReconstructingFromPath &&
        currentPathKey !== lastProcessedPath) {
      lastProcessedPath = currentPathKey;
      reconstructColumnsFromPath(selectedNodePath);
    }
  }
  
  async function reconstructColumnsFromPath(path: any[]) {
    if (path.length === 0 || isReconstructingFromPath) return;
    
    isReconstructingFromPath = true;
    
    try {
      // Start with root column
      await loadRootColumn();
      
      // Build columns for each node in the path (except the last one)
      for (let i = 0; i < path.length - 1; i++) {
        const currentNode = path[i];
        const nextNode = path[i + 1];
        
        // Find current node in the current column and expand it
        const currentColumnIndex = i;
        if (currentColumnIndex < columns.length) {
          const column = columns[currentColumnIndex];
          const nodeInColumn = column.nodes.find(n => n.id === currentNode.id);
          
          if (nodeInColumn && nextNode) {
            await loadChildrenColumn(nodeInColumn, currentColumnIndex);
          }
        }
      }
    } catch (error) {
      console.error('Failed to reconstruct columns from path:', error);
    } finally {
      isReconstructingFromPath = false;
    }
  }
  
  async function loadRootColumn() {
    try {
      const rootNode = await GetRootNode();
      const children = await ListChildren(rootNode.id);
      
      columns = [{
        id: 'root',
        parentId: null,
        nodes: children.map(transformNode),
        loading: false,
        selectedNodeId: null,
        title: 'Root'
      }];
    } catch (error) {
      console.error('Failed to load root nodes:', error);
      columns = [{
        id: 'root',
        parentId: null,
        nodes: [],
        loading: false,
        selectedNodeId: null,
        title: 'Root'
      }];
    }
  }
  
  async function loadChildrenColumn(parentNode: any, columnIndex: number) {
    // Remove columns after the selected one
    columns = columns.slice(0, columnIndex + 1);
    
    // Mark parent as selected
    columns[columnIndex].selectedNodeId = parentNode.id;
    
    if (!parentNode.hasChildren) {
      // Emit selection event for leaf nodes
      dispatch('nodeSelect', { 
        node: parentNode, 
        path: getNodePath(parentNode, columnIndex)
      });
      return;
    }
    
    // Add new loading column
    const newColumn: Column = {
      id: parentNode.id,
      parentId: parentNode.id,
      nodes: [],
      loading: true,
      selectedNodeId: null,
      title: parentNode.name.toUpperCase()
    };
    
    columns = [...columns, newColumn];
    
    try {
      const children = await ListChildren(parentNode.id);
      
      // Update the loading column with data
      columns[columns.length - 1] = {
        ...newColumn,
        nodes: children.map(transformNode),
        loading: false
      };
    } catch (error) {
      console.error('Failed to load children:', error);
      // Remove the failed column
      columns = columns.slice(0, -1);
    }
  }
  
  function getNodePath(node: any, columnIndex: number): any[] {
    const path = [];
    
    // Build path from column selections
    for (let i = 0; i <= columnIndex; i++) {
      const column = columns[i];
      if (i === columnIndex) {
        path.push(node);
      } else if (column.selectedNodeId) {
        const selectedNode = column.nodes.find(n => n.id === column.selectedNodeId);
        if (selectedNode) path.push(selectedNode);
      }
    }
    
    return path;
  }
  
  function handleNodeClick(node: any, columnIndex: number) {
    loadChildrenColumn(node, columnIndex);
  }
  
  function handleColumnHover(columnIndex: number, isEntering: boolean) {
    console.log(`Column ${columnIndex} hover: ${isEntering ? 'enter' : 'leave'}`);
    
    // Clear any existing timeout
    if (hoverTimeout) {
      clearTimeout(hoverTimeout);
      hoverTimeout = null;
    }
    
    if (isEntering) {
      isMouseOverColumns = true;
      // Immediate hover activation
      hoveredColumnIndex = columnIndex;
      console.log(`Hover activated for column ${columnIndex}`);
    } else {
      // Small delay to prevent flickering when moving between columns
      hoverTimeout = setTimeout(() => {
        hoveredColumnIndex = -1;
        isMouseOverColumns = false;
        console.log(`Hover deactivated for column ${columnIndex}`);
      }, 50); // Reduced delay
    }
  }
  
  // Reactive statement to ensure we get the current columns length
  $: currentColumnsLength = columns.length;

  // Compute simple layout (widths + left edges) based on container width and hover
  interface Layout { widths: number[]; lefts: number[] }

  function computeLayout(totalColumns: number, cw: number, hoveredIndex: number): Layout {
    const FW = COLUMN_WIDTH;
    const HW = HIDDEN_COLUMN_WIDTH;
    const PW = REVEAL_WIDTH;
    const n = totalColumns;
    const widths: number[] = new Array(n).fill(HW);
    const lefts: number[] = new Array(n).fill(0);

    let RW = cw;
    for (let i = n - 1; i >= 0; i--) {
      if (FW + i * HW <= RW) {
        widths[i] = FW;
        RW -= FW;
      } else {
        widths[i] = HW;
        RW -= HW;
      }
    }

    // base lefts
    for (let i = 1; i < n; i++) lefts[i] = lefts[i-1] + widths[i-1];

    // apply hover: if hovering a visible column while hidden exist, treat hover as last hidden index
    if (hoveredIndex >= 0 && hoveredIndex < n) {
      const lastHidden = widths.lastIndexOf(HW);
      let start = hoveredIndex;
      if (lastHidden >= 0 && widths[hoveredIndex] !== HW) {
        start = lastHidden; // reveal as if hovering last hidden column
      }
      for (let i = start; i < n; i++) {
        if (widths[i] === HW) widths[i] = PW;
      }
      // recompute lefts from start forward
      for (let i = Math.max(1, start); i < n; i++) {
        lefts[i] = lefts[i-1] + widths[i-1];
      }
    }

    return { widths, lefts };
  }

  $: layout = computeLayout(currentColumnsLength, containerWidth, hoveredColumnIndex);

  // Count fully hidden columns for UI indicators
  $: hiddenCount = layout?.widths?.filter((w) => w === HIDDEN_COLUMN_WIDTH).length || 0;
  $: hiddenTicks = Math.min(hiddenCount, 9);

  // Create reactive style strings for each column; explicitly depend on layout
  $: if (columns.length > 0) {
    columnStyles = columns.map((_, columnIndex) => getColumnStyle(columnIndex, currentColumnsLength, layout));
  }
  
  function containerWidthRef(): number {
    return containerWidth || containerRef?.clientWidth || 800;
  }
  
  function getColumnTransform(
    columnIndex: number,
    totalColumns: number,
    widthPx: number,
    hoveredIndex: number
  ): string {
    console.log(`getColumnTransform called for column ${columnIndex}, columns.length: ${totalColumns}`);
    const containerWidth = widthPx;
    const isHovered = hoveredIndex === columnIndex;
    
    // Calculate how many columns can fit in the container
    const maxVisibleColumns = Math.floor(containerWidth / (COLUMN_WIDTH - COLUMN_OVERLAP));
    
    // Only start stacking when we have more columns than can fit
    if (totalColumns <= maxVisibleColumns) {
      // Normal side-by-side positioning with overlap
      console.log(`Column ${columnIndex}: Normal positioning, translateX: ${columnIndex * (COLUMN_WIDTH - COLUMN_OVERLAP)}px`);
      return `translateX(${columnIndex * (COLUMN_WIDTH - COLUMN_OVERLAP)}px)`;
    }
    
    // Space is limited - keep NEWEST columns visible (rightmost), hide OLDEST as strips (leftmost)
    const hiddenColumns = totalColumns - maxVisibleColumns;
    const isHidden = columnIndex < hiddenColumns;
    
    // Debug logging for all columns
    console.log(`Column ${columnIndex} Debug:`, {
      containerWidth,
      maxVisibleColumns,
      totalColumns: totalColumns,
      willStack: totalColumns > maxVisibleColumns,
      hiddenColumns: hiddenColumns,
      isHidden: isHidden
    });
    
    let translateX = 0;
    
    if (isHidden && !isHovered) {
      // Hidden column (oldest) - stack on the extreme left, compact spacing
      translateX = columnIndex * (HIDDEN_COLUMN_WIDTH + 3);
    } else if (isHidden && isHovered) {
      // Hovered hidden column - stays in its original position, expands in place
      translateX = columnIndex * (HIDDEN_COLUMN_WIDTH + 3);
    } else if (isHidden) {
      // Hidden column that should be revealed due to hover (not the hovered one itself)
      // Check if this column should be revealed based on hover
      const hoveredHiddenIndex = hoveredIndex >= 0 && hoveredIndex < hiddenColumns ? hoveredIndex : -1;
      const shouldReveal = hoveredHiddenIndex >= 0 && columnIndex > hoveredHiddenIndex && columnIndex < hiddenColumns;

      if (shouldReveal) {
        // Lay out revealed hidden columns contiguously to the right of the hovered one,
        // separated by REVEAL_GAP, with a slightly narrower width (breadcrumb look).
        const step = (HIDDEN_COLUMN_WIDTH + 3);
        const revealedWidth = REVEAL_WIDTH - 20;
        const leftOfHovered = hoveredHiddenIndex * step;
        const rightOfHovered = leftOfHovered + REVEAL_WIDTH;
        const idxFromHovered = columnIndex - hoveredHiddenIndex - 1; // 0 for the first revealed after hovered
        translateX = rightOfHovered + REVEAL_GAP + idxFromHovered * (revealedWidth + REVEAL_GAP);
        console.log(`Column ${columnIndex}: contiguous reveal, translateX=${translateX}`);
      } else {
        // Normal hidden column positioning
        translateX = columnIndex * (HIDDEN_COLUMN_WIDTH + 3);
      }
    } else {
      // Visible column (newest) - positioned after hidden stack
      const visibleIndex = columnIndex - hiddenColumns;
      const base = hiddenColumns * (HIDDEN_COLUMN_WIDTH + 3);
      
      // Check if this visible column should be revealed due to hover
      const hoveredHiddenIndex = hoveredIndex >= 0 && hoveredIndex < hiddenColumns ? hoveredIndex : -1;

      // Compute the exact right edge of the hidden stack after reveal to avoid gaps
      let extraSpace = 0;
      if (hoveredHiddenIndex >= 0) {
        const h = hoveredHiddenIndex;
        const step = (HIDDEN_COLUMN_WIDTH + 3);
        const baseRight = hiddenColumns * step; // right edge with all hidden compact
        const revealedWidth = REVEAL_WIDTH - 20; // width used for non-hovered revealed hidden cols

        const leftOfHovered = h * step;
        const rightOfHovered = leftOfHovered + REVEAL_WIDTH;
        const numRevealedAfter = Math.max(0, hiddenColumns - h - 1);

        let lastHiddenRight = rightOfHovered;
        if (numRevealedAfter > 0) {
          // total width added by revealed-after columns plus gaps between them and after hovered
          const addedWidth = numRevealedAfter * revealedWidth;
          const addedGaps = (numRevealedAfter) * REVEAL_GAP; // one gap after hovered + between each revealed
          lastHiddenRight = rightOfHovered + REVEAL_GAP + (numRevealedAfter - 1) * (revealedWidth + REVEAL_GAP) + revealedWidth;
          // Simplify for clarity: rightOfHovered + REVEAL_GAP + sum_{i=0..n-1}(revealedWidth + REVEAL_GAP)
          lastHiddenRight = rightOfHovered + REVEAL_GAP + numRevealedAfter * revealedWidth + (numRevealedAfter - 1) * REVEAL_GAP;
        }

        // Shift visible columns so they start exactly at the revealed stack's edge
        extraSpace = Math.max(0, Math.round(lastHiddenRight - baseRight));
        console.log(`Column ${columnIndex}: computed extraSpace=${extraSpace}, lastHiddenRight=${lastHiddenRight}, baseRight=${baseRight}`);
      }

      // Clamp extraSpace so the rightmost visible column remains within container bounds
      const stepVisible = (COLUMN_WIDTH - COLUMN_OVERLAP);
      const numVisible = totalColumns - hiddenColumns;
      const maxExtraSpace = Math.max(0, widthPx - COLUMN_WIDTH - base - (numVisible - 1) * stepVisible);
      const clampedExtra = Math.min(extraSpace, maxExtraSpace);

      translateX = base + clampedExtra + visibleIndex * stepVisible;
    }
    
    console.log(`Column ${columnIndex} Final: translateX=${translateX}px, isHidden=${isHidden}, isHovered=${isHovered}, hoveredIndex=${hoveredIndex}`);
    
    return `translateX(${translateX}px)`;
  }
  
  function getColumnStyle(columnIndex: number, totalColumns: number, layoutArg: Layout): string {
    const width = layoutArg.widths[columnIndex] ?? COLUMN_WIDTH;
    const left = layoutArg.lefts[columnIndex] ?? columnIndex * (COLUMN_WIDTH - COLUMN_OVERLAP);
    const isHiddenWidth = width === HIDDEN_COLUMN_WIDTH;
    const isPartialWidth = width === REVEAL_WIDTH;
    const zIndex = 10 + columnIndex; // natural stacking

    let opacity = 1;
    // Background lightness ramps slightly with index for separation
    const baseL = 17 + Math.min(columnIndex, 8) * 1.2;
    const bgL = baseL + (isPartialWidth ? 1.2 : 2.4);
    const background = `hsl(215 28% ${bgL}%)`;

    let boxShadow = '0 8px 24px rgba(0,0,0,0.35), inset 1px 0 0 rgba(255,255,255,0.05), inset -1px 0 0 rgba(0,0,0,0.55)';
    if (isHiddenWidth) {
      opacity = 0.85;
      boxShadow = '0 4px 12px rgba(0,0,0,0.25), inset 1px 0 0 rgba(255,255,255,0.06), inset -1px 0 0 rgba(0,0,0,0.7)';
    } else if (isPartialWidth) {
      opacity = 0.95;
      boxShadow = '0 6px 18px rgba(0,0,0,0.30), inset 1px 0 0 rgba(255,255,255,0.08), inset -1px 0 0 rgba(0,0,0,0.6)';
    }
    if (columnIndex === hoveredColumnIndex) {
      boxShadow += ', 0 0 0 2px rgba(59,130,246,0.55) inset';
    }

    return `
      width: ${width}px;
      transform: translateX(${left}px);
      background-color: ${background};
      opacity: ${opacity};
      z-index: ${zIndex};
      box-shadow: ${boxShadow};
      border-right: 1px solid rgba(148,163,184,0.18);
      border-left: 1px solid rgba(15,23,42,0.85);
      box-sizing: border-box;
      border-radius: 10px 10px 0 0;
      transition: transform ${TRANSITION_DURATION}ms cubic-bezier(0.25, 0.46, 0.45, 0.94), width ${TRANSITION_DURATION}ms cubic-bezier(0.25, 0.46, 0.45, 0.94), opacity ${TRANSITION_DURATION}ms ease;
    `;
  }
  
  function getColumnTitle(column: Column, columnIndex: number): string {
    if (column.id === 'root') return 'ROOT';
    if (column.title) return column.title;
    
    // Find the parent node that led to this column
    const parentColumn = columns[columnIndex - 1];
    if (parentColumn && parentColumn.selectedNodeId) {
      const parentNode = parentColumn.nodes.find(n => n.id === parentColumn.selectedNodeId);
      return parentNode?.name?.toUpperCase() || 'UNKNOWN';
    }
    
    return 'UNKNOWN';
  }
</script>

<div 
  bind:this={containerRef}
  class="relative h-full overflow-hidden bg-slate-900"
  role="region"
  aria-label="Miller Columns Navigation"
  on:mouseleave={() => {
    if (hoverTimeout) {
      clearTimeout(hoverTimeout);
      hoverTimeout = null;
    }
    hoveredColumnIndex = -1;
    isMouseOverColumns = false;
    console.log('Mouse left container, clearing hover state');
  }}
>
  {#if hiddenCount > 0}
    <!-- Left hidden columns rail -->
    <div class="left-rail pointer-events-none absolute left-0 top-0 h-full" style="width: {HIDDEN_COLUMN_WIDTH}px;">
      <div class="rail-bg absolute inset-0"></div>
      <div class="badge absolute -top-3 left-1/2 -translate-x-1/2 text-[10px] px-2 py-0.5 rounded-full bg-blue-500/80 text-white shadow">
        {hiddenCount}
      </div>
      <div class="ticks absolute left-1/2 -translate-x-1/2 top-14 bottom-4 flex flex-col justify-between">
        {#each Array(hiddenTicks) as _, i}
          <div class="tick w-[6px] h-[10px] rounded-sm bg-slate-300/40"></div>
        {/each}
      </div>
    </div>
  {/if}
  {#each columns as column, columnIndex}
    <div
      class="miller-column absolute top-0 h-full bg-slate-800 shadow-lg cursor-pointer"
      style={columnStyles[columnIndex]}
      on:mouseenter={() => handleColumnHover(columnIndex, true)}
      on:mouseleave={() => handleColumnHover(columnIndex, false)}
      role="region"
      aria-label="Column {columnIndex + 1}"
    >
      <!-- Column Header -->
      <div class="column-header h-12 border-b border-slate-700 px-4 flex items-center">
        <h3 class="text-sm font-semibold text-slate-200 truncate">
          {getColumnTitle(column, columnIndex)}
        </h3>
      </div>
      
      <!-- Column Content -->
      <div class="h-full overflow-y-auto bg-slate-800">
        {#if column.loading}
          <!-- Loading State -->
          <div class="p-4 text-center">
            <div class="animate-pulse space-y-3">
              {#each Array(4) as _}
                <div class="h-8 bg-slate-700 rounded"></div>
              {/each}
            </div>
          </div>
        {:else if column.nodes.length === 0}
          <!-- Empty State -->
          <div class="p-6 text-center text-slate-400">
            <div class="text-3xl mb-3">📁</div>
            <div class="text-sm">No items</div>
          </div>
        {:else}
          <!-- Node List -->
          <div class="py-2">
            {#each column.nodes as node}
              <button
                class="w-full px-4 py-3 text-left hover:bg-slate-700/70 flex items-center gap-3 group transition-colors duration-150 border-b border-slate-700/40 {
                  column.selectedNodeId === node.id ? 'bg-blue-600 text-white ring-2 ring-blue-400/30' : 'text-slate-200 hover:text-white'
                }"
                on:click={() => handleNodeClick(node, columnIndex)}
              >
                <!-- Node Icon -->
                <span class="text-lg opacity-70 group-hover:opacity-100">
                  {node.hasChildren ? '📁' : '📄'}
                </span>
                
                <!-- Node Name -->
                <span class="flex-1 text-sm truncate font-medium">
                  {node.name}
                </span>
                
                <!-- Chevron for expandable nodes -->
                {#if node.hasChildren}
                  <span class="text-slate-400 group-hover:text-slate-200 transition-colors">
                    ›
                  </span>
                {/if}
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/each}
  
  <!-- Column count indicator -->
  {#if columns.length > 3}
    <div class="absolute bottom-4 right-4 text-xs text-slate-400 bg-slate-800 px-3 py-2 rounded-lg border border-slate-700">
      {columns.length} levels deep
    </div>
  {/if}
</div>

<style>
  /* Custom scrollbar for dark theme */
  :global(.overflow-y-auto::-webkit-scrollbar) {
    width: 6px;
  }
  
  :global(.overflow-y-auto::-webkit-scrollbar-track) {
    background: #1e293b;
  }
  
  :global(.overflow-y-auto::-webkit-scrollbar-thumb) {
    background: #475569;
    border-radius: 3px;
  }
  
  :global(.overflow-y-auto::-webkit-scrollbar-thumb:hover) {
    background: #64748b;
  }

  /* Column chrome */
  .miller-column::before,
  .miller-column::after {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    width: 10px;
    pointer-events: none;
  }
  .miller-column::before {
    left: -10px;
    background: linear-gradient(to right, rgba(0,0,0,0.28), rgba(0,0,0,0));
  }
  .miller-column::after {
    right: -10px;
    background: linear-gradient(to left, rgba(0,0,0,0.28), rgba(0,0,0,0));
  }

  .column-header {
    background: linear-gradient(to bottom, rgba(255,255,255,0.06), rgba(255,255,255,0));
    backdrop-filter: saturate(120%);
  }

  /* Hidden rail */
  .left-rail .rail-bg {
    background: linear-gradient(to bottom, rgba(15,23,42,0.9), rgba(15,23,42,0.65));
    box-shadow: inset -1px 0 0 rgba(255,255,255,0.06), inset 1px 0 0 rgba(0,0,0,0.7);
    border-right: 1px solid rgba(148,163,184,0.18);
  }
  .left-rail .badge {
    backdrop-filter: blur(2px);
  }
  .left-rail .ticks .tick {
    box-shadow: inset 0 -1px 0 rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.1);
  }
</style>
