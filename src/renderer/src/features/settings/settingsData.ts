import type { AppSettings, ScanFolderConfig } from "@shared";
import { folderKey } from "@lib/features/scanFolders";

/** The folder depth the backend uses when a folder is added without settings. */
export const DEFAULT_SCAN_DEPTH = 2;

/** What the page shows before the backend answers, and what fills missing fields. */
export const DEFAULT_APP_SETTINGS: AppSettings = {
	TransparentMode: true,
	NativeWayland: false,
	ScanFolders: [],
	ScanFolderConfigs: [],
	Blacklist: [],
	CustomPrefixDir: "",
	TrackPlaytime: true,
	DiscordRichPresence: true,
	DiscordClientId: ""
};

/**
 * Applies loaded settings over the settings already on screen. A failed load keeps
 * the current values, and null arrays in the file become empty lists.
 */
export function mergeAppSettings(loaded: AppSettings | null, current: AppSettings): AppSettings {
	if (!loaded) return current;
	return {
		...current,
		...loaded,
		ScanFolders: loaded.ScanFolders ?? current.ScanFolders ?? [],
		ScanFolderConfigs: loaded.ScanFolderConfigs ?? current.ScanFolderConfigs ?? [],
		Blacklist: loaded.Blacklist ?? current.Blacklist ?? [],
		CustomPrefixDir: loaded.CustomPrefixDir ?? current.CustomPrefixDir ?? "",
		TrackPlaytime: loaded.TrackPlaytime ?? current.TrackPlaytime ?? true,
		DiscordRichPresence: loaded.DiscordRichPresence ?? current.DiscordRichPresence ?? true,
		DiscordClientId: loaded.DiscordClientId ?? current.DiscordClientId ?? ""
	};
}

/** Builds the settings payload for a save without changing the settings on screen. */
export function appSettingsWith(settings: AppSettings, patch: Partial<AppSettings>): AppSettings {
	return { ...settings, ...patch };
}

/** Adds a watched folder unless the same folder is already in the list. */
export function withAddedScanFolder(settings: AppSettings, path: string): AppSettings {
	const folder = path.trim();
	const key = folderKey(folder);
	if (key === "") return settings;
	const folders = settings.ScanFolders ?? [];
	if (folders.some((item) => folderKey(item) === key)) return settings;
	return { ...settings, ScanFolders: [...folders, folder] };
}

/** Removes a watched folder and its scan settings, matching the folder the way the backend does. */
export function withRemovedScanFolder(settings: AppSettings, path: string): AppSettings {
	const key = folderKey(path);
	if (key === "") return settings;
	return {
		...settings,
		ScanFolders: (settings.ScanFolders ?? []).filter((item) => folderKey(item) !== key),
		ScanFolderConfigs: (settings.ScanFolderConfigs ?? []).filter(
			(config) => folderKey(config.Path) !== key
		)
	};
}

/** Replaces the scan settings of a folder, or adds them when the folder has none yet. */
export function withScanFolderConfig(
	settings: AppSettings,
	config: ScanFolderConfig
): AppSettings {
	const key = folderKey(config.Path);
	const configs = settings.ScanFolderConfigs ?? [];
	const index = configs.findIndex((item) => folderKey(item.Path) === key);
	const next =
		index < 0
			? [...configs, config]
			: configs.map((item, position) => (position === index ? config : item));
	return { ...settings, ScanFolderConfigs: next };
}

/** The saved scan settings of a folder, or the defaults the page shows for a new folder. */
export function scanFolderConfig(settings: AppSettings, folder: string): ScanFolderConfig {
	const key = folderKey(folder);
	const config = (settings.ScanFolderConfigs ?? []).find((item) => folderKey(item.Path) === key);
	return config ?? { Path: folder, Depth: DEFAULT_SCAN_DEPTH, ExcludeNames: [] };
}

export function toggledTheme(theme: "light" | "dark"): "light" | "dark" {
	return theme === "light" ? "dark" : "light";
}
