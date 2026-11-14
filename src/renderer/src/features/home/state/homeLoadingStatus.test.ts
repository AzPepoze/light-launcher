import { describe, expect, it } from "vitest";
import {
	homeStatus,
	runningFolderScan,
	scanPercent,
	scanStatusForFolder
} from "./homeLoadingStatus";

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

describe("homeStatus", () => {
	it("shows the library load with no progress value", () => {
		expect(homeStatus(true)).toEqual({ text: "Loading games...", percent: null });
	});

	it("shows nothing once the library is loaded", () => {
		expect(homeStatus(false)).toBeNull();
	});
});

describe("runningFolderScan", () => {
	it("names the folder being scanned with its progress", () => {
		expect(
			runningFolderScan({ state: "running", done: 3, total: 5, found: 18, current: "Games" })
		).toEqual({
			folderName: "Games",
			text: "Scanning... 3 of 5 · 18 found",
			percent: 0.6
		});
	});

	it("keeps the count when the total is unknown", () => {
		expect(
			runningFolderScan({ state: "running", done: 0, total: 0, found: 0, current: "Games" })
		).toEqual({ folderName: "Games", text: "Scanning... 0 of 0 · 0 found", percent: null });
	});

	it("returns null when the scan is done", () => {
		expect(runningFolderScan({ state: "done" })).toBeNull();
	});

	it("returns null when no scan has reported", () => {
		expect(runningFolderScan(null)).toBeNull();
	});

	it("returns null when no folder name is known yet", () => {
		expect(
			runningFolderScan({ state: "running", done: 0, total: 2, found: 0, current: "" })
		).toBeNull();
	});
});

describe("scanStatusForFolder", () => {
	const status = {
		folderName: "Games",
		text: "Scanning... 1 of 2 · 4 found",
		percent: 0.5
	};

	it("returns the status for the folder being scanned", () => {
		expect(scanStatusForFolder(status, "Games")).toBe(status);
	});

	it("returns null for any other folder", () => {
		expect(scanStatusForFolder(status, "Library")).toBeNull();
	});

	it("returns null when no folder is being scanned", () => {
		expect(scanStatusForFolder(null, "Games")).toBeNull();
	});
});
