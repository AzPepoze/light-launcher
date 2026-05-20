<script lang="ts">
	import { onMount } from "svelte";
	import GameCard from "@components/home/GameCard.svelte";
	import ContextMenu from "@components/shared/ContextMenu.svelte";
	import SidebarPanel from "@components/home/SidebarPanel.svelte";
	import FolderGroup from "@components/home/FolderGroup.svelte";
	import FolderSettingsModal from "@components/home/FolderSettingsModal.svelte";
	import { BlacklistGame, RemoveGame, RemoveScanFolder } from "@bindings/light-launcher/internal/app/app";
	import { notifications } from "@stores/notificationStore";
	import SidebarProfilesSection from "@components/home/SidebarProfilesSection.svelte";

	export let currentView: "grid" | "list-grid" | "sidebar-grid" = "grid";
	export let games: any[] = [];
	export let filteredGames: any[] = [];
	export let scannedFolderGroups: any[] = [];
	export let gameIcons: Record<string, string> = {};
	export let searchQuery = "";
	export let selectedPrefixFilter = "All Prefixes";
	export let isSelectionMode = false;
	export let selectedPaths = new Set<string>();
	
	export let isGameRunning: (game: any, sessionsList: any[]) => boolean;
	export let sessions: any[] = [];
	export let handleQuickLaunch: (game: any) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let toggleGameSelection: (game: any, shiftKey: boolean) => void;
	export let onRefresh: () => void = () => {};

	let menuX = 0;
	let menuY = 0;
	let menuVisible = false;
	let activeMenuGame: any = null;

	let activeFolderMenu: string | null = null;

	function toggleFolderMenu(event: MouseEvent, folderPath: string) {
		if (activeFolderMenu === folderPath) {
			activeFolderMenu = null;
		} else {
			activeFolderMenu = folderPath;
		}
	}

	async function handleRemoveFolder(folderPath: string) {
		try {
			await RemoveScanFolder(folderPath);
			notifications.add("Removed watched folder", "success");
			activeFolderMenu = null;
			onRefresh();
		} catch (err) {
			notifications.add(`Failed to remove watched folder: ${err}`, "error");
		}
	}

	function handleRescan() {
		notifications.add("Rescanning watched folders...", "info");
		activeFolderMenu = null;
		onRefresh();
	}

	let showFolderSettings = false;
	let folderSettingsPath = "";

	function handleConfigureFolder(folderPath: string) {
		folderSettingsPath = folderPath;
		showFolderSettings = true;
		activeFolderMenu = null;
	}

	onMount(() => {
		const handleGlobalClick = () => {
			activeFolderMenu = null;
		};
		window.addEventListener("click", handleGlobalClick);
		return () => {
			window.removeEventListener("click", handleGlobalClick);
		};
	});

	function handleRightClick(event: MouseEvent, game: any) {
		if (isSelectionMode) return;
		menuX = event.clientX;
		menuY = event.clientY;
		activeMenuGame = game;
		menuVisible = true;
	}

	async function handleAction() {
		if (!activeMenuGame) return;
		const path = activeMenuGame.path || activeMenuGame.config.LauncherPath;
		if (activeMenuGame.isAutoScanned) {
			try {
				await BlacklistGame(path);
				notifications.add("Game hidden from library", "success");
				onRefresh();
			} catch (err) {
				notifications.add(`Failed to hide game: ${err}`, "error");
			}
		} else {
			try {
				await RemoveGame(path);
				notifications.add("Game profile removed", "success");
				onRefresh();
			} catch (err) {
				notifications.add(`Failed to remove game: ${err}`, "error");
			}
		}
	}

	let selectedGroupKey = "no-folder";

	$: {
		if (selectedGroupKey !== "no-folder" && !scannedFolderGroups.some(g => g.folderPath === selectedGroupKey)) {
			selectedGroupKey = "no-folder";
		}
	}

	$: showCustomProfiles = currentView !== "sidebar-grid" ? filteredGames.length > 0 : selectedGroupKey === "no-folder";
	$: foldersToRender = currentView !== "sidebar-grid" ? scannedFolderGroups : scannedFolderGroups.filter(g => g.folderPath === selectedGroupKey);
</script>

<div
	class="games-container"
	class:grid-view={currentView === "grid"}
	class:list-view={currentView === "list-grid"}
	class:sidebar-layout-view={currentView === "sidebar-grid"}
