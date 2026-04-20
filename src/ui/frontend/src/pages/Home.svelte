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
	import { fly, fade } from "svelte/transition";
	import { notifications } from "../notificationStore";
	import { navigationCommand } from "../stores/navigationStore";
	import { runState } from "../stores/runState";
	import { loadExeIcon } from "../lib/iconService";
	import GameCard from "../components/home/GameCard.svelte";
	import StatusDrawer from "../components/StatusDrawer.svelte";
	import Modal from "../components/Modal.svelte";
	import AddGameModal from "../components/home/AddGame/AddGameModal.svelte";
	import Dropdown from "../components/Dropdown.svelte";

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
		const matchesSearch = game.name.toLowerCase().includes(searchQuery.toLowerCase());
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
			notifications.add(`Successfully removed ${count} games`, "success");
			selectedPaths.clear();
			selectedPaths = selectedPaths;
			isSelectionMode = false;
			showBulkRemoveModal = false;
			refreshData();
		} catch (err) {
			notifications.add(`Failed to remove some games: ${err}`, "error");
		}
	}
</script>

<div class="home-container">
	{#if sessions.length > 0}
		<div class="sessions-section">
			<h2 class="section-title">Running Sessions</h2>
			<div class="sessions-grid">
				{#each sessions as session}
					<div class="session-card" in:fly={{ y: -20, duration: 400 }}>
						<div class="session-info">
							<div class="session-title">{session.gameName}</div>
							<div class="session-pid">PID: {session.pid}</div>
						</div>
						<button
							class="kill-btn"
							on:click={() => handleKillSession(session.pid, session.gameName)}
						>
							Terminate
						</button>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<div class="quick-launch-section">
		<div class="section-header">
			<h2 class="section-title">Quick Launch</h2>

			{#if isSelectionMode}
				<div class="selection-actions" in:fade>
					<span class="selection-count">{selectedPaths.size} selected</span>
					<button
						class="bulk-remove-btn"
						on:click={handleBulkRemove}
						disabled={selectedPaths.size === 0}
					>
						<span class="material-icons" style="font-size: 18px;">delete</span>
						Remove Selected
					</button>
					<button class="cancel-selection-btn" on:click={toggleSelectionMode}> Cancel </button>
				</div>
			{:else}
				<button class="add-btn" on:click={() => (showAddModal = true)} title="Add Game">
					<span class="material-icons" style="font-size: 20px;">add</span>
				</button>

				<button
					class="select-mode-btn"
					on:click={toggleSelectionMode}
					title="Bulk Remove"
					class:active={isSelectionMode}
				>
					<span class="material-icons" style="font-size: 20px;">checklist</span>
				</button>

				<button class="help-btn" on:click={() => (showHelpModal = true)} title="How it works">
					<span class="material-icons" style="font-size: 24px;">help_outline</span>
				</button>
			{/if}

			<div class="prefix-filter-container">
				<Dropdown options={prefixes} bind:value={selectedPrefixFilter} placeholder="All Prefixes" />
			</div>

			<div class="search-container">
				<span class="material-icons search-icon">search</span>
				<input type="text" placeholder="Search games..." bind:value={searchQuery} class="search-input" />
				{#if searchQuery}
					<button class="clear-search" on:click={() => (searchQuery = "")} aria-label="Clear search">
						<span class="material-icons" style="font-size: 14px;">close</span>
					</button>
				{/if}
			</div>

			<div class="view-switcher">
				<button
					class="view-btn"
					class:active={currentView === "grid"}
					on:click={() => (currentView = "grid")}
					title="Grid View"
				>
					<span class="material-icons" style="font-size: 18px;">grid_view</span>
				</button>
				<button
					class="view-btn"
					class:active={currentView === "list-grid"}
					on:click={() => (currentView = "list-grid")}
					title="List View"
				>
					<span class="material-icons" style="font-size: 18px;">view_list</span>
				</button>
			</div>
		</div>

		{#if games.length === 0}
			<div class="empty-state">
				<p>
					No games configured yet. Go to <button
						class="link-btn"
						on:click={() => navigationCommand.set({ page: "run" })}>Run</button
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
								icon={gameIcons[game.path || game.config.LauncherPath]}
								isRunning={isGameRunning(game)}
								{isSelectionMode}
								isSelected={selectedPaths.has(game.path || game.config.LauncherPath)}
								view={currentView}
								onLaunch={() => handleQuickLaunch(game)}
								onConfigure={() => handleConfigure(game)}
								onSelect={() => toggleGameSelection(game)}
							/>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<Modal show={showHelpModal} title="How it works" onClose={() => (showHelpModal = false)}>
	<div class="help-content">
		<section>
			<h3>Adding Games</h3>
			<p>
				Go to the <strong>Run</strong> page, select your game executable (and launcher if applicable),
				configure your settings, and click <strong>LAUNCH GAME</strong>.
			</p>
			<p>
				After the first run, the game will automatically appear here in <strong>Quick Launch</strong>.
			</p>
		</section>

		<section>
			<h3>Quick Launch</h3>
			<p>Click on any game card in this section to start it immediately with its saved configuration.</p>
		</section>

		<section>
			<h3>Managing Sessions</h3>
			<p>Active game sessions are displayed at the top. You can terminate them if they become unresponsive.</p>
		</section>

		<section>
			<h3>CLI Usage</h3>
			<p>You can also launch games directly from your terminal or add them to your desktop entries:</p>
			<code class="help-code">light-launcher /path/to/game.exe</code>
		</section>
	</div>
</Modal>

<Modal show={showBulkRemoveModal} title="Remove Games" onClose={() => (showBulkRemoveModal = false)}>
	<div class="confirm-modal-content">
		<div class="warning-icon">
			<span class="material-icons" style="font-size: 48px; color: #ef4444;">warning</span>
		</div>
		<p>Are you sure you want to remove <strong>{selectedPaths.size}</strong> games from the library?</p>
		<p class="sub-text">This will only remove them from the Quick Launch list, not from your disk.</p>
	</div>

	<div slot="footer" class="modal-footer-actions">
		<button class="cancel-btn" on:click={() => (showBulkRemoveModal = false)}> Cancel </button>
		<button class="confirm-remove-btn" on:click={confirmBulkRemove}> Remove Games </button>
	</div>
</Modal>

<AddGameModal show={showAddModal} onClose={() => (showAddModal = false)} onRefresh={refreshData} />

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

	.section-title {
		font-size: 1.2rem;
		font-weight: 800;
		color: var(--text-dim);
		opacity: 0.4;
		text-transform: uppercase;
		letter-spacing: 2px;
		margin-bottom: 20px;
	}

	.sessions-section {
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		gap: 20px;
		background: linear-gradient(135deg, rgba(239, 68, 68, 0.1) 0%, rgba(239, 68, 68, 0.02) 100%);
		padding: 24px;
		border-radius: 24px;
		border: 1px solid rgba(239, 68, 68, 0.2);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
		animation: slide-down 0.5s cubic-bezier(0.23, 1, 0.32, 1);

		.section-title {
			margin-bottom: 10px;
			color: #ef4444;
		}
	}

	.sessions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 16px;
	}

	.session-card {
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 16px;
		padding: 14px 20px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: all 0.3s;

		&:hover {
			border-color: rgba(239, 68, 68, 0.4);
			background: var(--glass-surface);
			transform: translateX(4px);
		}

		.session-info {
			display: flex;
			flex-direction: column;
			gap: 2px;
		}

		.session-title {
			font-weight: 800;
			color: var(--text-main);
			font-size: 1rem;
			letter-spacing: -0.3px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			max-width: 200px;
		}

		.session-pid {
			font-size: 0.7rem;
			color: var(--text-dim);
			font-family: monospace;
			font-weight: 600;
		}

		.kill-btn {
			background: #ef4444;
			color: #fff;
			padding: 8px 16px;
			border: none;
			border-radius: 10px;
			font-size: 0.75rem;
			font-weight: 800;
			cursor: pointer;
			transition: all 0.2s;
			box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);

			&:hover {
				filter: brightness(1.2);
				transform: translateY(-2px);
				box-shadow: 0 6px 16px rgba(239, 68, 68, 0.4);
			}

			&:active {
				transform: translateY(0);
			}
		}
	}

	@keyframes slide-down {
		from {
			opacity: 0;
			transform: translateY(-20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.quick-launch-section {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;

		.section-header {
			display: flex;
			align-items: center;
			gap: 12px;
			margin-bottom: 20px;
			flex-wrap: wrap;

			.section-title {
				margin: 0;
				line-height: 1;
				white-space: nowrap;
			}
		}

		.view-switcher {
			display: flex;
			background: var(--glass-surface);
			padding: 4px;
			border-radius: 12px;
			gap: 4px;
			border: 1px solid var(--glass-border);

			.view-btn {
				background: none;
				border: none;
				color: var(--text-dim);
				opacity: 0.5;
				padding: 6px;
				cursor: pointer;
				border-radius: 8px;
				display: flex;
				align-items: center;
				justify-content: center;
				aspect-ratio: 1 / 1;
				transition: all 0.2s;

				&:hover {
					color: var(--text-main);
					background: var(--glass-border-bright);
				}

				&.active {
					color: var(--glass-bg);
					background: var(--accent-primary);
					box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
				}
			}
		}

		.help-btn {
			background: none;
			border: none;
			color: var(--text-dim);
			opacity: 0.4;
			cursor: pointer;
			padding: 4px;
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.2s;
			border-radius: 50%;

			&:hover {
				color: var(--accent-primary);
				opacity: 1;
				background: var(--glass-surface);
				transform: scale(1.1);
			}

		}

		.select-mode-btn {
			background: var(--glass-surface);
			border: 1px solid var(--glass-border);
			color: var(--text-dim);
			cursor: pointer;
			padding: 4px;
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.2s;
			border-radius: 50%;

			&:hover {
				color: var(--text-main);
				background: var(--glass-border-bright);
				transform: scale(1.1);
			}

			&.active {
				background: var(--accent-primary);
				color: var(--glass-bg);
				border-color: transparent;
			}

		}

		.selection-actions {
			display: flex;
			align-items: center;
			gap: 12px;
			background: var(--glass-surface);
			padding: 4px 12px;
			border-radius: 16px;
			border: 1px solid var(--glass-border);

			.selection-count {
				font-size: 0.85rem;
				font-weight: 700;
				color: var(--accent-primary);
			}

			.bulk-remove-btn {
				background: #ef4444;
				color: #fff;
				border: none;
				padding: 6px 12px;
				border-radius: 8px;
				font-size: 0.8rem;
				font-weight: 800;
				cursor: pointer;
				display: flex;
				align-items: center;
				gap: 6px;
				transition: all 0.2s;

				&:hover:not(:disabled) {
					filter: brightness(1.2);
					transform: translateY(-1px);
				}

				&:disabled {
					opacity: 0.5;
					cursor: not-allowed;
				}
			}

			.cancel-selection-btn {
				background: none;
				border: none;
				color: var(--text-dim);
				font-size: 0.8rem;
				font-weight: 700;
				cursor: pointer;

				&:hover {
					color: var(--text-main);
				}
			}
		}

		.add-btn {
			background: var(--accent-primary);
			border: none;
			color: var(--glass-bg);
			cursor: pointer;
			padding: 4px;
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.2s;
			border-radius: 50%;
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);

			&:hover {
				filter: brightness(1.2);
				transform: scale(1.1) rotate(90deg);
				box-shadow: 0 6px 16px rgba(0, 0, 0, 0.3);
			}

		}

		.prefix-filter-container {
			min-width: 160px;
			max-width: 200px;

			:global(.dropdown-trigger) {
				padding: 8px 12px;
				font-size: 0.8rem;
				background: rgba(255, 255, 255, 0.05);
				border-color: rgba(255, 255, 255, 0.05);

				&:hover {
					background: rgba(255, 255, 255, 0.1);
					border-color: rgba(255, 255, 255, 0.2);
				}
			}

			:global(.dropdown-menu) {
				background: #18181b;
				border: 1px solid rgba(255, 255, 255, 0.1);
				box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
			}
		}

		.search-container {
			display: flex;
			align-items: center;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.05);
			border-radius: 12px;
			padding: 4px 10px;
			gap: 8px;
			flex: 1;
			transition: all 0.3s;

			&:focus-within {
				background: rgba(255, 255, 255, 0.1);
				border-color: rgba(255, 255, 255, 0.2);
				box-shadow: 0 0 15px rgba(0, 0, 0, 0.2);
			}

			.search-icon {
				color: rgba(255, 255, 255, 0.3);
			}

			.search-input {
				background: none;
				border: none;
				color: #fff;
				font-size: 0.9rem;
				width: 100%;
				outline: none;

				&::placeholder {
					color: rgba(255, 255, 255, 0.2);
				}
			}

			.clear-search {
				background: none;
				border: none;
				color: rgba(255, 255, 255, 0.3);
				cursor: pointer;
				padding: 2px;
				display: flex;
				align-items: center;
				justify-content: center;
				border-radius: 4px;

				&:hover {
					color: #fff;
					background: rgba(255, 255, 255, 0.1);
				}
			}
		}
	}

	.help-content {
		display: flex;
		flex-direction: column;
		gap: 24px;
		color: var(--text-main, #eee);

		section {
			h3 {
				margin: 0 0 8px 0;
				font-size: 1.1rem;
				color: var(--accent-primary, #fff);
			}

			p {
				margin: 0;
				line-height: 1.6;
				font-size: 0.95rem;
				color: var(--text-dim, #aaa);

				strong {
					color: var(--text-main, #eee);
				}
			}

			.help-code {
				display: block;
				background: rgba(0, 0, 0, 0.3);
				padding: 12px;
				border-radius: 8px;
				font-family: monospace;
				font-size: 0.85rem;
				color: var(--accent-primary, #fff);
				margin-top: 10px;
				border: 1px solid rgba(255, 255, 255, 0.05);
			}

			& + section {
				padding-top: 16px;
				border-top: 1px solid rgba(255, 255, 255, 0.05);
			}
		}
	}

	.confirm-modal-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 16px;
		padding: 10px 0;

		p {
			margin: 0;
			font-size: 1.1rem;
			color: rgba(255, 255, 255, 0.9);

			strong {
				color: #ef4444;
			}
		}

		.sub-text {
			font-size: 0.9rem;
			color: rgba(255, 255, 255, 0.4);
		}

		.warning-icon {
			background: rgba(239, 68, 68, 0.1);
			padding: 20px;
			border-radius: 50%;
			margin-bottom: 8px;
		}
	}

	.modal-footer-actions {
		display: flex;
		gap: 12px;
		width: 100%;

		button {
			flex: 1;
			padding: 12px;
			border-radius: 12px;
			font-weight: 800;
			cursor: pointer;
			transition: all 0.2s;
		}

		.cancel-btn {
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			color: #fff;

			&:hover {
				background: rgba(255, 255, 255, 0.1);
			}
		}

		.confirm-remove-btn {
			background: #ef4444;
			border: none;
			color: #fff;
			box-shadow: 0 4px 15px rgba(239, 68, 68, 0.3);

			&:hover {
				filter: brightness(1.2);
				transform: translateY(-2px);
			}
		}
	}

	.games-container {
		flex: 1;
		min-height: 0;
		display: flex;
		flex-direction: column;

		&.list-view {
			.games-grid {
				grid-template-columns: 1fr;
				gap: 16px;
			}
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
		background: rgba(255, 255, 255, 0.02);
		border: 1px dashed var(--glass-border);
		border-radius: 12px;
		padding: 32px;
		text-align: center;
		color: var(--text-muted);

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
</style>
