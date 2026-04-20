<script lang="ts">
	import {
		GetAllGames,
		GetRunningSessions,
		KillSession,
		RunGame,
		ListPrefixes,
		RemoveGame,
	} from "@bindings/light-launcher-wails/backend/app";
	import { onMount, onDestroy } from "svelte";
	import { notifications } from "@stores/notificationStore";
	import { navigationCommand } from "@stores/navigationStore";
	import { runState } from "@stores/runState";
	import { loadExeIcon } from "@lib/iconService";

	import GameCard from "@components/home/GameCard.svelte";
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
				const path = game.path || game.config.LauncherPath;
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

	onMount(() => {
		refreshData();
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

	function isGameRunning(game) {
		const path = game.path || game.config.LauncherPath;
		return sessions.some((s) => s.gamePath === path);
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

<div class="home-container">
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
			<div
				class="games-container"
				class:grid-view={currentView === "grid"}
				class:list-view={currentView === "list-grid"}
			>
				{#if filteredGames.length === 0 && games.length > 0}
					<div class="no-results">
						<p>
							No games matching
							{#if searchQuery}"{searchQuery}"{/if}
							{#if selectedPrefixFilter !== "All Prefixes"}
								in prefix <b>{selectedPrefixFilter}</b>
							{/if}
						</p>
						<button
							class="link-btn"
							on:click={() => {
								searchQuery = "";
								selectedPrefixFilter = "All Prefixes";
							}}>Clear all filters</button
						>
					</div>
				{:else}
					<div class="games-grid">
						{#each filteredGames as game}
							<GameCard
								{game}
								icon={gameIcons[
									game.path ||
										game.config.LauncherPath
								]}
								isRunning={isGameRunning(game)}
								{isSelectionMode}
								isSelected={selectedPaths.has(
									game.path ||
										game.config.LauncherPath,
								)}
								view={currentView}
								onLaunch={() => handleQuickLaunch(game)}
								onConfigure={() =>
									handleConfigure(game)}
								onSelect={() =>
									toggleGameSelection(game)}
							/>
						{/each}
					</div>
				{/if}
			</div>
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
	onClose={() => (showAddModal = true)}
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

	.games-container {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding-right: 4px;

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: var(--glass-border);
			border-radius: 10px;
		}
	}

	.games-grid {
		display: grid;
		gap: 20px;
		width: 100%;
		padding-bottom: 40px;
	}

	.grid-view .games-grid {
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	}

	.list-view .games-grid {
		grid-template-columns: 1fr;
	}

	.empty-state,
	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60px 20px;
		color: var(--text-dim);
		text-align: center;
		background: var(--glass-surface);
		border-radius: 24px;
		border: 1px dashed var(--glass-border);
		margin-top: 20px;

		p {
			font-size: 1.1rem;
			margin-bottom: 16px;
		}
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

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 200px;
		color: rgba(255, 255, 255, 0.4);
		gap: 10px;

		p {
			margin: 0;
			font-size: 1rem;
		}

		.link-btn {
			background: none;
			border: none;
			color: var(--accent-color, #60a5fa);
			text-decoration: underline;
			cursor: pointer;
			padding: 0;
			font: inherit;
			font-weight: 600;

			&:hover {
				filter: brightness(1.2);
			}
		}
	}

	.games-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
		grid-auto-rows: min-content;
		gap: 32px;
		overflow-y: auto;
		overflow-x: hidden;
		padding: 10px;
		flex: 1;
		min-height: 0;
	}

	.empty-state {
		background: var(--glass-surface);
		border: 1px dashed var(--glass-border-bright);
		border-radius: 20px;
		padding: 48px;
		text-align: center;
		color: var(--text-muted);
		box-shadow: var(--glass-shadow);
		margin: 20px 0;

		.link-btn {
			background: none;
			border: none;
			color: var(--accent-primary);
			text-decoration: underline;
			cursor: pointer;
			padding: 0;
			font: inherit;
			font-weight: 800;

			&:hover {
				filter: brightness(1.2);
				color: var(--accent-secondary);
			}
		}
	}
</style>
