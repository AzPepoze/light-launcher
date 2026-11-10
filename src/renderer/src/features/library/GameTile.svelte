<script lang="ts">
	import type { Game } from "../../lib/features/types";

	let { game, onOpen }: { game: Game; onOpen: (game: Game) => void } = $props();
</script>

<button class="tile" onclick={() => onOpen(game)} aria-label="Open {game.title}">
	<div class="cover">
		{#if game.cover}
			<img src={game.cover} alt="" loading="lazy" />
		{:else}
			<span class="material-icons placeholder" aria-hidden="true">videogame_asset</span>
		{/if}
		<div class="badges">
			{#if game.favorite}
				<span class="material-icons badge" title="Favorite">favorite</span>
			{/if}
			{#if game.updateAvailable}
				<span class="badge update">Update</span>
			{/if}
		</div>
	</div>
	<p class="title">{game.title}</p>
	<p class="meta">{game.installed ? "Installed" : game.source}</p>
</button>

<style lang="scss">
	.tile {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 0;
		border: 2px solid transparent;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-main);
		text-align: left;
		cursor: pointer;
		font: inherit;
		transition: border-color var(--transition-fast), transform var(--transition-fast);

		&:hover {
			border-color: var(--glass-border-bright);
			transform: translateY(-2px);
		}

		&:focus-visible {
			outline: none;
			border-color: var(--accent-primary);
		}
	}

	.cover {
		position: relative;
		width: 100%;
		aspect-ratio: 3 / 4;
		border-radius: var(--radius-md);
		overflow: hidden;
		background: var(--bg-elevated);
		display: flex;
		align-items: center;
		justify-content: center;

		img {
			width: 100%;
			height: 100%;
			object-fit: cover;
		}
	}

	.placeholder {
		font-size: 40px;
		color: var(--text-dim);
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
			font-size: 11px;
			font-weight: 800;
			background: var(--accent-primary);
			color: var(--bg-base);
		}
	}

	.title {
		margin: 0;
		font-size: 0.85rem;
		font-weight: 700;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.meta {
		margin: 0;
		font-size: 0.72rem;
		color: var(--text-dim);
		text-transform: capitalize;
	}
</style>
