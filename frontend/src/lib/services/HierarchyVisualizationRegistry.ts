import type { IHierarchyVisualization, VisualizationId } from '$lib/types/visualization.js';

/**
 * Registry for hierarchy visualization components.
 * Provides a plugin-like system for registering and managing different visualization types.
 */
export class HierarchyVisualizationRegistry {
  private static instance: HierarchyVisualizationRegistry;
  private visualizations = new Map<VisualizationId, IHierarchyVisualization>();

  private constructor() {}

  static getInstance(): HierarchyVisualizationRegistry {
    if (!HierarchyVisualizationRegistry.instance) {
      HierarchyVisualizationRegistry.instance = new HierarchyVisualizationRegistry();
    }
    return HierarchyVisualizationRegistry.instance;
  }

  /**
   * Register a new visualization component
   */
  register(visualization: IHierarchyVisualization): void {
    if (this.visualizations.has(visualization.id as VisualizationId)) {
      console.warn(`Visualization with id '${visualization.id}' is already registered. Overriding.`);
    }
    this.visualizations.set(visualization.id as VisualizationId, visualization);
  }

  /**
   * Get a visualization by ID
   */
  get(id: VisualizationId): IHierarchyVisualization | undefined {
    return this.visualizations.get(id);
  }

  /**
   * Get all registered visualizations
   */
  getAll(): IHierarchyVisualization[] {
    return Array.from(this.visualizations.values());
  }

  /**
   * Get visualizations by category
   */
  getByCategory(category: 'spatial' | 'network' | 'linear'): IHierarchyVisualization[] {
    return Array.from(this.visualizations.values()).filter(viz => viz.category === category);
  }

  /**
   * Check if a visualization is registered
   */
  has(id: VisualizationId): boolean {
    return this.visualizations.has(id);
  }

  /**
   * Unregister a visualization
   */
  unregister(id: VisualizationId): boolean {
    const visualization = this.visualizations.get(id);
    if (visualization && visualization.onDestroy) {
      visualization.onDestroy();
    }
    return this.visualizations.delete(id);
  }

  /**
   * Get available visualization IDs
   */
  getIds(): VisualizationId[] {
    return Array.from(this.visualizations.keys());
  }

  /**
   * Initialize a visualization component
   */
  initialize(id: VisualizationId): void {
    const visualization = this.visualizations.get(id);
    if (visualization && visualization.onMount) {
      visualization.onMount();
    }
  }
}

// Export singleton instance
export const visualizationRegistry = HierarchyVisualizationRegistry.getInstance();