import type { ComponentType } from 'svelte';
import type { HierarchyNode } from 'd3-hierarchy';

export interface ArchonNode {
  id: string;
  name: string;
  hasChildren?: boolean;
  type?: string;
  metadata?: any;
  children?: string[];
}

export interface VisualizationLayoutConstraints {
  minWidth?: number;
  minHeight?: number;
  aspectRatio?: number;
  preferredWidth?: number;
  preferredHeight?: number;
}

export interface VisualizationCapabilities {
  supportsPanning?: boolean;
  supportsZooming?: boolean;
  supportsSelection?: boolean;
  supportsTooltips?: boolean;
  requiresFullHierarchy?: boolean;
}

export interface HierarchyVisualizationProps {
  projectId: string;
  selectedNodeId: string | null;
  selectedNodePath: ArchonNode[];
  data?: HierarchyNode<ArchonNode>;
  width?: number;
  height?: number;
}

export interface HierarchyVisualizationEvents {
  nodeSelect: { node: ArchonNode; path: ArchonNode[] };
  nodeHover?: { node: ArchonNode | null };
  viewportChange?: { zoom: number; pan: [number, number] };
}

export interface IHierarchyVisualization {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly category: 'spatial' | 'network' | 'linear';
  readonly icon: string;
  readonly component: ComponentType;
  readonly layoutConstraints: VisualizationLayoutConstraints;
  readonly capabilities: VisualizationCapabilities;
  
  // Lifecycle hooks
  onMount?(): void;
  onDestroy?(): void;
  onDataUpdate?(data: HierarchyNode<ArchonNode>): void;
}

export type VisualizationId = 'miller' | 'tree' | 'treemap' | 'sunburst' | 'pack' | 'node-link' | 'icicle' | 'debug' | 'simple-test' | 'direct-api';

export interface VisualizationSettings {
  [key: string]: any;
}