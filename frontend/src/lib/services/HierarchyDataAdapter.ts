import { hierarchy } from 'd3-hierarchy';
// Temporary type alias for d3-hierarchy (workaround for import issues)
type HierarchyNode<T> = any;
import { GetRootNode, ListChildren } from '../../../wailsjs/go/api/NodeService.js';
import type { ArchonNode } from '$lib/types/visualization.js';

// Internal interface for hierarchical data structure (with nested children instead of string IDs)
interface HierarchyArchonNode extends Omit<ArchonNode, 'children'> {
  children?: HierarchyArchonNode[];
}

/**
 * Transforms Archon node data into d3-hierarchy format for LayerChart components
 */
export class HierarchyDataAdapter {
  private cache = new Map<string, HierarchyNode<ArchonNode>>();
  private loadingNodes = new Set<string>();

  /**
   * Build a complete hierarchy starting from root
   */
  async buildFullHierarchy(): Promise<HierarchyNode<ArchonNode>> {
    const cacheKey = 'full_hierarchy';
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey)!;
    }

    // Check if Wails backend is available
    if (typeof window === 'undefined' || !window['go']) {
      console.log('Wails backend not available, using mock hierarchy data');
      const mockData = this.createMockHierarchy();
      const d3Hierarchy = hierarchy(mockData, (d: any) => d.children);
      d3Hierarchy.sum((d: any) => 1);
      d3Hierarchy.sort((a: any, b: any) => (a.data.name || '').localeCompare(b.data.name || ''));
      this.cache.set(cacheKey, d3Hierarchy);
      return d3Hierarchy;
    }

    try {
      console.log('Attempting to get root node from backend...');
      const rootNode = await GetRootNode();
      console.log('GetRootNode returned:', rootNode);

      // If no project is loaded, use mock data
      if (!rootNode) {
        console.log('No project loaded, using mock hierarchy data for testing');
        const mockData = this.createMockHierarchy();
        const d3Hierarchy = hierarchy(mockData, (d: any) => d.children);
        d3Hierarchy.sum((d: any) => 1);
        d3Hierarchy.sort((a: any, b: any) => (a.data.name || '').localeCompare(b.data.name || ''));
        this.cache.set(cacheKey, d3Hierarchy);
        return d3Hierarchy;
      }

      console.log('Building node hierarchy from root node...');
      const hierarchyData = await this.buildNodeHierarchy(this.transformArchonNode(rootNode));
      const d3Hierarchy = hierarchy(hierarchyData, (d: any) => d.children);

      // Add hierarchy-specific properties for LayerChart
      d3Hierarchy.sum((d: any) => 1); // Give each node a value of 1 for consistent visualization
      d3Hierarchy.sort((a: any, b: any) => (a.data.name || '').localeCompare(b.data.name || ''));

      this.cache.set(cacheKey, d3Hierarchy);
      return d3Hierarchy;
    } catch (error) {
      console.error('Failed to build full hierarchy:', error);
      console.error('Error details:', {
        message: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined
      });

      // If there's an error, try mock data as fallback
      console.log('Using mock hierarchy data as fallback due to error');
      const mockData = this.createMockHierarchy();
      const d3Hierarchy = hierarchy(mockData, (d: any) => d.children);
      d3Hierarchy.sum((d: any) => 1);
      d3Hierarchy.sort((a: any, b: any) => (a.data.name || '').localeCompare(b.data.name || ''));
      this.cache.set(cacheKey, d3Hierarchy);
      return d3Hierarchy;
    }
  }

  /**
   * Build hierarchy for a specific node and its descendants
   */
  async buildNodeHierarchy(node: ArchonNode, maxDepth = 10): Promise<HierarchyArchonNode> {
    console.log(`buildNodeHierarchy: Processing node ${node.name} (${node.id}), depth=${maxDepth}, hasChildren=${node.hasChildren}`);
    
    if (maxDepth <= 0 || !node.hasChildren) {
      console.log(`buildNodeHierarchy: Stopping at node ${node.name} - maxDepth=${maxDepth}, hasChildren=${node.hasChildren}`);
      return { ...node, children: [] };
    }

    if (this.loadingNodes.has(node.id)) {
      console.log(`buildNodeHierarchy: Already loading node ${node.name}, preventing loop`);
      return { ...node, children: [] }; // Prevent infinite loops
    }

    try {
      console.log(`buildNodeHierarchy: Adding ${node.name} to loading set`);
      this.loadingNodes.add(node.id);
      
      console.log(`buildNodeHierarchy: Calling ListChildren for ${node.name}...`);
      const children = await ListChildren(node.id);
      console.log(`buildNodeHierarchy: Got ${children.length} children for ${node.name}:`, children.map(c => c.name));
      
      console.log(`buildNodeHierarchy: Processing children for ${node.name}...`);
      const transformedChildren = await Promise.all(
        children.map(async (child) => {
          const transformedChild = this.transformArchonNode(child);
          console.log(`buildNodeHierarchy: Processing child ${transformedChild.name} of ${node.name}`);
          return this.buildNodeHierarchy(transformedChild, maxDepth - 1);
        })
      );

      console.log(`buildNodeHierarchy: Completed processing ${node.name} with ${transformedChildren.length} children`);
      return {
        ...node,
        children: transformedChildren
      };
    } catch (error) {
      console.error(`buildNodeHierarchy: Failed to load children for node ${node.id}:`, error);
      return { ...node, children: [] };
    } finally {
      console.log(`buildNodeHierarchy: Removing ${node.name} from loading set`);
      this.loadingNodes.delete(node.id);
    }
  }

  /**
   * Transform raw Archon node to standardized format
   */
  transformArchonNode(rawNode: any): ArchonNode {
    return {
      id: rawNode.id,
      name: rawNode.name || 'Untitled',
      hasChildren: typeof rawNode.hasChildren === 'boolean' 
        ? rawNode.hasChildren 
        : Boolean(rawNode.children && rawNode.children.length > 0),
      type: rawNode.type,
      metadata: rawNode.metadata,
      children: [], // Will be populated by buildNodeHierarchy
    };
  }

  /**
   * Build hierarchy from a selected path (for partial loading)
   */
  async buildHierarchyFromPath(path: ArchonNode[]): Promise<HierarchyNode<ArchonNode>> {
    if (path.length === 0) {
      return this.buildFullHierarchy();
    }

    // Start with the deepest selected node
    const targetNode = path[path.length - 1];
    const hierarchyData = await this.buildNodeHierarchy(targetNode, 3); // Limited depth for performance
    
    const d3Hierarchy = hierarchy(hierarchyData, (d: any) => d.children);
    d3Hierarchy.sum((d: any) => 1);
    d3Hierarchy.sort((a: any, b: any) => (a.data.name || '').localeCompare(b.data.name || ''));

    return d3Hierarchy;
  }

  /**
   * Create a lightweight hierarchy for visualizations that don't need full data
   */
  async buildLightweightHierarchy(): Promise<HierarchyNode<ArchonNode>> {
    console.log('buildLightweightHierarchy: Starting...');
    const cacheKey = 'lightweight_hierarchy';
    if (this.cache.has(cacheKey)) {
      console.log('buildLightweightHierarchy: Returning cached data');
      return this.cache.get(cacheKey)!;
    }

    try {
      console.log('buildLightweightHierarchy: Getting root node...');
      const rootNode = await GetRootNode();
      console.log('buildLightweightHierarchy: Root node received:', rootNode);
      
      console.log('buildLightweightHierarchy: Building node hierarchy...');
      const hierarchyData = await this.buildNodeHierarchy(this.transformArchonNode(rootNode), 2); // Only 2 levels deep
      console.log('buildLightweightHierarchy: Node hierarchy built:', hierarchyData);
      
      console.log('buildLightweightHierarchy: Creating d3 hierarchy...');
      const d3Hierarchy = hierarchy(hierarchyData, (d: any) => d.children);

      console.log('buildLightweightHierarchy: Adding sum and sort...');
      d3Hierarchy.sum((d: any) => 1);
      d3Hierarchy.sort((a: any, b: any) => (a.data.name || '').localeCompare(b.data.name || ''));
      
      console.log('buildLightweightHierarchy: Caching result...');
      this.cache.set(cacheKey, d3Hierarchy);
      console.log('buildLightweightHierarchy: Complete! Returning:', d3Hierarchy);
      return d3Hierarchy;
    } catch (error) {
      console.error('buildLightweightHierarchy: Error occurred:', error);
      console.error('buildLightweightHierarchy: Error details:', {
        message: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined
      });
      throw error;
    }
  }

  /**
   * Create mock hierarchy data for testing when no project is loaded
   */
  private createMockHierarchy(): HierarchyArchonNode {
    return {
      id: 'mock-root',
      name: 'Mock Project',
      hasChildren: true,
      type: 'project',
      metadata: { quantity: 100 },
      children: [
        {
          id: 'mock-docs',
          name: 'Documentation',
          hasChildren: true,
          type: 'folder',
          metadata: { quantity: 25 },
          children: [
            {
              id: 'mock-readme',
              name: 'README.md',
              hasChildren: false,
              type: 'file',
              metadata: { quantity: 1 },
              children: []
            },
            {
              id: 'mock-guide',
              name: 'User Guide',
              hasChildren: false,
              type: 'file',
              metadata: { quantity: 1 },
              children: []
            },
            {
              id: 'mock-api',
              name: 'API Reference',
              hasChildren: false,
              type: 'file',
              metadata: { quantity: 1 },
              children: []
            }
          ]
        },
        {
          id: 'mock-src',
          name: 'Source Code',
          hasChildren: true,
          type: 'folder',
          metadata: { quantity: 50 },
          children: [
            {
              id: 'mock-components',
              name: 'Components',
              hasChildren: true,
              type: 'folder',
              metadata: { quantity: 20 },
              children: [
                {
                  id: 'mock-button',
                  name: 'Button.svelte',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                },
                {
                  id: 'mock-input',
                  name: 'Input.svelte',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                },
                {
                  id: 'mock-modal',
                  name: 'Modal.svelte',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                }
              ]
            },
            {
              id: 'mock-services',
              name: 'Services',
              hasChildren: true,
              type: 'folder',
              metadata: { quantity: 15 },
              children: [
                {
                  id: 'mock-api-service',
                  name: 'api.ts',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                },
                {
                  id: 'mock-auth-service',
                  name: 'auth.ts',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                }
              ]
            },
            {
              id: 'mock-utils',
              name: 'Utils',
              hasChildren: true,
              type: 'folder',
              metadata: { quantity: 15 },
              children: [
                {
                  id: 'mock-helpers',
                  name: 'helpers.ts',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                },
                {
                  id: 'mock-constants',
                  name: 'constants.ts',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                }
              ]
            }
          ]
        },
        {
          id: 'mock-tests',
          name: 'Tests',
          hasChildren: true,
          type: 'folder',
          metadata: { quantity: 25 },
          children: [
            {
              id: 'mock-unit-tests',
              name: 'Unit Tests',
              hasChildren: true,
              type: 'folder',
              metadata: { quantity: 15 },
              children: [
                {
                  id: 'mock-button-test',
                  name: 'Button.test.ts',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                },
                {
                  id: 'mock-input-test',
                  name: 'Input.test.ts',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                }
              ]
            },
            {
              id: 'mock-integration-tests',
              name: 'Integration Tests',
              hasChildren: true,
              type: 'folder',
              metadata: { quantity: 10 },
              children: [
                {
                  id: 'mock-api-test',
                  name: 'api.test.ts',
                  hasChildren: false,
                  type: 'file',
                  metadata: { quantity: 1 },
                  children: []
                }
              ]
            }
          ]
        }
      ]
    };
  }

  /**
   * Clear the cache
   */
  clearCache(): void {
    this.cache.clear();
  }

  /**
   * Find a node in the hierarchy by ID
   */
  findNodeInHierarchy(hierarchy: HierarchyNode<ArchonNode>, nodeId: string): HierarchyNode<ArchonNode> | null {
    if (hierarchy.data.id === nodeId) {
      return hierarchy;
    }

    if (hierarchy.children) {
      for (const child of hierarchy.children) {
        const found = this.findNodeInHierarchy(child, nodeId);
        if (found) return found;
      }
    }

    return null;
  }

  /**
   * Get the path to a node in the hierarchy
   */
  getNodePath(hierarchy: HierarchyNode<ArchonNode>, nodeId: string): ArchonNode[] {
    const path: ArchonNode[] = [];

    function findPath(node: HierarchyNode<ArchonNode>): boolean {
      path.push(node.data);

      if (node.data.id === nodeId) {
        return true;
      }

      if (node.children) {
        for (const child of node.children) {
          if (findPath(child)) {
            return true;
          }
        }
      }

      path.pop();
      return false;
    }

    findPath(hierarchy);
    return path;
  }
}

// Export singleton instance
export const hierarchyDataAdapter = new HierarchyDataAdapter();