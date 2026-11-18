<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import CoverImage from "@components/shared/CoverImage.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import Marquee from "@components/shared/Marquee.svelte";
	import { loadGameIcon } from "../../lib/iconService";
	import type { Game, SourceCard } from "../../lib/features/types";
	import { LOCAL_SOURCE } from "./librarySections";
	import { cardView, type TileState } from "./tileCard";
	import type { TransferKind } from "./transferState";

	type CornerState = "favorite" | "notInstalled" | TransferKind;

	/** The pill shown in a tile corner. Only these states get one. */
	const CORNER_VIEW: Record<CornerState, { label: string; icon: string; tone: "danger" | "neutral" | "warning" }> = {
		favorite: { label: "Favorite", icon: "favorite", tone: "danger" },
		notInstalled: { label: "Not installed", icon: "download", tone: "neutral" },
		downloading: { label: "Downloading", icon: "downloading", tone: "warning" },
		installing: { label: "Installing", icon: "progress_activity", tone: "warning" }
	};

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
		/** Width divided by height of the cover. The name line sits under it. */
		ratio: number;
		card: SourceCard | null;
		/** The download or install running for this game, if any. */
		transfer: TransferKind | null;
		/** Covers stay blurred until the tile is hovered or focused. */
		blurUntilHover: boolean;
		/** Position in the first reveal, or undefined when the tile does not animate in. */
		revealIndex?: number;
		onOpen: (game: Game) => void;
	} = $props();

	const view = $derived(cardView(game, card, transfer));
	const corners = $derived(view.states.filter(isCornerState));
	const hasCover = $derived(Boolean(game.cover || game.coverLocal));
	const iconPath = $derived(exeIconPathFor(game));
	/** The custom or exe icon as a data URL, shown only for a local game with no cover. */
	let exeIcon = $state("");

	// Runs once per game path. Only a stale request is dropped, so a reused tile never shows the old icon.
	$effect(() => {
		const path = iconPath;
		exeIcon = "";
		if (!path) return;
		let current = true;
		void loadGameIcon(
			path,
			game.localConfig?.CustomIconPath ?? null,
			"low",
			false
		).then((result) => {
			if (current) exeIcon = result.icon;
		});
		return () => {
			current = false;
		};
	});

	/** The exe to read an icon from: a local game with no cover and a known exe. Otherwise null. */
	function exeIconPathFor(item: Game): string | null {
		const coverless = !(item.cover || item.coverLocal);
		if (item.source !== LOCAL_SOURCE || !coverless) return null;
		return item.exePath || item.localConfig?.GamePath || item.localConfig?.LauncherPath || null;
	}

	function isCornerState(state: TileState): state is CornerState {
		return state in CORNER_VIEW;
	}
</script>

<button
	class="tile"
	class:reveal={revealIndex !== undefined}
	style:--reveal-index={revealIndex}
	data-game-id={game.id}
	onclick={() => onOpen(game)}
	aria-label="Open {view.title}"
>
	<span class="cover" style:aspect-ratio={ratio}>
		{#if hasCover}
			<span class="art" class:blurred={blurUntilHover}>
				<CoverImage {game} />
			</span>
		{:else if exeIcon}
			<img class="exe-icon" src={exeIcon} alt="" />
		{:else}
			<Icon name="videogame_asset" size={40} class="placeholder" />
		{/if}

		{#if corners.length > 0}
			<span class="corners">
				{#each corners as state (state)}
					<Badge icon={CORNER_VIEW[state].icon} tone={CORNER_VIEW[state].tone}>{CORNER_VIEW[state].label}</Badge>
				{/each}
			</span>
		{/if}
	</span>

	<span class="label">
		<Marquee text={view.title} class="name" />
		{#if view.subtitle}<span class="subtitle">{view.subtitle}</span>{/if}
	</span>
</button>

<style lang="scss">
	@keyframes tile-reveal {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}

	/* Every tile has the same shell: a rounded cover, then one name line. */
	.tile {
		display: flex;
		flex-direction: column;
		gap: 8px;
		width: 100%;
		min-width: 0;
		box-sizing: border-box;
		padding: 0;
		border: none;
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
	}

	.cover {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		border-radius: var(--library-tile-radius);
		overflow: hidden;
		background: var(--bg-elevated);

		:global(.placeholder) {
			color: var(--text-muted);
		}

		/* The exe icon is small art: it keeps its shape and sits centred on the tile. */
		.exe-icon {
			width: 56%;
			height: 56%;
			object-fit: contain;
		}
	}

	.art {
		display: block;
		width: 100%;
		height: 100%;

		:global(img) {
			display: block;
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		&.blurred :global(img) {
			filter: blur(18px);
			transform: scale(1.1);
			transition:
				filter var(--motion-base) var(--ease-out),
				transform var(--motion-base) var(--ease-out);
		}
	}

	.tile:hover .art.blurred :global(img),
	.tile:focus-visible .art.blurred :global(img) {
		filter: none;
		transform: none;
	}

	.corners {
		position: absolute;
		top: 8px;
		left: 8px;
		right: 8px;
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	/* One line under the cover, the same height on every tile. libraryWindow.TILE_LABEL_PX matches it. */
	.label {
		display: flex;
		align-items: baseline;
		gap: 8px;
		min-width: 0;
		height: 24px;
		font-size: 0.9375rem;
		line-height: 24px;

		:global(.name) {
			flex: 1 1 auto;
			min-width: 0;
			font-weight: 600;
		}
	}

	.subtitle {
		flex-shrink: 0;
		max-width: 40%;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		color: var(--text-muted);
		font-size: 0.8125rem;
	}

	@media (prefers-reduced-motion: reduce) {
		.tile.reveal {
			animation: none;
		}
	}
</style>
