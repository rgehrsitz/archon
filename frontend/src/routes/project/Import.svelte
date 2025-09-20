<script lang="ts">
  import { params } from 'svelte-spa-router';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
  import ProjectHeader from '$lib/components/project/ProjectHeader.svelte';
  import ProjectSettingsDialog from '$lib/components/project/ProjectSettingsDialog.svelte';
  import RenameProjectDialog from '$lib/components/project/RenameProjectDialog.svelte';
  
  $: projectId = $params?.id || '';
  
  // Dialog states
  let showProjectSettings = false;
  let showRenameDialog = false;
  
  // Project action handlers
  function handleOpenSettings() {
    showProjectSettings = true;
  }
  
  function handleRenameProject() {
    showRenameDialog = true;
  }
  
  function handleCloseProject() {
    // TODO: Implement project close functionality
    console.log('Close project requested');
  }
  
  function handleProjectSettingsSaved(event: CustomEvent) {
    console.log('Project settings saved:', event.detail);
    showProjectSettings = false;
  }
  
  function handleProjectRenamed(event: CustomEvent) {
    console.log('Project renamed:', event.detail);
    showRenameDialog = false;
  }
</script>

<div class="min-h-screen bg-background text-foreground">
  <!-- Project Header -->
  <ProjectHeader 
    {projectId}
    on:openSettings={handleOpenSettings}
    on:renameProject={handleRenameProject}
    on:closeProject={handleCloseProject}
  />
  
  <div class="max-w-4xl mx-auto p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Import Wizard</h1>
      <p class="text-muted-foreground">Import data using plugins with validation and preview</p>
    </div>
    
    <Card>
      <CardHeader>
        <CardTitle>Plugin-Driven Import</CardTitle>
        <CardDescription>Import data using plugins with validation and preview</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground">Import wizard coming soon...</p>
      </CardContent>
    </Card>
  </div>
  
  <!-- Project Dialogs -->
  <ProjectSettingsDialog 
    bind:open={showProjectSettings}
    on:close={() => showProjectSettings = false}
    on:saved={handleProjectSettingsSaved}
  />
  
  <RenameProjectDialog 
    bind:open={showRenameDialog}
    on:close={() => showRenameDialog = false}
    on:renamed={handleProjectRenamed}
  />
</div>