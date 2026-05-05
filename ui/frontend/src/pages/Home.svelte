<script lang="ts">
	import {
		GetAllGames,
		GetRunningSessions,
		KillSession,
		RunGame,
		ListPrefixes,
		RemoveGame,
		SaveGameConfig,
		GetPrefixBaseDir,
	} from "@bindings/light-launcher/internal/app/app";
	import { onMount, onDestroy } from "svelte";
	import { Events } from "@wailsio/runtime";
	import { notifications } from "@stores/notificationStore";
	import { navigationCommand } from "@stores/navigationStore";
	import { runState } from "@stores/runState";
	import { loadExeIcon } from "@lib/iconService";

	import GameCard from "@components/home/GameCard.svelte";
	import GameGrid from "@components/home/GameGrid.svelte";
	import StatusDrawer from "@components/shared/StatusDrawer.svelte";
	import AddGameModal from "@components/home/addgame/AddGameModal.svelte";
	import RunningSessions from "@components/home/RunningSessions.svelte";
	import QuickLaunchHeader from "@components/home/QuickLaunchHeader.svelte";
	import HowItWorksModal from "@components/home/HowItWorksModal.svelte";
	import BulkRemoveModal from "@components/home/BulkRemoveModal.svelte";

	let games = [];
	let sessions = [];
	let prefixes = ["All Prefixes"];
	let selectedPrefixFilter = "All Prefixes";
	let sessionInterval;
	let gameIcons = {};
	let showHelpModal = false;
	let showAddModal = false;
	let showBulkRemoveModal = false;
	let currentView: "grid" | "list-grid" = "grid";
	let searchQuery = "";

	let isSelectionMode = false;
	let selectedPaths = new Set<string>();

	$: filteredGames = games.filter((game) => {
		const matchesSearch = game.name
			.toLowerCase()
			.includes(searchQuery.toLowerCase());
		const matchesPrefix =
			selectedPrefixFilter === "All Prefixes" ||
			game.config.PrefixPath.endsWith("/" + selectedPrefixFilter) ||
			game.config.PrefixPath.endsWith("\\" + selectedPrefixFilter);
		return matchesSearch && matchesPrefix;
	});

	async function refreshData() {
		try {
			const fetchedGames = await GetAllGames();
			games = fetchedGames || [];
			const fetchedSessions = await GetRunningSessions();
			sessions = fetchedSessions || [];

			const fetchedPrefixes = await ListPrefixes();
			prefixes = ["All Prefixes", ...(fetchedPrefixes || [])];

			// Fetch icons for games
			for (const game of games) {
				const path = game.path || game.config.RunnerPath;
				if (path && !gameIcons[path]) {
					loadExeIcon(path).then((icon) => {
						if (icon) {
							gameIcons = { ...gameIcons, [path]: icon };
						}
					});
				}
			}
		} catch (err) {
			console.error("Failed to refresh home data:", err);
		}
	}

	let dropUnsubscribe: () => void;

	onMount(() => {
		refreshData();

		dropUnsubscribe = Events.On("FilesDropped", async (event) => {
			const files = event.data as string[];
			let added = 0;
			const basePrefixDir = await GetPrefixBaseDir();
			const prefixPath = basePrefixDir + "/Default";

			for (const path of files) {
				if (path.toLowerCase().endsWith(".exe")) {
					const name =
						path.split("/").pop()?.replace(".exe", "") || "Game";
					const config: any = {
						Name: name,
						RunnerPath: path,
						GamePath: path,
						UseGamePath: false,
						PrefixPath: prefixPath,
						ProtonPath: "",
						ProtonPattern: "GE-Proton*",
						CustomArgs: "",
						Extras: {
							EnableMangoHud: false,
							EnableGamemode: false,
							Lsfg: {
								Enabled: false,
								Multiplier: "2",
								PerfMode: false,
								DllPath: "",
								Gpu: "",
								FlowScale: "0.8",
								Pacing: "none",
								AllowFp16: false,
							},
							Gamescope: {
								Enabled: false,
								Width: "1920",
								Height: "1080",
								RefreshRate: "60",
							},
							Memory: {
								Enabled: false,
								Value: "4G",
							},
						},
					};
					await SaveGameConfig(config);
					added++;
				}
			}
			if (added > 0) {
				notifications.add(
					`Successfully added ${added} game(s)`,
					"success",
				);
				refreshData();
			}
		});

		sessionInterval = setInterval(async () => {
			try {
				const fetchedSessions = await GetRunningSessions();
				sessions = fetchedSessions || [];
			} catch (err) {
				console.error("Failed to fetch sessions in interval:", err);
			}
		}, 3000);
	});

	onDestroy(() => {
		if (sessionInterval) clearInterval(sessionInterval);
		if (dropUnsubscribe) dropUnsubscribe();
	});

	async function handleQuickLaunch(game) {
		try {
			notifications.add(`Launching ${game.name}...`, "info");
			await RunGame(game.config, false); // No logs for quick launch
			refreshData();
		} catch (err) {
			notifications.add(`Launch failed: ${err}`, "error");
		}
	}

	function handleConfigure(game) {
		runState.update((s) => ({
			...s,
			options: game.config,
		}));
		navigationCommand.set({ page: "run" });
	}

	function isGameRunning(game, sessionsList) {
		const path = game.path || game.config.LauncherPath;
		return sessionsList.some((s) => s.gamePath === path);
	}

	async function handleKillSession(pid, name) {
		try {
			await KillSession(pid);
			notifications.add(`Terminated session: ${name}`, "success");
			refreshData();
		} catch (err) {
			notifications.add(`Failed to kill session: ${err}`, "error");
		}
	}

	function toggleSelectionMode() {
		isSelectionMode = !isSelectionMode;
		if (!isSelectionMode) {
			selectedPaths.clear();
			selectedPaths = selectedPaths; // trigger reactivity
		}
	}

	function toggleGameSelection(game) {
		const path = game.path || game.config.LauncherPath;
		if (selectedPaths.has(path)) {
			selectedPaths.delete(path);
		} else {
			selectedPaths.add(path);
		}
		selectedPaths = selectedPaths; // trigger reactivity
	}

	async function handleBulkRemove() {
		if (selectedPaths.size === 0) return;
		showBulkRemoveModal = true;
	}

	async function confirmBulkRemove() {
		try {
			let count = 0;
			for (const path of selectedPaths) {
				await RemoveGame(path);
				count++;
			}
			notifications.add(
				`Successfully removed ${count} games`,
				"success",
			);
			selectedPaths.clear();
			selectedPaths = selectedPaths;
			isSelectionMode = false;
			showBulkRemoveModal = false;
			refreshData();
		} catch (err) {
			notifications.add(
				`Failed to remove some games: ${err}`,
				"error",
			);
		}
	}
