import { GetAutoScannedGames, onEvent } from "@lib/api";
import * as service from "@lib/homeService";
import { navigationCommand } from "@stores/navigationStore";
import { notifications } from "@stores/notificationStore";
import { runState } from "@stores/runState";
import { IconLoaderState } from "./IconLoaderState.svelte";
import { SelectionState } from "./SelectionState.svelte";

export class HomePageState {
	games = $state<any[]>([]);
	scannedFolderGroups = $state<any[]>([]);
	sessions = $state<any[]>([]);
	prefixes = $state<string[]>(["All Prefixes"]);
	selectedPrefixFilter = $state("All Prefixes");
	sessionInterval: any = null;
	showHelpModal = $state(false);
	showAddModal = $state(false);
	currentView = $state<"grid" | "list-grid" | "sidebar-grid">("grid");
	searchQuery = $state("");

	// Sub-states
	icons = new IconLoaderState();
	selection = new SelectionState(() => this.getVisibleGames());

	filteredGames = $derived.by(() => {
		return this.games.filter((game) => {
			if (game.isAutoScanned) return false;
			const matchesSearch = game.name.toLowerCase().includes(this.searchQuery.toLowerCase());
			const matchesPrefix =
				this.selectedPrefixFilter === "All Prefixes" ||
				game.config.PrefixPath.endsWith("/" + this.selectedPrefixFilter) ||
				game.config.PrefixPath.endsWith("\\" + this.selectedPrefixFilter);
			return matchesSearch && matchesPrefix;
		});
	});

	filteredScannedFolderGroups = $derived.by(() => {
		return this.scannedFolderGroups
			.map((group) => {
				const filteredGames = group.games.filter((game: any) => {
					const matchesSearch = game.name.toLowerCase().includes(this.searchQuery.toLowerCase());
					const matchesPrefix = this.selectedPrefixFilter === "All Prefixes";
					return matchesSearch && matchesPrefix;
				});
				return {
					...group,
					games: filteredGames
				};
			})
			.filter((group) => {
				if (this.searchQuery) {
					return group.games.length > 0;
				}
				return true;
			});
	});

	dropUnsubscribe: (() => void) | null = null;

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
	}

	initialize() {
		this.refreshData(true);

		this.dropUnsubscribe = onEvent("FilesDropped", async (event: any) => {
			const files = (event?.data || event) as string[];
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
			options: game.config
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

	getVisibleGames() {
		const visible: any[] = [...this.filteredGames];
		for (const group of this.filteredScannedFolderGroups) {
			visible.push(...group.games);
		}
		return visible;
	}
}
