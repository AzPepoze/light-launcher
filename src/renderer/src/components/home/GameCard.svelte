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
				<div class="skeleton-icon-large">
					<span class="material-icons skeleton-logo">sports_esports</span>
				</div>
				<div class="skeleton-footer">
					<div class="skeleton-line name"></div>
					<div class="skeleton-button"></div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style lang="scss">
	.lazy-card-container {
		width: 100%;
	}

	// Shimmering pulse keyframe
	@keyframes skeleton-pulse {
		0%, 100% {
			opacity: 0.15;
		}
		50% {
			opacity: 0.35;
		}
	}

	.skeleton-logo {
		font-size: 32px;
		color: var(--text-muted, #fff);
		animation: skeleton-pulse 2s infinite ease-in-out;
	}

	// SKELETON DEFINITIONS

	// 1. Grid View Card Skeleton
	.skeleton-grid-card {
		display: flex;
		flex-direction: column;
		gap: 12px;
		width: 100%;
		max-width: 200px;
		margin: 6px;
		aspect-ratio: 1; // matches .game-icon-container aspect-ratio: 1
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

	// 2. List View Card Skeleton
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

	// 3. Perspective View Card Skeleton
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
</style>
