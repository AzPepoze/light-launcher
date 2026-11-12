import { loadGameIcon, type IconPriority } from "@lib/iconService";
import { createLogger } from "@lib/logger";
import { getGamePath } from "@lib/gameUtils";
import type { GameInfo } from "@shared";

const log = createLogger("IconLoader");

export class IconLoaderState {
	gameIcons = $state<Record<string, string>>({});
	loadingIcons = new Set<string>();
	iconSources = new Map<string, string>();

	async enqueueIconLoad(
		gamePath: string,
		customIconPath: string | null = null,
		priority: IconPriority = "high"
	) {
		const requestedSource = customIconPath ? `custom:${customIconPath}` : `exe:${gamePath}`;

		if (this.gameIcons[gamePath] && this.iconSources.get(gamePath) === requestedSource) return;
		if (this.loadingIcons.has(gamePath)) return;

		this.loadingIcons.add(gamePath);
		try {
			const { icon, source } = await loadGameIcon(gamePath, customIconPath, priority);

			if (icon && source) {
				this.gameIcons[gamePath] = icon;
				this.iconSources.set(
					gamePath,
					source === "custom" ? `custom:${customIconPath ?? ""}` : `exe:${gamePath}`
				);
			} else if (customIconPath && this.iconSources.get(gamePath) === requestedSource) {
				this.iconSources.delete(gamePath);
			}
		} catch (error) {
			log.error("Queue icon load error", error);
		} finally {
			this.loadingIcons.delete(gamePath);
		}
	}

	/** Pre-extracts icons for the given games; the queue pauses during user scroll. */
	warm(games: GameInfo[]) {
		for (const game of games) {
			const gamePath = getGamePath(game);
			if (gamePath) {
				void this.enqueueIconLoad(gamePath, game.config?.CustomIconPath ?? null, "low");
			}
		}
	}
}
