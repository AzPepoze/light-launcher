import { svelte } from "@sveltejs/vite-plugin-svelte";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

// https://vitejs.dev/config/
export default defineConfig({
	base: "./",
	plugins: [svelte()],
	resolve: {
		alias: {
			"@shared": fileURLToPath(new URL("../shared", import.meta.url)),
			"@components": fileURLToPath(
				new URL("./src/components", import.meta.url)
			),
			"@lib": fileURLToPath(new URL("./src/lib", import.meta.url)),
			"@stores": fileURLToPath(
				new URL("./src/stores", import.meta.url)
			),
			"@icons": fileURLToPath(new URL("./src/icons", import.meta.url)),
		},
	},
	server: {
		port: 9245,
		strictPort: true,
	},
	build: {
		target: "esnext",
		minify: "esbuild",
		sourcemap: false,
		cssMinify: true,
		cssCodeSplit: true,
		outDir: "dist",
		modulePreload: {
			polyfill: false,
		},
		rollupOptions: {
			treeshake: {
				preset: "smallest",
				moduleSideEffects: "no-external",
			},
		},
	},
});
