<script lang="ts">
	import { onMount } from "svelte";
	import { FrameHistory, summarize, type FrameStats } from "@lib/frameStats";

	const HISTORY_SIZE = 2048;
	const WINDOW_MS = 5000;
	const REFRESH_MS = 500;
	// Longer gaps mean the window was hidden or minimized, not a slow frame.
	const MAX_SAMPLE_MS = 1000;

	const history = new FrameHistory(HISTORY_SIZE);
	let stats: FrameStats = $state(summarize([]));

	function recordInterval(intervalMs: number) {
		if (intervalMs <= MAX_SAMPLE_MS) history.push(intervalMs);
	}

	function refresh() {
		stats = summarize(history.recentWithin(WINDOW_MS));
	}

	function formatMs(value: number): string {
		return `${value.toFixed(1)} ms`;
	}

	function displayLabel(hz: number): string {
		return hz > 0 ? `display ≈ ${hz} Hz` : "display: measuring";
	}

	onMount(() => {
		let frameId = 0;
		let lastTimestamp: number | null = null;

		function onFrame(timestamp: number) {
			if (lastTimestamp !== null) recordInterval(timestamp - lastTimestamp);
			lastTimestamp = timestamp;
			frameId = requestAnimationFrame(onFrame);
		}

		frameId = requestAnimationFrame(onFrame);
		const refreshTimer = setInterval(refresh, REFRESH_MS);

		return () => {
			cancelAnimationFrame(frameId);
			clearInterval(refreshTimer);
		};
	});
</script>

<div class="fps-meter" aria-hidden="true">
	<div class="row"><span class="label">fps</span><strong>{stats.fps}</strong></div>
	<div class="row"><span class="label">p50</span><strong>{formatMs(stats.p50)}</strong></div>
	<div class="row"><span class="label">p95</span><strong>{formatMs(stats.p95)}</strong></div>
	<div class="row"><span class="label">worst, 5 s</span><strong>{formatMs(stats.worst)}</strong></div>
	<div class="row"><span class="label">long frames, 5 s</span><strong>{stats.longFrames}</strong></div>
	<div class="display">{displayLabel(stats.displayHz)}</div>
	<p class="hint">fps near the display rate and no long frames = smooth</p>
</div>

<style lang="scss">
	.fps-meter {
		position: fixed;
		top: 12px;
		right: 12px;
		z-index: 9999;
		pointer-events: none;
		min-width: 220px;
		padding: 10px 14px;
		border-radius: var(--radius-md);
		background: var(--bg-elevated);
		color: var(--text-main);
		font-size: 0.875rem;
		font-variant-numeric: tabular-nums;
		box-shadow: 0 6px 24px rgba(0, 0, 0, 0.22);
	}

	.row {
		display: flex;
		justify-content: space-between;
		gap: 16px;
	}

	.label {
		color: var(--text-muted);
	}

	.display {
		margin-top: 6px;
		font-weight: 650;
	}

	.hint {
		margin: 8px 0 0;
		max-width: 240px;
		color: var(--text-muted);
	}
</style>
