import { describe, it, expect } from "vitest";
import { ConfigService } from "./config.service";

describe("ConfigService", () => {
	it("generates a valid 16-character hex ID", () => {
		const id1 = ConfigService.generateId();
		const id2 = ConfigService.generateId();
		expect(id1).toHaveLength(16);
		expect(id2).toHaveLength(16);
		expect(id1).not.toBe(id2);
		expect(/^[0-9a-f]{16}$/.test(id1)).toBe(true);
	});

	it("returns default LaunchOptions structure for unknown prefix", async () => {
		const prefixConfig = await ConfigService.loadPrefixConfig("NonExistentPrefix_123");
		expect(prefixConfig).toBeDefined();
		expect(prefixConfig.Name).toBe("NonExistentPrefix_123");
		expect(prefixConfig.Extras.Gamescope.Width).toBe("1280");
		expect(prefixConfig.Extras.Gamescope.Height).toBe("720");
		expect(prefixConfig.Extras.Memory.Value).toBe("4G");
		expect(prefixConfig.Extras.Lsfg.Multiplier).toBe("2");
	});

	it("saves and loads JSON data atomically", async () => {
		const os = await import("os");
		const path = await import("path");
		const fs = await import("fs/promises");

		const tempFile = path.join(os.tmpdir(), `test-config-${Date.now()}.json`);
		try {
			await ConfigService.saveJson(tempFile, { hello: "world", test: 123 });
			const loaded = await ConfigService.loadJson<{ hello: string; test: number }>(tempFile);
			expect(loaded.hello).toBe("world");
			expect(loaded.test).toBe(123);
		} finally {
			await fs.unlink(tempFile).catch(() => {});
		}
	});

	it("rejects empty JSON files cleanly", async () => {
		const os = await import("os");
		const path = await import("path");
		const fs = await import("fs/promises");

		const tempFile = path.join(os.tmpdir(), `test-empty-${Date.now()}.json`);
		try {
			await fs.writeFile(tempFile, "", "utf-8");
			await expect(ConfigService.loadJson(tempFile)).rejects.toThrow("is empty");
		} finally {
			await fs.unlink(tempFile).catch(() => {});
		}
	});
});
