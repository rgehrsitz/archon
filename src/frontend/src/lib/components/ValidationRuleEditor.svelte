<!-- ValidationRuleEditor.svelte -->
<script lang="ts">
  import { createEventDispatcher } from "svelte";

  let {
    rule,
    availableProperties,
  }: {
    rule: {
      property: string;
      required: boolean;
      type?: string;
      pattern?: string;
      min?: number;
      max?: number;
      enum?: string[];
      dependencies?: string[];
    };
    availableProperties: string[];
  } = $props();

  const dispatch = createEventDispatcher();

  const propertyTypes = [
    { value: "string", label: "String" },
    { value: "number", label: "Number" },
    { value: "boolean", label: "Boolean" },
    { value: "date", label: "Date" },
  ];

  $effect(() => {
    dispatch("change", { rule });
  });
  function addEnumValue() {
    rule.enum = [...(rule.enum || []), ""];
  }

  function removeEnumValue(index: number) {
    rule.enum = (rule.enum || []).filter((_: any, i: number) => i !== index);
  }

  function addDependency() {
    rule.dependencies = [...(rule.dependencies || []), ""];
  }

  function removeDependency(index: number) {
    rule.dependencies = (rule.dependencies || []).filter(
      (_: any, i: number) => i !== index
    );
  }
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4 space-y-4">
  <div class="grid grid-cols-2 gap-4">
    <!-- Property Name -->
    <div>
      <label
        for="property-select"
        class="block text-sm font-medium text-gray-700 dark:text-gray-300"
      >
        Property Name
      </label>
      <select
        id="property-select"
        bind:value={rule.property}
        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
      >
        <option value="">Select a property</option>
        {#each availableProperties as prop}
          <option value={prop}>{prop}</option>
        {/each}
      </select>
    </div>
    <!-- Required -->
    <div class="flex items-center">
      <input
        id="required-checkbox"
        type="checkbox"
        bind:checked={rule.required}
        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
      />
      <label
        for="required-checkbox"
        class="ml-2 block text-sm text-gray-700 dark:text-gray-300"
      >
        Required
      </label>
    </div>
  </div>
  <!-- Type -->
  <div>
    <label
      for="type-select"
      class="block text-sm font-medium text-gray-700 dark:text-gray-300"
    >
      Type
    </label>
    <select
      id="type-select"
      bind:value={rule.type}
      class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
    >
      <option value="">Select a type</option>
      {#each propertyTypes as type}
        <option value={type.value}>{type.label}</option>
      {/each}
    </select>
  </div>

  {#if rule.type === "string"}
    <!-- Pattern -->
    <div>
      <label
        for="pattern-input"
        class="block text-sm font-medium text-gray-700 dark:text-gray-300"
      >
        Pattern (Regex)
      </label>
      <input
        id="pattern-input"
        type="text"
        bind:value={rule.pattern}
        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
        placeholder="^[A-Za-z0-9]+$"
      />
    </div>
    <!-- Enum Values -->
    <fieldset>
      <legend
        class="block text-sm font-medium text-gray-700 dark:text-gray-300"
      >
        Allowed Values
      </legend>
      <div class="mt-1 space-y-2">
        {#if rule.enum}
          {#each rule.enum as value, i}
            <div class="flex gap-2">
              <input
                type="text"
                bind:value={rule.enum[i]}
                class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
              />
              <button
                type="button"
                onclick={() => removeEnumValue(i)}
                class="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              >
                Remove
              </button>
            </div>
          {/each}
        {/if}
        <button
          type="button"
          onclick={addEnumValue}
          class="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          Add Value
        </button>
      </div>
    </fieldset>
  {/if}

  {#if rule.type === "number"}
    <!-- Min/Max -->
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label
          for="min-input"
          class="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          Minimum
        </label>
        <input
          id="min-input"
          type="number"
          bind:value={rule.min}
          class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
        />
      </div>
      <div>
        <label
          for="max-input"
          class="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          Maximum
        </label>
        <input
          id="max-input"
          type="number"
          bind:value={rule.max}
          class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
        />
      </div>
    </div>
  {/if}
  <!-- Dependencies -->
  <fieldset>
    <legend class="block text-sm font-medium text-gray-700 dark:text-gray-300">
      Dependencies
    </legend>
    <div class="mt-1 space-y-2">
      {#if rule.dependencies}
        {#each rule.dependencies as dep, i}
          <div class="flex gap-2">
            <select
              bind:value={rule.dependencies[i]}
              class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-gray-700 dark:border-gray-600"
            >
              <option value="">Select a dependency</option>
              {#each availableProperties as prop}
                <option value={prop}>{prop}</option>
              {/each}
            </select>
            <button
              type="button"
              onclick={() => removeDependency(i)}
              class="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
            >
              Remove
            </button>
          </div>
        {/each}
      {/if}
      <button
        type="button"
        onclick={addDependency}
        class="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-indigo-700 bg-indigo-100 hover:bg-indigo-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
      >
        Add Dependency
      </button>
    </div>
  </fieldset>
</div>
