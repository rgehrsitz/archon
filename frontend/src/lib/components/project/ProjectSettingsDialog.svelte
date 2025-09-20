<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogFooter, 
    DialogHeader, 
    DialogTitle 
  } from '../ui/dialog/index.js';
  import { Button } from '../ui/button/index.js';
  import { Input } from '../ui/input/index.js';
  import { Label } from '../ui/label/index.js';
  import { Textarea } from '../ui/textarea/index.js';
  import { Separator } from '../ui/separator/index.js';
  import { Badge } from '../ui/badge/index.js';
  import { GetCurrentProject } from '../../../../wailsjs/go/api/ProjectService.js';
  
  export let open = false;
  
  const dispatch = createEventDispatcher<{
    close: void;
    saved: { name: string; description: string };
  }>();
  
  let projectName = '';
  let projectDescription = '';
  let projectPath = '';
  let loading = true;
  let saving = false;
  
  // Load project info when dialog opens
  $: if (open) {
    loadProjectInfo();
  }
  
  async function loadProjectInfo() {
    try {
      loading = true;
      const project = await GetCurrentProject();
      projectName = project.name || '';
      projectDescription = project.description || '';
      projectPath = project.path || '';
    } catch (error) {
      console.error('Failed to load project info:', error);
    } finally {
      loading = false;
    }
  }
  
  async function handleSave() {
    try {
      saving = true;
      // TODO: Implement project update API call
      // For now, just emit the saved event
      dispatch('saved', { name: projectName, description: projectDescription });
      dispatch('close');
    } catch (error) {
      console.error('Failed to save project:', error);
    } finally {
      saving = false;
    }
  }
  
  function handleCancel() {
    dispatch('close');
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-md">
    <DialogHeader>
      <DialogTitle>Project Settings</DialogTitle>
      <DialogDescription>
        Manage your project settings and metadata
      </DialogDescription>
    </DialogHeader>
    
    {#if loading}
      <div class="flex items-center justify-center py-8">
        <div class="text-center">
          <div class="text-2xl mb-2">⏳</div>
          <div class="text-sm text-muted-foreground">Loading project info...</div>
        </div>
      </div>
    {:else}
      <div class="space-y-6">
        <!-- Project Info -->
        <div class="space-y-4">
          <div>
            <Label for="project-name">Project Name</Label>
            <Input
              id="project-name"
              bind:value={projectName}
              placeholder="Enter project name"
              class="mt-1"
            />
          </div>
          
          <div>
            <Label for="project-description">Description</Label>
            <Textarea
              id="project-description"
              bind:value={projectDescription}
              placeholder="Describe your project..."
              class="mt-1 min-h-20"
            />
          </div>
        </div>
        
        <Separator />
        
        <!-- Project Details -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium">Project Path</span>
            <Badge variant="outline" class="text-xs">
              {projectPath}
            </Badge>
          </div>
          
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium">Status</span>
            <Badge variant="secondary" class="text-xs">
              Active
            </Badge>
          </div>
        </div>
        
        <Separator />
        
        <!-- Actions -->
        <div class="space-y-2">
          <div class="text-sm font-medium">Danger Zone</div>
          <div class="text-xs text-muted-foreground">
            These actions cannot be undone
          </div>
          <Button variant="destructive" size="sm" disabled>
            Delete Project
          </Button>
        </div>
      </div>
    {/if}
    
    <DialogFooter>
      <Button variant="outline" onclick={handleCancel} disabled={saving}>
        Cancel
      </Button>
      <Button onclick={handleSave} disabled={saving || loading}>
        {saving ? 'Saving...' : 'Save Changes'}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
