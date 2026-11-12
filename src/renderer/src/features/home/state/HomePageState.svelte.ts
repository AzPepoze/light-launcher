import {
	ClearPreferredLauncher,
	GetAutoScannedGames,
	GetRunningSessions,
	SetPreferredLauncher,
	onEvent
} from "@lib/api";
import * as service from "@lib/homeService";
import { getGamePath } from "@lib/gameUtils";
import { createLogger } from "@lib/logger";
import { navigationCommand } from "@stores/navigationStore";
import { notifications } from "@stores/notificationStore";
import { runState } from "@stores/runState";
import type { GameInfo, RunningSession, ScannedFolderGroup } from "@shared";
import { IconLoaderState } from "./IconLoaderState.svelte";
import { SelectionState } from "./SelectionState.svelte";
import { droppedFiles } from "./homeEvents";

const log = createLogger("HomePage");

function gameIdentity(game: GameInfo): string {
	return getGamePath(game) || game.name || "";
}

function sessionsSignature(sessions: RunningSession[]): string {
	return sessions
		.map((session) => `${session.pid}:${session.gamePath}`)
		.sort()
		.join("|");
}

function gamesSignature(games: GameInfo[]): string {
	return games
		.map(
			(game) =>
				`${gameIdentity(game)}|${game.name}|${game.config?.PrefixPath ?? ""}|${game.config?.CustomIconPath ?? ""}|${game.isAutoScanned ? "1" : "0"}`
		)
		.sort()
		.join(";");
}

function groupsSignature(groups: ScannedFolderGroup[]): string {
	return groups
		.map((group) => `${group.folderPath}#${gamesSignature(group.games)}`)
		.sort()
		.join(";");
}

function matchesPrefix(game: GameInfo, prefix: string): boolean {
	if (prefix === "All Prefixes") return true;
	const prefixPath = game.config?.PrefixPath ?? "";
	return prefixPath.endsWith("/" + prefix) || prefixPath.endsWith("\\" + prefix);
}

export class HomePageState {
	games = $state<GameInfo[]>([]);
	scannedFolderGroups = $state<ScannedFolderGroup[]>([]);
	sessions = $state<RunningSession[]>([]);
	prefixes = $state<string[]>(["All Prefixes"]);
	selectedPrefixFilter = $state("All Prefixes");
	sessionInterval: ReturnType<typeof setInterval> | null = null;
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
			return matchesSearch && matchesPrefix(game, this.selectedPrefixFilter);
		});
	});

	filteredScannedFolderGroups = $derived.by(() => {
		return this.scannedFolderGroups
			.map((group) => {
				const games = group.games.filter((game) => {
					const matchesSearch = game.name.toLowerCase().includes(this.searchQuery.toLowerCase());
					return matchesSearch && this.selectedPrefixFilter === "All Prefixes";
				});
				return { ...group, games };
			})
			.filter((group) => (this.searchQuery ? group.games.length > 0 : true));
	});

	dropUnsubscribe: (() => void) | null = null;
	sessionPollInFlight = false;
	iconWarmTimer: ReturnType<typeof setTimeout> | null = null;

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
		if (
			scannedGroups &&
			groupsSignature(scannedGroups) !== groupsSignature(this.scannedFolderGroups)
		) {
			this.scannedFolderGroups = scannedGroups;
		}

		this.scheduleIconWarm();
	}

	private scheduleIconWarm() {
		if (this.iconWarmTimer) return;
		this.iconWarmTimer = setTimeout(() => {
			this.iconWarmTimer = null;
			this.icons.warm([
				...this.games,
				...this.scannedFolderGroups.flatMap((group) => group.games || [])
			]);
		}, 1200);
	}

	async refreshSessions() {
		if (this.sessionPollInFlight) return;
		this.sessionPollInFlight = true;
		try {
			const sessions = (await GetRunningSessions()) || [];
			if (sessionsSignature(sessions) !== sessionsSignature(this.sessions)) {
				this.sessions = sessions;
			}
		} catch (error) {
			log.warn("Failed to refresh sessions", error);
		} finally {
			this.sessionPollInFlight = false;
		}
	}

	initialize() {
		this.refreshData(false);

		this.dropUnsubscribe = onEvent("FilesDropped", async (event: unknown) => {
			const files = droppedFiles(event);
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
		if (this.iconWarmTimer) clearTimeout(this.iconWarmTimer);
	}

	async handleQuickLaunch(game: GameInfo, showLogs = false) {
		try {
			await service.quickLaunchGame(game, showLogs);
			this.refreshData();
		} catch (error) {
			log.warn("Quick launch failed", error);
		}
	}

	handleConfigure(game: GameInfo) {
		runState.update((state) => ({
			...state,
			options: game.config
		}));
		navigationCommand.set({ page: "run" });
	}

	/** Remembers the picked launcher of a game folder, or clears it when launcherPath is null. */
	async chooseLauncher(game: GameInfo, launcherPath: string | null) {
		try {
			if (launcherPath) {
				await SetPreferredLauncher(game.folder, launcherPath);
			} else {
				await ClearPreferredLauncher(game.folder);
			}
			this.refreshData();
		} catch (error) {
			log.error("Failed to choose launcher", error);
			notifications.add(`Failed to choose launcher: ${error}`, "error");
		}
	}

	isGameRunning(game: GameInfo, sessionsList: RunningSession[]) {
		const path = getGamePath(game);
		return sessionsList.some((session) => session.gamePath === path);
	}

	getVisibleGames(): GameInfo[] {
		const visible: GameInfo[] = [...this.filteredGames];
		for (const group of this.filteredScannedFolderGroups) {
			visible.push(...group.games);
		}
		return visible;
	}
}
