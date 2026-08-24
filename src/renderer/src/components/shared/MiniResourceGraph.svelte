<script lang="ts">
	export let data: number[] = [];
	export let color: string = "var(--accent-primary, #64ffda)";
	export let height: number = 44;
	export let maxPoints: number = 30;

	const id = `graph-grad-${Math.random().toString(36).substring(2, 9)}`;

	const SVG_WIDTH = 200;
	const SVG_HEIGHT = 60;

	$: points = (() => {
		const arr = data.length > 0 ? data : [0];
		// Pad with initial values or zeros if less than maxPoints to ensure stable command count
		const fullData: number[] = [];
		const padCount = Math.max(0, maxPoints - arr.length);
		for (let i = 0; i < padCount; i++) {
			fullData.push(arr[0] ?? 0);
		}
		fullData.push(...arr);
		const slice = fullData.slice(-maxPoints);

		return slice.map((val, index) => {
			const x = (index / (maxPoints - 1)) * SVG_WIDTH;
			// Invert Y: 100% is top (y = 4), 0% is bottom (y = SVG_HEIGHT - 4)
			const clampVal = Math.min(100, Math.max(0, val));
			const y = SVG_HEIGHT - 4 - (clampVal / 100) * (SVG_HEIGHT - 8);
			return { x, y, val: clampVal };
		});
	})();

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

	$: linePath = getSmoothSplinePath(points);

	$: areaPath = (() => {
		if (!linePath || points.length === 0) return "";
		const startX = points[0].x.toFixed(1);
		const endX = points[points.length - 1].x.toFixed(1);
		return `${linePath} L ${endX},${SVG_HEIGHT} L ${startX},${SVG_HEIGHT} Z`;
	})();

	$: lastPoint = points.length > 0 ? points[points.length - 1] : { x: SVG_WIDTH, y: SVG_HEIGHT / 2 };
</script>

<div class="mini-graph" style="--graph-color: {color}; height: {height}px;">
	<svg
		viewBox="0 0 {SVG_WIDTH} {SVG_HEIGHT}"
		preserveAspectRatio="none"
		class="graph-svg"
		aria-hidden="true"
	>
		<defs>
			<linearGradient {id} x1="0" y1="0" x2="0" y2="1">
				<stop offset="0%" stop-color="var(--graph-color)" stop-opacity="0.32" />
				<stop offset="60%" stop-color="var(--graph-color)" stop-opacity="0.10" />
				<stop offset="100%" stop-color="var(--graph-color)" stop-opacity="0.0" />
			</linearGradient>
		</defs>

		<!-- Task Manager Style Background Grid -->
		<g class="grid-lines">
			<!-- Horizontal 25%, 50%, 75% gridlines -->
			<line x1="0" y1={SVG_HEIGHT * 0.25} x2={SVG_WIDTH} y2={SVG_HEIGHT * 0.25} />
			<line x1="0" y1={SVG_HEIGHT * 0.50} x2={SVG_WIDTH} y2={SVG_HEIGHT * 0.50} />
			<line x1="0" y1={SVG_HEIGHT * 0.75} x2={SVG_WIDTH} y2={SVG_HEIGHT * 0.75} />

			<!-- Vertical time columns -->
			<line x1={SVG_WIDTH * 0.2} y1="0" x2={SVG_WIDTH * 0.2} y2={SVG_HEIGHT} />
			<line x1={SVG_WIDTH * 0.4} y1="0" x2={SVG_WIDTH * 0.4} y2={SVG_HEIGHT} />
			<line x1={SVG_WIDTH * 0.6} y1="0" x2={SVG_WIDTH * 0.6} y2={SVG_HEIGHT} />
			<line x1={SVG_WIDTH * 0.8} y1="0" x2={SVG_WIDTH * 0.8} y2={SVG_HEIGHT} />
		</g>

		<!-- Gradient Fill -->
		{#if areaPath}
			<path d={areaPath} fill="url(#{id})" class="area-fill" />
		{/if}

		<!-- Active Line Stroke -->
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

		<!-- Current Value Endpoint Dot -->
		<circle
			cx={lastPoint.x}
			cy={lastPoint.y}
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
		background: rgba(0, 0, 0, 0.25);
		border-radius: var(--radius-sm, 6px);
		border: 1px solid rgba(255, 255, 255, 0.05);
		overflow: hidden;
		display: flex;
	}

	.graph-svg {
		width: 100%;
		height: 100%;
		display: block;
	}

	.grid-lines {
		line {
			stroke: rgba(255, 255, 255, 0.055);
			stroke-width: 1;
			stroke-dasharray: 2 2;
			vector-effect: non-scaling-stroke;
		}
	}

	.area-fill {
		transition: d 1000ms cubic-bezier(0.15, 0.85, 0.35, 1);
	}

	.stroke-line {
		filter: drop-shadow(0 0 3px var(--graph-color));
		transition: d 1000ms cubic-bezier(0.15, 0.85, 0.35, 1);
	}

	.endpoint-dot {
		filter: drop-shadow(0 0 4px var(--graph-color));
		transition: cx 1000ms cubic-bezier(0.15, 0.85, 0.35, 1), cy 1000ms cubic-bezier(0.15, 0.85, 0.35, 1);
	}
</style>
