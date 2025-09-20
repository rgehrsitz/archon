<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { Button } from '../ui/button/index.js';
  import { Badge } from '../ui/badge/index.js';
  import { 
    DropdownMenu, 
    DropdownMenuContent, 
    DropdownMenuItem, 
    DropdownMenuSeparator, 
    DropdownMenuTrigger 
  } from '../ui/dropdown-menu/index.js';
  import { GetCurrentProject } from '../../../../wailsjs/go/api/ProjectService.js';
  
  export const projectId: string = undefined!;
  
  const dispatch = createEventDispatcher<{
    openSettings: void;
    renameProject: void;
    closeProject: void;
  }>();
  
  let projectName = 'Loading...';
  let projectPath = '';
  let loading = true;
  
  // Load project info
  async function loadProjectInfo() {
    try {
      const project = await GetCurrentProject();
      projectName = project.name || 'Untitled Project';
      projectPath = project.path || '';
    } catch (error) {
      console.error('Failed to load project info:', error);
      projectName = 'Unknown Project';
    } finally {
      loading = false;
    }
  }
  
  function handleBackToProjects() {
    push('/');
  }
  
  function handleOpenSettings() {
    dispatch('openSettings');
  }
  
  function handleRenameProject() {
    dispatch('renameProject');
  }
  
  function handleCloseProject() {
    dispatch('closeProject');
  }
  
  function handleProjectDashboard() {
    push(`/project/${projectId}/dashboard`);
  }
  
  function handleProjectWorkbench() {
    push(`/project/${projectId}/workbench`);
  }
  
  function handleProjectHistory() {
    push(`/project/${projectId}/history`);
  }
  
  function handleProjectImport() {
    push(`/project/${projectId}/import`);
  }
  
  // Load project info on mount
  loadProjectInfo();
</script>

<div class="h-14 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
  <div class="flex h-full items-center px-4 gap-4">
    <!-- Back Button -->
    <Button
      variant="ghost"
      size="sm"
      onclick={handleBackToProjects}
      class="flex items-center gap-2"
    >
      {#snippet children()}
        ← Back to Projects
      {/snippet}
    </Button>
    
    <!-- Project Info -->
    <div class="flex items-center gap-3 min-w-0 flex-1">
      <div class="min-w-0">
        <h1 class="font-semibold text-lg truncate" title={projectName}>
          {projectName}
        </h1>
        {#if projectPath}
          <p class="text-xs text-muted-foreground truncate" title={projectPath}>
            {projectPath}
          </p>
        {/if}
      </div>
      <Badge variant="secondary" class="text-xs">
        Active
      </Badge>
    </div>
    
    <!-- Navigation Tabs -->
    <nav class="flex items-center gap-1">
      <Button
        variant="ghost"
        size="sm"
        onclick={handleProjectDashboard}
        class="text-sm"
      >
        Dashboard
      </Button>
      <Button
        variant="ghost"
        size="sm"
        onclick={handleProjectWorkbench}
        class="text-sm"
      >
        Workbench
      </Button>
      <Button
        variant="ghost"
        size="sm"
        onclick={handleProjectHistory}
        class="text-sm"
      >
        History
      </Button>
      <Button
        variant="ghost"
        size="sm"
        onclick={handleProjectImport}
        class="text-sm"
      >
        Import
      </Button>
    </nav>
    
    <!-- Project Actions -->
    <DropdownMenu>
      <DropdownMenuTrigger>
        <Button variant="outline" size="sm">
          {#snippet children()}
            ⚙️ Project
          {/snippet}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" class="w-56">
        <DropdownMenuItem onclick={handleRenameProject}>
          ✏️ Rename Project
        </DropdownMenuItem>
        <DropdownMenuItem onclick={handleOpenSettings}>
          ⚙️ Project Settings
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onclick={handleCloseProject} variant="destructive">
          🚪 Close Project
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
</div>
