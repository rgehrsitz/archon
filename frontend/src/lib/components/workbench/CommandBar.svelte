<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Button } from '../ui/button/index.js';
  import { Input } from '../ui/input/index.js';
  import VisualizationSwitcher from './VisualizationSwitcher.svelte';
  import type { VisualizationId } from '$lib/types/visualization.js';
  
  export const projectId: string = undefined!;
  export let viewMode: VisualizationId = 'miller';
  export let nodePath: any[] = [];
  
  const dispatch = createEventDispatcher<{
    viewModeChange: { mode: VisualizationId };
    search: { query: string };
    snapshot: void;
    sync: void;
    breadcrumbNavigate: { index: number };
  }>();
  
  let searchQuery = '';
  type Crumb = { id: string; name: string; index: number };
  let breadcrumbs: Crumb[] = [];

  $: breadcrumbs = [
    { id: 'root', name: 'ROOT', index: -1 },
    ...nodePath.map((n, i) => ({ id: n.id, name: n.name || 'Untitled', index: i }))
  ];
  
  function handleVisualizationChange(event: CustomEvent) {
    dispatch('viewModeChange', { mode: event.detail.mode });
  }
  
  function handleSearch() {
    if (searchQuery.trim()) {
      dispatch('search', { query: searchQuery });
    }
  }

  function handleCrumbClick(crumb: Crumb) {
    dispatch('breadcrumbNavigate', { index: crumb.index });
  }
  
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleSearch();
    } else if (event.key === '/' && event.target !== (event.currentTarget as Document)?.querySelector('input')) {
      event.preventDefault();
      (event.currentTarget as Document)?.querySelector('input')?.focus();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="h-12 border-b bg-muted/30 px-4 flex items-center gap-4">
  <!-- Breadcrumbs -->
  <nav class="flex items-center gap-1 text-sm min-w-0 flex-1">
    {#each breadcrumbs as crumb, i (crumb.id || i)}
      {#if i > 0}
        <span class="text-muted-foreground">/</span>
      {/if}
      <button 
        class="hover:text-primary truncate max-w-32"
        title={crumb.name}
        on:click={() => handleCrumbClick(crumb)}
      >
        {crumb.name}
      </button>
    {/each}
  </nav>
  
  <!-- Search -->
  <div class="relative">
    <Input
      bind:value={searchQuery}
      placeholder="Search nodes... (Press / to focus)"
      class="w-64 pr-8"
      oninput={handleSearch}
    />
    <kbd class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground pointer-events-none">
      /
    </kbd>
  </div>
  
  <!-- Visualization Switcher -->
  <VisualizationSwitcher
    currentViewMode={viewMode}
    on:change={handleVisualizationChange}
  />
  
  <!-- Actions -->
  <div class="flex items-center gap-2">
    <Button
      variant="outline"
      size="sm"
      onclick={() => dispatch('snapshot')}
    >
      {#snippet children()}
        📸 Snapshot
      {/snippet}
    </Button>
    
    <Button
      variant="outline" 
      size="sm"
      onclick={() => dispatch('sync')}
    >
      {#snippet children()}
        🔄 Sync
      {/snippet}
    </Button>
  </div>
</div>