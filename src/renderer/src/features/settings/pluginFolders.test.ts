import { describe, expect, it } from "vitest";
import { isInsideFolder } from "./pluginFolders";

describe("isInsideFolder", () => {
	it("is true for the parent folder itself", () => {
		expect(isInsideFolder("/games", "/games")).toBe(true);
	});

	it("is true for a folder inside the parent", () => {
		expect(isInsideFolder("/games/Sample", "/games")).toBe(true);
		expect(isInsideFolder("C:\\games\\Sample", "C:\\games")).toBe(true);
	});

	it("is false for a sibling that only starts with the same text", () => {
		expect(isInsideFolder("/games-old/Sample", "/games")).toBe(false);
	});

	it("is false for a folder above the parent", () => {
		expect(isInsideFolder("/", "/games")).toBe(false);
	});

	it("ignores case and trailing slashes", () => {
		expect(isInsideFolder("/Games/Sample/", "/games/")).toBe(true);
	});

	it("is false for an empty path", () => {
		expect(isInsideFolder("", "/games")).toBe(false);
		expect(isInsideFolder("/games/Sample", "")).toBe(false);
	});
});
