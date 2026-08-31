<script lang="ts">
	import { onMount } from "svelte";
	import GridCard from "@components/home/gamecards/GridCard.svelte";
	import ListGridCard from "@components/home/gamecards/ListGridCard.svelte";
	import PerspectiveCard from "@components/home/gamecards/PerspectiveCard.svelte";

	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let view: "grid" | "list-grid" | "perspective" = "grid";
	export let active: boolean = false; // For perspective view
	export let isSelectionMode: boolean = false;
	export let isSelected: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};
	export let onSelect: (game: any, shiftKey: boolean) => void = () => {};
	export let loadIcon: (path: string) => void = () => {};

	let containerElement: HTMLElement;
	let isIntersecting = false;

	onMount(() => {
		if (!containerElement) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting) {
					isIntersecting = true;
					const path = game.path || game.config.LauncherPath;
					if (path) {
						loadIcon(path);
					}
					observer.disconnect();
				}
			},
			{
				rootMargin: "250px", // Preload cards that are 250px below the viewport
			}
		);

		observer.observe(containerElement);
		return () => observer.disconnect();
	});

	// If the game path changes and we have already intersected, load the new icon
	$: if (isIntersecting) {
		const path = game.path || game.config.LauncherPath;
		if (path) {
			loadIcon(path);
		}
	}
</script>

<div bind:this={containerElement} class="lazy-card-container {view}">
	{#if isIntersecting}
		{#if view === "grid"}
			<GridCard
				{game}
				{icon}
				{isRunning}
				{isSelectionMode}
				{isSelected}
				{onLaunch}
				{onConfigure}
				{onSelect}
			/>
		{:else if view === "list-grid"}
			<ListGridCard
				{game}
				{icon}
				{isRunning}
				{isSelectionMode}
				{isSelected}
				{onLaunch}
				{onConfigure}
				{onSelect}
			/>
		{:else}
			<PerspectiveCard
				{game}
				{icon}
				{isRunning}
				{active}
				{onLaunch}
				{onConfigure}
			/>
		{/if}
	{:else}
		<!-- Beautiful dark-mode skeleton screens matching the exact layouts -->
		{#if view === "list-grid"}
			<div class="skeleton-list-card">
				<div class="skeleton-icon-box">
					<span class="material-icons skeleton-logo">sports_esports</span>
				</div>
				<div class="skeleton-info">
					<div class="skeleton-line name"></div>
					<div class="skeleton-line path"></div>
				</div>
			</div>
		{:else if view === "perspective"}
			<div class="skeleton-perspective-card">
				<div class="skeleton-icon-large">
					<span class="material-icons skeleton-logo">sports_esports</span>
				</div>
			</div>
		{:else}
			<div class="skeleton-grid-card">
				<div class="skeleton-icon-large skeleton-shimmer">
					<span class="material-icons skeleton-logo">sports_esports</span>
				</div>
				<div class="skeleton-footer">
					<div class="skeleton-line name skeleton-shimmer"></div>
					<div class="skeleton-button skeleton-shimmer"></div>
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

	.skeleton-logo {
		font-size: 32px;
		color: var(--text-muted, #fff);
		animation: skeleton-pulse 2s infinite ease-in-out;
	}

	.skeleton-grid-card {
		display: flex;
		flex-direction: column;
		gap: 12px;
		width: 100%;
		max-width: 200px;
		margin: 6px;
		aspect-ratio: 1;
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

		.skeleton-footer {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 0 4px;
			height: 32px;
		}

		.skeleton-line.name {
			height: 14px;
			background: rgba(255, 255, 255, 0.04);
			border-radius: var(--radius-sm, 4px);
			width: 65%;
			animation: skeleton-pulse 2s infinite ease-in-out;
		}

		.skeleton-button {
			width: 32px;
			height: 32px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.03);
			border: 2px solid rgba(255, 255, 255, 0.05);
		}
	}

	.skeleton-list-card {
		display: flex;
		align-items: center;
		background: var(--bg-surface, rgba(255, 255, 255, 0.02));
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg, 12px);
		padding: 14px 24px;
		gap: 20px;
		width: 100%;
		height: 110px;
		box-sizing: border-box;

		.skeleton-icon-box {
			height: 80px;
			width: 80px;
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

	.skeleton-perspective-card {
		width: 220px;
		height: 310px;
		background: var(--bg-surface, rgba(255, 255, 255, 0.02));
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg, 12px);
		perspective: 1000px;
		flex-shrink: 0;
		transform: rotateY(-15deg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-sizing: border-box;

		.skeleton-icon-large {
			width: 100%;
			height: 100%;
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}

	// Apple-style shimmer — only grid skeleton (placed after base rules so it overrides)
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
