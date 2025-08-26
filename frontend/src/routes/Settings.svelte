<script lang="ts">
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Switch } from '$lib/components/ui/switch';
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="max-w-4xl mx-auto p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Settings</h1>
      <p class="text-muted-foreground">Configure your Archon preferences</p>
    </div>
    
    <div class="grid gap-6">
      <!-- Appearance -->
      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>Customize the look and feel of Archon</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="flex items-center justify-between">
            <Label for="dark-mode">Dark Mode</Label>
            <Switch id="dark-mode" />
          </div>
          <div class="flex items-center justify-between">
            <Label for="cozy-density">Cozy Density</Label>
            <Switch id="cozy-density" />
          </div>
        </CardContent>
      </Card>
      
      <!-- Git Settings -->
      <Card>
        <CardHeader>
          <CardTitle>Git Configuration</CardTitle>
          <CardDescription>Git remote and credential settings</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label for="git-remote">Default Remote URL</Label>
            <Input id="git-remote" placeholder="https://github.com/user/repo.git" />
          </div>
          <div class="space-y-2">
            <Label for="git-user">Git User Name</Label>
            <Input id="git-user" placeholder="Your Name" />
          </div>
          <div class="space-y-2">
            <Label for="git-email">Git Email</Label>
            <Input id="git-email" type="email" placeholder="your.email@example.com" />
          </div>
        </CardContent>
      </Card>
      
      <!-- LFS Settings -->
      <Card>
        <CardHeader>
          <CardTitle>Git LFS</CardTitle>
          <CardDescription>Large file storage configuration</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label for="lfs-threshold">LFS Threshold (MB)</Label>
            <Input id="lfs-threshold" type="number" value="1" />
          </div>
          <div class="flex items-center justify-between">
            <Label for="lfs-enabled">Enable Git LFS</Label>
            <Switch id="lfs-enabled" checked />
          </div>
        </CardContent>
      </Card>
      
      <!-- About -->
      <Card>
        <CardHeader>
          <CardTitle>About Archon</CardTitle>
          <CardDescription>Version and system information</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-1">
            <p class="text-sm font-medium">Version: 1.0.0-alpha</p>
            <p class="text-sm text-muted-foreground">Desktop knowledge workbench with hierarchical projects</p>
          </div>
          <div class="flex gap-2">
            <Button variant="outline" size="sm">View Logs</Button>
            <Button variant="outline" size="sm">Check for Updates</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</div>