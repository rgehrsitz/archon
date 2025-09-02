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
  let isReconstructingFromPath = false;
  let lastProcessedPath: string = '';
  
  // Hover state management
  let hoveredColumnIndex = -1;
  let hoverTimeout: NodeJS.Timeout | null = null;
  
  // Layout constants
  const COLUMN_WIDTH = 280;
  const COLUMN_OVERLAP = 40; // How much columns overlap
  const HIDDEN_COLUMN_WIDTH = 60; // Width when partially hidden
  const TRANSITION_DURATION = 300;
  const HOVER_DELAY = 150; // Delay before hover reveal
  
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
    if (isEntering) {
      // Clear any existing timeout
      if (hoverTimeout) {
        clearTimeout(hoverTimeout);
        hoverTimeout = null;
      }
      
      // Set hover state with delay for smooth UX
      hoverTimeout = setTimeout(() => {
        hoveredColumnIndex = columnIndex;
      }, HOVER_DELAY);
    } else {
      // Clear timeout and reset hover state
      if (hoverTimeout) {
        clearTimeout(hoverTimeout);
        hoverTimeout = null;
      }
      hoveredColumnIndex = -1;
    }
  }
  
  function getColumnTransform(columnIndex: number): string {
    const containerWidth = containerRef?.clientWidth || 800;
    const maxVisibleColumns = Math.floor(containerWidth / (COLUMN_WIDTH - COLUMN_OVERLAP));
    
    // Calculate how many columns are hidden
    const hiddenColumns = Math.max(0, columns.length - maxVisibleColumns);
    const isHidden = columnIndex < hiddenColumns;
    const isHovered = hoveredColumnIndex === columnIndex;
    
    let translateX = 0;
    let width = COLUMN_WIDTH;
    let zIndex = 10 + columnIndex;
    
    if (isHidden && !isHovered) {
      // Hidden column - show only a small portion
      const hiddenWidth = hiddenColumns * HIDDEN_COLUMN_WIDTH;
      const visibleColumns = Math.min(maxVisibleColumns, columns.length);
      const visibleWidth = visibleColumns * (COLUMN_WIDTH - COLUMN_OVERLAP);
      
      translateX = hiddenWidth + visibleWidth - COLUMN_WIDTH;
      width = HIDDEN_COLUMN_WIDTH;
      zIndex = 5 + columnIndex;
    } else if (isHidden && isHovered) {
      // Hovered hidden column - reveal it
      const hiddenWidth = hiddenColumns * HIDDEN_COLUMN_WIDTH;
      const visibleColumns = Math.min(maxVisibleColumns, columns.length);
      const visibleWidth = visibleColumns * (COLUMN_WIDTH - COLUMN_OVERLAP);
      
      translateX = hiddenWidth + visibleWidth - COLUMN_WIDTH;
      width = COLUMN_WIDTH;
      zIndex = 100; // Bring to front
    } else {
      // Visible column
      const hiddenWidth = hiddenColumns * HIDDEN_COLUMN_WIDTH;
      const visibleIndex = columnIndex - hiddenColumns;
      translateX = hiddenWidth + visibleIndex * (COLUMN_WIDTH - COLUMN_OVERLAP);
      zIndex = 10 + columnIndex;
    }
    
    return `translateX(${translateX}px)`;
  }
  
  function getColumnStyle(columnIndex: number): string {
    const containerWidth = containerRef?.clientWidth || 800;
    const maxVisibleColumns = Math.floor(containerWidth / (COLUMN_WIDTH - COLUMN_OVERLAP));
    
    const isHidden = columnIndex < Math.max(0, columns.length - maxVisibleColumns);
    const isHovered = hoveredColumnIndex === columnIndex;
    
    let width = COLUMN_WIDTH;
    let opacity = 1;
    let zIndex = 10 + columnIndex;
    
    if (isHidden && !isHovered) {
      width = HIDDEN_COLUMN_WIDTH;
      opacity = 0.6;
      zIndex = 5 + columnIndex;
    } else if (isHidden && isHovered) {
      width = COLUMN_WIDTH;
      opacity = 1;
      zIndex = 100;
    }
    
    return `
      width: ${width}px;
      transform: ${getColumnTransform(columnIndex)};
      opacity: ${opacity};
      z-index: ${zIndex};
      transition: all ${TRANSITION_DURATION}ms cubic-bezier(0.25, 0.46, 0.45, 0.94);
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
      class="absolute top-0 h-full bg-slate-800 border-r border-slate-700 shadow-lg"
      style={getColumnStyle(columnIndex)}
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
