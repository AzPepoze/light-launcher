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
			'@components': fileURLToPath(new URL('./src/components', import.meta.url)),
			'@lib': fileURLToPath(new URL('./src/lib', import.meta.url)),
			'@stores': fileURLToPath(new URL('./src/stores', import.meta.url)),
			'@icons': fileURLToPath(new URL('./src/icons', import.meta.url)),
		},
	},
	server: {
		port: 9245,
		strictPort: true,
	}
})
