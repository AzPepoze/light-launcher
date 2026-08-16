import { defineConfig } from "vitest/config";
import path from "path";

export default defineConfig({
	test: {
		globals: true,
		environment: "node",
		include: ["src/**/*.test.ts"],
		coverage: {
			reporter: ["text", "json", "html"],
			include: ["src/main/services/**/*.ts", "src/shared/**/*.ts"]
		}
	},
	resolve: {
		alias: {
			"@shared": path.resolve(__dirname, "src/shared/index.ts"),
			"@shared/*": path.resolve(__dirname, "src/shared/*"),
			"@services/*": path.resolve(__dirname, "src/main/services/*"),
			"@ipc/*": path.resolve(__dirname, "src/main/ipc/*"),
			"@lib/*": path.resolve(__dirname, "src/renderer/src/lib/*")
		}
	}
});
