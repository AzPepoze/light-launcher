import { getGamePath } from "@lib/gameUtils";
import type { GameInfo, ScannedFolderGroup } from "@shared";

/** A scanned folder on the Home page. `loading` is true until the folder's result arrives. */
export type HomeFolderGroup = ScannedFolderGroup & { loading: boolean };

/** Cover fields a library game may carry. */
export interface LibraryCoverFields {
	id?: string | null;
	cover?: string | null;
	coverLocal?: string | null;
}

export type CoveredGame = GameInfo & LibraryCoverFields;

/** Lower-case path with forward slashes and no trailing slash, for matching paths. */
export function pathMatchKey(path: string): string {
	return path.trim().replace(/\\/g, "/").replace(/\/+$/, "").toLowerCase();
}

/** Copies id, cover and coverLocal from the library game with the same path onto each scanned game. */
export function withLibraryCovers(
	groups: HomeFolderGroup[],
	libraryGames: CoveredGame[]
): HomeFolderGroup[] {
	const libraryByPath = new Map<string, CoveredGame>();
	for (const libraryGame of libraryGames) {
		const key = pathMatchKey(getGamePath(libraryGame));
		if (key) libraryByPath.set(key, libraryGame);
	}
	return groups.map((group) => ({ ...group, games: group.games.map((game) => joinCovers(game, libraryByPath)) }));
}

function joinCovers(game: GameInfo, libraryByPath: Map<string, CoveredGame>): CoveredGame {
	const libraryGame = libraryByPath.get(pathMatchKey(getGamePath(game)));
	if (!libraryGame) return game;
	return { ...game, id: libraryGame.id, cover: libraryGame.cover, coverLocal: libraryGame.coverLocal };
}

/** Last segment of a folder path, for Windows and POSIX separators alike. */
export function folderBaseName(folderPath: string): string {
	const parts = folderPath.split(/[\\/]/).filter((part) => part !== "");
	return parts[parts.length - 1] ?? folderPath;
}

/** Adds an empty loading group for each folder that is not listed yet. */
export function addLoadingGroups(groups: HomeFolderGroup[], folderPaths: string[]): HomeFolderGroup[] {
	const listed = new Set(groups.map((group) => group.folderPath));
	const added = folderPaths
		.filter((folderPath) => !listed.has(folderPath))
		.map((folderPath) => ({
			folderPath,
			folderName: folderBaseName(folderPath),
			games: [],
			loading: true
		}));
	return [...groups, ...added];
}

/** Puts a finished folder in place of its group with the same folderPath, or appends it. */
export function upsertGroup(groups: HomeFolderGroup[], group: ScannedFolderGroup): HomeFolderGroup[] {
	const loaded: HomeFolderGroup = { ...group, loading: false };
	if (!groups.some((existing) => existing.folderPath === group.folderPath)) {
		return [...groups, loaded];
	}
	return groups.map((existing) => (existing.folderPath === group.folderPath ? loaded : existing));
}

/** Marks every folder of a full scan result as loaded. */
export function loadedGroups(groups: ScannedFolderGroup[]): HomeFolderGroup[] {
	return groups.map((group) => ({ ...group, loading: false }));
}
