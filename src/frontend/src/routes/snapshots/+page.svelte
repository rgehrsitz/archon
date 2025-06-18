<!-- snapshots/+page.svelte -->
<script lang="ts">
  import { onMount } from "svelte";
  import type { Snapshot } from "../../lib/types/wails.d.ts";
  import {
    GetSnapshots,
    CreateSnapshot,
  } from "../../../wailsjs/go/main/App.js";

  let snapshots = $state<Snapshot[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let creating = $state(false);
  let newSnapshotMessage = $state("");

  async function loadSnapshots() {
    loading = true;
    error = null;
    try {
      snapshots = await GetSnapshots();
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : "Failed to load snapshots";
    } finally {
      loading = false;
    }
  }

  async function createSnapshot() {
    if (!newSnapshotMessage.trim()) return;

    creating = true;
    try {
      await CreateSnapshot(newSnapshotMessage);
      newSnapshotMessage = "";
      await loadSnapshots();
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : "Failed to create snapshot";
    } finally {
      creating = false;
    }
  }

  onMount(loadSnapshots);

  async function handleDeleteSnapshot(id: string) {
    if (!confirm("Are you sure you want to delete this snapshot?")) return;
    try {
      // Mock deletion logic
      snapshots = snapshots.filter((snapshot) => snapshot.id !== id);
    } catch (error) {
      console.error("Failed to delete snapshot:", error);
    }
  }
</script>

<div class="space-y-6">
  <div class="bg-white rounded-lg shadow">
    <div class="p-4 border-b border-slate-200">
      <h1 class="text-xl font-semibold text-slate-800">Snapshot Management</h1>
      <p class="text-sm text-slate-600 mt-1">
        Create and manage configuration snapshots
      </p>
    </div>

    <div class="p-4">
      <form class="flex gap-4 mb-6" onsubmit={createSnapshot}>
        <input
          type="text"
          placeholder="Snapshot description..."
          class="flex-1 rounded border-slate-300 bg-white text-slate-900"
          bind:value={newSnapshotMessage}
          required
        />
        <button
          type="submit"
          class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
          disabled={creating || !newSnapshotMessage.trim()}
        >
          {creating ? "Creating..." : "Create Snapshot"}
        </button>
      </form>

      {#if loading}
        <div class="text-slate-500">Loading snapshots...</div>
      {:else if error}
        <div class="text-red-500">Error: {error}</div>
      {:else if snapshots.length === 0}
        <div class="text-slate-500">
          No snapshots found. Create your first snapshot above.
        </div>
      {:else}
        <div class="space-y-3">
          {#each snapshots as snapshot (snapshot.id)}
            <div
              class="border border-slate-200 rounded-lg p-4 hover:bg-slate-50"
            >
              <div class="flex items-start justify-between">
                <div>
                  <h3 class="font-medium text-slate-900">{snapshot.name}</h3>
                  {#if snapshot.description}
                    <p class="text-sm text-slate-600 mt-1">
                      {snapshot.description}
                    </p>
                  {/if}
                  <div
                    class="flex items-center space-x-4 mt-2 text-xs text-slate-500"
                  >
                    <span>ID: {snapshot.id}</span>
                    <span>Author: {snapshot.author}</span>
                    <span
                      >Created: {new Date(
                        snapshot.timestamp
                      ).toLocaleString()}</span
                    >
                  </div>
                </div>
                <div class="flex space-x-2">
                  <button
                    class="px-3 py-1 text-sm bg-slate-100 text-slate-700 rounded hover:bg-slate-200"
                    onclick={() => console.log("View diff for", snapshot.id)}
                  >
                    View Diff
                  </button>
                  <button
                    class="px-3 py-1 text-sm bg-indigo-100 text-indigo-700 rounded hover:bg-indigo-200"
                    onclick={() => console.log("Restore", snapshot.id)}
                  >
                    Restore
                  </button>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>
