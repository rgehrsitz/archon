import { visualizationRegistry } from './HierarchyVisualizationRegistry.js';
import TreeView from '$lib/components/workbench/TreeView.svelte';
import MillerColumns from '$lib/components/workbench/MillerColumns.svelte';
import LayerChart2PackVisualization from '$lib/components/workbench/visualizations/LayerChart2PackVisualization.svelte';
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

  // LayerChart 2.0 visualization (Svelte 5 compatible)
  visualizationRegistry.register({
    id: 'pack',
    name: 'Pack',
    description: 'Nested circles representing hierarchical structure',
    category: 'spatial',
    icon: '🎯',
    component: LayerChart2PackVisualization,
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