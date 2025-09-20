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
  import { GetCurrentProject } from '../../../../wailsjs/go/api/ProjectService.js';
  
  export let open = false;
  
  const dispatch = createEventDispatcher<{
    close: void;
    renamed: { name: string };
  }>();
  
  let projectName = '';
  let loading = true;
  let saving = false;
  
  // Load project name when dialog opens
  $: if (open) {
    loadProjectName();
  }
  
  async function loadProjectName() {
    try {
      loading = true;
      const project = await GetCurrentProject();
      projectName = project.name || '';
    } catch (error) {
      console.error('Failed to load project name:', error);
    } finally {
      loading = false;
    }
  }
  
  async function handleRename() {
    if (!projectName.trim()) {
      return;
    }
    
    try {
      saving = true;
      // TODO: Implement project rename API call
      // For now, just emit the renamed event
      dispatch('renamed', { name: projectName.trim() });
      dispatch('close');
    } catch (error) {
      console.error('Failed to rename project:', error);
    } finally {
      saving = false;
    }
  }
  
  function handleCancel() {
    dispatch('close');
  }
  
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && !saving) {
      handleRename();
    }
  }
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-md">
    <DialogHeader>
      <DialogTitle>Rename Project</DialogTitle>
      <DialogDescription>
        Enter a new name for your project
      </DialogDescription>
    </DialogHeader>
    
    {#if loading}
      <div class="flex items-center justify-center py-8">
        <div class="text-center">
          <div class="text-2xl mb-2">⏳</div>
          <div class="text-sm text-muted-foreground">Loading...</div>
        </div>
      </div>
    {:else}
      <div class="space-y-4">
        <div>
          <Label for="project-name">Project Name</Label>
          <Input
            id="project-name"
            bind:value={projectName}
            placeholder="Enter project name"
            class="mt-1"
            onkeydown={handleKeydown}
            disabled={saving}
          />
        </div>
      </div>
    {/if}
    
    <DialogFooter>
      <Button variant="outline" onclick={handleCancel} disabled={saving}>
        Cancel
      </Button>
      <Button 
        onclick={handleRename} 
        disabled={saving || loading || !projectName.trim()}
      >
        {saving ? 'Renaming...' : 'Rename'}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
