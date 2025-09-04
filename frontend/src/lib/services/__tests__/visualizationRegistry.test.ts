import { describe, it, expect, beforeEach } from 'vitest';
import { visualizationRegistry } from '../HierarchyVisualizationRegistry.js';

describe('Visualization Registry', () => {
  beforeEach(() => {
    // Clear registry before each test
    visualizationRegistry.getIds().forEach(id => {
      visualizationRegistry.unregister(id);
    });
  });

  it('should register and retrieve visualizations', () => {
    const mockViz = {
      id: 'test',
      name: 'Test Visualization',
      description: 'A test visualization',
      category: 'spatial' as const,
      icon: '🧪',
      component: null as any,
      layoutConstraints: {},
      capabilities: {}
    };

    visualizationRegistry.register(mockViz);
    
    expect(visualizationRegistry.has('test')).toBe(true);
    expect(visualizationRegistry.get('test')).toEqual(mockViz);
    expect(visualizationRegistry.getAll()).toHaveLength(1);
  });

  it('should group visualizations by category', () => {
    const spatialViz = {
      id: 'spatial1',
      name: 'Spatial Viz',
      description: 'Spatial',
      category: 'spatial' as const,
      icon: '🎯',
      component: null as any,
      layoutConstraints: {},
      capabilities: {}
    };

    const linearViz = {
      id: 'linear1', 
      name: 'Linear Viz',
      description: 'Linear',
      category: 'linear' as const,
      icon: '📊',
      component: null as any,
      layoutConstraints: {},
      capabilities: {}
    };

    visualizationRegistry.register(spatialViz);
    visualizationRegistry.register(linearViz);

    const spatialVizs = visualizationRegistry.getByCategory('spatial');
    const linearVizs = visualizationRegistry.getByCategory('linear');

    expect(spatialVizs).toHaveLength(1);
    expect(linearVizs).toHaveLength(1);
    expect(spatialVizs[0].id).toBe('spatial1');
    expect(linearVizs[0].id).toBe('linear1');
  });

  it('should handle unregistering visualizations', () => {
    const mockViz = {
      id: 'test',
      name: 'Test',
      description: 'Test',
      category: 'spatial' as const,
      icon: '🧪',
      component: null as any,
      layoutConstraints: {},
      capabilities: {}
    };

    visualizationRegistry.register(mockViz);
    expect(visualizationRegistry.has('test')).toBe(true);

    visualizationRegistry.unregister('test');
    expect(visualizationRegistry.has('test')).toBe(false);
    expect(visualizationRegistry.get('test')).toBeUndefined();
  });
});