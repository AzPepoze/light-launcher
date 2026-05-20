import { loadExeIcon } from "@lib/iconService";

export class IconLoaderState {
	gameIcons = $state<Record<string, string>>({});
	loadingIcons = new Set<string>();
	iconQueue: string[] = [];
	isProcessingIconQueue = false;

	enqueueIconLoad(path: string) {
		if (this.gameIcons[path] || this.loadingIcons.has(path)) {
			return;
		}
		this.loadingIcons.add(path);
		this.iconQueue.push(path);
		this.processIconQueue();
	}

	async processIconQueue() {
		if (this.isProcessingIconQueue) return;
		this.isProcessingIconQueue = true;

		while (this.iconQueue.length > 0) {
			const path = this.iconQueue.shift();
			if (!path) continue;

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
			// Small delay between icon loads to keep the system responsive
			await new Promise((resolve) => setTimeout(resolve, 50));
		}

		this.isProcessingIconQueue = false;
	}
}
