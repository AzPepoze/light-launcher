<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { fade } from "svelte/transition";
	import { cubicOut } from "svelte/easing";
	import { navigateToSettingsSection } from "@stores/navigationStore";
	import { GAME_FOLDERS_SECTION } from "@lib/features/settingsSection";
	import { HomePageState } from "./state/HomePageState.svelte";

	import GameGrid from "./GameGrid.svelte";
	import HomeSkeleton from "./HomeSkeleton.svelte";
	import AddGameModal from "./addgame/AddGameModal.svelte";
	import QuickLaunchHeader from "./shared/QuickLaunchHeader.svelte";
	import HowItWorksModal from "./shared/HowItWorksModal.svelte";
	import BulkRemoveModal from "./shared/BulkRemoveModal.svelte";
	import ProgressBar from "@components/shared/ProgressBar.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import Button from "@components/shared/Button.svelte";

	const state = new HomePageState();

	const prefersReducedMotion =
		typeof window !== "undefined" &&
		window.matchMedia("(prefers-reduced-motion: reduce)").matches;

	onMount(() => {
		state.initialize();
	});

	onDestroy(() => {
		state.destroy();
	});
</script>

<div class="home-container" data-file-drop-target>
	<div class="quick-launch-section">
		<QuickLaunchHeader
			isSelectionMode={state.selection.isSelectionMode}
			selectedCount={state.selection.selectedPaths.size}
			prefixes={state.prefixes}
			bind:selectedPrefixFilter={state.selectedPrefixFilter}
			bind:searchQuery={state.searchQuery}
			bind:currentView={state.currentView}
			onBulkRemove={() => state.selection.handleBulkRemove()}
			onToggleSelectionMode={() => state.selection.toggleSelectionMode()}
			onSelectAll={() => state.selection.selectAll()}
			onShowAddModal={() => (state.showAddModal = true)}
			onShowHelpModal={() => (state.showHelpModal = true)}
		/>

		{#if state.status}
			<div class="status-line" out:fade={{ duration: prefersReducedMotion ? 0 : 240, easing: cubicOut }}>
				<span class="status-text">{state.status.text}</span>
				<div class="status-bar">
					<ProgressBar value={state.status.percent} label={state.status.text} active height={4} />
				</div>
			</div>
		{/if}

		{#if state.isLoading}
			<div class="loading-area">
				<HomeSkeleton view={state.currentView} />
			</div>
		{:else if state.showEmptyState}
			<div class="empty-area">
				<EmptyState icon="sports_esports" title="No games yet" hint="Add a game or scan a folder to see it here.">
					<div class="empty-actions">
						<Button variant="primary" onclick={() => (state.showAddModal = true)}>Add game</Button>
						<Button onclick={() => navigateToSettingsSection(GAME_FOLDERS_SECTION)}>Scan folders</Button>
					</div>
				</EmptyState>
			</div>
		{:else}
			<GameGrid
				currentView={state.currentView}
				games={state.games}
				filteredGames={state.filteredGames}
				scannedFolderGroups={state.filteredScannedFolderGroups}
				gameIcons={state.icons.gameIcons}
				searchQuery={state.searchQuery}
				selectedPrefixFilter={state.selectedPrefixFilter}
				isSelectionMode={state.selection.isSelectionMode}
				selectedPaths={state.selection.selectedPaths}
				sessions={state.sessions}
				isGameRunning={state.isGameRunning}
				handleQuickLaunch={(game, showLogs) => state.handleQuickLaunch(game, showLogs)}
				handleConfigure={(game) => state.handleConfigure(game)}
				handleChooseLauncher={(game, launcherPath) => state.chooseLauncher(game, launcherPath)}
				toggleGameSelection={(game, shiftKey, ctrlKey) => state.selection.toggleGameSelection(game, shiftKey, ctrlKey)}
				onRefresh={() => state.refreshData(true)}
				loadIcon={(path, customIconPath) => state.icons.enqueueIconLoad(path, customIconPath)}
				onMarqueeSelect={(paths, additive) => state.selection.applyMarquee(paths, additive)}
				onSelectAll={() => state.selection.selectAll()}
				onCancelSelection={() => state.selection.toggleSelectionMode()}
				modalOpen={state.showAddModal || state.showHelpModal || state.selection.showBulkRemoveModal}
			/>
		{/if}
	</div>
</div>

<HowItWorksModal show={state.showHelpModal} onClose={() => (state.showHelpModal = false)} />

<BulkRemoveModal
	show={state.selection.showBulkRemoveModal}
	selectedCount={state.selection.selectedPaths.size}
	onClose={() => (state.selection.showBulkRemoveModal = false)}
	onConfirm={() => state.selection.confirmBulkRemove(() => state.refreshData())}
/>

<AddGameModal
	show={state.showAddModal}
	onClose={() => (state.showAddModal = false)}
	onRefresh={() => state.refreshData()}
/>

<style lang="scss">
	.home-container {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
		height: 100%;
		max-height: 100%;
		width: 100%;
		padding: 10px 0 0 0;
		background-color: transparent;
		gap: 24px;
		box-sizing: border-box;
		overflow: hidden;
	}

	.quick-launch-section {
		position: relative;
		display: flex;
		flex-direction: column;
		flex: 1;
		min-height: 0;
		height: 100%;
		max-height: 100%;
		overflow: hidden;
	}

	.status-line {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-shrink: 0;
		padding: 0 20px;
	}

	.status-text {
		font-size: 0.85rem;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.status-bar {
		flex: 1;
		max-width: 260px;
	}

	/* Scrolls so the placeholder grid never clips in a short window. */
	.loading-area {
		display: flex;
		flex: 1;
		min-height: 0;
		min-width: 0;
		max-width: 100%;
		padding-right: 8px;
		box-sizing: border-box;
		overflow-y: auto;
		overflow-x: hidden;
	}

	/* "safe" keeps the content reachable by scrolling when the window is too short to centre it. */
	.empty-area {
		display: flex;
		flex-direction: column;
		justify-content: safe center;
		flex: 1;
		min-height: 0;
		overflow-y: auto;
	}

	.empty-actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 12px;
	}
</style>
