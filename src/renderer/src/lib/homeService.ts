import {
	GetAllGames,
	GetConfig,
	GetRunningSessions,
	ListPrefixes,
	RemoveGame,
	GetPrefixBaseDir,
	SaveGameConfig,
	IsDir,
	AddScanFolder
} from "@lib/api";
import { notifications } from "@stores/notificationStore";
import { plainMessage } from "./plainMessage";
import { createLaunchOptions } from "./formService";
import { createLogger } from "./logger";
import { launchGame } from "./gameLaunchService";
import { getGamePath, normalizeGamePath } from "./gameUtils";
import type { GameInfo, RunningSession } from "@shared";

const log = createLogger("homeService");

export interface HomeData {
	games: GameInfo[];
	sessions: RunningSession[];
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
			ListPrefixes()
		]);

		return {
			games: fetchedGames || [],
			sessions: fetchedSessions || [],
			prefixes: ["All Prefixes", ...(fetchedPrefixes || [])]
		};
	} catch (error) {
		log.error("Failed to refresh home data", error);
		return { games: [], sessions: [], prefixes: ["All Prefixes"] };
	}
}

/** Quick launch via the shared pipeline with fresh saved config. */
export async function quickLaunchGame(game: GameInfo, showLogs = false): Promise<void> {
	const gamePath = getGamePath(game);
	let options = game.config;
	if (gamePath) {
		try {
			const fresh = await GetConfig(gamePath);
			if (fresh) options = fresh;
		} catch {
			// Keep stored config when fresh load fails.
		}
	}
	await launchGame(options, { showLogs });
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
		notifications.add(`Failed to remove some games: ${plainMessage(error)}`, "error");
	}
	return removedCount;
}

async function knownGamePaths(): Promise<Set<string>> {
	const games = (await GetAllGames()) ?? [];
	return new Set(games.map((game) => normalizeGamePath(game.path)));
}

/**
 * Processes dropped files and registers them as games.
 * Executables already in the library are skipped, so their saved config is kept.
 */
export async function processDroppedFiles(filePaths: string[]): Promise<number> {
	let addedCount = 0;
	try {
		const basePrefixDirectory = await GetPrefixBaseDir();
		const defaultPrefixPath = `${basePrefixDirectory}/Default`;
		const libraryPaths = await knownGamePaths();

		for (const filePath of filePaths) {
			const isFolder = await IsDir(filePath);
			if (isFolder) {
				await AddScanFolder(filePath);
				addedCount++;
			} else if (filePath.toLowerCase().endsWith(".exe")) {
				const normalizedPath = normalizeGamePath(filePath);
				if (libraryPaths.has(normalizedPath)) continue;

				const gameName = filePath.split("/").pop()?.replace(".exe", "") || "Game";

				const gameConfig = createLaunchOptions();
				gameConfig.Name = gameName;
				gameConfig.LauncherPath = filePath;
				gameConfig.GamePath = filePath;
				gameConfig.PrefixPath = defaultPrefixPath;

				await SaveGameConfig(gameConfig);
				libraryPaths.add(normalizedPath);
				addedCount++;
			}
		}
	} catch (error) {
		log.error("Failed to process dropped files", error);
	}
	return addedCount;
}
