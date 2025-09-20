<script lang="ts">
  import { params } from 'svelte-spa-router';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '$lib/components/ui/breadcrumb';
  import MillerColumns from '$lib/components/workbench/MillerColumns.svelte';
  import InspectorPanel from '$lib/components/workbench/InspectorPanel.svelte';
  
  // Get project ID from route params
  $: projectId = $params?.id || '';
  
  // State for Miller columns and inspector
  let selectedNodeId: string | null = null;
  let selectedNodePath: any[] = [];
  let selectedNode: any = null;
  
  function handleNodeSelect(event: CustomEvent<{ node: any, path: any[] }>) {
    selectedNode = event.detail.node;
    selectedNodePath = event.detail.path;
    selectedNodeId = event.detail.node.id;
  }
</script>

<div class="flex h-screen bg-background text-foreground">
  <!-- Main Content Area -->
  <div class="flex flex-1 flex-col">
    <!-- Command Bar -->
    <header class="sticky top-0 z-40 border-b bg-background/95 backdrop-blur">
      <div class="flex h-14 items-center px-4 gap-4">
        <!-- Breadcrumbs -->
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbLink href="#/">Archon</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbLink href="#/project/{projectId}/dashboard">Project</BreadcrumbLink>
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>Workbench</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
        
        <div class="flex flex-1"></div>
        
        <!-- Search -->
        <div class="hidden md:flex items-center gap-2">
          <Input placeholder="Search... (/ to focus)" class="w-64" />
          <Button variant="outline" size="sm">Filters</Button>
        </div>
        
        <!-- Actions -->
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm">View</Button>
          <Button size="sm">Snapshot</Button>
          <Button variant="outline" size="sm">Sync</Button>
        </div>
      </div>
    </header>

    <!-- Miller Columns + Inspector Layout -->
    <main class="flex flex-1 overflow-hidden">
      <!-- Miller Columns Area -->
      <div class="flex flex-1 border-r">
        <MillerColumns 
          bind:selectedNodeId 
          bind:selectedNodePath 
          on:nodeSelect={handleNodeSelect}
        />
      </div>

      <!-- Inspector Panel -->
      <InspectorPanel {selectedNode} />
    </main>
  </div>
</div>