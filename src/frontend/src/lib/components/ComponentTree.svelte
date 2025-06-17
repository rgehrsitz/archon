<!-- ComponentTree.svelte -->
<script lang="ts">
  import { onMount } from "svelte";

  // Proper Svelte 5 props interface
  interface Props {
    onselect?: (event: { id: string }) => void;
  }

  let { onselect }: Props = $props();

  // Use $state for reactive state management
  let expandedNodes = $state(new Set<string>());
  let components = $state<
    Array<{ id: string; name: string; type: string; parentId: string | null }>
  >([]);
  let loading = $state(true);
  let error = $state("");

  // New component form state
  let showNewComponentForm = $state(false);
  let newComponent = $state({
    name: "",
    type: "",
    parentId: null as string | null,
  });
  let creating = $state(false);
  let createError = $state("");
  async function fetchComponents() {
    loading = true;
    error = "";
    try {
      // Import the Wails bindings with .js extension
      const { GetComponentTree } = await import(
        "../../../wailsjs/go/main/App.js"
      );
      const response = await GetComponentTree();
      // The response should have a Components map and RootIDs array
      if (response && response.Components) {
        // Convert the Components map to an array for easier handling
        const componentMap = response.Components;
        components = Object.values(componentMap).map((comp: any) => ({
          id: comp.id,
          name: comp.name,
          type: comp.type,
          parentId: comp.parentId || null,
        }));
      } else {
        components = [];
      }
      loading = false;
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : "Failed to load components";
      loading = false;
    }
  }

  onMount(fetchComponents);

  function toggleNode(id: string) {
    if (expandedNodes.has(id)) {
      expandedNodes.delete(id);
    } else {
      expandedNodes.add(id);
    }
    // Trigger reactivity by reassigning
    expandedNodes = new Set(expandedNodes);
  }

  function getChildComponents(parentId: string | null) {
    return components.filter((c) => c.parentId === parentId);
  }

  function handleSelect(componentId: string) {
    onselect?.({ id: componentId });
  }

  function handleDragStart(event: DragEvent, componentId: string) {
    if (event.dataTransfer) {
      event.dataTransfer.setData("text/plain", componentId);
    }
  }

  function handleDrop(event: DragEvent, targetId: string) {
    event.preventDefault();
    const sourceId = event.dataTransfer?.getData("text/plain");
    if (sourceId && sourceId !== targetId) {
      // Handle drop logic here - will integrate with backend
      console.log("Dropped", sourceId, "onto", targetId);
    }
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
  }
  async function createComponent() {
    creating = true;
    createError = "";
    try {
      // Import the Wails bindings with .js extension
      const { CreateComponentSimple } = await import(
        "../../../wailsjs/go/main/App.js"
      );
      await CreateComponentSimple(
        newComponent.name,
        newComponent.type,
        newComponent.parentId || ""
      );
      showNewComponentForm = false;
      newComponent = { name: "", type: "", parentId: null };
      await fetchComponents();
    } catch (e: unknown) {
      createError =
        e instanceof Error ? e.message : "Failed to create component";
    } finally {
      creating = false;
    }
  }
</script>

