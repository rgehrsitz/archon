<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { GetRootNode, ListChildren } from '../../../../wailsjs/go/api/NodeService.js';
  
  export const projectId: string = undefined!;
  export let selectedNodeId: string | null = null;
  export let selectedNodePath: any[] = [];
  
  const dispatch = createEventDispatcher<{
    nodeSelect: { node: any, path: any[] };
  }>();
  
  // Tree node interface
  interface TreeNode {
    id: string;
    name: string;
    type?: string;
    metadata?: any;
    children?: TreeNode[];
    expanded: boolean;
    loading: boolean;
    level: number;
    hasChildren: boolean;
  }
  
  let rootNodes: TreeNode[] = [];
  let loading = true;
  let isExpandingFromPath = false;
  let lastProcessedPath: string = '';
  
  // Helper function to transform backend nodes to frontend format
  function transformNode(node: any, level: number = 0): TreeNode {
    return {
      id: node.id,
      name: node.name,
      type: node.type,
      metadata: node.metadata,
      children: undefined, // Will be loaded on expansion
      expanded: false,
      loading: false,
      level,
      hasChildren: node.children && node.children.length > 0
    };
  }
  
  onMount(async () => {
    await loadRootNodes();
  });
  
  // Expand nodes when selectedNodePath changes (e.g., when switching from MillerColumns)
  $: {
    const currentPathKey = selectedNodePath.map(n => n.id).join('->');
    if (selectedNodePath.length > 0 && 
        rootNodes.length > 0 && 
        !isExpandingFromPath && 
        currentPathKey !== lastProcessedPath) {
      lastProcessedPath = currentPathKey;
      expandNodesFromPath(selectedNodePath);
    }
  }
  
  async function expandNodesFromPath(path: any[]) {
    if (path.length === 0 || isExpandingFromPath) return;
    
    isExpandingFromPath = true;
    
    // Helper function to find and expand a node by ID
    async function findAndExpandNode(nodes: TreeNode[], nodeId: string): Promise<TreeNode | null> {
      for (const node of nodes) {
        if (node.id === nodeId) {
          // Found the node, expand it if it has children
          if (node.hasChildren && !node.expanded) {
            await loadChildren(node);
            node.expanded = true;
          }
          return node;
        }
        
        // Search in children if this node is expanded
        if (node.expanded && node.children) {
          const found = await findAndExpandNode(node.children, nodeId);
          if (found) return found;
        }
      }
      return null;
    }
    
    try {
      // Expand each node in the path (except the last one, which is just selected)
      for (let i = 0; i < path.length - 1; i++) {
        const nodeToExpand = path[i];
        await findAndExpandNode(rootNodes, nodeToExpand.id);
      }
      rootNodes = rootNodes; // Trigger reactivity once at the end
    } catch (error) {
      console.error('Failed to expand nodes from path:', error);
    } finally {
      isExpandingFromPath = false;
    }
  }
  
  async function loadRootNodes() {
    loading = true;
    try {
      const rootNode = await GetRootNode();
      const children = await ListChildren(rootNode.id);
      
      rootNodes = children.map(node => transformNode(node, 0));
    } catch (error) {
      console.error('Failed to load root nodes:', error);
      rootNodes = [];
    } finally {
      loading = false;
    }
  }
  
  async function loadChildren(parentNode: TreeNode) {
    if (parentNode.children || !parentNode.hasChildren) return;
    
    parentNode.loading = true;
    rootNodes = rootNodes; // Trigger reactivity
    
    try {
      const children = await ListChildren(parentNode.id);
      parentNode.children = children.map(node => transformNode(node, parentNode.level + 1));
    } catch (error) {
      console.error('Failed to load children:', error);
      parentNode.children = [];
    } finally {
      parentNode.loading = false;
      rootNodes = rootNodes; // Trigger reactivity
    }
  }
  
  async function handleNodeClick(node: TreeNode) {
    if (node.hasChildren) {
      if (!node.expanded) {
        // Expand the node
        await loadChildren(node);
        node.expanded = true;
      } else {
        // Collapse the node
        node.expanded = false;
      }
      rootNodes = rootNodes; // Trigger reactivity
    }
    
    // Emit selection event
    const path = getNodePath(node);
    dispatch('nodeSelect', { node, path });
  }
  
  function getNodePath(targetNode: TreeNode): TreeNode[] {
    const path: TreeNode[] = [];
    
    function findPath(nodes: TreeNode[], target: TreeNode): boolean {
      for (const node of nodes) {
        path.push(node);
        
        if (node.id === target.id) {
          return true;
        }
        
        if (node.children && node.expanded) {
          if (findPath(node.children, target)) {
            return true;
          }
        }
        
        path.pop();
      }
      return false;
    }
    
    findPath(rootNodes, targetNode);
    return path;
  }
  
  // Recursive function to render tree nodes
  function renderNodes(nodes: TreeNode[]): TreeNode[] {
    const flatNodes: TreeNode[] = [];
    
    for (const node of nodes) {
      flatNodes.push(node);
      
      if (node.expanded && node.children) {
        flatNodes.push(...renderNodes(node.children));
      }
    }
    
    return flatNodes;
  }
  
  $: visibleNodes = renderNodes(rootNodes);
