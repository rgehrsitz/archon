<!-- snapshots/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';

  let newTag = '';
  let newMessage = '';
  let snapshots: Array<{id: string, tag: string, message: string, timestamp: string, components: number}> = [];
  let loading = true;
  let error = '';

  onMount(async () => {
    // Mock data for now
    snapshots = [
      { 
        id: '1', 
        tag: 'v1.0.0', 
        message: 'Initial release', 
        timestamp: new Date().toISOString(),
        components: 5
      },
      { 
        id: '2', 
        tag: 'v1.1.0', 
        message: 'Added new features', 
        timestamp: new Date(Date.now() - 86400000).toISOString(),
        components: 7
      },
    ];
    loading = false;
  });

  async function handleCreateSnapshot() {
    if (!newTag || !newMessage) return;
    try {
      // Mock creation logic
      const newSnapshot = {
        id: Date.now().toString(),
        tag: newTag,
        message: newMessage,
        timestamp: new Date().toISOString(),
        components: Math.floor(Math.random() * 10) + 1
      };
      snapshots = [newSnapshot, ...snapshots];
      newTag = '';
      newMessage = '';
    } catch (error) {
      console.error('Failed to create snapshot:', error);
    }
  }

  async function handleDeleteSnapshot(id: string) {
    if (!confirm('Are you sure you want to delete this snapshot?')) return;
    try {
      // Mock deletion logic
      snapshots = snapshots.filter(snapshot => snapshot.id !== id);
    } catch (error) {
      console.error('Failed to delete snapshot:', error);
    }
  }

  async function handleRestoreSnapshot(id: string) {
    if (!confirm('Are you sure you want to restore this snapshot? This will overwrite current state.')) return;
    try {
      // Mock restore logic
      console.log('Restoring snapshot:', id);
    } catch (error) {
      console.error('Failed to restore snapshot:', error);
    }
  }
</script>

<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-6">Snapshots</h1>

  <!-- Create Snapshot Form -->
  <div class="bg-white rounded-lg shadow p-6 mb-6">
    <h2 class="text-lg font-semibold mb-4">Create New Snapshot</h2>
    <form onsubmit={handleCreateSnapshot} class="space-y-4">
      <div>
        <label for="tag" class="block text-sm font-medium text-slate-600">Tag</label>
        <input
          type="text"
          id="tag"
          bind:value={newTag}
          class="mt-1 block w-full rounded-md border-slate-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          placeholder="e.g., v1.0.0"
          required
        />
      </div>
      <div>
        <label for="message" class="block text-sm font-medium text-slate-600">Message</label>
        <textarea
          id="message"
          bind:value={newMessage}
          class="mt-1 block w-full rounded-md border-slate-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
          rows="3"
          placeholder="Describe the changes in this snapshot"
          required
        ></textarea>
      </div>
      <button
        type="submit"
        class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
      >
        Create Snapshot
      </button>
    </form>
  </div>

  <!-- Snapshots List -->
  <div class="bg-white rounded-lg shadow">
    <div class="px-6 py-4 border-b border-slate-200">
      <h2 class="text-lg font-semibold">Snapshots History</h2>
    </div>
    {#if loading}
      <div class="p-4 text-slate-500">Loading snapshots...</div>
    {:else if error}
      <div class="p-4 text-red-500">Error: {error}</div>
    {:else if snapshots.length === 0}
      <div class="p-4 text-slate-500">No snapshots found</div>
    {:else}
      <div class="divide-y divide-slate-200">
        {#each snapshots as snapshot (snapshot.id)}
          <div class="p-4 hover:bg-slate-50">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-lg font-medium text-slate-900">{snapshot.tag}</h3>
                <p class="text-sm text-slate-500">{snapshot.message}</p>
                <div class="mt-1 text-xs text-slate-400">
                  Created on {new Date(snapshot.timestamp).toLocaleString()} • {snapshot.components} components
                </div>
              </div>
              <div class="flex space-x-2">
                <button
                  onclick={() => handleRestoreSnapshot(snapshot.id)}
                  class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-blue-700 bg-blue-100 hover:bg-blue-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Restore
                </button>
                <button
                  onclick={() => handleDeleteSnapshot(snapshot.id)}
                  class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>