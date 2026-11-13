<script lang="ts">
	import { eta as formatEta, percentLabel, sizeRange, speed as formatSpeed } from "@lib/features/format";
	import Icon from "./Icon.svelte";

	let {
		speed,
		eta,
		done,
		total,
		percent
	}: {
		/** Bytes per second. */
		speed?: number;
		/** Seconds left. */
		eta?: number | null;
		done?: number;
		total?: number;
		percent?: number | null;
	} = $props();

	const size = $derived(sizeRange(done, total));
</script>

<div class="meta">
	{#if speed !== undefined}
		<span class="item">
			<Icon name="speed" size={16} />
			{formatSpeed(speed)}
		</span>
	{/if}
	{#if eta !== undefined && eta !== null}
		<span class="item">
			<Icon name="schedule" size={16} />
			{formatEta(eta)}
		</span>
	{/if}
	{#if size}
		<span class="item">{size}</span>
	{/if}
	{#if percent !== undefined && percent !== null}
		<span class="item">{percentLabel(percent)}</span>
	{/if}
</div>

<style lang="scss">
	.meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 16px;
		color: var(--text-muted);
		font-size: 0.85rem;
		font-variant-numeric: tabular-nums;
	}

	.item {
		display: inline-flex;
		align-items: center;
		gap: 6px;
	}
</style>
