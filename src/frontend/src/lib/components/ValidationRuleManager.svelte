<!-- ValidationRuleManager.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import ValidationRuleSet from './ValidationRuleSet.svelte';
  
  const { componentType, isGeneric } = $props<{
    componentType: string;
    isGeneric: boolean;
  }>();
  
  let rules = $state<{
    property: string;
    required: boolean;
    type?: string;
    pattern?: string;
    min?: number;
    max?: number;
    enum?: string[];
    dependencies?: string[];
  }[]>([]);
  
  let availableProperties = $state<string[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  
  onMount(async () => {
    try {
      // Fetch available properties for this component type
      const response = await fetch(`/api/component-types/${componentType}/properties`);
      if (!response.ok) throw new Error('Failed to fetch properties');
      availableProperties = await response.json();
      
      // Fetch existing rules
      const rulesResponse = await fetch(`/api/validation-rules/${componentType}`);
      if (!rulesResponse.ok) throw new Error('Failed to fetch rules');
      rules = await rulesResponse.json();
      
      loading = false;
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : 'An unknown error occurred';
      loading = false;
    }
  });
  
  async function saveRules() {
    try {
      loading = true;
      const response = await fetch(`/api/validation-rules/${componentType}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          rules,
          isGeneric
        }),
      });
      
      if (!response.ok) throw new Error('Failed to save rules');
      
      loading = false;
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : 'An unknown error occurred';
      loading = false;
    }
  }
  
  function handleRulesChange(event: CustomEvent) {
    rules = event.detail.rules;
  }
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
  <div class="bg-white dark:bg-gray-800 shadow rounded-lg">
    <div class="px-4 py-5 sm:p-6">
      <div class="flex justify-between items-center mb-6">
        <div>
          <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {isGeneric ? 'Generic' : 'Specific'} Validation Rules
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {componentType}
          </p>
        </div>
        <button
          type="button"
          on:click={saveRules}
          disabled={loading}
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
        >
          {#if loading}
            <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Saving...
          {:else}
            Save Rules
          {/if}
        </button>
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
        <ValidationRuleSet
          {rules}
          {availableProperties}
          on:change={handleRulesChange}
        />
      {/if}
    </div>
  </div>
</div> 