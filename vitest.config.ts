import { defineConfig } from "vitest/config";
import path from "path";

export default defineConfig({
	test: {
		globals: true,
		environment: "node",
		include: ["src/**/*.test.ts"],
		coverage: {
			reporter: ["text", "json", "html"],
			include: ["src/shared/**/*.ts", "src/renderer/src/lib/transport.ts"]
		}
	},
	resolve: {
		alias: {
			"@shared": path.resolve(__dirname, "src/shared"),
			"@lib": path.resolve(__dirname, "src/renderer/src/lib")
		}
	}
});
