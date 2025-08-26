<script lang="ts">
  import { params } from 'svelte-spa-router';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  
  $: projectId = $params?.id || '';
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="max-w-4xl mx-auto p-8 space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Snapshots & History</h1>
        <p class="text-muted-foreground">Project: {projectId}</p>
      </div>
      <Button>Create Snapshot</Button>
    </div>
    
    <Card>
      <CardHeader>
        <CardTitle>History Timeline</CardTitle>
        <CardDescription>Linear timeline of snapshots and commits</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground">History view coming soon...</p>
      </CardContent>
    </Card>
  </div>
</div>