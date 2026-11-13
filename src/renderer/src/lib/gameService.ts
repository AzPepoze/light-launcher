import { SearchExecutables, SaveGameConfig, DetectLosslessDll } from "@lib/api";
import { notifications } from "@stores/notificationStore";
import { plainMessage } from "./plainMessage";
import { createLaunchOptions } from "./formService";
import { createLogger } from "./logger";

const log = createLogger("gameService");

export interface ScannedExecutable {
	path: string;
	name: string;
	icon: string | null;
	alreadyExists?: boolean;
}

/**
 * Scans a folder for game executables
 */
export async function scanFolderForExecutables(
	folderPath: string,
	recursionDepth: number,
	excludedNames: string[]
): Promise<ScannedExecutable[]> {
	try {
		const executables = await SearchExecutables(folderPath, recursionDepth, excludedNames);
		if (executables && executables.length > 0) {
			return executables.map((path) => ({
				path,
				name: path.split("/").pop()?.replace(".exe", "") || "Game",
				icon: null
			}));
		}
	} catch (error) {
		log.error("Failed to search folder", error);
		notifications.add(`Failed to search folder: ${plainMessage(error)}`, "error");
	}
	return [];
}

/**
 * Registers a game in the library with its initial configuration
 */
export async function registerGame(executablePath: string, prefixPath: string): Promise<void> {
	const gameName = executablePath.split("/").pop()?.replace(".exe", "") || "Game";

	let losslessDllPath = "";
	try {
		losslessDllPath = await DetectLosslessDll();
	} catch (error) {
		// Dll detection optional
	}

	const gameConfig = createLaunchOptions();
	gameConfig.Name = gameName;
	gameConfig.LauncherPath = executablePath;
	gameConfig.GamePath = executablePath;
	gameConfig.PrefixPath = prefixPath;

	if (losslessDllPath) {
		gameConfig.Extras.Lsfg.DllPath = losslessDllPath;
	}

	try {
		await SaveGameConfig(gameConfig);
	} catch (error) {
		log.error(`Failed to save game config for ${gameName}`, error);
		throw error;
	}
}
