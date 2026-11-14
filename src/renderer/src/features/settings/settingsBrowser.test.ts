import { describe, expect, it } from "vitest";
import { DEFAULT_APP_SETTINGS, appSettingsWith, mergeAppSettings } from "./settingsData";

describe("Open in browser setting", () => {
	it("is off by default", () => {
		expect(DEFAULT_APP_SETTINGS.OpenInBrowser).toBe(false);
	});

	it("takes the saved value when settings load", () => {
		const loaded = { ...DEFAULT_APP_SETTINGS, OpenInBrowser: true };

		expect(mergeAppSettings(loaded, DEFAULT_APP_SETTINGS).OpenInBrowser).toBe(true);
	});

	it("builds the save payload with only this value changed", () => {
		const next = appSettingsWith(DEFAULT_APP_SETTINGS, { OpenInBrowser: true });

		expect(next.OpenInBrowser).toBe(true);
		expect(next.NativeWayland).toBe(DEFAULT_APP_SETTINGS.NativeWayland);
		expect(DEFAULT_APP_SETTINGS.OpenInBrowser).toBe(false);
	});
});
