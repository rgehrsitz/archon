<script lang="ts">
  import Layout from "./lib/components/Layout.svelte";
  import ComponentTree from "./lib/components/ComponentTree.svelte";

  let selectedId = $state<string | null>(null);
  let selectedComponent = $state<any>(null);
  let loadingDetails = $state(false);
  let errorDetails = $state<string | null>(null);
  async function handleSelect(event: { id: string }) {
    selectedId = event.id;
    selectedComponent = null;
    errorDetails = null;
    if (selectedId) {
      loadingDetails = true;
      try {
        const { GetComponent } = await import("../wailsjs/go/main/App.js");
        selectedComponent = await GetComponent(selectedId);
      } catch (e: unknown) {
        errorDetails =
          e instanceof Error ? e.message : "Failed to load details";
      } finally {
        loadingDetails = false;
      }
    }
  }
</script>

<Layout>
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <!-- Component Tree -->
    <div class="lg:col-span-1">
      <ComponentTree onselect={handleSelect} />
    </div>

    <!-- Component Details -->
    <div class="lg:col-span-2">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div class="p-4 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-800 dark:text-white">
            {#if selectedId}
              Component Details
            {:else}
              Select a Component
            {/if}
          </h2>
        </div>

        <div class="p-4">
          {#if selectedId}
            {#if loadingDetails}
              <div class="text-gray-500 dark:text-gray-400">Loading...</div>
            {:else if errorDetails}
              <div class="text-red-500">{errorDetails}</div>
            {:else if selectedComponent}
              <dl class="grid grid-cols-1 gap-4">
                <div>
                  <dt
                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                  >
                    Name
                  </dt>
                  <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                    {selectedComponent.name}
                  </dd>
                </div>
                <div>
                  <dt
                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                  >
                    Type
                  </dt>
                  <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                    {selectedComponent.type}
                  </dd>
                </div>
                <div>
                  <dt
                    class="text-sm font-medium text-gray-500 dark:text-gray-400"
                  >
                    ID
                  </dt>
                  <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                    {selectedComponent.id}
                  </dd>
                </div>
                {#if selectedComponent.description}
                  <div>
                    <dt
                      class="text-sm font-medium text-gray-500 dark:text-gray-400"
                    >
                      Description
                    </dt>
                    <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                      {selectedComponent.description}
                    </dd>
                  </div>
                {/if}
                {#if selectedComponent.parentId}
                  <div>
                    <dt
                      class="text-sm font-medium text-gray-500 dark:text-gray-400"
                    >
                      Parent ID
                    </dt>
                    <dd class="mt-1 text-sm text-gray-900 dark:text-white">
                      {selectedComponent.parentId}
                    </dd>
                  </div>
                {/if}
              </dl>
            {/if}
          {:else}
            <p class="text-gray-500 dark:text-gray-400">
              Select a component to view its details.
            </p>
          {/if}
        </div>
      </div>
    </div>
  </div>
</Layout>

<style>
</style>
