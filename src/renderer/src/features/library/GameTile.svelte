<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import CoverImage from "@components/shared/CoverImage.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import Marquee from "@components/shared/Marquee.svelte";
	import type { Game, SourceCard } from "../../lib/features/types";
	import { cardView, type TileState } from "./tileCard";
	import type { TransferKind } from "./transferState";

	let {
		game,
		ratio,
		card,
		transfer,
		blurUntilHover,
		revealIndex,
		onOpen
	}: {
		game: Game;
		/** Width divided by height of the cover, which is the whole tile. */
		ratio: number;
		card: SourceCard | null;
		/** The download or install running for this game, if any. */
		transfer: TransferKind | null;
		/** Covers stay blurred until the tile is hovered or focused. The overlay is never blurred. */
		blurUntilHover: boolean;
		/** Position in the first reveal, or undefined when the tile does not animate in. */
		revealIndex?: number;
		onOpen: (game: Game) => void;
	} = $props();

	const view = $derived(cardView(game, card, transfer));
	const cornerStart = $derived(view.states.filter((state) => state === "favorite"));
	const cornerEnd = $derived(view.states.filter((state) => state !== "favorite"));

	const STATE_VIEW: Record<TileState, { label: string; icon: string; tone: "danger" | "accent" | "neutral" | "warning" }> = {
		favorite: { label: "Favorite", icon: "favorite", tone: "danger" },
		updateAvailable: { label: "Update", icon: "system_update", tone: "accent" },
		notInstalled: { label: "Not installed", icon: "download", tone: "neutral" },
		downloading: { label: "Downloading", icon: "downloading", tone: "warning" },
		installing: { label: "Installing", icon: "progress_activity", tone: "warning" }
	};
</script>

<button
	class="tile"
	class:reveal={revealIndex !== undefined}
	style:--reveal-index={revealIndex}
	data-game-id={game.id}
	onclick={() => onOpen(game)}
	aria-label="Open {view.title}"
>
	<span class="cover" class:blurred={blurUntilHover} style:aspect-ratio={ratio}>
		{#if game.cover || game.coverLocal}
			<CoverImage {game} />
		{:else}
			<Icon name="videogame_asset" size={40} class="placeholder" />
		{/if}

		{#if cornerStart.length > 0}
			<span class="corner start">
				{#each cornerStart as state (state)}
					<Badge icon={STATE_VIEW[state].icon} tone={STATE_VIEW[state].tone}>{STATE_VIEW[state].label}</Badge>
				{/each}
			</span>
		{/if}
		{#if cornerEnd.length > 0}
			<span class="corner end">
				{#each cornerEnd as state (state)}
					<Badge icon={STATE_VIEW[state].icon} tone={STATE_VIEW[state].tone}>{STATE_VIEW[state].label}</Badge>
				{/each}
			</span>
		{/if}

		<span class="overlay">
			<Marquee text={view.title} class="name" />
			{#if view.subtitle}<span class="subtitle">{view.subtitle}</span>{/if}
			{#if view.details.length > 0}<span class="details">{view.details.join(" · ")}</span>{/if}
		</span>
	</span>
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
		width: 100%;
		min-width: 0;
		box-sizing: border-box;
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

		&:hover .cover.blurred :global(img),
		&:focus-visible .cover.blurred :global(img) {
			filter: none;
			transform: none;
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

		&.blurred :global(img) {
			filter: blur(18px);
			transform: scale(1.1);
			transition:
				filter var(--motion-base) var(--ease-out),
				transform var(--motion-base) var(--ease-out),
				opacity var(--motion-base) var(--ease-out);
		}

		:global(.frame) {
			position: absolute;
			inset: 0;
		}

		:global(.placeholder) {
			color: var(--text-muted);
		}
	}

	.corner {
		position: absolute;
		top: 6px;
		display: flex;
		flex-wrap: wrap;
		gap: 4px;

		&.start {
			left: 6px;
		}

		&.end {
			right: 6px;
			justify-content: flex-end;
		}
	}

	/* Bottom overlay: fades in with a 6px slide on hover and focus, always shown on touch. */
	.overlay {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 28px 10px 10px;
		background: linear-gradient(to top, rgb(0 0 0 / 0.75), transparent);
		color: #fff;
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 160ms var(--ease-out),
			transform 160ms var(--ease-out);
		pointer-events: none;

		:global(.name) {
			display: block;
			overflow: hidden;
			white-space: nowrap;
			text-overflow: ellipsis;
			font-size: 0.95rem;
			font-weight: 600;
			line-height: 1.25;
		}

		:global(.name.scrolling) {
			overflow: visible;
			text-overflow: clip;
			width: max-content;
		}
	}

	.tile:hover .overlay,
	.tile:focus-visible .overlay {
		opacity: 1;
		transform: none;
	}

	// Hold the slide until the tile is hovered or focused, as before.
	.tile:not(:hover):not(:focus-visible) :global(.name.scrolling) {
		animation-play-state: paused;
	}

	.subtitle,
	.details {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		font-size: 0.8125rem;
	}

	.subtitle {
		color: rgb(255 255 255 / 0.8);
	}

	.details {
		color: rgb(255 255 255 / 0.6);
	}

	/* Touch screens have no hover, so the overlay stays on. */
	@media (hover: none) {
		.overlay {
			opacity: 1;
			transform: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.overlay {
			transform: none;
			transition: opacity 160ms linear;
		}
	}
</style>