>
	{#if currentView === "sidebar-grid"}
		<SidebarPanel
			{filteredGames}
			{scannedFolderGroups}
			bind:selectedGroupKey
		/>
	{/if}

	<div class="main-content-panel">
		<!-- 1. Render Custom Profiles if visible -->
		{#if showCustomProfiles}
			{#if currentView !== "sidebar-grid" && scannedFolderGroups.length > 0}
				<h2 class="scan-section-title">
					<span class="material-icons">library_books</span>
					Custom Profiles <span class="badge">{filteredGames.length}</span>
				</h2>
			{/if}

			{#if currentView === "sidebar-grid"}
				<SidebarProfilesSection
					{filteredGames}
					{gameIcons}
					{isGameRunning}
					{sessions}
					{isSelectionMode}
					{selectedPaths}
					{handleRightClick}
					{handleQuickLaunch}
					{handleConfigure}
					{toggleGameSelection}
				/>
			{:else if filteredGames.length > 0}
				<div class="games-grid">
					{#each filteredGames as game}
						<div on:contextmenu|preventDefault|stopPropagation={(e) => handleRightClick(e, game)}>
							<GameCard
								{game}
								icon={gameIcons[game.path || game.config.LauncherPath]}
								isRunning={isGameRunning(game, sessions)}
								{isSelectionMode}
								isSelected={selectedPaths.has(game.path || game.config.LauncherPath)}
								view={currentView}
								onLaunch={() => handleQuickLaunch(game)}
								onConfigure={() => handleConfigure(game)}
								onSelect={toggleGameSelection}
							/>
						</div>
					{/each}
				</div>
			{/if}
		{/if}

		<!-- 2. Render Folder Groups -->
		{#each foldersToRender as group}
			<FolderGroup
				{group}
				{currentView}
				{gameIcons}
				{isGameRunning}
				{sessions}
				{isSelectionMode}
				{selectedPaths}
				bind:activeFolderMenu
				{toggleFolderMenu}
				{handleRescan}
				{handleRemoveFolder}
				{handleConfigureFolder}
				{handleRightClick}
				{handleQuickLaunch}
				{handleConfigure}
				{toggleGameSelection}
			/>
		{/each}

		<!-- 3. Show "no results" if both custom profiles and folder groups are empty -->
		{#if filteredGames.length === 0 && scannedFolderGroups.every(g => g.games.length === 0) && (games.length > 0 || scannedFolderGroups.length > 0)}
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
		{/if}
	</div>
</div>

{#if activeMenuGame}
	<ContextMenu
		bind:x={menuX}
		bind:y={menuY}
		bind:visible={menuVisible}
		isAutoScanned={activeMenuGame.isAutoScanned}
		isRunning={isGameRunning(activeMenuGame, sessions)}
		onLaunch={() => handleQuickLaunch(activeMenuGame)}
		onConfigure={() => handleConfigure(activeMenuGame)}
		onAction={handleAction}
		onClose={() => { menuVisible = false; activeMenuGame = null; }}
	/>
{/if}

<FolderSettingsModal
	show={showFolderSettings}
	folderPath={folderSettingsPath}
	onClose={() => { showFolderSettings = false; }}
	onSave={onRefresh}
/>

<style lang="scss">
	.games-container {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding-right: 8px;
	}

	.games-grid {
		display: grid;
		gap: 28px;
		width: 100%;
		padding: 12px;
		padding-bottom: 40px;
	}

	.grid-view .games-grid,
	.sidebar-layout-view .games-grid {
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
	}

	.list-view .games-grid {
		grid-template-columns: 1fr;
		gap: 16px;
	}

	.scan-section-title {
		display: flex;
		align-items: center;
		gap: 10px;
		margin: 28px 12px 14px 12px;
		font-size: 1.1rem;
		font-weight: 800;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		padding-bottom: 8px;

		.material-icons {
			font-size: 20px;
			color: var(--accent-secondary);
		}

		.badge {
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: var(--radius-sm);
			padding: 2px 8px;
			font-size: 0.75rem;
			font-weight: 700;
			color: var(--text-muted);
		}
	}

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 80px 20px;
		color: var(--text-muted);
		text-align: center;
		background: var(--bg-surface);
		border-radius: var(--radius-xl);
		border: 2px dashed rgba(255, 255, 255, 0.08);
		margin-top: 20px;

		p {
			font-size: 1.15rem;
			font-weight: 700;
			margin-bottom: 20px;
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



	.games-container.sidebar-layout-view {
		overflow: hidden;
		padding-right: 0;
		display: flex;
		flex-direction: row;
		gap: 24px;

		.main-content-panel {
			flex: 1;
			overflow-y: auto;
			padding-right: 8px;
			display: flex;
			flex-direction: column;
		}
	}
</style>
