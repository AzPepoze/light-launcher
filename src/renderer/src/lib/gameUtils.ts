/**
 * Resolves the primary executable or launcher path for a game item or configuration object.
 */
export function getGamePath(game: any): string {
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

/**
 * Resolves the display name for a game.
 */
export function getGameName(game: any): string {
	if (!game) return "";
	return game.name || game.config?.Name || game.Name || "Unknown Game";
}

/**
 * Resolves the unique identifier for a game.
 */
export function getGameId(game: any): string {
	if (!game) return "";
	return game.id || game.config?.ID || game.ID || getGamePath(game);
}
