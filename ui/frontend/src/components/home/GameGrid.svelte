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
	
	export let isGameRunning: (game: any, sessionsList: any[]) => boolean;
	export let sessions: any[] = [];
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
					isRunning={isGameRunning(game, sessions)}
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
		padding-right: 8px;

		&::-webkit-scrollbar {
			width: 8px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 102, 171, 0.2);
			border-radius: 10px;
		}
	}

	.games-grid {
		display: grid;
		gap: 28px;
		width: 100%;
		padding: 12px;
		padding-bottom: 40px;
	}

	.grid-view .games-grid {
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
	}

	.list-view .games-grid {
		grid-template-columns: 1fr;
		gap: 16px;
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
</style>
