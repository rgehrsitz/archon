<script lang="ts">
  import { params } from 'svelte-spa-router';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  
  $: projectId = $params?.id || '';
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="max-w-6xl mx-auto p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Semantic Diff Viewer</h1>
      <p class="text-muted-foreground">Project: {projectId}</p>
    </div>
    
    <Card>
      <CardHeader>
        <CardTitle>Semantic Changes</CardTitle>
        <CardDescription>Structured view of changes between snapshots</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground">Diff viewer coming soon...</p>
      </CardContent>
    </Card>
  </div>
</div>