<svg class="check" viewBox="0 0 24 24" aria-hidden="true">
	<path pathLength="1" d="M5 12.5 9.5 17 19 7.5" />
</svg>

<style lang="scss">
	.check {
		width: 22px;
		height: 22px;
		flex-shrink: 0;
		fill: none;
		stroke: var(--success);
		stroke-width: 2.5;
		stroke-linecap: round;
		stroke-linejoin: round;

		path {
			stroke-dasharray: 1;
			stroke-dashoffset: 1;
			animation: draw-check var(--motion-base) var(--ease-out) forwards;
		}
	}

	@keyframes draw-check {
		to {
			stroke-dashoffset: 0;
		}
	}
</style>
