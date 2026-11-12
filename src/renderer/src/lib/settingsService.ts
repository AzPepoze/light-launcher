import {
	GetAppSettings,
	SaveAppSettings,
	RestartApp,
	PickFileCustom,
	PickFolder,
	AddScanFolder,
	RemoveScanFolder,
	UpdateScanFolderConfig,
	UnblacklistGame,
	IsDir,
	GetAutoScannedGames
} from "@lib/api";
import { createLogger } from "./logger";
import type { AppSettings, ScannedFolderGroup } from "@shared";

const log = createLogger("settingsService");

/**
 * Loads application settings from the backend. Null means the call failed; the
 * caller keeps whatever defaults it already shows.
 */
export async function loadAppSettings(): Promise<AppSettings | null> {
	try {
		return await GetAppSettings();
	} catch (err) {
		log.error("Failed to load app settings", err);
		return null;
	}
}

export function saveAppSettings(settings: AppSettings): Promise<void> {
	return SaveAppSettings(settings);
}

export function restartApp(): Promise<void> {
	return RestartApp();
}

/** Opens the background image picker. Null means the user cancelled. */
export async function pickBackgroundImage(): Promise<string | null> {
	const path = await PickFileCustom("Select Background Image", [
		{
			displayName: "Images",
			pattern: "*.png;*.jpg;*.jpeg;*.svg;*.webp"
		}
	]);
	return path || null;
}

/** Opens the folder picker for the Wine prefix location. Null means the user cancelled. */
export async function pickPrefixDirectory(): Promise<string | null> {
	const path = await PickFolder();
	return path || null;
}

export function loadScannedFolders(force: boolean): Promise<ScannedFolderGroup[]> {
	return GetAutoScannedGames(force);
}

export function isDirectory(path: string): Promise<boolean> {
	return IsDir(path);
}

export function addScanFolder(folderPath: string): Promise<void> {
	return AddScanFolder(folderPath);
}

export function removeScanFolder(folderPath: string): Promise<void> {
	return RemoveScanFolder(folderPath);
}

export function updateScanFolderConfig(
	folderPath: string,
	depth: number,
	excludeNames: string[]
): Promise<void> {
	return UpdateScanFolderConfig(folderPath, depth, excludeNames);
}

export function unblacklistGame(path: string): Promise<void> {
	return UnblacklistGame(path);
}
