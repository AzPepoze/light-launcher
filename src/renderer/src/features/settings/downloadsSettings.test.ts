import { describe, expect, it } from "vitest";
import {
	DEFAULT_DOWNLOAD_SETTINGS,
	DOWNLOAD_SAVE_DELAY_MS,
	clampDownloadNumber,
	clampDownloadSettings,
	parseDownloadNumber,
	withDownloadNumber
} from "./downloadsSettings";

describe("clampDownloadNumber", () => {
	it("clamps connections to 1 to 32", () => {
		expect(clampDownloadNumber("connections", 0)).toBe(1);
		expect(clampDownloadNumber("connections", 40)).toBe(32);
		expect(clampDownloadNumber("connections", 12)).toBe(12);
	});

	it("clamps simultaneous downloads to 1 to 10", () => {
		expect(clampDownloadNumber("maxActive", -3)).toBe(1);
		expect(clampDownloadNumber("maxActive", 11)).toBe(10);
	});

	it("keeps retries at 0 to 10", () => {
		expect(clampDownloadNumber("retries", -1)).toBe(0);
		expect(clampDownloadNumber("retries", 11)).toBe(10);
	});

	it("keeps the speed limit at 0 or more, with no upper limit", () => {
		expect(clampDownloadNumber("speedLimitKBps", -5)).toBe(0);
		expect(clampDownloadNumber("speedLimitKBps", 0)).toBe(0);
		expect(clampDownloadNumber("speedLimitKBps", 50000)).toBe(50000);
	});

	it("rounds to whole numbers", () => {
		expect(clampDownloadNumber("connections", 4.6)).toBe(5);
	});

	it("falls back to the minimum for values that are not finite", () => {
		expect(clampDownloadNumber("connections", Number.NaN)).toBe(1);
		expect(clampDownloadNumber("speedLimitKBps", Number.POSITIVE_INFINITY)).toBe(0);
	});
});

describe("clampDownloadSettings", () => {
	it("clamps every number and keeps the switch", () => {
		const clamped = clampDownloadSettings({
			connections: 99,
			maxActive: 0,
			hostConnections: 0,
			retries: 50,
			speedLimitKBps: -1,
			keepArchives: true
		});
		expect(clamped).toEqual({
			connections: 32,
			maxActive: 1,
			hostConnections: 1,
			retries: 10,
			speedLimitKBps: 0,
			keepArchives: true
		});
	});

	it("leaves valid defaults unchanged", () => {
		expect(clampDownloadSettings(DEFAULT_DOWNLOAD_SETTINGS)).toEqual(DEFAULT_DOWNLOAD_SETTINGS);
	});
});

describe("parseDownloadNumber", () => {
	it("reads a number from field text", () => {
		expect(parseDownloadNumber("16")).toBe(16);
		expect(parseDownloadNumber(" 4 ")).toBe(4);
	});

	it("returns null for empty or non-number text", () => {
		expect(parseDownloadNumber("")).toBeNull();
		expect(parseDownloadNumber("   ")).toBeNull();
		expect(parseDownloadNumber("abc")).toBeNull();
	});
});

describe("withDownloadNumber", () => {
	it("returns the settings with the clamped number", () => {
		const next = withDownloadNumber(DEFAULT_DOWNLOAD_SETTINGS, "connections", "64");
		expect(next?.connections).toBe(32);
		expect(next?.maxActive).toBe(DEFAULT_DOWNLOAD_SETTINGS.maxActive);
	});

	it("returns null when the text gives no number, so nothing is saved", () => {
		expect(withDownloadNumber(DEFAULT_DOWNLOAD_SETTINGS, "retries", "")).toBeNull();
	});

	it("keeps 0 as unlimited for the speed limit", () => {
		const next = withDownloadNumber(DEFAULT_DOWNLOAD_SETTINGS, "speedLimitKBps", "0");
		expect(next?.speedLimitKBps).toBe(0);
	});
});

describe("save delay", () => {
	it("uses the same 300 ms pause as the card size", () => {
		expect(DOWNLOAD_SAVE_DELAY_MS).toBe(300);
	});
});
