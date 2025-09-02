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
    return () => ro.disconnect();
  });

  
  // Layout constants
  const COLUMN_WIDTH = 280;
  const COLUMN_OVERLAP = 40; // How much columns overlap
  const HIDDEN_COLUMN_WIDTH = 16; // Width when partially hidden
  const REVEAL_WIDTH = 180; // Width to reveal when hovering a hidden column
  const REVEAL_GAP = 12; // Small gap between revealed hidden column and visible stack
  const TRANSITION_DURATION = 300;
  const HOVER_DELAY = 50; // Delay before hover reveal
  
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
    if (isEntering) {
      // Clear any existing timeout
      if (hoverTimeout) {
        clearTimeout(hoverTimeout);
        hoverTimeout = null;
      }
      
      // Set hover state with delay for smooth UX
      hoverTimeout = setTimeout(() => {
        hoveredColumnIndex = columnIndex;
        console.log(`Hover activated for column ${columnIndex}`);
      }, HOVER_DELAY);
    } else {
      // Clear timeout and reset hover state with a small delay to prevent flickering
      if (hoverTimeout) {
        clearTimeout(hoverTimeout);
        hoverTimeout = null;
      }
      hoverTimeout = setTimeout(() => {
        hoveredColumnIndex = -1;
        console.log(`Hover deactivated for column ${columnIndex}`);
      }, 100); // Small delay to prevent flickering
    }
  }
  
  // Reactive statement to ensure we get the current columns length
  $: currentColumnsLength = columns.length;
  
  // Create reactive style objects for each column - explicitly depend on all relevant variables
  $: if (columns.length > 0) {
    columnStyles = columns.map((_, columnIndex) => {
      console.log(`Recalculating style for column ${columnIndex}`, { containerWidth, hoveredColumnIndex, currentColumnsLength });
      return getColumnStyle(columnIndex, currentColumnsLength);
    });
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
    } else {
      // Visible column (newest) - positioned after hidden stack
      const visibleIndex = columnIndex - hiddenColumns;
      const base = hiddenColumns * (HIDDEN_COLUMN_WIDTH + 3);
      
      // If there's a hovered hidden column, add extra space to push visible columns right
      const hoveredHiddenIndex = hoveredIndex >= 0 && hoveredIndex < hiddenColumns ? hoveredIndex : -1;
      const extraSpace = hoveredHiddenIndex >= 0 ? (REVEAL_WIDTH - HIDDEN_COLUMN_WIDTH) : 0;
      
      translateX = base + extraSpace + visibleIndex * (COLUMN_WIDTH - COLUMN_OVERLAP);
    }
    
    console.log(`Column ${columnIndex} Final: translateX=${translateX}px, isHidden=${isHidden}, isHovered=${isHovered}, hoveredIndex=${hoveredIndex}`);
    
    return `translateX(${translateX}px)`;
  }
  
  function getColumnStyle(columnIndex: number, totalColumns: number): string {
    const containerWidth = containerWidthRef();
    const maxVisibleColumns = Math.floor(containerWidth / (COLUMN_WIDTH - COLUMN_OVERLAP));
    
    // Only start stacking when we have more columns than can fit
    if (totalColumns <= maxVisibleColumns) {
      // Normal side-by-side positioning with overlap
      return `
        width: ${COLUMN_WIDTH}px;
        transform: ${getColumnTransform(columnIndex, totalColumns, containerWidth, hoveredColumnIndex)};
        opacity: 1;
        z-index: ${10 + columnIndex};
        transition: all ${TRANSITION_DURATION}ms cubic-bezier(0.25, 0.46, 0.45, 0.94);
      `;
    }
    
    // Space is limited - handle stacking
    const hiddenColumns = totalColumns - maxVisibleColumns;
    const isHidden = columnIndex < hiddenColumns;
    const isHovered = hoveredColumnIndex === columnIndex;
    
    console.log(`getColumnStyle for column ${columnIndex}: isHidden=${isHidden}, isHovered=${isHovered}, hoveredColumnIndex=${hoveredColumnIndex}`);
    
    let width = COLUMN_WIDTH;
    let opacity = 1;
    let zIndex = 10 + columnIndex;
    let boxShadow = 'inset -1px 0 0 rgba(15, 23, 42, 0.92), inset 1px 0 0 rgba(148, 163, 184, 0.08)';
    
    if (isHidden && !isHovered) {
      width = HIDDEN_COLUMN_WIDTH;
      opacity = 0.8;
      zIndex = 5 + columnIndex;
      boxShadow = 'inset -2px 0 0 rgba(15, 23, 42, 0.98), inset 2px 0 0 rgba(148, 163, 184, 0.6), inset 0 -1px 0 rgba(148, 163, 184, 0.3)';
    } else if (isHidden && isHovered) {
      // Reveal at a narrower, readable width so layout remains stable
      width = REVEAL_WIDTH;
      opacity = 1;
      zIndex = 100;
      boxShadow = '0 0 0 2px rgba(59, 130, 246, 0.7), 0 12px 32px rgba(0,0,0,0.8)';
    }
    
    return `
      width: ${width}px;
      transform: ${getColumnTransform(columnIndex, totalColumns, containerWidth, hoveredColumnIndex)};
      opacity: ${opacity};
      z-index: ${zIndex};
      box-shadow: ${boxShadow};
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
>
  {#each columns as column, columnIndex}
    <div
      class="absolute top-0 h-full bg-slate-800 shadow-lg cursor-pointer"
      style={columnStyles[columnIndex]}
      onmouseenter={() => handleColumnHover(columnIndex, true)}
      onmouseleave={() => handleColumnHover(columnIndex, false)}
      role="region"
      aria-label="Column {columnIndex + 1}"
    >
      <!-- Column Header -->
      <div class="h-12 border-b border-slate-700 bg-slate-800 px-4 flex items-center">
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
                class="w-full px-4 py-3 text-left hover:bg-slate-700 flex items-center gap-3 group transition-colors duration-150 {
                  column.selectedNodeId === node.id ? 'bg-blue-600 text-white' : 'text-slate-200 hover:text-white'
                }"
                onclick={() => handleNodeClick(node, columnIndex)}
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
</style>
