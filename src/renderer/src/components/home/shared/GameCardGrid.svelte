<script lang="ts">
	import GameCard from "@components/home/GameCard.svelte";
	import { getGamePath } from "@lib/gameUtils";

	export let games: any[] = [];
	export let view: "grid" | "list-grid" | "sidebar-grid" = "grid";
	export let gameIcons: Record<string, string> = {};
	export let isGameRunning: (game: any, sessionsList: any[]) => boolean;
	export let sessions: any[] = [];
	export let isSelectionMode: boolean = false;
	export let selectedPaths: Set<string> = new Set();

	export let handleRightClick: (event: MouseEvent, game: any) => void = () => {};
	export let handleQuickLaunch: (game: any) => Promise<void> = async () => {};
	export let handleConfigure: (game: any) => void = () => {};
	export let toggleGameSelection: (game: any, shiftKey: boolean) => void = () => {};
	export let loadIcon: (path: string) => void = () => {};

	$: cardView = view === "sidebar-grid" ? "grid" : view;
</script>

<div
	class="games-grid"
	class:grid-view={view === "grid" || view === "sidebar-grid"}
	class:list-view={view === "list-grid"}
>
	{#each games as game}
		{@const gamePath = getGamePath(game)}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div on:contextmenu|preventDefault|stopPropagation={(e) => handleRightClick(e, game)}>
			<GameCard
				{game}
				icon={gameIcons[gamePath]}
				isRunning={isGameRunning(game, sessions)}
				{isSelectionMode}
				isSelected={selectedPaths.has(gamePath)}
				view={cardView}
				onLaunch={() => handleQuickLaunch(game)}
				onConfigure={() => handleConfigure(game)}
				onSelect={toggleGameSelection}
				{loadIcon}
			/>
		</div>
	{/each}
</div>

<style lang="scss">
	.games-grid {
		display: grid;
		gap: 28px;
		width: 100%;
		padding: 12px;
		padding-bottom: 40px;

		&.grid-view {
			grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		}

		&.list-view {
			grid-template-columns: 1fr;
			gap: 16px;
		}
	}
</style>
