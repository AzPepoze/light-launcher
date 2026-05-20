<script lang="ts">
	import GameCard from "@components/home/GameCard.svelte";

	export let filteredGames: any[] = [];
	export let gameIcons: Record<string, string> = {};
	export let isGameRunning: (game: any, sessions: any[]) => boolean;
	export let sessions: any[] = [];
	export let isSelectionMode: boolean = false;
	export let selectedPaths = new Set<string>();
	export let handleRightClick: (event: MouseEvent, game: any) => void;
	export let handleQuickLaunch: (game: any) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let toggleGameSelection: (game: any, shiftKey: boolean) => void;
</script>

<div class="folder-group-header">
	<div class="folder-title">
		<span class="material-icons folder-icon-main">sports_esports</span>
		<div class="folder-metadata">
			<span class="folder-name">No Folder</span>
			<span class="folder-path">Manually registered custom profiles</span>
		</div>
		<span class="badge">{filteredGames.length}</span>
	</div>
</div>

{#if filteredGames.length > 0}
	<div class="games-grid">
		{#each filteredGames as game}
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div on:contextmenu|preventDefault|stopPropagation={(e) => handleRightClick(e, game)}>
				<GameCard
					{game}
					icon={gameIcons[game.path || game.config.LauncherPath]}
					isRunning={isGameRunning(game, sessions)}
					{isSelectionMode}
					isSelected={selectedPaths.has(game.path || game.config.LauncherPath)}
					view="grid"
					onLaunch={() => handleQuickLaunch(game)}
					onConfigure={() => handleConfigure(game)}
					onSelect={toggleGameSelection}
				/>
			</div>
		{/each}
	</div>
{:else}
	<div class="folder-empty-placeholder">
		<span class="material-icons placeholder-icon">sports_esports</span>
		<span class="placeholder-text">No custom profiles in this section.</span>
		<span class="placeholder-subtext"
			>Add specific games or configure monitored watched directories to list games.</span
		>
	</div>
{/if}

<style lang="scss">
	.folder-group-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
		padding-bottom: 16px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.folder-title {
		display: flex;
		align-items: center;
		gap: 14px;
		min-width: 0;
		flex: 1;
	}

	.folder-icon-main {
		font-size: 28px;
		color: var(--accent-secondary);
		filter: drop-shadow(0 0 8px rgba(255, 102, 171, 0.3));
	}

	.folder-metadata {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	.folder-name {
		font-size: 1.15rem;
		font-weight: 900;
		color: var(--text-main);
		letter-spacing: -0.3px;
	}

	.folder-path {
		font-size: 0.75rem;
		color: var(--text-muted);
		opacity: 0.6;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.folder-group-header .badge {
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-sm);
		padding: 2px 8px;
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--text-muted);
		flex-shrink: 0;
	}

	.games-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 28px;
		width: 100%;
		padding: 12px;
		padding-bottom: 40px;
	}

	.folder-empty-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		color: var(--text-muted);
		text-align: center;
		background: rgba(0, 0, 0, 0.15);
		border-radius: var(--radius-md);
		border: 2px dashed rgba(255, 255, 255, 0.04);
		margin: 8px 12px;
		width: calc(100% - 24px);

		.placeholder-icon {
			font-size: 36px;
			color: var(--text-muted);
			opacity: 0.3;
			margin-bottom: 12px;
		}

		.placeholder-text {
			font-size: 0.95rem;
			font-weight: 700;
			color: var(--text-muted);
			margin-bottom: 4px;
		}

		.placeholder-subtext {
			font-size: 0.75rem;
			opacity: 0.6;
			max-width: 400px;
			line-height: 1.4;
		}
	}
</style>
