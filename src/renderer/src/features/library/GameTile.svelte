<script lang="ts">
	import CoverImage from "@components/shared/CoverImage.svelte";
	import type { Game } from "../../lib/features/types";
	import type { TileShape } from "./sourceTileShape";

	let {
		game,
		shape,
		revealIndex,
		onOpen
	}: {
		game: Game;
		shape: TileShape;
		/** Position in the first reveal, or undefined when the tile does not animate in. */
		revealIndex?: number;
		onOpen: (game: Game) => void;
	} = $props();

	const statusLine = $derived(
		[
			game.version ? `Version ${game.version}` : null,
			game.installed ? "Installed" : "Not installed"
		]
			.filter(Boolean)
			.join(" · ")
	);
</script>

<button
	class="tile"
	class:reveal={revealIndex !== undefined}
	style:--reveal-index={revealIndex}
	data-game-id={game.id}
	onclick={() => onOpen(game)}
	aria-label="Open {game.title}"
>
	<span class="cover" style:aspect-ratio={shape.ratio}>
		{#if game.cover || game.coverLocal}
			<CoverImage {game} framed />
		{:else}
			<span class="material-icons placeholder" aria-hidden="true">videogame_asset</span>
		{/if}
		<span class="badges">
			{#if game.favorite}
				<span class="material-icons badge" title="Favorite">favorite</span>
			{/if}
			{#if game.updateAvailable}
				<span class="badge update">Update</span>
			{/if}
		</span>
	</span>
	<span class="title">{game.title}</span>
	<span class="meta">{statusLine}</span>
</button>

<style lang="scss">
	@keyframes tile-reveal {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}

	.tile {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 0;
		border: none;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-main);
		text-align: left;
		cursor: pointer;
		font: inherit;

		&.reveal {
			animation: tile-reveal var(--motion-base) var(--ease-out) both;
			animation-delay: calc(var(--reveal-index) * var(--motion-stagger));
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}

		&:hover .cover,
		&:focus-visible .cover {
			transform: scale(1.03);
		}

		&:active .cover {
			transform: scale(0.97);
		}
	}

	.cover {
		display: flex;
		position: relative;
		width: 100%;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-md);
		overflow: hidden;
		background: var(--bg-elevated);
		transition: transform var(--motion-quick) var(--ease-out);
	}

	.placeholder {
		font-size: 40px;
		color: var(--text-muted);
	}

	.badges {
		position: absolute;
		top: 6px;
		right: 6px;
		display: flex;
		gap: 4px;
	}

	.badge {
		font-size: 16px;
		padding: 2px 6px;
		border-radius: var(--radius-pill);
		background: rgba(0, 0, 0, 0.7);
		color: var(--text-main);

		&.update {
			font-size: 0.75rem;
			font-weight: 800;
			background: var(--accent-primary);
			color: var(--bg-base);
		}
	}

	.title,
	.meta {
		display: block;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.title {
		font-size: 1.05rem;
		font-weight: 600;
		line-height: 1.3;
	}

	.meta {
		font-size: 0.875rem;
		color: var(--text-muted);
	}
</style>
