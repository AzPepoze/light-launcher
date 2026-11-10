import { describe, expect, it } from "vitest";
import { isFileForOs } from "./platform";

describe("isFileForOs", () => {
	it("matches the OS names used by the platform store", () => {
		expect(isFileForOs("windows", "windows")).toBe(true);
		expect(isFileForOs("linux", "linux")).toBe(true);
		expect(isFileForOs("mac", "darwin")).toBe(true);
	});

	it("rejects a file built for another OS", () => {
		expect(isFileForOs("windows", "linux")).toBe(false);
		expect(isFileForOs("mac", "windows")).toBe(false);
	});

	it("accepts files with no platform or an unknown one", () => {
		expect(isFileForOs("", "linux")).toBe(true);
		expect(isFileForOs("unknown", "linux")).toBe(true);
	});

	it("accepts every file until the OS is known", () => {
		expect(isFileForOs("windows", "")).toBe(true);
	});
});
