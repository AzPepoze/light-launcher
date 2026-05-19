<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { navigationCommand } from "@stores/navigationStore";
	import { HomePageState } from "@lib/HomePageState.svelte";

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
			isSelectionMode={state.isSelectionMode}
			selectedCount={state.selectedPaths.size}
			prefixes={state.prefixes}
			bind:selectedPrefixFilter={state.selectedPrefixFilter}
			bind:searchQuery={state.searchQuery}
			bind:currentView={state.currentView}
			onBulkRemove={() => state.handleBulkRemove()}
			onToggleSelectionMode={() => state.toggleSelectionMode()}
			onShowAddModal={() => (state.showAddModal = true)}
			onShowHelpModal={() => (state.showHelpModal = true)}
		/>

		{#if state.games.length === 0}
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
				currentView={state.currentView}
				games={state.games}
				filteredGames={state.filteredGames}
				gameIcons={state.gameIcons}
				searchQuery={state.searchQuery}
				selectedPrefixFilter={state.selectedPrefixFilter}
				isSelectionMode={state.isSelectionMode}
				selectedPaths={state.selectedPaths}
				sessions={state.sessions}
				isGameRunning={state.isGameRunning}
				handleQuickLaunch={(game) => state.handleQuickLaunch(game)}
				handleConfigure={(game) => state.handleConfigure(game)}
				toggleGameSelection={(game) => state.toggleGameSelection(game)}
			/>
		{/if}
	</div>
</div>

<HowItWorksModal show={state.showHelpModal} onClose={() => (state.showHelpModal = false)} />

<BulkRemoveModal
	show={state.showBulkRemoveModal}
	selectedCount={state.selectedPaths.size}
	onClose={() => (state.showBulkRemoveModal = false)}
	onConfirm={() => state.confirmBulkRemove()}
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
