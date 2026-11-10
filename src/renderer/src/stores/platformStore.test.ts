import { describe, expect, it } from "vitest";
import { isPageAvailable } from "./platformStore";
import type { PlatformFeatures } from "@shared";

const linuxFeatures: PlatformFeatures = {
	proton: true,
	prefix: true,
	lsfg: true,
	gamescope: true,
	mangoHud: true,
	gameMode: true,
	liveTerminalLogs: true
};

const windowsFeatures: PlatformFeatures = {
	proton: false,
	prefix: false,
	lsfg: false,
	gamescope: false,
	mangoHud: false,
	gameMode: false,
	liveTerminalLogs: false
};

describe("isPageAvailable", () => {
	it("shows every page on linux", () => {
		for (const page of ["home", "run", "versions", "prefix", "utils", "editlsfg"]) {
			expect(isPageAvailable(page, linuxFeatures)).toBe(true);
		}
	});

	it("hides the linux-only pages on windows", () => {
		expect(isPageAvailable("versions", windowsFeatures)).toBe(false);
		expect(isPageAvailable("prefix", windowsFeatures)).toBe(false);
		expect(isPageAvailable("editlsfg", windowsFeatures)).toBe(false);
	});

	it("keeps the shared pages on windows", () => {
		for (const page of ["home", "manager", "run", "utils", "settings"]) {
			expect(isPageAvailable(page, windowsFeatures)).toBe(true);
		}
	});
});
