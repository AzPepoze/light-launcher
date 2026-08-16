import * as service from "@lib/homeService";

export class SelectionState {
	isSelectionMode = $state(false);
	selectedPaths = $state(new Set<string>());
	lastSelectedPath = $state("");
	showBulkRemoveModal = $state(false);

	getVisibleGames: () => any[];

	constructor(getVisibleGames: () => any[]) {
		this.getVisibleGames = getVisibleGames;
	}

	toggleSelectionMode() {
		this.isSelectionMode = !this.isSelectionMode;
		if (!this.isSelectionMode) {
			this.selectedPaths = new Set<string>();
			this.lastSelectedPath = "";
		}
	}

	toggleGameSelection(game: any, shiftKey: boolean = false) {
		const path = game.path || game.config.LauncherPath;

		if (shiftKey && this.lastSelectedPath) {
			const visible = this.getVisibleGames();
			const lastIdx = visible.findIndex(
				(g) => (g.path || g.config.LauncherPath) === this.lastSelectedPath
			);
			const currentIdx = visible.findIndex((g) => (g.path || g.config.LauncherPath) === path);

			if (lastIdx !== -1 && currentIdx !== -1) {
				const start = Math.min(lastIdx, currentIdx);
				const end = Math.max(lastIdx, currentIdx);
				const shouldSelect = this.selectedPaths.has(this.lastSelectedPath);

				for (let i = start; i <= end; i++) {
					const p = visible[i].path || visible[i].config.LauncherPath;
					if (shouldSelect) {
						this.selectedPaths.add(p);
					} else {
						this.selectedPaths.delete(p);
					}
				}
				this.selectedPaths = new Set(this.selectedPaths);
				this.lastSelectedPath = path;
				return;
			}
		}

		if (this.selectedPaths.has(path)) {
			this.selectedPaths.delete(path);
		} else {
			this.selectedPaths.add(path);
		}
		this.selectedPaths = new Set(this.selectedPaths);
		this.lastSelectedPath = path;
	}

	handleBulkRemove() {
		if (this.selectedPaths.size === 0) return;
		this.showBulkRemoveModal = true;
	}

	async confirmBulkRemove(onSuccess: () => void) {
		const count = await service.removeGamesBulk(this.selectedPaths);
		if (count > 0) {
			this.selectedPaths = new Set<string>();
			this.isSelectionMode = false;
			this.showBulkRemoveModal = false;
			onSuccess();
		}
	}
}
