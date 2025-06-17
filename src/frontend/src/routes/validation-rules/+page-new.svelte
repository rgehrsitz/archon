<!-- +page.svelte -->
<script lang="ts">
  import { onMount } from "svelte";

  let rules = $state([
    {
      id: "1",
      name: "Required Properties",
      description: "Ensure all components have required properties",
      enabled: true,
      category: "validation",
    },
    {
      id: "2",
      name: "Naming Convention",
      description: "Component names follow naming conventions",
      enabled: false,
      category: "validation",
    },
  ]);

  let newRule = $state({
    name: "",
    description: "",
    enabled: true,
    category: "validation",
  });

  let showNewRuleForm = $state(false);
  let loading = $state(false);
  let error = $state<string | null>(null);

  function toggleRule(ruleId: string) {
    const rule = rules.find((r) => r.id === ruleId);
    if (rule) {
      rule.enabled = !rule.enabled;
      rules = [...rules]; // Trigger reactivity
    }
  }

  async function createRule() {
    if (!newRule.name.trim()) return;

    const rule = {
      id: Date.now().toString(),
      ...newRule,
    };

    rules = [rule, ...rules];
    newRule = {
      name: "",
      description: "",
      enabled: true,
      category: "validation",
    };
    showNewRuleForm = false;
  }

  async function runValidation() {
    loading = true;
    error = null;
    try {
      // Mock validation run
      await new Promise((resolve) => setTimeout(resolve, 1000));
      console.log(
        "Running validation with enabled rules:",
        rules.filter((r) => r.enabled)
      );
    } catch (e: unknown) {
      error = e instanceof Error ? e.message : "Validation failed";
    } finally {
      loading = false;
    }
  }
</script>

<div class="space-y-6">
  <div class="bg-white rounded-lg shadow">
    <div
      class="p-4 border-b border-slate-200 flex items-center justify-between"
    >
      <div>
        <h1 class="text-xl font-semibold text-slate-800">Validation Rules</h1>
        <p class="text-sm text-slate-600 mt-1">
          Manage component validation rules
        </p>
      </div>
      <div class="flex space-x-2">
        <button
          class="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
          onclick={runValidation}
          disabled={loading}
        >
          {loading ? "Running..." : "Run Validation"}
        </button>
        <button
          class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          onclick={() => (showNewRuleForm = !showNewRuleForm)}
        >
          {showNewRuleForm ? "Cancel" : "New Rule"}
        </button>
      </div>
    </div>

    {#if showNewRuleForm}
      <form
        class="p-4 space-y-4 border-b border-slate-200"
        onsubmit={createRule}
      >
        <div>
          <label for="ruleName" class="block text-sm font-medium text-slate-700"
            >Name</label
          >
          <input
            type="text"
            id="ruleName"
            class="mt-1 block w-full rounded border-slate-300 bg-white text-slate-900"
            bind:value={newRule.name}
            required
          />
        </div>
        <div>
          <label
            for="ruleDescription"
            class="block text-sm font-medium text-slate-700">Description</label
          >
          <textarea
            id="ruleDescription"
            class="mt-1 block w-full rounded border-slate-300 bg-white text-slate-900"
            rows="3"
            bind:value={newRule.description}
          ></textarea>
        </div>
        <div class="flex items-center space-x-2">
          <input
            type="checkbox"
            id="ruleEnabled"
            class="rounded border-slate-300"
            bind:checked={newRule.enabled}
          />
          <label for="ruleEnabled" class="text-sm text-slate-700"
            >Enable by default</label
          >
        </div>
        <div class="flex space-x-2">
          <button
            type="submit"
            class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
          >
            Create Rule
          </button>
          <button
            type="button"
            class="px-4 py-2 bg-slate-200 text-slate-700 rounded hover:bg-slate-300"
            onclick={() => (showNewRuleForm = false)}
          >
            Cancel
          </button>
        </div>
      </form>
    {/if}

    <div class="p-4">
      {#if error}
        <div
          class="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded"
        >
          {error}
        </div>
      {/if}

      {#if rules.length === 0}
        <div class="text-slate-500">
          No validation rules defined. Create your first rule above.
        </div>
      {:else}
        <div class="space-y-3">
          {#each rules as rule (rule.id)}
            <div
              class="border border-slate-200 rounded-lg p-4 hover:bg-slate-50"
            >
              <div class="flex items-start justify-between">
                <div class="flex items-start space-x-3">
                  <div class="flex items-center mt-1">
                    <input
                      type="checkbox"
                      checked={rule.enabled}
                      onchange={() => toggleRule(rule.id)}
                      class="rounded border-slate-300"
                    />
                  </div>
                  <div>
                    <h3
                      class="font-medium text-slate-900"
                      class:text-slate-500={!rule.enabled}
                    >
                      {rule.name}
                    </h3>
                    {#if rule.description}
                      <p
                        class="text-sm text-slate-600 mt-1"
                        class:text-slate-400={!rule.enabled}
                      >
                        {rule.description}
                      </p>
                    {/if}
                    <div
                      class="flex items-center space-x-4 mt-2 text-xs text-slate-500"
                    >
                      <span class="bg-slate-100 px-2 py-1 rounded">
                        {rule.category}
                      </span>
                      <span
                        class:text-green-600={rule.enabled}
                        class:text-slate-400={!rule.enabled}
                      >
                        {rule.enabled ? "Enabled" : "Disabled"}
                      </span>
                    </div>
                  </div>
                </div>
                <div class="flex space-x-2">
                  <button
                    class="px-3 py-1 text-sm bg-slate-100 text-slate-700 rounded hover:bg-slate-200"
                    onclick={() => console.log("Edit rule", rule.id)}
                  >
                    Edit
                  </button>
                  <button
                    class="px-3 py-1 text-sm bg-red-100 text-red-700 rounded hover:bg-red-200"
                    onclick={() => console.log("Delete rule", rule.id)}
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
</div>
