<script lang="ts">
  import { onMount } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import CreateProjectDialog from '$lib/components/project/CreateProjectDialog.svelte';
  import OpenProjectDialog from '$lib/components/project/OpenProjectDialog.svelte';
  import { IsProjectOpen, GetCurrentProject } from '../../wailsjs/go/api/ProjectService.js';
  
  let showCreateDialog = false;
  let showOpenDialog = false;
  let currentProjectOpen = false;
  let currentProject: any = null;
  let loading = true;
  
  // Mock recent projects for now - in a real app this would come from local storage or settings
  let recentProjects = [
    { name: 'Lab Equipment Catalog', path: '/Users/jane/archon-projects/lab-catalog', lastOpened: '2 hours ago' },
    { name: 'Manufacturing Plant', path: '/Users/jane/archon-projects/manufacturing', lastOpened: 'Yesterday' },
    { name: 'Research Notes', path: '/Users/jane/archon-projects/research', lastOpened: '3 days ago' }
  ];
  
  onMount(async () => {
    try {
      // Use Wails-generated bindings directly (no context needed)
      currentProjectOpen = await IsProjectOpen();
      if (currentProjectOpen) {
        currentProject = await GetCurrentProject();
      }
    } catch (error) {
      console.error('Failed to check project status:', error);
      // Set defaults if API calls fail
      currentProjectOpen = false;
      currentProject = null;
    } finally {
      loading = false;
    }
  });
  
  function handleProjectCreated(event: CustomEvent) {
    const { path, project } = event.detail;
    console.log('Project created:', { path, project });
    
    // Add to recent projects
    recentProjects = [
      { name: project.settings?.name || 'New Project', path, lastOpened: 'Just now' },
      ...recentProjects.slice(0, 4) // Keep only 5 recent projects
    ];
    
    // Navigate to the workbench to start working
    // Extract project ID from path or use a hash of the path for now
    const projectId = path.split('/').pop() || 'project';
    push(`/project/${projectId}/workbench`);
  }
  
  function handleProjectOpened(event: CustomEvent) {
    const { path, project } = event.detail;
    console.log('Project opened:', { path, project });
    
    // Update recent projects (move to top)
    const existingIndex = recentProjects.findIndex(p => p.path === path);
    const projectEntry = {
      name: project.settings?.name || path.split('/').pop() || 'Project',
      path,
      lastOpened: 'Just now'
    };
    
    if (existingIndex >= 0) {
      recentProjects = [projectEntry, ...recentProjects.filter((_, i) => i !== existingIndex)];
    } else {
      recentProjects = [projectEntry, ...recentProjects.slice(0, 4)];
    }
    
    // Navigate to the workbench to start working
    const projectId = path.split('/').pop() || 'project';
    push(`/project/${projectId}/workbench`);
  }
  
  function openRecentProject(project: any) {
    console.log('Opening recent project:', project);
    const projectId = project.path.split('/').pop() || 'project';
    push(`/project/${projectId}/workbench`);
  }
  
  function handleCloneProject() {
    // TODO: Implement Git clone functionality
    alert('Git clone functionality coming soon!');
  }
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="max-w-6xl mx-auto p-8 space-y-8">
    <!-- Header -->
    <div class="text-center space-y-4">
      <div class="flex items-center justify-center gap-3 mb-6">
        <div class="h-12 w-12 rounded-xl bg-primary shadow-lg"></div>
        <div class="text-left">
          <h1 class="text-4xl font-bold tracking-tight">Archon</h1>
          <p class="text-lg text-muted-foreground">Hierarchical knowledge workbench</p>
        </div>
      </div>
      
      {#if currentProjectOpen && currentProject}
        <div class="bg-primary/10 border border-primary/20 rounded-lg p-4 max-w-md mx-auto">
          <p class="text-sm font-medium">Project Currently Open</p>
          <p class="text-primary font-semibold">{currentProject.settings?.name || 'Current Project'}</p>
          <Button size="sm" class="mt-2" on:click={() => push('/project/current/workbench')}>
            Open Workbench
          </Button>
        </div>
      {/if}
    </div>
    
    <div class="grid md:grid-cols-2 gap-6">
      <!-- Recent Projects -->
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            Recent Projects
            <Badge variant="secondary">{recentProjects.length}</Badge>
          </CardTitle>
          <CardDescription>Continue working on your recent projects</CardDescription>
        </CardHeader>
        <CardContent>
          {#if loading}
            <div class="space-y-3">
              {#each Array(3) as _}
                <div class="h-12 bg-muted/50 rounded-md animate-pulse"></div>
              {/each}
            </div>
          {:else if recentProjects.length > 0}
            <div class="space-y-2">
              {#each recentProjects as project}
                <button
                  class="w-full text-left p-3 rounded-md border hover:bg-accent hover:text-accent-foreground transition-colors"
                  on:click={() => openRecentProject(project)}
                >
                  <div class="flex items-center justify-between">
                    <div>
                      <p class="font-medium">{project.name}</p>
                      <p class="text-xs text-muted-foreground">{project.path}</p>
                    </div>
                    <div class="text-xs text-muted-foreground">
                      {project.lastOpened}
                    </div>
                  </div>
                </button>
              {/each}
            </div>
          {:else}
            <div class="text-center py-8 text-muted-foreground">
              <div class="h-12 w-12 rounded-full bg-muted/50 mx-auto mb-3 flex items-center justify-center">
                📁
              </div>
              <p class="text-sm">No recent projects yet</p>
              <p class="text-xs">Create or open a project to get started</p>
            </div>
          {/if}
        </CardContent>
      </Card>
      
      <!-- Create/Open Project -->
      <Card>
        <CardHeader>
          <CardTitle>Get Started</CardTitle>
          <CardDescription>Create a new project or open an existing one</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3">
          <Button 
            class="w-full" 
            variant="default"
            onclick={() => {
              showCreateDialog = true;
            }}
          >
            {#snippet children()}
              📝 Create New Project
            {/snippet}
          </Button>
          <Button 
            class="w-full" 
            variant="outline"
            onclick={() => {
              showOpenDialog = true;
            }}
          >
            {#snippet children()}
              📂 Open Existing Project
            {/snippet}
          </Button>
          <Button 
            class="w-full" 
            variant="outline"
            onclick={handleCloneProject}
          >
            {#snippet children()}
              🔄 Clone from Git
            {/snippet}
          </Button>
        </CardContent>
      </Card>
    </div>
    
    <!-- Features Overview -->
    <Card>
      <CardHeader>
        <CardTitle>What's Inside</CardTitle>
        <CardDescription>Powerful features for managing hierarchical knowledge</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid md:grid-cols-3 gap-4">
          <div class="text-center p-4">
            <div class="h-12 w-12 rounded-full bg-primary/10 mx-auto mb-3 flex items-center justify-center">
              🌳
            </div>
            <h3 class="font-semibold mb-1">Hierarchical Nodes</h3>
            <p class="text-xs text-muted-foreground">Organize your knowledge in meaningful tree structures</p>
          </div>
          <div class="text-center p-4">
            <div class="h-12 w-12 rounded-full bg-primary/10 mx-auto mb-3 flex items-center justify-center">
              📸
            </div>
            <h3 class="font-semibold mb-1">Git Snapshots</h3>
            <p class="text-xs text-muted-foreground">Version control with semantic diffs and merge resolution</p>
          </div>
          <div class="text-center p-4">
            <div class="h-12 w-12 rounded-full bg-primary/10 mx-auto mb-3 flex items-center justify-center">
              🔌
            </div>
            <h3 class="font-semibold mb-1">Plugin System</h3>
            <p class="text-xs text-muted-foreground">Extend functionality with sandboxed plugins</p>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</div>

<!-- Dialogs -->
<CreateProjectDialog bind:open={showCreateDialog} on:created={handleProjectCreated} />
<OpenProjectDialog bind:open={showOpenDialog} on:opened={handleProjectOpened} />