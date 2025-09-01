<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Badge } from '$lib/components/ui/badge';
  
  export let projectId: string;
  export let viewMode: 'miller' | 'tree' = 'miller';
  
  const dispatch = createEventDispatcher<{
    viewModeChange: { mode: 'miller' | 'tree' };
    search: { query: string };
    snapshot: void;
    sync: void;
  }>();
  
  let searchQuery = '';
  let breadcrumbs = [
    { id: 'root', name: 'Project Root' },
    { id: 'lab-a', name: 'Lab A' },
    { id: 'bench-3', name: 'Bench 3' }
  ];
  
  function toggleViewMode() {
    const newMode = viewMode === 'miller' ? 'tree' : 'miller';
    dispatch('viewModeChange', { mode: newMode });
  }
  
  function handleSearch() {
    if (searchQuery.trim()) {
      dispatch('search', { query: searchQuery });
    }
  }
  
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      handleSearch();
    } else if (event.key === '/' && event.target !== event.currentTarget?.querySelector('input')) {
      event.preventDefault();
      event.currentTarget?.querySelector('input')?.focus();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

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
      on:input={handleSearch}
    />
    <kbd class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground pointer-events-none">
      /
    </kbd>
  </div>
  
  <!-- View Mode Toggle -->
  <Button 
    variant="outline" 
    size="sm"
    onclick={toggleViewMode}
    class="w-24"
  >
    {#snippet children()}
      {viewMode === 'miller' ? '📋 Miller' : '🌳 Tree'}
    {/snippet}
  </Button>
  
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