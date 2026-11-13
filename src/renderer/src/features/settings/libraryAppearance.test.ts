import { describe, expect, it } from "vitest";
import type { AppSettings } from "@shared";
import {
	LIBRARY_CARD_SIZE_DEFAULT,
	LIBRARY_CARD_SIZE_MAX,
	LIBRARY_CARD_SIZE_MIN,
	blurCoversUntilHover,
	clampLibraryCardSize,
	libraryCardSize
} from "./libraryAppearance";
import { DEFAULT_APP_SETTINGS, mergeAppSettings } from "./settingsData";

function settingsWith(patch: Partial<AppSettings>): AppSettings {
	return { ...DEFAULT_APP_SETTINGS, ...patch };
}

describe("library card size defaults", () => {
	it("uses 160 px and no blur when the setting is missing", () => {
		const settings = settingsWith({ LibraryCardSize: undefined, BlurCoversUntilHover: undefined });
		expect(libraryCardSize(settings)).toBe(160);
		expect(blurCoversUntilHover(settings)).toBe(false);
	});

	it("keeps the documented limits", () => {
		expect(LIBRARY_CARD_SIZE_MIN).toBe(100);
		expect(LIBRARY_CARD_SIZE_MAX).toBe(260);
		expect(LIBRARY_CARD_SIZE_DEFAULT).toBe(160);
	});
});

describe("clampLibraryCardSize", () => {
	it("clamps below and above the range", () => {
		expect(clampLibraryCardSize(10)).toBe(100);
		expect(clampLibraryCardSize(999)).toBe(260);
	});

	it("keeps values in range and rounds to whole px", () => {
		expect(clampLibraryCardSize(180)).toBe(180);
		expect(clampLibraryCardSize(180.4)).toBe(180);
		expect(clampLibraryCardSize(180.6)).toBe(181);
	});

	it("falls back to the default for NaN and infinity", () => {
		expect(clampLibraryCardSize(Number.NaN)).toBe(160);
		expect(clampLibraryCardSize(Number.POSITIVE_INFINITY)).toBe(160);
	});
});

describe("libraryCardSize reader", () => {
	it("clamps a stored value that is out of range", () => {
		expect(libraryCardSize(settingsWith({ LibraryCardSize: 40 }))).toBe(100);
		expect(libraryCardSize(settingsWith({ LibraryCardSize: 400 }))).toBe(260);
	});

	it("returns a stored value in range", () => {
		expect(libraryCardSize(settingsWith({ LibraryCardSize: 220 }))).toBe(220);
	});
});

describe("blurCoversUntilHover reader", () => {
	it("is true only when the setting is true", () => {
		expect(blurCoversUntilHover(settingsWith({ BlurCoversUntilHover: true }))).toBe(true);
		expect(blurCoversUntilHover(settingsWith({ BlurCoversUntilHover: false }))).toBe(false);
	});
});

describe("mergeAppSettings for library settings", () => {
	it("keeps the current values when the loaded file has none", () => {
		const current = settingsWith({ LibraryCardSize: 200, BlurCoversUntilHover: true });
		const merged = mergeAppSettings(settingsWith({ LibraryCardSize: undefined, BlurCoversUntilHover: undefined }), current);
		expect(merged.LibraryCardSize).toBe(200);
		expect(merged.BlurCoversUntilHover).toBe(true);
	});

	it("takes the loaded values when present", () => {
		const current = settingsWith({ LibraryCardSize: 200, BlurCoversUntilHover: true });
		const merged = mergeAppSettings(settingsWith({ LibraryCardSize: 130, BlurCoversUntilHover: false }), current);
		expect(merged.LibraryCardSize).toBe(130);
		expect(merged.BlurCoversUntilHover).toBe(false);
	});

	it("fills the defaults when neither side has the value", () => {
		const current = settingsWith({ LibraryCardSize: undefined, BlurCoversUntilHover: undefined });
		const merged = mergeAppSettings(settingsWith({ LibraryCardSize: undefined, BlurCoversUntilHover: undefined }), current);
		expect(merged.LibraryCardSize).toBe(160);
		expect(merged.BlurCoversUntilHover).toBe(false);
	});

	it("has the defaults in the page defaults", () => {
		expect(DEFAULT_APP_SETTINGS.LibraryCardSize).toBe(160);
		expect(DEFAULT_APP_SETTINGS.BlurCoversUntilHover).toBe(false);
	});
});
