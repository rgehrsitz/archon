<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "$lib/components/ui/card";
</script>

<main class="p-8 bg-background text-foreground min-h-screen">
  <div class="max-w-4xl mx-auto space-y-8">
    <h1 class="text-4xl font-bold text-center">Wails + Svelte 5 + TailwindCSS 4 + shadcn-svelte</h1>
    
    <Card class="w-full max-w-md mx-auto">
      <CardHeader>
        <CardTitle>Welcome!</CardTitle>
        <CardDescription>Your shadcn-svelte setup is working perfectly</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <p class="text-muted-foreground">
          This demonstrates that Svelte 5, TailwindCSS 4, and shadcn-svelte are all configured correctly.
        </p>
        <div class="flex gap-2">
          <Button variant="default">Primary Button</Button>
          <Button variant="secondary">Secondary Button</Button>
          <Button variant="outline">Outline Button</Button>
        </div>
      </CardContent>
    </Card>
    
    <div class="text-center">
      <p class="text-lg text-muted-foreground">
        Ready to build beautiful UIs with modern tools! 🎉
      </p>
    </div>
  </div>
</main>

<style>
</style>
