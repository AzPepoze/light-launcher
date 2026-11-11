import { svelte } from "@sveltejs/vite-plugin-svelte";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";

// The Go launcher serves /api in dev (started with --port 9246 --dev-token dev).
const GO_API_TARGET = "http://127.0.0.1:9246";

// material-icons is installed in the repository root, outside this Vite root.
const REPO_ROOT = fileURLToPath(new URL("../..", import.meta.url));

// https://vitejs.dev/config/
export default defineConfig({
	base: "./",
	plugins: [svelte()],
	resolve: {
		alias: {
			"@shared": fileURLToPath(new URL("../shared", import.meta.url)),
			"@components": fileURLToPath(new URL("./src/components", import.meta.url)),
			"@lib": fileURLToPath(new URL("./src/lib", import.meta.url)),
			"@stores": fileURLToPath(new URL("./src/stores", import.meta.url)),
			"@icons": fileURLToPath(new URL("./src/icons", import.meta.url))
		}
	},
	server: {
		fs: {
			allow: [REPO_ROOT]
		},
		port: 9245,
		strictPort: true,
		proxy: {
			"/api": {
				target: GO_API_TARGET,
				changeOrigin: true,
				configure: (proxy) => {
					// The launcher only accepts its own origin, so make proxied calls look like direct ones.
					proxy.on("proxyReq", (request) => request.setHeader("Origin", GO_API_TARGET));
				}
			}
		}
	},
	build: {
		target: "esnext",
		minify: "esbuild",
		sourcemap: false,
		cssMinify: true,
		cssCodeSplit: true,
		outDir: "dist",
		modulePreload: {
			polyfill: false
		},
		rollupOptions: {
			treeshake: {
				preset: "smallest",
				moduleSideEffects: "no-external"
			}
		}
	}
});
