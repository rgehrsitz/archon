<!-- +page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import ValidationRuleManager from '$lib/components/ValidationRuleManager.svelte';
  
  let componentTypes = $state<string[]>([]);
  let selectedType = $state<string | null>(null);
  let isGeneric = $state(false);
  let loading = $state(true);
  let error = $state<string | null>(null);
  
  onMount(async () => {
    try {
      const response = await fetch('/api/component-types');
      if (!response.ok) throw new Error('Failed to fetch component types');
      componentTypes = await response.json();
      loading = false;
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : 'An unknown error occurred';
      loading = false;
    }
  });
</script>

<div class="min-h-screen bg-gray-100 dark:bg-gray-900">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="bg-white dark:bg-gray-800 shadow rounded-lg">
      <div class="px-4 py-5 sm:p-6">
        <div class="flex justify-between items-center mb-6">
          <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
            Validation Rules
          </h1>
        </div>

        {#if error}
          <div class="rounded-md bg-red-50 dark:bg-red-900 p-4 mb-6">
            <div class="flex">
              <div class="flex-shrink-0">
                <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                </svg>
              </div>
              <div class="ml-3">
                <h3 class="text-sm font-medium text-red-800 dark:text-red-200">
                  Error
                </h3>
                <div class="mt-2 text-sm text-red-700 dark:text-red-300">
                  {error}
                </div>
              </div>
            </div>
          </div>
        {/if}

        {#if loading}
          <div class="flex justify-center items-center py-12">
            <svg class="animate-spin h-8 w-8 text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>
        {:else}
          <div class="space-y-6">
            <!-- Component Type Selection -->
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Component Type
              </label>
              <select
                bind:value={selectedType}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
              >
                <option value="">Select a component type</option>
                {#each componentTypes as type}
                  <option value={type}>{type}</option>
                {/each}
              </select>
            </div>

            <!-- Generic/Specific Toggle -->
            <div class="flex items-center">
              <input
                type="checkbox"
                bind:checked={isGeneric}
                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <label class="ml-2 block text-sm text-gray-700 dark:text-gray-300">
                Generic Rules
              </label>
            </div>

            {#if selectedType}
              <ValidationRuleManager
                componentType={selectedType}
                {isGeneric}
              />
            {/if}
          </div>
        {/if}
      </div>
    </div>
  </div>
</div> 