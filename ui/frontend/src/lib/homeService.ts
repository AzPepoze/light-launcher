import {
	GetAllGames,
	GetRunningSessions,
	ListPrefixes,
	RunGame,
	KillSession,
	RemoveGame,
	GetPrefixBaseDir,
	SaveGameConfig,
} from "@bindings/light-launcher/internal/app/app";
import { notifications } from "@stores/notificationStore";
import { createLaunchOptions } from "./formService";

export interface HomeData {
	games: any[];
	sessions: any[];
	prefixes: string[];
}

/**
 * Fetches all necessary data for the Home page
 */
export async function refreshHomeData(): Promise<HomeData> {
	try {
		const [fetchedGames, fetchedSessions, fetchedPrefixes] = await Promise.all([
			GetAllGames(),
			GetRunningSessions(),
			ListPrefixes(),
		]);

		return {
			games: fetchedGames || [],
			sessions: fetchedSessions || [],
			prefixes: ["All Prefixes", ...(fetchedPrefixes || [])],
		};
	} catch (error) {
		console.error("Failed to refresh home data:", error);
		return { games: [], sessions: [], prefixes: ["All Prefixes"] };
	}
}

/**
 * Handles quick launch of a game
 */
export async function quickLaunchGame(game: any): Promise<void> {
	try {
		notifications.add(`Launching ${game.name}...`, "info");
		await RunGame(game.config, false);
	} catch (error) {
		notifications.add(`Launch failed: ${error}`, "error");
		throw error;
	}
}

/**
 * Terminates a running game session
 */
export async function terminateSession(processId: number, gameName: string): Promise<void> {
	try {
		await KillSession(processId);
		notifications.add(`Terminated session: ${gameName}`, "success");
	} catch (error) {
		notifications.add(`Failed to kill session: ${error}`, "error");
		throw error;
	}
}

/**
 * Removes multiple games in bulk
 */
export async function removeGamesBulk(gamePaths: Set<string>): Promise<number> {
	let removedCount = 0;
	try {
		for (const path of gamePaths) {
			await RemoveGame(path);
			removedCount++;
		}
		notifications.add(`Successfully removed ${removedCount} games`, "success");
	} catch (error) {
		notifications.add(`Failed to remove some games: ${error}`, "error");
	}
	return removedCount;
}

/**
 * Processes dropped files and registers them as games
 */
export async function processDroppedFiles(filePaths: string[]): Promise<number> {
	let addedCount = 0;
	try {
		const basePrefixDirectory = await GetPrefixBaseDir();
		const defaultPrefixPath = `${basePrefixDirectory}/Default`;

		for (const filePath of filePaths) {
			if (filePath.toLowerCase().endsWith(".exe")) {
				const gameName = filePath.split("/").pop()?.replace(".exe", "") || "Game";
				
				const gameConfig = createLaunchOptions();
				gameConfig.Name = gameName;
				gameConfig.RunnerPath = filePath;
				gameConfig.GamePath = filePath;
				gameConfig.PrefixPath = defaultPrefixPath;

				await SaveGameConfig(gameConfig);
				addedCount++;
			}
		}
	} catch (error) {
		console.error("Failed to process dropped files:", error);
	}
	return addedCount;
}
