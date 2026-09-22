<script lang="ts">
	export let rect: { x: number; y: number; w: number; h: number } | null = null;
</script>

{#if rect && rect.w > 0 && rect.h > 0}
	<div
		class="marquee"
		style="left: {rect.x}px; top: {rect.y}px; width: {rect.w}px; height: {rect.h}px;"
	></div>
{/if}

<style lang="scss">
	.marquee {
		position: fixed;
		z-index: 9000;
		pointer-events: none;
		border: 1px solid var(--accent-primary);
		background: color-mix(in srgb, var(--accent-primary) 12%, transparent);
		border-radius: var(--radius-sm);
		animation: marquee-in 120ms ease-out;
	}

	@keyframes marquee-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.marquee {
			animation: none;
		}
	}
</style>
