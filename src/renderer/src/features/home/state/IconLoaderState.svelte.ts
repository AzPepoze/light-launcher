import {
	loadGameIcon,
	needsExeIcon,
	type CoverFields,
	type IconPriority
} from "@lib/iconService";
import { createLogger } from "@lib/logger";
import { getGamePath } from "@lib/gameUtils";
import type { GameInfo } from "@shared";
import { IconLoadSharing } from "./iconLoadSharing";

const log = createLogger("IconLoader");

export class IconLoaderState {
	gameIcons = $state<Record<string, string>>({});
	iconSources = new Map<string, string>();
	private iconLoads = new IconLoadSharing();

	/**
	 * Loads an icon for the path and returns the shared pending promise.
	 *
	 * A path that is already queued or in flight (for example by `warm`) returns
	 * the same promise, so callers wait for the real outcome instead of reading an
	 * immediate "no icon" and flashing the fallback.
	 */
	enqueueIconLoad(
		gamePath: string,
		customIconPath: string | null = null,
		priority: IconPriority = "high",
		hasCover = false
	): Promise<void> {
		const requestedSource = customIconPath ? `custom:${customIconPath}` : `exe:${gamePath}`;

		if (this.gameIcons[gamePath] && this.iconSources.get(gamePath) === requestedSource) {
			return Promise.resolve();
		}

		return this.iconLoads.share(gamePath, () =>
			this.loadIcon(gamePath, customIconPath, requestedSource, priority, hasCover)
		);
	}

	private async loadIcon(
		gamePath: string,
		customIconPath: string | null,
		requestedSource: string,
		priority: IconPriority,
		hasCover: boolean
	): Promise<void> {
		try {
			const { icon, source } = await loadGameIcon(gamePath, customIconPath, priority, hasCover);

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
		}
	}

	/** Pre-extracts exe icons for games without a cover; the queue pauses during user scroll. */
	warm(games: (GameInfo & CoverFields)[]) {
		for (const game of games) {
			const gamePath = getGamePath(game);
			if (gamePath && needsExeIcon(game)) {
				void this.enqueueIconLoad(gamePath, game.config?.CustomIconPath ?? null, "low");
			}
		}
	}
}
