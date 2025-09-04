<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Button } from '../ui/button/index.js';
  import { visualizationRegistry } from '$lib/services/visualizationRegistry.js';
  import type { VisualizationId } from '$lib/types/visualization.js';

  export let currentViewMode: VisualizationId = 'miller';

  const dispatch = createEventDispatcher<{
    change: { mode: VisualizationId };
  }>();

  // Get all available visualizations
  $: visualizations = visualizationRegistry.getAll();
  $: groupedVisualizations = {
    linear: visualizations.filter(v => v.category === 'linear'),
    spatial: visualizations.filter(v => v.category === 'spatial'),
    network: visualizations.filter(v => v.category === 'network')
  };

  let isOpen = false;

  function handleVisualizationSelect(mode: VisualizationId) {
    dispatch('change', { mode });
    isOpen = false;
  }

  function toggleDropdown() {
    isOpen = !isOpen;
  }

  // Close dropdown when clicking outside
  function handleClickOutside(event: MouseEvent) {
    const target = event.target as Element;
    if (!target.closest('.visualization-switcher')) {
      isOpen = false;
    }
  }

  $: currentVisualization = visualizationRegistry.get(currentViewMode);
</script>

<svelte:window onclick={handleClickOutside} />

<div class="visualization-switcher relative">
  <Button
    variant="outline"
    size="sm"
    onclick={toggleDropdown}
    class="gap-2 min-w-32"
  >
    <span class="text-base">{currentVisualization?.icon || '📊'}</span>
    <span class="text-sm font-medium">{currentVisualization?.name || 'Unknown'}</span>
    <span class="text-xs opacity-60">{isOpen ? '▲' : '▼'}</span>
  </Button>

  {#if isOpen}
    <div class="absolute top-full left-0 mt-1 z-50 bg-background border border-border rounded-lg shadow-lg min-w-80">
      <div class="p-3">
        <div class="text-xs font-semibold text-muted-foreground mb-3">Choose Visualization</div>
        
        <!-- Linear Visualizations -->
        {#if groupedVisualizations.linear.length > 0}
          <div class="mb-4">
            <div class="text-xs font-medium text-foreground mb-2">📊 Navigation Views</div>
            <div class="grid gap-1">
              {#each groupedVisualizations.linear as viz}
                <button
                  class="flex items-center gap-3 w-full p-2 rounded hover:bg-accent hover:text-accent-foreground text-left transition-colors {
                    currentViewMode === viz.id ? 'bg-accent text-accent-foreground' : ''
                  }"
                  onclick={() => handleVisualizationSelect(viz.id)}
                >
                  <span class="text-lg">{viz.icon}</span>
                  <div class="flex-1">
                    <div class="text-sm font-medium">{viz.name}</div>
                    <div class="text-xs text-muted-foreground line-clamp-2">{viz.description}</div>
                  </div>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Spatial Visualizations -->
        {#if groupedVisualizations.spatial.length > 0}
          <div class="mb-4">
            <div class="text-xs font-medium text-foreground mb-2">🎯 Spatial Views</div>
            <div class="grid gap-1">
              {#each groupedVisualizations.spatial as viz}
                <button
                  class="flex items-center gap-3 w-full p-2 rounded hover:bg-accent hover:text-accent-foreground text-left transition-colors {
                    currentViewMode === viz.id ? 'bg-accent text-accent-foreground' : ''
                  }"
                  onclick={() => handleVisualizationSelect(viz.id)}
                >
                  <span class="text-lg">{viz.icon}</span>
                  <div class="flex-1">
                    <div class="text-sm font-medium">{viz.name}</div>
                    <div class="text-xs text-muted-foreground line-clamp-2">{viz.description}</div>
                  </div>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Network Visualizations -->
        {#if groupedVisualizations.network.length > 0}
          <div class="mb-2">
            <div class="text-xs font-medium text-foreground mb-2">🌐 Network Views</div>
            <div class="grid gap-1">
              {#each groupedVisualizations.network as viz}
                <button
                  class="flex items-center gap-3 w-full p-2 rounded hover:bg-accent hover:text-accent-foreground text-left transition-colors {
                    currentViewMode === viz.id ? 'bg-accent text-accent-foreground' : ''
                  }"
                  onclick={() => handleVisualizationSelect(viz.id)}
                >
                  <span class="text-lg">{viz.icon}</span>
                  <div class="flex-1">
                    <div class="text-sm font-medium">{viz.name}</div>
                    <div class="text-xs text-muted-foreground line-clamp-2">{viz.description}</div>
                  </div>
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Footer info -->
        <div class="border-t border-border pt-2 mt-3">
          <div class="text-xs text-muted-foreground text-center">
            {visualizations.length} visualizations available
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>