<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { fade } from "svelte/transition";
	import { cubicOut } from "svelte/easing";
	import { navigationCommand } from "@stores/navigationStore";
	import { HomePageState } from "@components/home/HomePageState.svelte";

	import GameGrid from "@components/home/GameGrid.svelte";
	import AddGameModal from "@components/home/addgame/AddGameModal.svelte";
	import QuickLaunchHeader from "@components/home/QuickLaunchHeader.svelte";
	import HowItWorksModal from "@components/home/HowItWorksModal.svelte";
	import BulkRemoveModal from "@components/home/BulkRemoveModal.svelte";

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

		{#if state.isLoading}
			<div class="loading-layer" out:fade={{ duration: prefersReducedMotion ? 0 : 240, easing: cubicOut }}>
				<div class="spinner"></div>
			</div>
		{:else if state.games.length === 0 && state.scannedFolderGroups.length === 0}
			<div class="empty-state">
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
				handleQuickLaunch={(game, showLogs) => state.handleQuickLaunch(game, showLogs)}
				handleConfigure={(game) => state.handleConfigure(game)}
				toggleGameSelection={(game, shiftKey, ctrlKey) => state.selection.toggleGameSelection(game, shiftKey, ctrlKey)}
				onRefresh={() => state.refreshData(true)}
				loadIcon={(path) => state.icons.enqueueIconLoad(path)}
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

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	.loading-layer {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: none;
	}

	.spinner {
		width: 46px;
		height: 46px;
		border-radius: 50%;
		border: 3px solid rgba(255, 255, 255, 0.08);
		border-top-color: var(--accent-primary);
		animation: spin 0.8s linear infinite;
	}

	@keyframes empty-pop {
		0% {
			opacity: 0;
			transform: scale(0.7);
		}
		60% {
			opacity: 1;
			transform: scale(1.05);
		}
		100% {
			opacity: 1;
			transform: scale(1);
		}
	}

	@keyframes empty-rise {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding: 48px 24px;
		gap: 16px;
		background: transparent;
		border: none;
		box-shadow: none;

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
			animation: empty-pop 400ms var(--ease-spring) both;

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
			opacity: 0;
			animation: empty-rise 300ms var(--ease-out) 60ms both;
		}

		p {
			margin: 0;
			font-size: 0.95rem;
			color: var(--text-muted);
			max-width: 440px;
			line-height: 1.6;
			opacity: 0;
			animation: empty-rise 300ms var(--ease-out) 120ms both;
		}

		.empty-actions {
			display: flex;
			gap: 16px;
			margin-top: 12px;
			opacity: 0;
			animation: empty-rise 300ms var(--ease-out) 180ms both;

			.btn {
				display: inline-flex;
				align-items: center;
				gap: 8px;
				padding: 12px 24px;
			}
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.empty-state .empty-icon-bg,
		.empty-state h2,
		.empty-state p,
		.empty-state .empty-actions {
			opacity: 1;
			animation: none;
			transform: none;
		}
	}
</style>
