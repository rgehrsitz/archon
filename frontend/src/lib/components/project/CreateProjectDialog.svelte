<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';
  import { Badge } from '$lib/components/ui/badge';
  import { CreateProject } from '../../../../wailsjs/go/api/ProjectService.js';
  
  export let open = false;
  
  const dispatch = createEventDispatcher<{
    created: { path: string; project: any };
    cancel: void;
  }>();
  
  let projectName = '';
  let projectDescription = '';
  let projectPath = '';
  let loading = false;
  let error = '';
  
  // Project settings
  let enableGitLFS = true;
  let enableIndex = true;
  let enablePlugins = true;
  
  async function handleSelectDirectory() {
    try {
      // For now, use prompt until we implement proper file dialog
      const path = prompt('Enter project directory path:');
      if (path) {
        projectPath = path;
        // Auto-generate project name from directory if not set
        if (!projectName && path) {
          const parts = path.split(/[\/\\]/);
          projectName = parts[parts.length - 1] || 'Archon Project';
        }
      }
    } catch (err) {
      error = `Failed to select directory: ${err}`;
    }
  }
  
  async function handleCreate() {
    if (!projectName.trim()) {
      error = 'Project name is required';
      return;
    }
    
    if (!projectPath.trim()) {
      error = 'Project path is required';
      return;
    }
    
    loading = true;
    error = '';
    
    try {
      const settings = {
        name: projectName,
        description: projectDescription,
        enableGitLFS,
        enableIndex,
        enablePlugins,
        createdAt: new Date().toISOString()
      };
      
      // Use Wails-generated binding directly (no context needed)
      const project = await CreateProject(projectPath, settings);
      dispatch('created', { path: projectPath, project });
      
      // Reset form
      projectName = '';
      projectDescription = '';
      projectPath = '';
      error = '';
      open = false;
      
    } catch (err) {
      error = `Failed to create project: ${err}`;
    } finally {
      loading = false;
    }
  }
  
  function handleCancel() {
    projectName = '';
    projectDescription = '';
    projectPath = '';
    error = '';
    open = false;
    dispatch('cancel');
  }
  
  // Reset error when inputs change
  $: if (projectName || projectPath) {
    error = '';
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[500px]">
    <DialogHeader>
      <DialogTitle>Create New Project</DialogTitle>
      <DialogDescription>
        Create a new Archon project with hierarchical nodes, Git-backed snapshots, and plugin support.
      </DialogDescription>
    </DialogHeader>
    
    <div class="space-y-4">
      <!-- Project Name -->
      <div class="space-y-2">
        <Label for="project-name">Project Name</Label>
        <Input
          id="project-name"
          bind:value={projectName}
          placeholder="My Archon Project"
          disabled={loading}
        />
      </div>
      
      <!-- Project Description -->
      <div class="space-y-2">
        <Label for="project-description">Description (Optional)</Label>
        <Textarea
          id="project-description"
          bind:value={projectDescription}
          placeholder="Describe your project..."
          class="min-h-[80px]"
          disabled={loading}
        />
      </div>
      
      <!-- Project Path -->
      <div class="space-y-2">
        <Label for="project-path">Project Location</Label>
        <div class="flex gap-2">
          <Input
            id="project-path"
            bind:value={projectPath}
            placeholder="/path/to/project"
            disabled={loading}
            class="flex-1"
          />
          <Button
            type="button"
            variant="outline"
            onclick={handleSelectDirectory}
            disabled={loading}
          >
            {#snippet children()}
              Browse
            {/snippet}
          </Button>
        </div>
      </div>
      
      <!-- Features -->
      <div class="space-y-3">
        <Label>Features</Label>
        <div class="flex flex-wrap gap-2">
          <Badge variant={enableGitLFS ? "default" : "secondary"}>
            Git LFS {enableGitLFS ? '✓' : '✗'}
          </Badge>
          <Badge variant={enableIndex ? "default" : "secondary"}>
            Search Index {enableIndex ? '✓' : '✗'}
          </Badge>
          <Badge variant={enablePlugins ? "default" : "secondary"}>
            Plugins {enablePlugins ? '✓' : '✗'}
          </Badge>
        </div>
        <p class="text-xs text-muted-foreground">
          Features can be configured later in project settings
        </p>
      </div>
      
      <!-- Error Display -->
      {#if error}
        <div class="rounded-md bg-destructive/15 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      {/if}
    </div>
    
    <DialogFooter>
      <Button
        type="button"
        variant="outline"
        onclick={handleCancel}
        disabled={loading}
      >
        {#snippet children()}
          Cancel
        {/snippet}
      </Button>
      <Button
        type="button"
        onclick={handleCreate}
        disabled={loading || !projectName.trim() || !projectPath.trim()}
      >
        {#snippet children()}
          {#if loading}
            Creating...
          {:else}
            Create Project
          {/if}
        {/snippet}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>