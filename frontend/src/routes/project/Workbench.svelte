<script lang="ts">
  import { params } from 'svelte-spa-router';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '$lib/components/ui/breadcrumb';
  
  // Get project ID from route params
  $: projectId = $params?.id || '';
  
  // Mock data for Miller columns
  const mockColumns = [
    { title: 'Sites', items: ['Lab A', 'Lab B', 'Storage', 'Calibration'] },
    { title: 'Lab A', items: ['Rooms', 'Benches', 'Storage'] },
    { title: 'Benches', items: ['Bench 1', 'Bench 2', 'Bench 3', 'Bench 4'] },
  ];
  
  let selectedItem = 'Bench 3';
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
        <div class="grid grid-cols-3 flex-1">
          {#each mockColumns as column, i}
            <div class="border-r last:border-r-0 overflow-hidden">
              <!-- Column Header -->
              <div class="sticky top-0 bg-muted/50 border-b px-3 py-2">
                <div class="flex items-center justify-between">
                  <h3 class="font-medium text-sm">{column.title}</h3>
                  <Button variant="ghost" size="sm" class="h-6 px-2 text-xs">
                    New
                  </Button>
                </div>
              </div>
              
              <!-- Column Content -->
              <div class="overflow-y-auto h-full">
                <div class="divide-y">
                  {#each column.items as item}
                    <button 
                      class="w-full px-3 py-2 text-left text-sm hover:bg-muted/50 transition-colors border-0 bg-transparent focus:bg-muted outline-none focus-visible:ring-2 focus-visible:ring-ring"
                      class:bg-accent={item === selectedItem}
                      class:text-accent-foreground={item === selectedItem}
                      on:click={() => selectedItem = item}
                    >
                      <div class="flex items-center gap-2">
                        <div class="h-1.5 w-1.5 rounded-full bg-primary"></div>
                        <span>{item}</span>
                      </div>
                    </button>
                  {/each}
                </div>
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Inspector Panel -->
      <aside class="w-80 flex flex-col">
        <div class="border-b px-4 py-3">
          <h2 class="font-semibold">Inspector</h2>
        </div>
        
        <div class="flex-1 overflow-y-auto p-4 space-y-4">
          <!-- Node Info -->
          <Card>
            <CardHeader class="pb-3">
              <CardTitle class="text-base">Node Details</CardTitle>
            </CardHeader>
            <CardContent class="space-y-3">
              <div>
                <label for="node-name" class="text-xs font-medium text-muted-foreground">Name</label>
                <Input id="node-name" value={selectedItem} class="mt-1" />
              </div>
              
              <div>
                <label for="node-description" class="text-xs font-medium text-muted-foreground">Description</label>
                <Input id="node-description" placeholder="Add description..." class="mt-1" />
              </div>
            </CardContent>
          </Card>
          
          <!-- Properties -->
          <Card>
            <CardHeader class="pb-3">
              <CardTitle class="text-base">Properties</CardTitle>
            </CardHeader>
            <CardContent class="space-y-3">
              <div class="grid grid-cols-[120px_1fr] gap-2 items-center">
                <span class="text-sm text-muted-foreground">max_voltage</span>
                <Input value="60" class="h-8" />
              </div>
              <div class="grid grid-cols-[120px_1fr] gap-2 items-center">
                <span class="text-sm text-muted-foreground">serial</span>
                <Input value="ABC123" class="h-8" />
              </div>
              <Button variant="outline" size="sm" class="w-full">
                Add Property
              </Button>
            </CardContent>
          </Card>
          
          <!-- Attachments -->
          <Card>
            <CardHeader class="pb-3">
              <CardTitle class="text-base">Attachments</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="text-sm text-muted-foreground text-center py-4">
                No attachments
              </div>
              <Button variant="outline" size="sm" class="w-full">
                Add Attachment
              </Button>
            </CardContent>
          </Card>
        </div>
      </aside>
    </main>
  </div>
</div>