import { describe, it, expect } from "vitest";
import path from "path";
import os from "os";
import { MigrationService } from "./migration.service";

describe("MigrationService", () => {
	it("returns ~/LightLauncher as legacy directory path", () => {
		expect(MigrationService.getLegacyDirectory()).toBe(path.join(os.homedir(), "LightLauncher"));
	});

	it("safely handles non-existent legacy directory without errors", async () => {
		const result = await MigrationService.migrateIfNeeded();
		expect(typeof result).toBe("boolean");
	});
});
