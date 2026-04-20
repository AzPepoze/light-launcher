import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import wails from "@wailsio/runtime/plugins/vite";
import { fileURLToPath, URL } from 'node:url';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte(), wails("./bindings")],
  resolve: {
    alias: {
      '@bindings': fileURLToPath(new URL('./bindings', import.meta.url)),
    },
  },
  server: {
    port: 9245,
    strictPort: true,
  }
})
