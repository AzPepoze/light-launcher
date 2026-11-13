<script lang="ts">
	import { onDestroy, tick } from "svelte";
	import GameCard from "../cards/GameCard.svelte";
	import { findScrollableAncestor } from "@lib/autoscroll";
	import { getGameId, getGamePath } from "@lib/gameUtils";
	import type { GameInfo, RunningSession } from "@shared";
	import { GRID_CHUNK_SIZE, nextRenderedCount } from "./gridChunk";

	export let games: GameInfo[] = [];
	export let view: "grid" | "list-grid" | "sidebar-grid" = "grid";
	export let gameIcons: Record<string, string> = {};
	export let isGameRunning: (game: GameInfo, sessionsList: RunningSession[]) => boolean;
	export let sessions: RunningSession[] = [];
	export let isSelectionMode: boolean = false;
	export let selectedPaths: Set<string> = new Set();

	export let handleRightClick: (event: MouseEvent, game: GameInfo) => void = () => {};
	export let handleQuickLaunch: (game: GameInfo, showLogs?: boolean) => Promise<void> = async () => {};
	export let handleConfigure: (game: GameInfo) => void = () => {};
	export let handleChooseLauncher: (game: GameInfo, launcherPath: string | null) => void = () => {};
	export let toggleGameSelection: (game: GameInfo, shiftKey: boolean, ctrlKey?: boolean) => void = () => {};
	export let loadIcon: (path: string, customIconPath?: string | null) => void = () => {};

	$: cardView = view === "sidebar-grid" ? "grid" : view;

	let renderedCount = GRID_CHUNK_SIZE;
	let sentinelElement: HTMLElement;
	let observer: IntersectionObserver | null = null;

	// Same exe can appear twice (e.g. duplicate profiles); keys must stay unique.
	function dedupeGames(list: GameInfo[]): GameInfo[] {
		const seen = new Set<string>();
		return list.filter((game) => {
			const key = getGameId(game);
			if (seen.has(key)) return false;
			seen.add(key);
			return true;
		});
	}

	$: uniqueGames = dedupeGames(games);
	$: visibleGames = uniqueGames.slice(0, renderedCount);

	// A new list (filter or search) starts at one chunk again.
	$: {
		games;
		renderedCount = GRID_CHUNK_SIZE;
	}

	function handleSentinelIntersect(entries: IntersectionObserverEntry[]): void {
		if (!entries.some((entry) => entry.isIntersecting)) return;
		renderedCount = nextRenderedCount(renderedCount, uniqueGames.length);
	}

	async function connectSentinel(sentinel: HTMLElement): Promise<void> {
		// Wait for the DOM to update so the scroll root is measured on the new list.
		await tick();
		observer?.disconnect();
		observer = new IntersectionObserver(handleSentinelIntersect, {
			root: findScrollableAncestor(sentinel),
			rootMargin: "400px"
		});
		observer.observe(sentinel);
	}

	// Reconnect when the list changes; the scroll container may change with it.
	$: if (sentinelElement) {
		games;
		void connectSentinel(sentinelElement);
	}

	onDestroy(() => observer?.disconnect());
</script>

<div
	class="games-grid"
	class:grid-view={view === "grid" || view === "sidebar-grid"}
	class:list-view={view === "list-grid"}
>
	{#each visibleGames as game (getGameId(game))}
		{@const gamePath = getGamePath(game)}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div on:contextmenu|preventDefault|stopPropagation={(e) => handleRightClick(e, game)} data-game-path={gamePath}>
			<GameCard
				{game}
				icon={gameIcons[gamePath]}
				isRunning={isGameRunning(game, sessions)}
				{isSelectionMode}
				isSelected={selectedPaths.has(gamePath)}
				view={cardView}
				onLaunch={() => handleQuickLaunch(game)}
				onConfigure={() => handleConfigure(game)}
				onChooseLauncher={(launcherPath) => handleChooseLauncher(game, launcherPath)}
				onSelect={toggleGameSelection}
				{loadIcon}
			/>
		</div>
	{/each}
</div>
<div bind:this={sentinelElement} class="grid-sentinel" aria-hidden="true"></div>

<style lang="scss">
	.grid-sentinel {
		height: 1px;
	}

	.games-grid {
		display: grid;
		gap: 4px;
		width: 100%;
		max-width: 100%;
		padding: 18px;
		padding-bottom: 40px;
		box-sizing: border-box;
		min-width: 0;

		& > div {
			min-width: 0;
			max-width: 100%;
			overflow: visible;
		}

		&.grid-view {
			grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		}

		&.list-view {
			grid-template-columns: 1fr;
			gap: 4px;
			min-width: 0;
			overflow: visible;
		}
	}
</style>
