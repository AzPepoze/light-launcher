<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { navigationCommand } from "@stores/navigationStore";
	import { HomePageState } from "@components/home/HomePageState.svelte";

	import GameGrid from "@components/home/GameGrid.svelte";
	import StatusDrawer from "@components/shared/StatusDrawer.svelte";
	import AddGameModal from "@components/home/addgame/AddGameModal.svelte";
	import RunningSessions from "@components/home/RunningSessions.svelte";
	import QuickLaunchHeader from "@components/home/QuickLaunchHeader.svelte";
	import HowItWorksModal from "@components/home/HowItWorksModal.svelte";
	import BulkRemoveModal from "@components/home/BulkRemoveModal.svelte";

	const state = new HomePageState();

	onMount(() => {
		state.initialize();
	});

	onDestroy(() => {
		state.destroy();
	});
</script>

<div class="home-container" data-file-drop-target>
	<RunningSessions sessions={state.sessions} onKill={(pid, name) => state.handleKillSession(pid, name)} />

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
			onShowAddModal={() => (state.showAddModal = true)}
			onShowHelpModal={() => (state.showHelpModal = true)}
		/>

		{#if state.games.length === 0 && state.scannedFolderGroups.length === 0}
			<div class="empty-state glass">
				<div class="empty-icon-bg">
					<span class="material-icons empty-icon">sports_esports</span>
				</div>
				<h2>Your Library is Empty</h2>
				<p>Import your existing game profiles or create custom configurations to start launching with LSFG frame generation.</p>
				<div class="empty-actions">
					<button class="btn primary" on:click={() => (state.showAddModal = true)}>
						<span class="material-icons">add</span> Add Game Profile
					</button>
					<button class="btn secondary" on:click={() => navigationCommand.set({ page: "run" })}>
						<span class="material-icons">settings</span> Create Custom Config
					</button>
				</div>
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
				handleQuickLaunch={(game) => state.handleQuickLaunch(game)}
				handleConfigure={(game) => state.handleConfigure(game)}
				toggleGameSelection={(game, shiftKey) => state.selection.toggleGameSelection(game, shiftKey)}
				onRefresh={() => state.refreshData(true)}
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

<StatusDrawer />

<style lang="scss">
	.home-container {
		display: flex;
		flex-direction: column;
		height: 100%;
		width: 100%;
		padding: 10px 0;
		background-color: transparent;
		gap: 36px;
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

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		padding: 60px 40px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		max-width: 600px;
		margin: 80px auto;
		gap: 20px;
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.35);

		.empty-icon-bg {
			width: 88px;
			height: 88px;
			background: rgba(255, 255, 255, 0.03);
			border: 2px solid rgba(255, 255, 255, 0.08);
			border-radius: 50%;
			display: flex;
			align-items: center;
			justify-content: center;
			margin-bottom: 8px;
			box-shadow: 0 0 24px rgba(255, 255, 255, 0.02);

			.empty-icon {
				font-size: 44px;
				color: var(--text-main);
			}
		}

		h2 {
			margin: 0;
			font-size: 1.5rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 1px;
		}

		p {
			margin: 0;
			font-size: 0.95rem;
			color: var(--text-muted);
			max-width: 440px;
			line-height: 1.6;
		}

		.empty-actions {
			display: flex;
			gap: 16px;
			margin-top: 12px;

			.btn {
				display: inline-flex;
				align-items: center;
				gap: 8px;
				padding: 12px 24px;
			}
		}
	}
</style>
