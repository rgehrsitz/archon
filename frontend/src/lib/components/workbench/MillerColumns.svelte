<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { GetRootNode, ListChildren } from '../../../../wailsjs/go/api/NodeService.js';
  
  export let projectId: string;
  
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
  }
  
  let columns: Column[] = [];
  let containerRef: HTMLDivElement;
  let columnRefs: HTMLDivElement[] = [];
  let visibleColumnCount = 3;
  let columnOffset = 0;
  let hoveredColumnIndex = -1;
  let isHovering = false;
  
  // Column width and transition settings
  const COLUMN_WIDTH = 240;
  const COLUMN_MIN_WIDTH = 60; // When partially hidden
  const TRANSITION_DURATION = 200;
  
  // Helper function to transform backend nodes to frontend format
  function transformNode(node: any) {
    return {
      id: node.id,
      name: node.name,
      hasChildren: node.children && node.children.length > 0,
      // Include additional node properties as needed
      type: node.type,
      metadata: node.metadata
    };
  }
  
  onMount(async () => {
    await loadRootColumn();
    updateLayout();
  });
  
  async function loadRootColumn() {
    try {
      const rootNode = await GetRootNode();
      const children = await ListChildren(rootNode.id);
      
      columns = [{
        id: 'root',
        parentId: null,
        nodes: children.map(transformNode),
        loading: false,
        selectedNodeId: null
      }];
    } catch (error) {
      console.error('Failed to load root nodes:', error);
      // Show empty state on error
      columns = [{
        id: 'root',
        parentId: null,
        nodes: [],
        loading: false,
        selectedNodeId: null
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
      selectedNodeId: null
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
    
    // Update layout after adding new column
    updateLayout();
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
  
  function updateLayout() {
    if (!containerRef) return;
    
    const containerWidth = containerRef.clientWidth;
    visibleColumnCount = Math.floor(containerWidth / COLUMN_WIDTH);
    visibleColumnCount = Math.max(2, Math.min(5, visibleColumnCount)); // 2-5 columns
    
    // Adjust offset if we have more columns than can fit
    if (columns.length > visibleColumnCount) {
      columnOffset = Math.max(0, columns.length - visibleColumnCount);
    } else {
      columnOffset = 0;
    }
  }
  
  function handleNodeClick(node: any, columnIndex: number) {
    loadChildrenColumn(node, columnIndex);
  }
  
  function handleColumnHover(columnIndex: number, isEntering: boolean) {
    if (isEntering) {
      hoveredColumnIndex = columnIndex;
      isHovering = true;
    } else {
      hoveredColumnIndex = -1;
      isHovering = false;
    }
  }
  
  function getColumnStyle(columnIndex: number): string {
    const actualIndex = columnIndex - columnOffset;
    const isVisible = actualIndex >= 0 && actualIndex < visibleColumnCount;
    const isPartiallyHidden = columnIndex < columnOffset;
    const isHovered = isHovering && hoveredColumnIndex === columnIndex;
    
    let translateX = 0;
    let width = COLUMN_WIDTH;
    let opacity = 1;
    let zIndex = 10 + columnIndex;
    
    if (isPartiallyHidden && !isHovered) {
      // Partially hidden column
      width = COLUMN_MIN_WIDTH;
      translateX = actualIndex * COLUMN_MIN_WIDTH;
      opacity = 0.7;
    } else if (isPartiallyHidden && isHovered) {
      // Revealed column on hover
      width = COLUMN_WIDTH;
      translateX = actualIndex * COLUMN_MIN_WIDTH;
      zIndex = 100; // Bring to front
    } else if (isVisible) {
      // Fully visible column
      const hiddenColumnsWidth = Math.max(0, columnOffset) * COLUMN_MIN_WIDTH;
      translateX = hiddenColumnsWidth + (actualIndex - Math.max(0, columnOffset)) * COLUMN_WIDTH;
    } else {
      // Hidden column
      opacity = 0;
      translateX = containerRef?.clientWidth || 0;
    }
    
    return `
      width: ${width}px;
      transform: translateX(${translateX}px);
      opacity: ${opacity};
      z-index: ${zIndex};
      transition: all ${TRANSITION_DURATION}ms cubic-bezier(0.2, 0, 0.2, 1);
    `;
  }
  
  // Handle window resize
  let resizeTimeout: NodeJS.Timeout;
  function handleResize() {
    clearTimeout(resizeTimeout);
    resizeTimeout = setTimeout(updateLayout, 100);
  }
</script>

<svelte:window onresize={handleResize} />

<div 
  bind:this={containerRef}
  class="relative h-full overflow-hidden bg-background"
>
  {#each columns as column, columnIndex}
    <div
      bind:this={columnRefs[columnIndex]}
      class="absolute top-0 h-full border-r border-border bg-background"
      style={getColumnStyle(columnIndex)}
      onmouseenter={() => handleColumnHover(columnIndex, true)}
      onmouseleave={() => handleColumnHover(columnIndex, false)}
      role="region"
      aria-label="Column {columnIndex + 1}"
    >
      <!-- Column Header -->
      <div class="h-10 border-b border-border bg-muted/30 px-3 flex items-center">
        <h3 class="text-sm font-medium truncate">
          {column.id === 'root' ? 'Root' : 
           columns[columnIndex - 1]?.nodes.find(n => n.id === column.selectedNodeId)?.name || 'Unknown'}
        </h3>
      </div>
      
      <!-- Column Content -->
      <div class="h-full overflow-y-auto">
        {#if column.loading}
          <!-- Loading State -->
          <div class="p-4 text-center">
            <div class="animate-pulse">
              <div class="space-y-2">
                {#each Array(3) as _}
                  <div class="h-8 bg-muted rounded"></div>
                {/each}
              </div>
            </div>
          </div>
        {:else if column.nodes.length === 0}
          <!-- Empty State -->
          <div class="p-4 text-center text-muted-foreground">
            <div class="text-2xl mb-2">📁</div>
            <div class="text-sm">No items</div>
          </div>
        {:else}
          <!-- Node List -->
          <div class="py-1">
            {#each column.nodes as node}
              <button
                class="w-full px-3 py-2 text-left hover:bg-accent hover:text-accent-foreground flex items-center gap-2 group {
                  node.id === column.selectedNodeId ? 'bg-accent text-accent-foreground' : ''
                }"
                onclick={() => handleNodeClick(node, columnIndex)}
              >
                <!-- Node Icon -->
                <span class="text-sm opacity-60">
                  {node.hasChildren ? '📁' : '📄'}
                </span>
                
                <!-- Node Name -->
                <span class="flex-1 text-sm truncate">
                  {node.name}
                </span>
                
                <!-- Chevron for expandable nodes -->
                {#if node.hasChildren}
                  <span class="text-xs opacity-40 group-hover:opacity-70">
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
  
  <!-- Column count indicator (bottom right) -->
  <div class="absolute bottom-2 right-2 text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
    {columns.length} column{columns.length !== 1 ? 's' : ''}
  </div>
</div>