import { GetAutoScannedGames, GetRunningSessions, onEvent } from "@lib/api";
import * as service from "@lib/homeService";
import { createLogger } from "@lib/logger";
import { navigationCommand } from "@stores/navigationStore";
import { notifications } from "@stores/notificationStore";
import { runState } from "@stores/runState";
import { IconLoaderState } from "./IconLoaderState.svelte";
import { SelectionState } from "./SelectionState.svelte";

const log = createLogger("HomePage");

function gameIdentity(game: any): string {
	return game?.path || game?.config?.LauncherPath || game?.config?.GamePath || game?.name || "";
}

function sessionsSignature(sessions: any[]): string {
	return sessions
		.map((s) => `${s?.pid ?? "?"}:${s?.gamePath ?? ""}`)
		.sort()
		.join("|");
}

function gamesSignature(games: any[]): string {
	return games
		.map(
			(game) =>
				`${gameIdentity(game)}|${game?.name ?? ""}|${game?.config?.PrefixPath ?? ""}|${game?.config?.CustomIconPath ?? ""}|${game?.isAutoScanned ? "1" : "0"}`
		)
		.sort()
		.join(";");
}

function groupsSignature(groups: any[]): string {
	return groups
		.map((group) => `${group?.folderPath ?? ""}#${gamesSignature(group?.games || [])}`)
		.sort()
		.join(";");
}

export class HomePageState {
	games = $state<any[]>([]);
	scannedFolderGroups = $state<any[]>([]);
	sessions = $state<any[]>([]);
	prefixes = $state<string[]>(["All Prefixes"]);
	selectedPrefixFilter = $state("All Prefixes");
	sessionInterval: any = null;
	showHelpModal = $state(false);
	showAddModal = $state(false);
	isLoading = $state(true);
	currentView = $state<"grid" | "list-grid" | "sidebar-grid">("grid");
	searchQuery = $state("");

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
	sessionPollInFlight = false;

	async refreshData(forceScan = false) {
		// Kick the heavier folder scan off now so it overlaps the library fetch.
		const scannedPromise = GetAutoScannedGames(forceScan).catch((error) => {
			log.error("Failed to load scanned folders", error);
			return null;
		});

		try {
			const data = await service.refreshHomeData();

			if (gamesSignature(data.games) !== gamesSignature(this.games)) {
				this.games = data.games;
			}
			if (sessionsSignature(data.sessions) !== sessionsSignature(this.sessions)) {
				this.sessions = data.sessions;
			}
			const nextPrefixes = data.prefixes || ["All Prefixes"];
			if (nextPrefixes.join("|") !== this.prefixes.join("|")) {
				this.prefixes = nextPrefixes;
			}
		} catch (error) {
			log.error("Failed to load library", error);
			notifications.add("Failed to load your library", "error");
		}

		// Render the library before waiting on the scan.
		this.isLoading = false;

		const scannedGroups = await scannedPromise;
		if (scannedGroups && groupsSignature(scannedGroups) !== groupsSignature(this.scannedFolderGroups)) {
			this.scannedFolderGroups = scannedGroups;
		}
	}

	async refreshSessions() {
		if (this.sessionPollInFlight) return;
		this.sessionPollInFlight = true;
		try {
			const sessions = (await GetRunningSessions()) || [];
			if (sessionsSignature(sessions) !== sessionsSignature(this.sessions)) {
				this.sessions = sessions;
			}
		} catch {
		} finally {
			this.sessionPollInFlight = false;
		}
	}

	initialize() {
		this.refreshData(false);

		this.dropUnsubscribe = onEvent("FilesDropped", async (event: any) => {
			const files = (event?.data || event) as string[];
			const added = await service.processDroppedFiles(files);
			if (added > 0) {
				notifications.add(`Successfully added ${added} game(s)`, "success");
				this.refreshData(true);
			}
		});

		this.sessionInterval = setInterval(() => this.refreshSessions(), 3000);
	}

	destroy() {
		if (this.sessionInterval) clearInterval(this.sessionInterval);
		if (this.dropUnsubscribe) this.dropUnsubscribe();
	}

	async handleQuickLaunch(game: any, showLogs = false) {
		try {
			await service.quickLaunchGame(game, showLogs);
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
		const path = game.path || game.config.GamePath || game.config.LauncherPath;
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
