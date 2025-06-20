<!-- Layout.svelte -->
<script lang="ts">
  import {
    InitializeSampleProject,
    LoadProject,
    CreateProject,
  } from "../../../wailsjs/go/main/App.js";

  let isSidebarOpen = $state(true);
  let currentPath = $state("/"); // Mock for demonstration

  function toggleSidebar() {
    isSidebarOpen = !isSidebarOpen;
  }

  async function handleNewProject() {
    await InitializeSampleProject();
    window.location.reload(); // Simple way to refresh state
  }

  async function handleOpenProject() {
    // Use browser file picker
    const input = document.createElement("input");
    input.type = "file";
    input.accept = ".archon";
    input.onchange = async (e: any) => {
      const file = e.target.files[0];
      if (file) {
        await LoadProject(file.path || file.name);
        window.location.reload();
      }
    };
    input.click();
  }

  async function handleSaveProject() {
    const name = prompt("Enter project name:");
    if (!name) return;
    await CreateProject(name, name + ".archon");
    alert("Project saved!");
  }
</script>

<div class="h-screen flex overflow-hidden bg-slate-100">
  <!-- Sidebar -->
  <aside class="flex flex-col w-64 bg-white border-r border-slate-200">
    <!-- Sidebar Header -->
    <div
      class="flex items-center justify-between h-16 px-4 border-b border-slate-200"
    >
      <h1 class="text-xl font-semibold text-slate-800">Archon</h1>
      <button
        class="p-2 rounded-md text-slate-500 hover:text-slate-600"
        onclick={toggleSidebar}
        aria-label="Toggle sidebar"
      >
        <svg
          class="w-6 h-6"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 6h16M4 12h16M4 18h16"
          />
        </svg>
      </button>
    </div>
    <!-- Project Actions -->
    <div class="flex flex-col gap-2 p-4 border-b border-slate-200">
      <button
        class="px-3 py-1 rounded bg-slate-200 text-slate-700 text-sm font-medium"
        onclick={handleNewProject}>New Project</button
      >
      <button
        class="px-3 py-1 rounded bg-slate-200 text-slate-700 text-sm font-medium"
        onclick={handleOpenProject}>Open Project</button
      >
      <button
        class="px-3 py-1 rounded bg-slate-200 text-slate-700 text-sm font-medium"
        onclick={handleSaveProject}>Save Project</button
      >
    </div>

    <!-- Navigation -->
    <nav class="flex-1 px-2 py-4 space-y-1">
      <a
        href="/"
        class="flex items-center px-4 py-2 text-slate-600 rounded-md hover:bg-slate-100"
        class:bg-slate-100={currentPath === "/"}
      >
        <svg
          class="w-5 h-5 mr-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
          />
        </svg>
        Components
      </a>
      <a
        href="/snapshots"
        class="flex items-center px-4 py-2 text-slate-600 rounded-md hover:bg-slate-100"
        class:bg-slate-100={currentPath === "/snapshots"}
      >
        <svg
          class="w-5 h-5 mr-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        Snapshots
      </a>
      <a
        href="/plugins"
        class="flex items-center px-4 py-2 text-slate-600 rounded-md hover:bg-slate-100"
        class:bg-slate-100={currentPath === "/plugins"}
      >
        <svg
          class="w-5 h-5 mr-3"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
          />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
        Plugins
      </a>
    </nav>
  </aside>

  <!-- Main Content -->
  <main class="flex-1 overflow-auto">
    <div class="container mx-auto px-6 py-8">
      <slot />
    </div>
  </main>
</div>
