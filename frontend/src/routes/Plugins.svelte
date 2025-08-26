<script lang="ts">
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  import { Badge } from '$lib/components/ui/badge';
  import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
</script>

<div class="min-h-screen bg-background text-foreground">
  <div class="max-w-6xl mx-auto p-8 space-y-8">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Plugin Manager</h1>
      <p class="text-muted-foreground">Manage your Archon plugins and permissions</p>
    </div>
    
    <Tabs value="installed" class="w-full">
      <TabsList>
        <TabsTrigger value="installed">Installed</TabsTrigger>
        <TabsTrigger value="permissions">Permissions</TabsTrigger>
        <TabsTrigger value="discover" disabled>Discover</TabsTrigger>
      </TabsList>
      
      <TabsContent value="installed" class="space-y-4">
        <Card>
          <CardHeader>
            <div class="flex items-center justify-between">
              <div>
                <CardTitle>CSV Importer</CardTitle>
                <CardDescription>Import data from CSV files</CardDescription>
              </div>
              <div class="flex items-center gap-2">
                <Badge variant="secondary">Importer</Badge>
                <Button variant="outline" size="sm">Configure</Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div class="flex items-center justify-between">
              <div class="text-sm text-muted-foreground">
                Version: 0.1.0 • Permissions: readRepo
              </div>
              <Button variant="outline" size="sm">Enable</Button>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader>
            <div class="flex items-center justify-between">
              <div>
                <CardTitle>Data Validator</CardTitle>
                <CardDescription>Validate node data against schemas</CardDescription>
              </div>
              <div class="flex items-center gap-2">
                <Badge variant="secondary">Validator</Badge>
                <Button variant="outline" size="sm">Configure</Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div class="flex items-center justify-between">
              <div class="text-sm text-muted-foreground">
                Version: 0.2.0 • Permissions: readRepo, writeRepo
              </div>
              <Button variant="outline" size="sm">Enable</Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
      
      <TabsContent value="permissions" class="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Permission Management</CardTitle>
            <CardDescription>Control what plugins can access</CardDescription>
          </CardHeader>
          <CardContent>
            <p class="text-muted-foreground">Permission management coming soon...</p>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  </div>
</div>