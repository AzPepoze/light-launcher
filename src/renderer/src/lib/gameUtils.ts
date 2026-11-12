import type { LaunchOptions } from "@shared";

/** Minimal shape shared by library entries, launch configs, and scan results. */
export interface GamePathSource {
	path?: string | null;
	config?: Partial<Pick<LaunchOptions, "LauncherPath" | "GamePath" | "Name" | "ID">> | null;
	name?: string | null;
	Name?: string | null;
	id?: string | null;
	ID?: string | null;
	LauncherPath?: string | null;
	GamePath?: string | null;
}

/**
 * Resolves the primary executable path. The library entry path wins over the
 * launch config, which wins over legacy flat config fields.
 */
export function getGamePath(game: GamePathSource | null | undefined): string {
	if (!game) return "";
	return (
		game.path ||
		game.config?.LauncherPath ||
		game.config?.GamePath ||
		game.LauncherPath ||
		game.GamePath ||
		""
	);
}

/** Resolves the display name for a game. */
export function getGameName(game: GamePathSource | null | undefined): string {
	if (!game) return "";
	return game.name || game.config?.Name || game.Name || "Unknown Game";
}

/** Resolves the unique identifier for a game. */
export function getGameId(game: GamePathSource | null | undefined): string {
	if (!game) return "";
	return game.id || game.config?.ID || game.ID || getGamePath(game);
}
