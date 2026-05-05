import {
	SearchExecutables,
	SaveGameConfig,
	DetectLosslessDll,
} from "@bindings/light-launcher/internal/app/app";
import { notifications } from "@stores/notificationStore";
import { createLaunchOptions } from "./formService";

export interface ScannedExecutable {
	path: string;
	name: string;
	icon: string | null;
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
				icon: null,
			}));
		}
	} catch (error) {
		console.error("Failed to search folder:", error);
		notifications.add(`Failed to search folder: ${error}`, "error");
	}
	return [];
}

/**
 * Registers a game in the library with its initial configuration
 */
export async function registerGame(
	executablePath: string,
	prefixPath: string
): Promise<void> {
	const gameName = executablePath.split("/").pop()?.replace(".exe", "") || "Game";
	
	let losslessDllPath = "";
	try {
		losslessDllPath = await DetectLosslessDll();
	} catch (error) {
		// Dll detection optional
	}

	const gameConfig = createLaunchOptions();
	gameConfig.Name = gameName;
	gameConfig.RunnerPath = executablePath;
	gameConfig.GamePath = executablePath;
	gameConfig.PrefixPath = prefixPath;
	
	if (losslessDllPath) {
		gameConfig.Extras.Lsfg.DllPath = losslessDllPath;
	}
	
	try {
		await SaveGameConfig(gameConfig);
	} catch (error) {
		console.error(`Failed to save game config for ${gameName}:`, error);
		throw error;
	}
}

/**
 * Registers multiple games at once
 */
export async function batchRegisterGames(
	scannedExecutables: ScannedExecutable[],
	discardedPaths: Set<string>,
	targetPrefixPath: string
): Promise<number> {
	const gamesToRegister = scannedExecutables.filter((exe) => !discardedPaths.has(exe.path));
	let successfullyAddedCount = 0;
	
	for (const game of gamesToRegister) {
		try {
			await registerGame(game.path, targetPrefixPath);
			successfullyAddedCount++;
		} catch (error) {
			console.error(`Failed to register ${game.name}:`, error);
		}
	}
	
	if (successfullyAddedCount > 0) {
		notifications.add(`Successfully added ${successfullyAddedCount} games to library`, "success");
	}
	
	return successfullyAddedCount;
}
