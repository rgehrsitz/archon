<script lang="ts">
  // Simple test without LayerChart imports first
  const testData = [
    { x: 0, y: 0, width: 100, height: 50, fill: '#3b82f6' },
    { x: 100, y: 0, width: 100, height: 50, fill: '#ef4444' },
    { x: 0, y: 50, width: 100, height: 50, fill: '#10b981' },
    { x: 100, y: 50, width: 100, height: 50, fill: '#f59e0b' }
  ];
</script>

<div class="w-full h-full bg-slate-900 p-4 text-white">
  <h3 class="text-lg font-semibold mb-4 text-white">LayerChart Test</h3>
  
  <!-- Basic test -->
  <div class="mb-4 p-4 bg-blue-900 border border-blue-600 rounded">
    <h4 class="text-sm font-medium mb-2 text-blue-200">Basic Component Test</h4>
    <p class="text-blue-100">If you can see this text, the component is rendering!</p>
    <p class="text-blue-200 text-xs font-mono">Test data: {JSON.stringify(testData)}</p>
  </div>
  
  <!-- Simple SVG test -->
  <div class="border border-slate-600 rounded mb-4">
    <h4 class="text-sm font-medium mb-2 text-slate-200">Simple SVG Test</h4>
    <div class="bg-slate-800 p-4">
      <svg width="400" height="200" viewBox="0 0 400 200">
        {#each testData as item}
          <rect 
            x={item.x} 
            y={item.y} 
            width={item.width} 
            height={item.height} 
            fill={item.fill}
            stroke="#ffffff"
            stroke-width="2"
          />
        {/each}
      </svg>
    </div>
  </div>

  <!-- LayerChart import test -->
  <div class="border border-slate-600 rounded">
    <h4 class="text-sm font-medium mb-2 text-slate-200">LayerChart Import Test</h4>
    <div class="bg-slate-800 p-4">
      <p class="text-slate-200">Testing LayerChart imports...</p>
      <div id="layerchart-test">
        <!-- This will be populated by LayerChart if it works -->
        <p class="text-slate-300">LayerChart components should appear here</p>
      </div>
    </div>
  </div>
</div>
