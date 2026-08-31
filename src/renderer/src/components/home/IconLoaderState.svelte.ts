import { GetImageBase64 } from "@lib/api";
import { loadExeIcon } from "@lib/iconService";
import type { GameInfo } from "@shared";

export class IconLoaderState {
	gameIcons = $state<Record<string, string>>({});
	loadingIcons = new Set<string>();
	iconSources = new Map<string, string>();

	async syncGames(games: GameInfo[]) {
		for (const game of games) {
			const gamePath = game?.path || game?.config?.LauncherPath;
			if (!gamePath) continue;
			await this.enqueueIconLoad(gamePath, game?.config?.CustomIconPath || null);
		}
	}

	async enqueueIconLoad(path: string, customIconPath?: string | null) {
		const explicitSource = customIconPath !== undefined;
		const source = customIconPath ? `custom:${customIconPath}` : `exe:${path}`;

		// Lazy card requests do not know the custom path. If syncGames already loaded a
		// custom icon, keep it instead of replacing it with the executable icon.
		if (!explicitSource && this.iconSources.get(path)?.startsWith("custom:")) return;
		if (this.gameIcons[path] && this.iconSources.get(path) === source) return;
		if (this.loadingIcons.has(source)) return;

		this.loadingIcons.add(source);
		try {
			let icon = "";
			if (customIconPath) {
				try {
					icon = (await GetImageBase64(customIconPath)) || "";
				} catch {
					// Missing custom files gracefully fall back to the executable icon.
				}
			}
			if (!icon) icon = (await loadExeIcon(path)) || "";

			if (icon) {
				this.gameIcons[path] = icon;
				this.iconSources.set(path, icon && customIconPath ? source : `exe:${path}`);
			}
		} catch (err) {
			console.error("Queue icon load error:", err);
		} finally {
			this.loadingIcons.delete(source);
		}
	}
}
