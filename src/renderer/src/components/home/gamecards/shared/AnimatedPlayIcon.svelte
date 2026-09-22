<script lang="ts">
	export let size: number = 64;
</script>

<svg
	viewBox="0 0 64 64"
	width={size}
	height={size}
	class="animated-play"
	aria-hidden="true"
>
	<!-- single Google-loader style dash tracing the play shape — no fill, only edge -->
	<path
		class="play-edge"
		d="M 20 14 L 52 32 L 20 50 Z"
		fill="none"
		stroke-width="3.6"
		stroke-linejoin="round"
		stroke-linecap="round"
		pathLength="100"
	/>
</svg>

<style lang="scss">
	.animated-play {
		display: block;
		overflow: visible;
		filter: drop-shadow(0 0 12px rgba(var(--spotlight-rgb, 255, 255, 255), 0.55));
	}

	.play-edge {
		opacity: 1;
		stroke: rgb(var(--spotlight-rgb, 255, 255, 255));
		stroke-dasharray: 82 18;
		stroke-dashoffset: 0;
		animation: play-edge-dash 1.4s cubic-bezier(0.65, 0, 0.35, 1) infinite;
		animation-play-state: paused;
	}

	@keyframes play-edge-dash {
		to {
			stroke-dashoffset: -100;
		}
	}

	:global(.game-card:hover) .play-edge,
	:global(.game-card:focus-within) .play-edge {
		animation-play-state: running;
	}

	@media (prefers-reduced-motion: reduce) {
		.play-edge {
			animation: none;
			stroke-dasharray: none;
			opacity: 0.6;
		}
	}
</style>
