import { visualizationRegistry } from './HierarchyVisualizationRegistry.js';
import TreeView from '$lib/components/workbench/TreeView.svelte';
import MillerColumns from '$lib/components/workbench/MillerColumns.svelte';
import TreemapVisualization from '$lib/components/workbench/visualizations/TreemapVisualization.svelte';
import SunburstVisualization from '$lib/components/workbench/visualizations/SunburstVisualization.svelte';
import CirclePackVisualization from '$lib/components/workbench/visualizations/CirclePackVisualization.svelte';
import NodeLinkTreeVisualization from '$lib/components/workbench/visualizations/NodeLinkTreeVisualization.svelte';
import DebugVisualization from '$lib/components/workbench/visualizations/DebugVisualization.svelte';
import DirectAPITest from '$lib/components/workbench/visualizations/DirectAPITest.svelte';
import LayerChartTest from '$lib/components/workbench/visualizations/LayerChartTest.svelte';
import MinimalTest from '$lib/components/workbench/visualizations/MinimalTest.svelte';
import type { IHierarchyVisualization } from '$lib/types/visualization.js';

/**
 * Register all available hierarchy visualizations
 */
export function registerAllVisualizations() {
  // Traditional visualizations (existing)
  visualizationRegistry.register({
    id: 'miller',
    name: 'Miller Columns',
    description: 'Multi-column file browser view for deep navigation',
    category: 'linear',
    icon: '🗂️',
    component: MillerColumns,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 300,
      aspectRatio: 16/9
    },
    capabilities: {
      supportsPanning: true,
      supportsSelection: true,
      requiresFullHierarchy: false
    }
  } as IHierarchyVisualization);

  visualizationRegistry.register({
    id: 'tree',
    name: 'Tree View',
    description: 'Traditional expandable tree structure',
    category: 'linear',
    icon: '🌳',
    component: TreeView,
    layoutConstraints: {
      minWidth: 300,
      minHeight: 400
    },
    capabilities: {
      supportsPanning: true,
      supportsSelection: true,
      requiresFullHierarchy: false
    }
  } as IHierarchyVisualization);

  // LayerChart visualizations (new)
  visualizationRegistry.register({
    id: 'treemap',
    name: 'Treemap',
    description: 'Space-filling rectangular hierarchy visualization',
    category: 'spatial',
    icon: '🔲',
    component: TreemapVisualization,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 300,
      aspectRatio: 4/3
    },
    capabilities: {
      supportsZooming: true,
      supportsSelection: true,
      supportsTooltips: true,
      requiresFullHierarchy: true
    }
  } as IHierarchyVisualization);

  visualizationRegistry.register({
    id: 'sunburst',
    name: 'Sunburst',
    description: 'Radial partition showing hierarchy as concentric rings',
    category: 'spatial',
    icon: '☀️',
    component: SunburstVisualization,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 400,
      aspectRatio: 1,
      preferredWidth: 600,
      preferredHeight: 600
    },
    capabilities: {
      supportsZooming: true,
      supportsSelection: true,
      supportsTooltips: true,
      requiresFullHierarchy: true
    }
  } as IHierarchyVisualization);

  visualizationRegistry.register({
    id: 'pack',
    name: 'Circle Packing',
    description: 'Nested circles representing hierarchical structure',
    category: 'spatial',
    icon: '🎯',
    component: CirclePackVisualization,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 400,
      aspectRatio: 1,
      preferredWidth: 600,
      preferredHeight: 600
    },
    capabilities: {
      supportsZooming: true,
      supportsSelection: true,
      supportsTooltips: true,
      requiresFullHierarchy: true
    }
  } as IHierarchyVisualization);

  visualizationRegistry.register({
    id: 'node-link',
    name: 'Node-Link Tree',
    description: 'Traditional tree diagram with nodes and connecting lines',
    category: 'network',
    icon: '🌲',
    component: NodeLinkTreeVisualization,
    layoutConstraints: {
      minWidth: 600,
      minHeight: 400,
      aspectRatio: 3/2
    },
    capabilities: {
      supportsPanning: true,
      supportsZooming: true,
      supportsSelection: true,
      supportsTooltips: true,
      requiresFullHierarchy: true
    }
  } as IHierarchyVisualization);

  // Debug visualization for troubleshooting
  visualizationRegistry.register({
    id: 'debug',
    name: 'Debug View',
    description: 'Debug visualization for troubleshooting hierarchy data',
    category: 'network',
    icon: '🐛',
    component: DebugVisualization,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 300
    },
    capabilities: {
      supportsSelection: true,
      requiresFullHierarchy: false
    }
  } as IHierarchyVisualization);


  // Direct API test (bypasses HierarchyDataAdapter)
  visualizationRegistry.register({
    id: 'direct-api',
    name: 'Direct API Test',
    description: 'Direct test of Wails API calls without data adapter',
    category: 'network',
    icon: '🔧',
    component: DirectAPITest,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 300
    },
    capabilities: {
      supportsSelection: true,
      requiresFullHierarchy: false
    }
  } as IHierarchyVisualization);

  // LayerChart test (basic functionality test)
  visualizationRegistry.register({
    id: 'layerchart-test',
    name: 'LayerChart Test',
    description: 'Basic LayerChart functionality test with simple rectangles',
    category: 'network',
    icon: '🧪',
    component: LayerChartTest,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 300
    },
    capabilities: {
      supportsSelection: false,
      requiresFullHierarchy: false
    }
  } as IHierarchyVisualization);

  // Minimal test (basic Svelte functionality test)
  visualizationRegistry.register({
    id: 'minimal-test',
    name: 'Minimal Test',
    description: 'Minimal Svelte component test to verify basic rendering',
    category: 'network',
    icon: '🔬',
    component: MinimalTest,
    layoutConstraints: {
      minWidth: 400,
      minHeight: 300
    },
    capabilities: {
      supportsSelection: false,
      requiresFullHierarchy: false
    }
  } as IHierarchyVisualization);


  console.log(`Registered ${visualizationRegistry.getAll().length} hierarchy visualizations`);
}

/**
 * Initialize the visualization registry on first import
 */
if (typeof window !== 'undefined') {
  registerAllVisualizations();
}

// Re-export the singleton for convenience
export { visualizationRegistry };