</script>

<div class="h-full overflow-auto bg-background">
  {#if loading}
    <!-- Loading State -->
    <div class="p-4">
      <div class="animate-pulse space-y-2">
        {#each Array(5) as _, i (i)}
          <div class="h-8 bg-muted rounded flex items-center px-3">
            <div class="w-4 h-4 bg-muted-foreground/20 rounded mr-2"></div>
            <div class="flex-1 h-4 bg-muted-foreground/20 rounded"></div>
          </div>
        {/each}
      </div>
    </div>
  {:else if rootNodes.length === 0}
    <!-- Empty State -->
    <div class="p-8 text-center text-muted-foreground">
      <div class="text-4xl mb-4">🌳</div>
      <div class="text-lg font-medium mb-2">No items</div>
      <div class="text-sm">This project doesn't contain any nodes yet.</div>
    </div>
  {:else}
    <!-- Tree Content -->
    <div class="py-2">
      {#each visibleNodes as node (`${node.id}-${node.level}`)}
        <button
          class="w-full px-3 py-2 text-left hover:bg-accent hover:text-accent-foreground flex items-center gap-2 group {
            node.id === selectedNodeId ? 'bg-accent text-accent-foreground' : ''
          }"
          onclick={() => handleNodeClick(node)}
          style="padding-left: {12 + node.level * 20}px"
        >
          <!-- Expand/Collapse Indicator -->
          <div class="w-4 h-4 flex items-center justify-center">
            {#if node.loading}
              <div class="w-3 h-3 border-2 border-current border-t-transparent rounded-full animate-spin opacity-60"></div>
            {:else if node.hasChildren}
              <span class="text-xs opacity-60 transition-transform {
                node.expanded ? 'rotate-90' : 'rotate-0'
              }">
                ▶
              </span>
            {:else}
              <span class="w-1 h-1 bg-current opacity-30 rounded-full"></span>
            {/if}
          </div>
          
          <!-- Node Icon -->
          <span class="text-sm opacity-60">
            {node.hasChildren ? (node.expanded ? '📂' : '📁') : '📄'}
          </span>
          
          <!-- Node Name -->
          <span class="flex-1 text-sm truncate">
            {node.name}
          </span>
          
          <!-- Node Type Badge (if available) -->
          {#if node.type && node.type !== 'folder'}
            <span class="text-xs px-1.5 py-0.5 bg-muted text-muted-foreground rounded">
              {node.type}
            </span>
          {/if}
        </button>
      {/each}
    </div>
  {/if}
</div>