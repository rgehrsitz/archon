<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '$lib/components/ui/dialog';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { OpenProject, ProjectExists } from '../../../../wailsjs/go/api/ProjectService.js';
  import { OpenDirectoryDialog } from '../../../../wailsjs/go/api/DialogService.js';
  
  export let open = false;
  
  const dispatch = createEventDispatcher<{
    opened: { path: string; project: any };
    cancel: void;
  }>();
  
  let projectPath = '';
  let loading = false;
  let error = '';
  let validating = false;
  
  async function handleSelectDirectory() {
    console.log('Browse button clicked!');
    try {
      const selectedPath = await OpenDirectoryDialog();
      if (selectedPath && selectedPath.trim()) {
        projectPath = selectedPath.trim();
        await validateProject(projectPath);
      }
    } catch (err) {
      console.error('Error opening directory dialog:', err);
      error = `Failed to open directory dialog: ${err}`;
    }
  }
  
  async function validateProject(path: string) {
    if (!path.trim()) return;
    
    validating = true;
    error = '';
    
    try {
      console.log('Validating project at path:', path);
      // Use Wails-generated binding to check if project exists
      const exists = await ProjectExists(path);
      console.log('ProjectExists result:', exists);
      
      if (!exists) {
        error = 'No Archon project found at this location';
      }
    } catch (err) {
      console.error('Validation error:', err);
      error = `Failed to validate project: ${err}`;
    } finally {
      validating = false;
    }
  }
  
  async function handleOpen() {
    if (!projectPath.trim()) {
      error = 'Project path is required';
      return;
    }
    
    loading = true;
    error = '';
    
    try {
      console.log('Opening project at path:', projectPath);
      // Use Wails-generated binding to open project
      const result = await OpenProject(projectPath);
      console.log('OpenProject result:', result);
      
      dispatch('opened', { path: projectPath, project: result });
      
      // Reset form
      projectPath = '';
      error = '';
      open = false;
      
    } catch (err) {
      console.error('Open project error:', err);
      error = `Failed to open project: ${err}`;
    } finally {
      loading = false;
    }
  }
  
  function handleCancel() {
    projectPath = '';
    error = '';
    open = false;
    dispatch('cancel');
  }
  
  // Validate when path changes with debouncing
  let validationTimeout: NodeJS.Timeout;
  $: {
    if (projectPath && !loading) {
      clearTimeout(validationTimeout);
      validationTimeout = setTimeout(() => validateProject(projectPath), 500);
    }
  }
  
  // Reset error when path changes
  $: if (projectPath) {
    error = '';
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[500px]">
    <DialogHeader>
      <DialogTitle>Open Existing Project</DialogTitle>
      <DialogDescription>
        Open an existing Archon project from your local filesystem.
      </DialogDescription>
    </DialogHeader>
    
    <div class="space-y-4">
      <!-- Project Path -->
      <div class="space-y-2">
        <Label for="open-project-path">Project Location</Label>
        <div class="flex gap-2">
          <Input
            id="open-project-path"
            bind:value={projectPath}
            placeholder="/path/to/existing-project"
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
        {#if validating}
          <p class="text-xs text-muted-foreground">Validating project...</p>
        {/if}
      </div>
      
      <!-- Project Info Preview -->
      {#if projectPath && !error && !validating}
        <div class="rounded-md bg-muted/50 p-3">
          <p class="text-sm">
            <span class="font-medium">Valid Archon project</span>
          </p>
          <p class="text-xs text-muted-foreground mt-1">
            Ready to open
          </p>
        </div>
      {/if}
      
      <!-- Error Display -->
      {#if error}
        <div class="rounded-md bg-destructive/15 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      {/if}
      
      <!-- Help Text -->
      <div class="rounded-md bg-muted/30 p-3 text-xs text-muted-foreground">
        <p class="font-medium mb-1">Looking for your project?</p>
        <p>Archon projects contain a <code>project.json</code> file and a <code>nodes/</code> directory.</p>
      </div>
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
        onclick={handleOpen}
        disabled={loading || !projectPath.trim() || !!error || validating}
      >
        {#snippet children()}
          {#if loading}
            Opening...
          {:else}
            Open Project
          {/if}
        {/snippet}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>

