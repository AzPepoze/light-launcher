import { 
	GetAppSettings, 
	SaveAppSettings, 
	RestartApp,
	PickFileCustom
} from "@bindings/light-launcher/internal/app/app";
import { notifications } from "@stores/notificationStore";
import { settingsStore } from "@stores/settingsStore";

/**
 * Loads application settings from the backend
 */
export async function loadAppSettings(): Promise<any> {
	try {
		const settings = await GetAppSettings();
		return settings;
	} catch (err) {
		console.error("Failed to load app settings", err);
		return null;
	}
}

/**
 * Toggles the transparent window mode and restarts the app
 */
export async function toggleTransparentMode(currentSettings: any): Promise<void> {
	try {
		const newSettings = { ...currentSettings, TransparentMode: !currentSettings.TransparentMode };
		await SaveAppSettings(newSettings);
		notifications.add("Restarting app to apply transparency changes...", "info");
		setTimeout(async () => {
			await RestartApp();
		}, 1500);
	} catch (err) {
		notifications.add("Failed to save setting", "error");
		throw err;
	}
}

/**
 * Opens a file picker for background images
 */
export async function browseBackgroundImage(): Promise<string | null> {
	try {
		const path = await PickFileCustom("Select Background Image", [
			{
				DisplayName: "Images",
				Pattern: "*.png;*.jpg;*.jpeg;*.svg;*.webp",
			},
		]);
		if (path) {
			settingsStore.update((s) => ({
				...s,
				backgroundImagePath: path,
			}));
			notifications.add("Background image updated", "success");
			return path;
		}
	} catch (err) {
		notifications.add("Failed to pick image", "error");
	}
	return null;
}

/**
 * Clears the custom background image
 */
export function clearBackgroundImage(): void {
	settingsStore.update((s) => ({
		...s,
		backgroundImagePath: "",
	}));
	notifications.add("Background image cleared", "info");
}

/**
 * Toggles the theme between light and dark
 */
export function toggleTheme(): void {
	settingsStore.update((s) => ({
		...s,
		theme: s.theme === "light" ? "dark" : "light",
	}));
}

/**
 * Updates the window transparency level
 */
export function updateTransparency(value: number): void {
	settingsStore.update((s) => ({
		...s,
		transparency: value,
	}));
}
