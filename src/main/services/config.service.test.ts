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
});
