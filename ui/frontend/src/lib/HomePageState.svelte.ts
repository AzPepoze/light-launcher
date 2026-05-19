import {
	KillSession,
	RunGame,
	RemoveGame,
} from "@bindings/light-launcher/internal/app/app";
import { Events } from "@wailsio/runtime";
import { notifications } from "@stores/notificationStore";
import { navigationCommand } from "@stores/navigationStore";
import { runState } from "@stores/runState";
import { loadExeIcon } from "@lib/iconService";
import * as service from "@lib/homeService";

export class HomePageState {
	games = $state<any[]>([]);
	sessions = $state<any[]>([]);
	prefixes = $state<string[]>(["All Prefixes"]);
	selectedPrefixFilter = $state("All Prefixes");
	sessionInterval: any = null;
	gameIcons = $state<Record<string, string>>({});
	showHelpModal = $state(false);
	showAddModal = $state(false);
	showBulkRemoveModal = $state(false);
	currentView = $state<"grid" | "list-grid">("grid");
	searchQuery = $state("");

	isSelectionMode = $state(false);
	selectedPaths = $state(new Set<string>());

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

	dropUnsubscribe: (() => void) | null = null;

	async refreshData() {
		const data = await service.refreshHomeData();
		this.games = data.games;
		this.sessions = data.sessions;
		this.prefixes = data.prefixes;

		// Fetch icons for games
		for (const game of this.games) {
			const path = game.path || game.config.LauncherPath;
			if (path && !this.gameIcons[path]) {
				loadExeIcon(path).then((icon) => {
					if (icon) {
						this.gameIcons[path] = icon;
					}
				});
			}
		}
	}

	initialize() {
		this.refreshData();

		this.dropUnsubscribe = Events.On("FilesDropped", async (event) => {
			const files = event.data as string[];
			const added = await service.processDroppedFiles(files);
			if (added > 0) {
				notifications.add(`Successfully added ${added} game(s)`, "success");
				this.refreshData();
			}
		});

		this.sessionInterval = setInterval(() => this.refreshData(), 3000);
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
			this.selectedPaths.clear();
			this.selectedPaths = this.selectedPaths; // trigger reactivity
		}
	}

	toggleGameSelection(game: any) {
		const path = game.path || game.config.LauncherPath;
		if (this.selectedPaths.has(path)) {
			this.selectedPaths.delete(path);
		} else {
			this.selectedPaths.add(path);
		}
		this.selectedPaths = this.selectedPaths; // trigger reactivity
	}

	async handleBulkRemove() {
		if (this.selectedPaths.size === 0) return;
		this.showBulkRemoveModal = true;
	}

	async confirmBulkRemove() {
		const count = await service.removeGamesBulk(this.selectedPaths);
		if (count > 0) {
			this.selectedPaths.clear();
			this.selectedPaths = this.selectedPaths;
			this.isSelectionMode = false;
			this.showBulkRemoveModal = false;
			this.refreshData();
		}
	}
}
