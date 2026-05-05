import { GetExeIcon } from "@bindings/light-launcher/internal/app/app";

/**
 * Loads an executable icon asynchronously
 * Returns the icon data URL on success, empty string on failure
 */
export async function loadExeIcon(filePath: string): Promise<string> {
	if (!filePath) return "";

	try {
		const icon = await GetExeIcon(filePath);
		return icon || "";
	} catch (err) {
		console.error("Failed to load exe icon:", err);
		return "";
	}
}

/**
 * Loads multiple exe icons in parallel
 */
export async function loadExeIcons(...filePaths: string[]): Promise<(string | null)[]> {
	return Promise.all(filePaths.map((path) => loadExeIcon(path).catch(() => "")));
}
