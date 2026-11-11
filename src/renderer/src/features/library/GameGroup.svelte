<script lang="ts">
	import type { Game } from "../../lib/features/types";
	import GameTile from "./GameTile.svelte";
	import type { GameGroup } from "./libraryGroups";

	let { group, onOpen }: { group: GameGroup; onOpen: (game: Game) => void } = $props();
</script>

<section class="group" aria-label={group.label}>
	<h2>
		{group.label}
		<span class="count">{group.games.length}</span>
	</h2>
	<ul class="grid" style:--tile-min="{group.shape.minColumnPx}px">
		{#each group.games as game (game.id)}
			<li>
				<GameTile {game} shape={group.shape} {onOpen} />
			</li>
		{/each}
	</ul>
</section>

<style lang="scss">
	.group {
		margin-bottom: 24px;
	}

	h2 {
		display: flex;
		align-items: center;
		gap: 8px;
		margin: 0 0 12px;
		font-size: 1rem;
		font-weight: 700;
	}

	.count {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--text-dim);
	}

	.grid {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(var(--tile-min), 1fr));
		gap: 18px 14px;
	}
</style>
