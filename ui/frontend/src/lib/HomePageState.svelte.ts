import {
	KillSession,
	RunGame,
	RemoveGame,
	GetAutoScannedGames,
} from "@bindings/light-launcher/internal/app/app";
import { Events } from "@wailsio/runtime";
import { notifications } from "@stores/notificationStore";
import { navigationCommand } from "@stores/navigationStore";
import { runState } from "@stores/runState";
import { loadExeIcon } from "@lib/iconService";
import * as service from "@lib/homeService";

export class HomePageState {
	games = $state<any[]>([]);
	scannedFolderGroups = $state<any[]>([]);
	sessions = $state<any[]>([]);
	prefixes = $state<string[]>(["All Prefixes"]);
	selectedPrefixFilter = $state("All Prefixes");
	sessionInterval: any = null;
	gameIcons = $state<Record<string, string>>({});
	showHelpModal = $state(false);
	showAddModal = $state(false);
	showBulkRemoveModal = $state(false);
	currentView = $state<"grid" | "list-grid" | "sidebar-grid">("grid");
	searchQuery = $state("");

	isSelectionMode = $state(false);
	selectedPaths = $state(new Set<string>());
	lastSelectedPath = $state("");

	filteredGames = $derived.by(() => {
		return this.games.filter((game) => {
			const matchesSearch = game.name
				.toLowerCase()
				.includes(this.searchQuery.toLowerCase());
			const matchesPrefix =
				this.selectedPrefixFilter === "All Prefixes" ||
				game.config.PrefixPath.endsWith("/" + this.selectedPrefixFilter) ||
				game.config.PrefixPath.endsWith("\\" + this.selectedPrefixFilter);
			return matchesSearch && matchesPrefix;
		});
	});

	filteredScannedFolderGroups = $derived.by(() => {
		return this.scannedFolderGroups.map(group => {
			const filteredGames = group.games.filter((game: any) => {
				const matchesSearch = game.name
					.toLowerCase()
					.includes(this.searchQuery.toLowerCase());
				const matchesPrefix = this.selectedPrefixFilter === "All Prefixes";
				return matchesSearch && matchesPrefix;
			});
			return {
				...group,
				games: filteredGames
			};
		}).filter(group => {
			if (this.searchQuery) {
				return group.games.length > 0;
			}
			return true;
		});
	});

	dropUnsubscribe: (() => void) | null = null;
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

	async refreshData(forceScan = false) {
		const shouldScan = forceScan || this.scannedFolderGroups.length === 0;

		const [data, scannedGroups] = await Promise.all([
			service.refreshHomeData(),
			shouldScan ? GetAutoScannedGames() : Promise.resolve(this.scannedFolderGroups)
		]);

		this.games = data.games;
		this.sessions = data.sessions;
		this.prefixes = data.prefixes;
		this.scannedFolderGroups = scannedGroups || [];

		// Fetch icons for games
		for (const game of this.games) {
			const path = game.path || game.config.LauncherPath;
			if (path) {
				this.enqueueIconLoad(path);
			}
		}

		// Also fetch icons for scanned games
		for (const group of this.scannedFolderGroups) {
			for (const game of group.games) {
				const path = game.path || game.config.LauncherPath;
				if (path) {
					this.enqueueIconLoad(path);
				}
			}
		}
	}

	initialize() {
		this.refreshData(true);

		this.dropUnsubscribe = Events.On("FilesDropped", async (event) => {
			const files = event.data as string[];
			const added = await service.processDroppedFiles(files);
			if (added > 0) {
				notifications.add(`Successfully added ${added} game(s)`, "success");
				this.refreshData(true);
			}
		});

		this.sessionInterval = setInterval(() => this.refreshData(false), 3000);
	}

	destroy() {
		if (this.sessionInterval) clearInterval(this.sessionInterval);
		if (this.dropUnsubscribe) this.dropUnsubscribe();
	}

	async handleQuickLaunch(game: any) {
		try {
			await service.quickLaunchGame(game);
			this.refreshData();
		} catch (err) {
			// Error handled in service
		}
	}

	handleConfigure(game: any) {
		runState.update((s) => ({
			...s,
			options: game.config,
		}));
		navigationCommand.set({ page: "run" });
	}

	isGameRunning(game: any, sessionsList: any[]) {
		const path = game.path || game.config.LauncherPath;
		return sessionsList.some((s) => s.gamePath === path);
	}

	async handleKillSession(pid: number, name: string) {
		try {
			await service.terminateSession(pid, name);
			this.refreshData();
		} catch (err) {
			// Error handled in service
		}
	}

	toggleSelectionMode() {
		this.isSelectionMode = !this.isSelectionMode;
		if (!this.isSelectionMode) {
			this.selectedPaths = new Set<string>();
			this.lastSelectedPath = "";
		}
	}

	getVisibleGames() {
		const visible: any[] = [...this.filteredGames];
		for (const group of this.filteredScannedFolderGroups) {
			visible.push(...group.games);
		}
		return visible;
	}

	toggleGameSelection(game: any, shiftKey: boolean = false) {
		const path = game.path || game.config.LauncherPath;

		if (shiftKey && this.lastSelectedPath) {
			const visible = this.getVisibleGames();
			const lastIdx = visible.findIndex((g) => (g.path || g.config.LauncherPath) === this.lastSelectedPath);
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

	async handleBulkRemove() {
		if (this.selectedPaths.size === 0) return;
		this.showBulkRemoveModal = true;
	}

	async confirmBulkRemove() {
		const count = await service.removeGamesBulk(this.selectedPaths);
		if (count > 0) {
			this.selectedPaths = new Set<string>();
			this.isSelectionMode = false;
			this.showBulkRemoveModal = false;
			this.refreshData();
		}
	}
}
