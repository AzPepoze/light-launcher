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

	let {
		game,
		ratio,
		card,
		variant,
		transfer,
		blurUntilHover,
		onOpen
	}: {
		game: Game;
		/** Width divided by height of the cover, which is the whole tile. */
		ratio: number;
		card: SourceCard | null;
		/** The card look: media covers, or the square Home-style frame for local games. */
		variant: "cover" | "home";
		/** The download or install running for this game, if any. */
		transfer: TransferKind | null;
		/** Covers stay blurred until the tile is hovered or focused. The overlay is never blurred. */
		blurUntilHover: boolean;
		onOpen: (game: Game) => void;
	} = $props();

	const view = $derived(cardView(game, card, transfer));
	const cornerStart = $derived(view.states.filter((state) => state === "favorite"));
	const cornerEnd = $derived(view.states.filter((state) => state !== "favorite"));
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
		const hasCover = Boolean(item.cover || item.coverLocal);
		if (item.source !== LOCAL_SOURCE || hasCover) return null;
		return item.exePath || item.localConfig?.GamePath || item.localConfig?.LauncherPath || null;
	}

	const STATE_VIEW: Record<TileState, { label: string; icon: string; tone: "danger" | "accent" | "neutral" | "warning" }> = {
		favorite: { label: "Favorite", icon: "favorite", tone: "danger" },
		updateAvailable: { label: "Update", icon: "system_update_alt", tone: "warning" },
		notInstalled: { label: "Not installed", icon: "download", tone: "success" },
		downloading: { label: "Downloading", icon: "downloading", tone: "warning" },
		installing: { label: "Installing", icon: "progress_activity", tone: "warning" }
	};
</script>

{#snippet stateBadge(state: TileState)}
	<span class="state-badge {state}" role="img" aria-label={STATE_VIEW[state].label} title={STATE_VIEW[state].label}>
		<Badge icon={STATE_VIEW[state].icon} tone={STATE_VIEW[state].tone} />
	</span>
{/snippet}

<button
	class="tile"
	class:home-card={variant === "home"}
	data-game-id={game.id}
	onclick={() => onOpen(game)}
	aria-label="Open {view.title}"
>
	<span
		class="cover"
		class:blurred={blurUntilHover && variant === "cover"}
		style:aspect-ratio={ratio}
	>
		{#if game.cover || game.coverLocal}
			<CoverImage {game} />
		{:else if exeIcon}
			<img class="exe-icon" src={exeIcon} alt="" />
		{:else}
			<Icon name="videogame_asset" size={variant === "home" ? 56 : 40} class="placeholder" />
		{/if}

		{#if cornerStart.length > 0}
			<span class="corner start">
				{#each cornerStart as state (state)}
					{@render stateBadge(state)}
				{/each}
			</span>
		{/if}
		{#if cornerEnd.length > 0}
			<span class="corner end">
				{#each cornerEnd as state (state)}
					{@render stateBadge(state)}
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

		/* The exe icon is small art: it keeps its shape and sits centred on the tile. */
		.exe-icon {
			width: 48%;
			height: 48%;
			object-fit: contain;
		}
	}

	/* Local games use the Home card's padded square frame, minus its play overlay. */
	.tile.home-card {
		padding: 8px;
		box-sizing: border-box;
		border-radius: var(--radius-lg);

		.cover {
			border: 2px solid rgb(255 255 255 / 0.06);
			border-radius: var(--radius-lg);
			background: var(--bg-base);
			transition:
				transform var(--motion-base) var(--ease-out),
				border-color var(--motion-base) var(--ease-out),
				box-shadow var(--motion-base) var(--ease-out);

			:global(img) {
				transition:
					transform var(--motion-quick) var(--ease-out),
					filter var(--motion-quick) var(--ease-out);
			}
		}

		.exe-icon {
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		&:hover .cover,
		&:focus-visible .cover {
			transform: translateY(-4px) scale(1.015);
			border-color: rgb(255 255 255 / 0.32);
			box-shadow: 0 0 16px rgb(255 255 255 / 0.28);
		}

		&:hover :global(img),
		&:focus-visible :global(img) {
			transform: scale(1.06);
			filter: brightness(1.05);
		}

		&:active .cover {
			transform: scale(0.98);
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

	/* Solid round icon badges, coloured by the tone the badge gets from its state. */
	.corner :global(.badge) {
		width: 26px;
		height: 26px;
		padding: 0;
		justify-content: center;
		border: none;
		border-radius: 50%;
		background: var(--tone);
		color: var(--bg-base);
		backdrop-filter: none;
	}

	/* Update is yellow and install is green, both with a dark icon. */
	.state-badge.updateAvailable :global(.badge) {
		--tone: var(--tile-update);
	}

	.state-badge.notInstalled :global(.badge) {
		--tone: var(--success);
	}

	/* Badge sets its icon to 14px inline, so this must win over that inline style. */
	.corner :global(.badge .material-symbols-rounded) {
		font-size: 16px !important;
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

		.tile.home-card {
			&:hover .cover,
			&:focus-visible .cover,
			&:active .cover {
				transform: none;
			}

			&:hover :global(img),
			&:focus-visible :global(img) {
				transform: none;
				filter: none;
			}
		}
	}
</style>
