<script lang="ts">
	interface Props {
		data?: number[];
		color?: string;
		height?: number;
		maxPoints?: number;
	}

	let {
		data = [],
		color = "var(--accent-primary, #64ffda)",
		height = 46,
		maxPoints = 30
	}: Props = $props();

	const gradId = `graph-grad-${Math.random().toString(36).substring(2, 9)}`;
	const patId = `grid-pat-${Math.random().toString(36).substring(2, 9)}`;

	const SVG_WIDTH = 200;
	const SVG_HEIGHT = 60;
	const STEP_X = $derived(SVG_WIDTH / Math.max(1, maxPoints - 1));

	const points = $derived.by(() => {
		const arr = data.length > 0 ? data : [0];
		const initialVal = arr[0] ?? 0;

		let fullData = arr;
		if (arr.length < maxPoints) {
			const pad = Array(maxPoints - arr.length).fill(initialVal);
			fullData = [...pad, ...arr];
		}
		const slice = fullData.slice(-maxPoints);
		const extended = [...slice, slice[slice.length - 1] ?? initialVal];

		return extended.map((val, index) => {
			const x = index * STEP_X;
			const clampVal = Math.min(100, Math.max(0, val));
			const y = SVG_HEIGHT - 4 - (clampVal / 100) * (SVG_HEIGHT - 8);
			return { x, y };
		});
	});

	function getSmoothSplinePath(pts: { x: number; y: number }[]): string {
		if (pts.length === 0) return "";
		if (pts.length === 1) return `M ${pts[0].x.toFixed(1)},${pts[0].y.toFixed(1)}`;
		if (pts.length === 2) {
			return `M ${pts[0].x.toFixed(1)},${pts[0].y.toFixed(1)} L ${pts[1].x.toFixed(1)},${pts[1].y.toFixed(1)}`;
		}

		let path = `M ${pts[0].x.toFixed(1)},${pts[0].y.toFixed(1)}`;
		const tension = 0.22;

		for (let i = 0; i < pts.length - 1; i++) {
			const p0 = pts[Math.max(0, i - 1)];
			const p1 = pts[i];
			const p2 = pts[i + 1];
			const p3 = pts[Math.min(pts.length - 1, i + 2)];

			const cp1x = p1.x + (p2.x - p0.x) * tension;
			const cp1y = p1.y + (p2.y - p0.y) * tension;
			const cp2x = p2.x - (p3.x - p1.x) * tension;
			const cp2y = p2.y - (p3.y - p1.y) * tension;

			path += ` C ${cp1x.toFixed(1)},${cp1y.toFixed(1)} ${cp2x.toFixed(1)},${cp2y.toFixed(1)} ${p2.x.toFixed(1)},${p2.y.toFixed(1)}`;
		}
		return path;
	}

	const linePath = $derived(getSmoothSplinePath(points));

	const areaPath = $derived.by(() => {
		if (!linePath || points.length === 0) return "";
		const startX = points[0].x.toFixed(1);
		const endX = points[points.length - 1].x.toFixed(1);
		return `${linePath} L ${endX},${SVG_HEIGHT} L ${startX},${SVG_HEIGHT} Z`;
	});

	const currentLiveY = $derived.by(() => {
		const latestVal = data.length > 0 ? data[data.length - 1] : 0;
		const clampVal = Math.min(100, Math.max(0, latestVal));
		return SVG_HEIGHT - 4 - (clampVal / 100) * (SVG_HEIGHT - 8);
	});
</script>

<div
	class="mini-graph"
	style="--graph-color: {color}; --step-x: {STEP_X}px; height: {height}px;"
>
	<svg
		viewBox="0 0 {SVG_WIDTH} {SVG_HEIGHT}"
		preserveAspectRatio="none"
		class="graph-svg"
		aria-hidden="true"
	>
		<defs>
			<linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
				<stop offset="0%" stop-color="var(--graph-color)" stop-opacity="0.36" />
				<stop offset="60%" stop-color="var(--graph-color)" stop-opacity="0.12" />
				<stop offset="100%" stop-color="var(--graph-color)" stop-opacity="0.0" />
			</linearGradient>

			<pattern
				id={patId}
				width={STEP_X}
				height={SVG_HEIGHT}
				patternUnits="userSpaceOnUse"
			>
				<line x1="0" y1={SVG_HEIGHT * 0.25} x2={STEP_X} y2={SVG_HEIGHT * 0.25} class="grid-line" />
				<line x1="0" y1={SVG_HEIGHT * 0.50} x2={STEP_X} y2={SVG_HEIGHT * 0.50} class="grid-line" />
				<line x1="0" y1={SVG_HEIGHT * 0.75} x2={STEP_X} y2={SVG_HEIGHT * 0.75} class="grid-line" />
				<line x1={STEP_X} y1="0" x2={STEP_X} y2={SVG_HEIGHT} class="grid-line" />
			</pattern>
		</defs>

		<g class="grid-scroll">
			<rect
				x={-STEP_X}
				y="0"
				width={SVG_WIDTH + STEP_X * 2}
				height={SVG_HEIGHT}
				fill="url(#{patId})"
			/>
		</g>

		{#key data}
			<g class="wave-scroll">
				{#if areaPath}
					<path d={areaPath} fill="url(#{gradId})" />
				{/if}

				{#if linePath}
					<path
						d={linePath}
						fill="none"
						stroke="var(--graph-color)"
						stroke-width="1.8"
						stroke-linecap="round"
						stroke-linejoin="round"
						vector-effect="non-scaling-stroke"
						class="stroke-line"
					/>
				{/if}
			</g>
		{/key}

		<circle
			cx={SVG_WIDTH - 2}
			cy={currentLiveY}
			r="2.5"
			fill="var(--graph-color)"
			class="endpoint-dot"
		/>
	</svg>
</div>

<style lang="scss">
	.mini-graph {
		width: 100%;
		position: relative;
		background: rgba(0, 0, 0, 0.3);
		border-radius: var(--radius-sm, 6px);
		border: 1px solid rgba(255, 255, 255, 0.06);
		overflow: hidden;
		display: flex;
	}

	.graph-svg {
		width: 100%;
		height: 100%;
		display: block;
	}

	.grid-line {
		stroke: rgba(255, 255, 255, 0.06);
		stroke-width: 0.8;
		stroke-dasharray: 2 3;
		vector-effect: non-scaling-stroke;
	}

	.grid-scroll {
		animation: slideOneStep 1000ms linear infinite;
		will-change: transform;
	}

	.wave-scroll {
		animation: slideOneStep 1000ms linear forwards;
		will-change: transform;
	}

	@keyframes slideOneStep {
		from {
			transform: translateX(0);
		}
		to {
			transform: translateX(calc(-1 * var(--step-x)));
		}
	}

	.stroke-line {
		filter: drop-shadow(0 0 4px var(--graph-color));
	}

	.endpoint-dot {
		filter: drop-shadow(0 0 5px var(--graph-color));
		animation: dotPulse 1.6s ease-in-out infinite;
		transform-box: fill-box;
		transform-origin: center;
		transition: cy 260ms ease-out;
	}

	@keyframes dotPulse {
		0%,
		100% {
			opacity: 1;
			transform: scale(1);
		}
		50% {
			opacity: 0.82;
			transform: scale(1.35);
		}
	}
</style>
