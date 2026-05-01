<script lang="ts">
	import GameCard from "@components/home/GameCard.svelte";

	export let currentView: "grid" | "list-grid" = "grid";
	export let games: any[] = [];
	export let filteredGames: any[] = [];
	export let gameIcons: Record<string, string> = {};
	export let searchQuery = "";
	export let selectedPrefixFilter = "All Prefixes";
	export let isSelectionMode = false;
	export let selectedPaths = new Set<string>();
	
	export let isGameRunning: (game: any) => boolean;
	export let handleQuickLaunch: (game: any) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let toggleGameSelection: (game: any) => void;
</script>

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

<style lang="scss">
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
		gap: 12px;
		width: 100%;
		padding: 8px;
		padding-bottom: 40px;
	}

	.grid-view .games-grid {
		grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
	}

	.list-view .games-grid {
		grid-template-columns: 1fr;
		gap: 10px;
	}

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60px 20px;
		color: var(--text-dim);
		text-align: center;
		background: var(--glass-surface);
		border-radius: var(--radius-xl);
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
</style>
