<script lang="ts">
  import { params } from 'svelte-spa-router';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  
  $: projectId = $params?.id || '';
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="max-w-4xl mx-auto p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Import Wizard</h1>
      <p class="text-muted-foreground">Project: {projectId}</p>
    </div>
    
    <Card>
      <CardHeader>
        <CardTitle>Plugin-Driven Import</CardTitle>
        <CardDescription>Import data using plugins with validation and preview</CardDescription>
      </CardHeader>
      <CardContent>
        <p class="text-muted-foreground">Import wizard coming soon...</p>
      </CardContent>
    </Card>
  </div>
</div>