import type { LaunchOptions } from "@shared";

/** Minimal shape shared by library entries, launch configs, and scan results. */
export interface GamePathSource {
	path?: string | null;
	/** The library install folder. */
	installDir?: string | null;
	exePath?: string | null;
	config?: Partial<Pick<LaunchOptions, "LauncherPath" | "GamePath" | "Name" | "ID">> | null;
	localConfig?: Partial<Pick<LaunchOptions, "LauncherPath" | "GamePath">> | null;
	name?: string | null;
	Name?: string | null;
	id?: string | number | null;
	ID?: string | null;
	LauncherPath?: string | null;
	GamePath?: string | null;
}

/**
 * Resolves the game's folder or executable path, in this order: the library
 * entry path, the install folder, the exe, the launch config, then legacy flat fields.
 */
export function getGamePath(game: GamePathSource | null | undefined): string {
	if (!game) return "";
	return (
		game.path ||
		game.installDir ||
		game.exePath ||
		game.config?.LauncherPath ||
		game.config?.GamePath ||
		game.localConfig?.LauncherPath ||
		game.localConfig?.GamePath ||
		game.LauncherPath ||
		game.GamePath ||
		""
	);
}

/** Makes paths comparable: forward slashes, lower case, no surrounding spaces. */
export function normalizeGamePath(path: string): string {
	return path.replace(/\\/g, "/").toLowerCase().trim();
}

/** Resolves the unique identifier for a game. */
export function getGameId(game: GamePathSource | null | undefined): string {
	if (!game) return "";
	return String(game.id || game.config?.ID || game.ID || getGamePath(game));
}
