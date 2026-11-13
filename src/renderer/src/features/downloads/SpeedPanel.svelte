<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import Sparkline from "@components/shared/Sparkline.svelte";
	import { speed as formatSpeed } from "../../lib/features/format";
	import { speedLimitLabel } from "./speedView";

	let {
		speed,
		samples,
		active,
		queued,
		limitKBps
	}: {
		/** Bytes per second, all downloads together. */
		speed: number;
		samples: number[];
		active: number;
		queued: number;
		limitKBps: number;
	} = $props();

	const limit = $derived(speedLimitLabel(limitKBps));
</script>

<section class="speed-panel" aria-label="Total speed">
	<div class="figures">
		<p class="value">{formatSpeed(speed)}</p>
		<p class="caption">Total speed</p>
		<div class="counts">
			<span class="count">
				<Icon name="downloading" size={16} />
				{active} active
			</span>
			<span class="count">
				<Icon name="queue" size={16} />
				{queued} queued
			</span>
			{#if limit}<Badge icon="speed">{limit}</Badge>{/if}
		</div>
	</div>

	<div class="chart">
		<Sparkline {samples} label="Total speed over the last minute" />
	</div>
</section>

<style lang="scss">
	.speed-panel {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: 20px 32px;
		margin-bottom: 20px;
		padding: 20px 24px;
		border-radius: var(--radius-lg);
		background: var(--bg-surface);
	}

	.figures {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 180px;
	}

	.value {
		margin: 0;
		color: var(--text-main);
		font-size: 2.25rem;
		font-weight: 800;
		line-height: 1.1;
		font-variant-numeric: tabular-nums;
	}

	.caption {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.counts {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
		margin-top: 8px;
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.count {
		display: inline-flex;
		align-items: center;
		gap: 6px;
	}

	.chart {
		flex: 1 1 260px;
		min-width: 0;
	}
</style>
