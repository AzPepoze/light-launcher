<script lang="ts">
	import { sparkShape } from "@lib/features/sparkline";

	let {
		samples,
		height = 56,
		label
	}: {
		samples: number[];
		height?: number;
		/** Accessible name, such as "Speed over the last minute". */
		label: string;
	} = $props();

	const VIEW_WIDTH = 100;
	const shape = $derived(sparkShape(samples, VIEW_WIDTH, height));
</script>

<div class="sparkline" style:height="{height}px">
	<svg viewBox="0 0 {VIEW_WIDTH} {height}" preserveAspectRatio="none" role="img" aria-label={label}>
		<path class="area" d={shape.area} />
		<path class="line" d={shape.line} />
	</svg>
	{#if shape.last}
		<span
			class="dot"
			style:left="{shape.last.x}%"
			style:top="{(shape.last.y / height) * 100}%"
		></span>
	{/if}
</div>

<style lang="scss">
	.sparkline {
		position: relative;
		width: 100%;
	}

	svg {
		display: block;
		width: 100%;
		height: 100%;
		overflow: visible;
	}

	.area {
		fill: var(--accent-primary);
		fill-opacity: 0.12;
	}

	.line {
		fill: none;
		stroke: var(--accent-primary);
		stroke-width: 1.5;
		stroke-linejoin: round;
		vector-effect: non-scaling-stroke;
	}

	.dot {
		position: absolute;
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--accent-primary);
		box-shadow: 0 0 0 3px var(--accent-glow);
		transform: translate(-50%, -50%);
	}
</style>
