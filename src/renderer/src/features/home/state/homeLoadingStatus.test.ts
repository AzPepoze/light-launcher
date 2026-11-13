import { describe, expect, it } from "vitest";
import { homeStatus, scanPercent, scanStatusText } from "./homeLoadingStatus";

describe("scanPercent", () => {
	it("returns the done share of the total", () => {
		expect(scanPercent(0, 4)).toBe(0);
		expect(scanPercent(1, 4)).toBe(0.25);
		expect(scanPercent(4, 4)).toBe(1);
	});

	it("clamps to the 0 to 1 range", () => {
		expect(scanPercent(9, 4)).toBe(1);
		expect(scanPercent(-1, 4)).toBe(0);
	});

	it("returns null when the total is unknown", () => {
		expect(scanPercent(0, 0)).toBeNull();
	});
});

describe("scanStatusText", () => {
	it("counts folders done", () => {
		expect(scanStatusText(3, 5)).toBe("Scanning folders... 3 of 5");
	});

	it("names the folder being scanned", () => {
		expect(scanStatusText(2, 5, "Games")).toBe("Scanning Games... 2 of 5");
	});

	it("falls back to the folder count when current is empty", () => {
		expect(scanStatusText(2, 5, "")).toBe("Scanning folders... 2 of 5");
	});

	it("drops the count when the total is unknown", () => {
		expect(scanStatusText(0, 0)).toBe("Scanning folders...");
	});
});

describe("homeStatus", () => {
	it("shows the library load first, with no progress value", () => {
		expect(homeStatus(true, { state: "running", done: 1, total: 3, current: "Games" })).toEqual({
			text: "Loading games...",
			percent: null
		});
	});

	it("shows scan progress once the library is loaded", () => {
		expect(homeStatus(false, { state: "running", done: 3, total: 5, current: "" })).toEqual({
			text: "Scanning folders... 3 of 5",
			percent: 0.6
		});
	});

	it("names the folder being scanned", () => {
		expect(homeStatus(false, { state: "running", done: 1, total: 4, current: "Library" })).toEqual({
			text: "Scanning Library... 1 of 4",
			percent: 0.25
		});
	});

	it("shows nothing when the scan is done", () => {
		expect(homeStatus(false, { state: "done" })).toBeNull();
	});

	it("shows nothing when no scan has reported", () => {
		expect(homeStatus(false, null)).toBeNull();
	});
});
