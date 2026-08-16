import { describe, it, expect } from "vitest";
import {
	AppName,
	ConfigDirName,
	PrefixesDirName,
	DefaultWidth,
	DefaultHeight,
	DefaultMultiplier,
	DefaultExcludeNames
} from "./constants";

describe("Shared Constants", () => {
	it("has correct application directory constants", () => {
		expect(AppName).toBe("light-launcher");
		expect(ConfigDirName).toBe("games");
		expect(PrefixesDirName).toBe("prefixes");
	});

	it("has standard default fallback values", () => {
		expect(DefaultWidth).toBe("1280");
		expect(DefaultHeight).toBe("720");
		expect(DefaultMultiplier).toBe("2");
	});

	it("has default exclusion patterns for executable scans", () => {
		expect(DefaultExcludeNames.length).toBeGreaterThan(5);
		expect(DefaultExcludeNames).toContain("unins*.exe");
		expect(DefaultExcludeNames).toContain("DXSETUP.exe");
	});
});
