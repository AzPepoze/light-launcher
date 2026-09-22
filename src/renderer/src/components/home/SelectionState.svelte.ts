import * as service from "@lib/homeService";
import { getGamePath } from "@lib/gameUtils";

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

	toggleGameSelection(game: any, shiftKey: boolean = false, ctrlKey: boolean = false) {
		const path = getGamePath(game);

		// Ctrl+click only ever adds, so it never flips an existing selection off.
		if (ctrlKey && !shiftKey) {
			if (!this.selectedPaths.has(path)) {
				this.selectedPaths.add(path);
				this.selectedPaths = new Set(this.selectedPaths);
			}
			this.lastSelectedPath = path;
			return;
		}

		if (shiftKey && this.lastSelectedPath) {
			const visible = this.getVisibleGames();
			const lastIdx = visible.findIndex((g) => getGamePath(g) === this.lastSelectedPath);
			const currentIdx = visible.findIndex((g) => getGamePath(g) === path);

			if (lastIdx !== -1 && currentIdx !== -1) {
				const start = Math.min(lastIdx, currentIdx);
				const end = Math.max(lastIdx, currentIdx);
				const shouldSelect = this.selectedPaths.has(this.lastSelectedPath);

				for (let i = start; i <= end; i++) {
					const p = getGamePath(visible[i]);
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

	applyMarquee(paths: string[], additive: boolean) {
		const next = additive ? new Set(this.selectedPaths) : new Set<string>();
		for (const p of paths) next.add(p);
		this.selectedPaths = next;
		if (paths.length > 0) this.lastSelectedPath = paths[paths.length - 1];
	}

	selectAll() {
		const visible = this.getVisibleGames();
		if (visible.length === 0) return;
		const next = new Set<string>();
		for (const g of visible) next.add(getGamePath(g));
		this.selectedPaths = next;
		this.lastSelectedPath = getGamePath(visible[visible.length - 1]);
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
