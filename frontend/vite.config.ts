import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      // Ensure wailsjs imports can be resolved from the project root
      'wailsjs': './wailsjs'
    }
  },  base: './', // Ensure assets resolve correctly in Wails
  build: {
    outDir: 'dist', // Output directory for Wails asset handler
    emptyOutDir: true
  }
})
