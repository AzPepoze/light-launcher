<script lang="ts">
	export let view: "grid" | "list-grid" | "sidebar-grid" = "grid";

	const placeholders = Array.from({ length: 12 }, (_, index) => index);
</script>

<div
	class="home-skeleton"
	class:list-view={view === "list-grid"}
	aria-hidden="true"
	data-testid="home-skeleton"
>
	{#each placeholders as index (index)}
		{#if view === "list-grid"}
			<div class="skeleton-list-card">
				<div class="skeleton-cover"></div>
				<div class="skeleton-lines">
					<div class="skeleton-line name"></div>
					<div class="skeleton-line path"></div>
				</div>
			</div>
		{:else}
			<div class="skeleton-grid-card">
				<div class="skeleton-cover"></div>
			</div>
		{/if}
	{/each}
</div>

<style lang="scss">
	// Mirrors GameCardGrid: the same auto-fill grid and gutters as the real cards.
	.home-skeleton {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 4px;
		width: 100%;
		max-width: 100%;
		padding: 18px;
		padding-bottom: 40px;
		box-sizing: border-box;
		min-width: 0;

		&.list-view {
			grid-template-columns: 1fr;
		}
	}

	// Same footprint as GridCard: 8px padding inside a 190px card, square cover.
	.skeleton-grid-card {
		width: 100%;
		max-width: 190px;
		padding: 8px;
		box-sizing: border-box;
		min-width: 0;
	}

	// Same footprint as ListGridCard: 108px square icon and two text lines.
	.skeleton-list-card {
		display: flex;
		align-items: center;
		gap: 12px;
		width: 100%;
		max-width: 100%;
		min-width: 0;
		padding: 8px 16px;
		box-sizing: border-box;
		background: var(--bg-surface);
		border: 2px solid var(--bg-elevated);
		border-radius: var(--radius-lg);

		.skeleton-cover {
			width: 108px;
			height: 108px;
			flex-shrink: 0;
			border: none;
			background: var(--bg-base);
		}
	}

	.skeleton-cover {
		position: relative;
		width: 100%;
		aspect-ratio: 1;
		overflow: hidden;
		box-sizing: border-box;
		background: var(--bg-surface);
		border: 2px solid var(--bg-elevated);
		border-radius: var(--radius-lg);
	}

	.skeleton-lines {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 8px;
		min-width: 0;
	}

	.skeleton-line {
		position: relative;
		overflow: hidden;
		border-radius: var(--radius-sm);
		background: var(--bg-elevated);

		&.name {
			width: 30%;
			height: 16px;
		}

		&.path {
			width: 50%;
			height: 10px;
		}
	}

	.skeleton-cover::after,
	.skeleton-line::after {
		content: "";
		position: absolute;
		inset: 0;
		background: linear-gradient(
			100deg,
			transparent 30%,
			var(--accent-glow) 50%,
			transparent 70%
		);
		background-size: 250% 100%;
		background-repeat: no-repeat;
		animation: home-skeleton-shimmer 1.8s ease-in-out infinite;
	}

	@keyframes home-skeleton-shimmer {
		from {
			background-position: 150% 0;
		}
		to {
			background-position: -150% 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.skeleton-cover::after,
		.skeleton-line::after {
			animation: none;
			background: none;
		}
	}
</style>