</script>

<div class="home-container" data-file-drop-target>
	<RunningSessions {sessions} onKill={handleKillSession} />

	<div class="quick-launch-section">
		<QuickLaunchHeader
			{isSelectionMode}
			selectedCount={selectedPaths.size}
			{prefixes}
			bind:selectedPrefixFilter
			bind:searchQuery
			bind:currentView
			onBulkRemove={handleBulkRemove}
			onToggleSelectionMode={toggleSelectionMode}
			onShowAddModal={() => (showAddModal = true)}
			onShowHelpModal={() => (showHelpModal = true)}
		/>

		{#if games.length === 0}
			<div class="empty-state">
				<p>
					No games configured yet. Go to <button
						class="link-btn"
						on:click={() =>
							navigationCommand.set({ page: "run" })}
						>Run</button
					> to add one.
				</p>
			</div>
		{:else}
			<GameGrid
				{currentView}
				{games}
				{filteredGames}
				{gameIcons}
				{searchQuery}
				{selectedPrefixFilter}
				{isSelectionMode}
				{selectedPaths}
				{sessions}
				{isGameRunning}
				handleQuickLaunch={handleQuickLaunch}
				handleConfigure={handleConfigure}
				toggleGameSelection={toggleGameSelection}
			/>
		{/if}
	</div>
</div>

<HowItWorksModal show={showHelpModal} onClose={() => (showHelpModal = false)} />

<BulkRemoveModal
	show={showBulkRemoveModal}
	selectedCount={selectedPaths.size}
	onClose={() => (showBulkRemoveModal = false)}
	onConfirm={confirmBulkRemove}
/>

<AddGameModal
	show={showAddModal}
	onClose={() => (showAddModal = false)}
	onRefresh={refreshData}
/>

<StatusDrawer />

<style lang="scss">
	.home-container {
		display: flex;
		flex-direction: column;
		height: 100%;
		width: 100%;
		padding: 0;
		background-color: transparent;
		gap: 20px;
		box-sizing: border-box;
		min-height: 0;
		overflow-x: hidden;
	}

	.quick-launch-section {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
	}



	.link-btn {
		background: none;
		border: none;
		color: var(--accent-primary);
		font-weight: 800;
		text-decoration: underline;
		cursor: pointer;
		padding: 0;
		font-size: inherit;

		&:hover {
			filter: brightness(1.2);
		}
	}


</style>
