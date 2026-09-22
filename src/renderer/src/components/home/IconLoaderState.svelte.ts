import { GetImageBase64 } from "@lib/api";
import { loadExeIcon } from "@lib/iconService";
import { createLogger } from "@lib/logger";

const log = createLogger("IconLoader");

export class IconLoaderState {
	gameIcons = $state<Record<string, string>>({});
	loadingIcons = new Set<string>();
	iconSources = new Map<string, string>();

	async enqueueIconLoad(gamePath: string, customIconPath: string | null = null) {
		const requestedSource = customIconPath ? `custom:${customIconPath}` : `exe:${gamePath}`;

		if (this.gameIcons[gamePath] && this.iconSources.get(gamePath) === requestedSource) return;
		if (this.loadingIcons.has(gamePath)) return;

		this.loadingIcons.add(gamePath);
		try {
			let icon = "";
			let resolvedSource: string | null = null;

			if (customIconPath) {
				try {
					icon = (await GetImageBase64(customIconPath)) || "";
					if (icon) resolvedSource = requestedSource;
				} catch {
					// Custom file missing; fall through to the exe icon.
				}
			}

			if (!icon) {
				icon = (await loadExeIcon(gamePath)) || "";
				if (icon) resolvedSource = `exe:${gamePath}`;
			}

			if (icon && resolvedSource) {
				this.gameIcons[gamePath] = icon;
				this.iconSources.set(gamePath, resolvedSource);
			} else if (customIconPath && this.iconSources.get(gamePath) === requestedSource) {
				this.iconSources.delete(gamePath);
			}
		} catch (error) {
			log.error("Queue icon load error", error);
		} finally {
			this.loadingIcons.delete(gamePath);
		}
	}
}
