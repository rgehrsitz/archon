<!-- ValidationRuleSet.svelte -->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ValidationRuleEditor from './ValidationRuleEditor.svelte';
  
  const { rules, availableProperties } = $props<{
    rules: {
      property: string;
      required: boolean;
      type?: string;
      pattern?: string;
      min?: number;
      max?: number;
      enum?: string[];
      dependencies?: string[];
    }[];
    availableProperties: string[];
  }>();
  
  const dispatch = createEventDispatcher();
  
  function addRule() {
    rules.push({
      property: '',
      required: false
    });
  }
  
  function removeRule(index: number) {
    rules.splice(index, 1);
  }
  
  $effect(() => {
    dispatch('change', { rules });
  });
</script>

<div class="space-y-4">
  <div class="flex justify-between items-center">
    <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
      Validation Rules
    </h3>
    <button
      type="button"
      on:click={addRule}
      class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
    >
      Add Rule
    </button>
  </div>

  <div class="space-y-4">
    {#each rules as rule, i}
      <div class="relative">
        <ValidationRuleEditor
          {rule}
          {availableProperties}
        />
        <button
          type="button"
          on:click={() => removeRule(i)}
          class="absolute top-2 right-2 inline-flex items-center p-1.5 border border-transparent text-xs font-medium rounded-full text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    {/each}
  </div>
</div> 