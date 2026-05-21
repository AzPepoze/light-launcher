import { loadExeIcon } from "@lib/iconService";

export class IconLoaderState {
	gameIcons = $state<Record<string, string>>({});
	loadingIcons = new Set<string>();

	enqueueIconLoad(path: string) {
		if (this.gameIcons[path] || this.loadingIcons.has(path)) {
			return;
		}
		this.loadingIcons.add(path);

		(async () => {
			try {
				const icon = await loadExeIcon(path);
				if (icon) {
					this.gameIcons[path] = icon;
				}
			} catch (err) {
				console.error("Queue icon load error:", err);
			} finally {
				this.loadingIcons.delete(path);
			}
		})();
	}
}
