<script lang="ts">
	import type { Game } from "../../lib/features/types";
	import GameTile from "./GameTile.svelte";
	import type { GameGroup } from "./libraryGroups";

	let {
		group,
		reveal,
		onOpen
	}: { group: GameGroup; reveal: Map<number, number>; onOpen: (game: Game) => void } = $props();

	const uid = $props.id();
	const bodyId = `${uid}-games`;
	let expanded = $state(true);
</script>

<section class="group" aria-label={group.label}>
	<h2>
		<button
			class="toggle"
			aria-expanded={expanded}
			aria-controls={bodyId}
			onclick={() => (expanded = !expanded)}
		>
			<span class="label">{group.label}</span>
			<span class="count">{group.games.length}</span>
			<span class="material-icons chevron" class:turned={!expanded} aria-hidden="true">expand_more</span>
		</button>
	</h2>
	<div id={bodyId} class="body" class:collapsed={!expanded} inert={!expanded}>
		<div class="clip">
			<ul class="grid" style:--tile-min="{group.shape.minColumnPx}px">
				{#each group.games as game (game.id)}
					<li>
						<GameTile
							{game}
							shape={group.shape}
							sourceLabel={group.label}
							revealIndex={reveal.get(game.id)}
							{onOpen}
						/>
					</li>
				{/each}
			</ul>
		</div>
	</div>
</section>

<style lang="scss">
	.group {
		margin-bottom: 28px;
	}

	h2 {
		margin: 0 0 14px;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.toggle {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		padding: 4px 8px 4px 0;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--text-main);
		font: inherit;
		cursor: pointer;

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	.count {
		padding: 1px 10px;
		border-radius: var(--radius-pill);
		background: var(--bg-elevated);
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-muted);
	}

	.chevron {
		color: var(--text-muted);
		transition: transform var(--motion-quick) var(--ease-out);

		&.turned {
			transform: rotate(-90deg);
		}
	}

	.body {
		display: grid;
		grid-template-rows: 1fr;
		transition: grid-template-rows var(--motion-base) var(--ease-out);

		&.collapsed {
			grid-template-rows: 0fr;
		}
	}

	.clip {
		min-height: 0;
		overflow: hidden;
	}

	/* Room inside the clip for the focus ring of the edge tiles. */
	.grid {
		list-style: none;
		margin: 0;
		padding: 4px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(var(--tile-min), 1fr));
		gap: 18px 14px;
	}
</style>
