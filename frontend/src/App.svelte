<script lang="ts">
  import router from 'svelte-spa-router';
  import { ModeWatcher } from 'mode-watcher';
  
  // Import routes
  import Welcome from './routes/Welcome.svelte';
  import Dashboard from './routes/project/Dashboard.svelte';
  import Workbench from './routes/project/Workbench.svelte';
  import History from './routes/project/History.svelte';
  import Diff from './routes/project/Diff.svelte';
  import Merge from './routes/project/Merge.svelte';
  import Import from './routes/project/Import.svelte';
  import Settings from './routes/Settings.svelte';
  import Plugins from './routes/Plugins.svelte';
  
  // Define routes
  const routes = {
    '/': Welcome,
    '/project/:id/dashboard': Dashboard,
    '/project/:id/workbench': Workbench,
    '/project/:id/history': History,
    '/project/:id/diff': Diff,
    '/project/:id/merge': Merge,
    '/project/:id/import': Import,
    '/settings': Settings,
    '/plugins': Plugins
  };
</script>

<!-- Mode watcher for theme management -->
<ModeWatcher />

<main class="min-h-screen bg-background text-foreground">
  <svelte:component this={router} {routes} />
</main>

<style>
  /* Global app styles */
  :global(html) {
    font-size: 14px;
  }
  
  :global(html[data-density="cozy"]) {
    font-size: 15px;
  }
</style>