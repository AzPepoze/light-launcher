<script lang="ts">
	import { onMount, tick } from "svelte";
	import { fade } from "svelte/transition";
	import { cubicOut } from "svelte/easing";
	import { observeOnce } from "@lib/lazyObserver";
	import Icon from "@components/shared/Icon.svelte";
	import GridCard from "./GridCard.svelte";
	import ListGridCard from "./ListGridCard.svelte";
	import { useCover } from "@lib/features/useMedia.svelte";
	import { hasLibraryCover, needsExeIcon } from "@lib/iconService";
	import { getGamePath } from "@lib/gameUtils";

	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let view: "grid" | "list-grid" = "grid";
	export let isSelectionMode: boolean = false;
	export let isSelected: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};
	export let onChooseLauncher: (launcherPath: string | null) => void = () => {};
	export let onSelect: (game: any, shiftKey: boolean, ctrlKey?: boolean) => void = () => {};
	export let loadIcon: (
		path: string,
		customIconPath?: string | null,
		hasCover?: boolean
	) => void | Promise<unknown> = () => {};

	let containerElement: HTMLElement;
	let isIntersecting = false;
	/** True while the icon request started by this card has not settled. */
	let iconPending = false;
	let iconRequestId = 0;
	const prefersReducedMotion =
		typeof window !== "undefined" && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

	// The library cover, used in the icon slot when there is no custom icon.
	const cover = useCover(() => ({
		id: game.id,
		cover: game.cover ?? null,
		coverLocal: game.coverLocal ?? null
	}));

	$: hasCover = hasLibraryCover(game);
	$: displayIcon = icon || (hasCover && !iconPending ? (cover.url ?? "") : "");
	$: displayPending = !displayIcon && (iconPending || (hasCover && cover.pending));

	onMount(() => {
		if (!containerElement) return;
		return observeOnce(containerElement, () => {
			isIntersecting = true;
		});
	});

	$: if (isIntersecting) {
		const path = getGamePath(game);
		const customIconPath = game.config?.CustomIconPath ?? null;
		// A covered game without a custom icon never asks for an exe icon.
		if (path && (customIconPath || needsExeIcon(game))) {
			void requestIcon(path, customIconPath, hasCover);
		} else {
			// Nothing to load from: the card shows its cover or the rocket fallback.
			iconPending = false;
		}
	}

	async function requestIcon(
		path: string,
		customIconPath: string | null,
		hasCoverForIcon: boolean
	): Promise<void> {
		const request = ++iconRequestId;
		iconPending = true;
		try {
			// The shared loader resolves once the icon cache has an answer, including "no icon".
			await Promise.resolve(loadIcon(path, customIconPath, hasCoverForIcon));
		} catch {
			// A failed load ends in the same fallback as a game without an icon.
		}
		// Let the card's `icon` prop catch up with the cache before declaring the icon missing.
		await tick();
		if (request !== iconRequestId) return;
		iconPending = false;
	}
</script>

<div bind:this={containerElement} class="lazy-card-container {view}">
	{#if isIntersecting}
		<div
			class="card-enter"
			in:fade={{ duration: prefersReducedMotion ? 0 : 180, easing: cubicOut }}
		>
			{#if view === "grid"}
				<GridCard
					{game}
					icon={displayIcon}
					iconPending={displayPending}
					{isRunning}
					{isSelectionMode}
					{isSelected}
					{onLaunch}
					{onConfigure}
					{onChooseLauncher}
					{onSelect}
				/>
			{:else}
				<ListGridCard
					{game}
					icon={displayIcon}
					iconPending={displayPending}
					{isRunning}
					{isSelectionMode}
					{isSelected}
					{onLaunch}
					{onConfigure}
					{onChooseLauncher}
					{onSelect}
				/>
			{/if}
		</div>
	{:else}
		{#if view === "list-grid"}
			<div class="skeleton-list-card">
				<div class="skeleton-icon-box">
					<Icon name="sports_esports" size={32} class="skeleton-logo" />
				</div>
				<div class="skeleton-info">
					<div class="skeleton-line name"></div>
					<div class="skeleton-line path"></div>
				</div>
			</div>
		{:else}
			<div class="skeleton-grid-card">
				<div class="skeleton-icon-large skeleton-shimmer">
					<Icon name="sports_esports" size={32} class="skeleton-logo" />
				</div>
			</div>
		{/if}
	{/if}
</div>

<style lang="scss">
	.lazy-card-container {
		width: 100%;
		max-width: 100%;
		min-width: 0;
		box-sizing: border-box;
		overflow: visible;

		&.list-grid {
			min-width: 0;
			max-width: 100%;
			overflow: visible;
		}
	}

	.card-enter {
		width: 100%;
		min-width: 0;
	}

	@keyframes skeleton-pulse {
		0%, 100% {
			opacity: 0.15;
		}
		50% {
			opacity: 0.35;
		}
	}

	@keyframes shimmer-sweep {
		0% {
			background-position: 200% 0;
		}
		100% {
			background-position: -200% 0;
		}
	}

	:global(.skeleton-logo) {
		font-size: 32px;
		color: var(--text-muted, #fff);
		animation: skeleton-pulse 2s infinite ease-in-out;
	}

	.skeleton-grid-card {
		width: calc(100% - 12px);
		max-width: 190px;
		margin: 6px;
		padding: 8px;
		box-sizing: border-box;

		.skeleton-icon-large {
			width: 100%;
			aspect-ratio: 1;
			background: rgba(255, 255, 255, 0.03);
			border: 2px solid rgba(255, 255, 255, 0.05);
			border-radius: var(--radius-lg, 12px);
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}

	.skeleton-list-card {
		display: flex;
		align-items: center;
		background: var(--bg-surface, rgba(255, 255, 255, 0.02));
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg, 12px);
		padding: 8px 16px;
		gap: 12px;
		width: 100%;
		box-sizing: border-box;

		.skeleton-icon-box {
			height: 108px;
			width: 108px;
			border-radius: var(--radius-md, 8px);
			background: rgba(0, 0, 0, 0.2);
			border: 2px solid rgba(255, 255, 255, 0.05);
			display: flex;
			align-items: center;
			justify-content: center;
			flex-shrink: 0;
		}

		.skeleton-info {
			flex: 1;
			display: flex;
			flex-direction: column;
			gap: 8px;
			min-width: 0;
		}

		.skeleton-line {
			background: rgba(255, 255, 255, 0.04);
			border-radius: var(--radius-sm, 4px);
			animation: skeleton-pulse 2s infinite ease-in-out;

			&.name {
				height: 16px;
				width: 30%;
			}

			&.path {
				height: 10px;
				width: 50%;
			}
		}
	}

	// Placed after base rules so it overrides
	.skeleton-grid-card .skeleton-shimmer {
		background: linear-gradient(
			90deg,
			rgba(255, 255, 255, 0.04) 25%,
			rgba(255, 255, 255, 0.09) 37%,
			rgba(255, 255, 255, 0.04) 63%
		) !important;
		background-size: 400% 100% !important;
		animation: shimmer-sweep 1.6s infinite linear !important;
	}

	@media (prefers-reduced-motion: reduce) {
		.skeleton-grid-card .skeleton-shimmer {
			animation: none !important;
			background: rgba(255, 255, 255, 0.04) !important;
		}
	}
</style>
