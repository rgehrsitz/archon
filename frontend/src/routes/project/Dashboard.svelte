<script lang="ts">
  import { params } from 'svelte-spa-router';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  
  // Get project ID from route params
  $: projectId = $params?.id || '';
</script>

<div class="min-h-screen bg-background text-foreground">
  <!-- Header -->
  <header class="sticky top-0 z-40 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
    <div class="flex h-14 items-center px-4">
      <div class="mr-4 flex">
        <a class="mr-6 flex items-center space-x-2" href="#/">
          <div class="h-6 w-6 rounded bg-primary"></div>
          <span class="hidden font-bold sm:inline-block">Archon</span>
        </a>
      </div>
      <div class="flex flex-1 items-center space-x-2">
        <h1 class="text-lg font-semibold">Project Dashboard</h1>
      </div>
      <div class="flex items-center space-x-2">
        <Button variant="outline" size="sm">Settings</Button>
        <Button size="sm">Open Workbench</Button>
      </div>
    </div>
  </header>

  <!-- Main Content -->
  <main class="container mx-auto p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Project Dashboard</h1>
        <p class="text-muted-foreground">Project ID: {projectId}</p>
      </div>
    </div>
    
    <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
      <!-- Quick Stats -->
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Recent Snapshots</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">3</div>
          <p class="text-xs text-muted-foreground">+1 from last week</p>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Total Nodes</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">127</div>
          <p class="text-xs text-muted-foreground">+12 from last week</p>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Pending Sync</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">2</div>
          <p class="text-xs text-muted-foreground">commits ahead</p>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Index Health</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">100%</div>
          <p class="text-xs text-muted-foreground">all indexed</p>
        </CardContent>
      </Card>
    </div>
    
    <!-- Main Dashboard Content -->
    <div class="grid gap-6 md:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Latest changes to your project</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="h-2 w-2 rounded-full bg-primary"></div>
              <div class="flex-1">
                <p class="text-sm">Added new equipment to Lab A</p>
                <p class="text-xs text-muted-foreground">2 hours ago</p>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <div class="h-2 w-2 rounded-full bg-muted"></div>
              <div class="flex-1">
                <p class="text-sm">Updated calibration data</p>
                <p class="text-xs text-muted-foreground">Yesterday</p>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <div class="h-2 w-2 rounded-full bg-muted"></div>
              <div class="flex-1">
                <p class="text-sm">Created snapshot v1.2</p>
                <p class="text-xs text-muted-foreground">2 days ago</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>Common tasks for your project</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3">
          <Button class="w-full justify-start" variant="outline">
            📁 Open Workbench
          </Button>
          <Button class="w-full justify-start" variant="outline">
            📸 Create Snapshot
          </Button>
          <Button class="w-full justify-start" variant="outline">
            📥 Import Data
          </Button>
          <Button class="w-full justify-start" variant="outline">
            🔍 Search Nodes
          </Button>
        </CardContent>
      </Card>
    </div>
  </main>
</div>