<div class="bg-white rounded-lg shadow">
  <div class="p-4 border-b border-slate-200 flex items-center justify-between">
    <h2 class="text-lg font-semibold text-slate-800">Component Hierarchy</h2>
    <button
      class="px-3 py-1 rounded bg-indigo-600 text-white hover:bg-indigo-700 text-sm font-medium"
      onclick={() => (showNewComponentForm = !showNewComponentForm)}
    >
      {showNewComponentForm ? "Cancel" : "New Component"}
    </button>
  </div>

  {#if showNewComponentForm}
    <form
      class="p-4 space-y-3 border-b border-slate-200"
      onsubmit={createComponent}
    >
      <div>
        <label
          for="componentName"
          class="block text-sm font-medium text-slate-700">Name</label
        >
        <input
          type="text"
          id="componentName"
          class="mt-1 block w-full rounded border-slate-300 bg-white text-slate-900"
          bind:value={newComponent.name}
          required
        />
      </div>
      <div>
        <label
          for="componentType"
          class="block text-sm font-medium text-slate-700">Type</label
        >
        <input
          type="text"
          id="componentType"
          class="mt-1 block w-full rounded border-slate-300 bg-white text-slate-900"
          bind:value={newComponent.type}
          required
        />
      </div>
      <div>
        <label
          for="componentParent"
          class="block text-sm font-medium text-slate-700">Parent</label
        >
        <select
          id="componentParent"
          class="mt-1 block w-full rounded border-slate-300 bg-white text-slate-900"
          bind:value={newComponent.parentId}
        >
          <option value={null}>None (root)</option>
          {#each components as c}
            <option value={c.id}>{c.name} ({c.type})</option>
          {/each}
        </select>
      </div>
      {#if createError}
        <div class="text-red-500 text-sm">{createError}</div>
      {/if}
      <div class="flex gap-2">
        <button
          type="submit"
          class="px-3 py-1 rounded bg-indigo-600 text-white hover:bg-indigo-700 text-sm font-medium"
          disabled={creating}
        >
          {creating ? "Creating..." : "Create"}
        </button>
        <button
          type="button"
          class="px-3 py-1 rounded bg-slate-200 text-slate-700 text-sm font-medium"
          onclick={() => (showNewComponentForm = false)}
        >
          Cancel
        </button>
      </div>
    </form>
  {/if}

  <div class="p-4">
    {#if loading}
      <div class="text-slate-500">Loading components...</div>
    {:else if error}
      <div class="text-red-500">Error: {error}</div>
    {:else}
      <div class="space-y-2">
        {#each getChildComponents(null) as component (component.id)}
          <div class="component-tree">
            <div class="flex items-center py-1">
              <button
                class="p-1 rounded hover:bg-slate-100"
                onclick={() => toggleNode(component.id)}
                aria-label="Toggle component {component.name}"
              >
                <svg
                  class="w-4 h-4 transform transition-transform"
                  class:rotate-90={expandedNodes.has(component.id)}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </button>

              <button
                class="flex-1 px-2 py-1 text-left rounded hover:bg-slate-100"
                draggable="true"
                ondragstart={(e) => handleDragStart(e, component.id)}
                ondrop={(e) => handleDrop(e, component.id)}
                ondragover={handleDragOver}
                onclick={() => handleSelect(component.id)}
                aria-label="Select component {component.name}"
              >
                <span class="font-medium">{component.name}</span>
                <span class="ml-2 text-sm text-slate-500"
                  >({component.type})</span
                >
              </button>
            </div>

            {#if expandedNodes.has(component.id)}
              <div class="ml-6">
                {#each getChildComponents(component.id) as child (child.id)}
                  <div class="flex items-center py-1">
                    <button
                      class="p-1 rounded hover:bg-slate-100"
                      onclick={() => toggleNode(child.id)}
                      aria-label="Toggle component {child.name}"
                    >
                      <svg
                        class="w-4 h-4 transform transition-transform"
                        class:rotate-90={expandedNodes.has(child.id)}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M9 5l7 7-7 7"
                        />
                      </svg>
                    </button>

                    <button
                      class="flex-1 px-2 py-1 text-left rounded hover:bg-slate-100"
                      draggable="true"
                      ondragstart={(e) => handleDragStart(e, child.id)}
                      ondrop={(e) => handleDrop(e, child.id)}
                      ondragover={handleDragOver}
                      onclick={() => handleSelect(child.id)}
                      aria-label="Select component {child.name}"
                    >
                      <span class="font-medium">{child.name}</span>
                      <span class="ml-2 text-sm text-slate-500"
                        >({child.type})</span
                      >
                    </button>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .component-tree {
    color: #64748b;
  }
</style>
