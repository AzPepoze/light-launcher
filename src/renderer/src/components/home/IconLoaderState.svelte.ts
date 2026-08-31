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

	async enqueueIconLoad(gamePath: string, customIconPath?: string | null) {
		const isExplicitSync = customIconPath !== undefined;
		const requestedSource = customIconPath ? `custom:${customIconPath}` : `exe:${gamePath}`;
		const exeSource = `exe:${gamePath}`;
		const loadingKey = gamePath;

		if (!isExplicitSync && this.iconSources.get(gamePath)?.startsWith("custom:")) return;
		if (this.gameIcons[gamePath] && this.iconSources.get(gamePath) === requestedSource) return;
		if (this.loadingIcons.has(loadingKey)) return;

		this.loadingIcons.add(loadingKey);
		try {
			let icon = "";
			let resolvedSource: string | null = null;

			if (customIconPath) {
				try {
					icon = (await GetImageBase64(customIconPath)) || "";
					if (icon) resolvedSource = requestedSource;
				} catch {
					// custom file missing — clear stale marker and fall through to exe
				}
			}

			if (!icon) {
				icon = (await loadExeIcon(gamePath)) || "";
				if (icon) resolvedSource = exeSource;
			}

			if (icon && resolvedSource) {
				this.gameIcons[gamePath] = icon;
				this.iconSources.set(gamePath, resolvedSource);
			} else if (customIconPath && this.iconSources.get(gamePath) === requestedSource) {
				this.iconSources.delete(gamePath);
			}
		} catch (error) {
			console.error("Queue icon load error:", error);
		} finally {
			this.loadingIcons.delete(loadingKey);
		}
	}
}
