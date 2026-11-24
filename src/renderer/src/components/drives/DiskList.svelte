<script lang="ts">
	import type { Partition, SetupEntry } from "@shared";
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import ProgressBar from "@components/shared/ProgressBar.svelte";
	import { formatBytes } from "@lib/features/format";
	import { letterForPath, letterLabel, usedBytes, usedRatio } from "@lib/driveSetup";

	let {
		partitions,
		entries,
		nextLetter,
		onAdd
	}: {
		partitions: Partition[];
		entries: SetupEntry[];
		/** Empty when every drive letter is taken. */
		nextLetter: string;
		onAdd: (partition: Partition) => void;
	} = $props();
</script>

{#if partitions.length === 0}
	<EmptyState icon="storage" title="No disks found" />
{:else}
	<ul class="disk-list">
		{#each partitions as partition (partition.device)}
			{@const setupLetter = letterForPath(entries, partition.mount)}
			<li class="disk-row">
				<div class="disk-info">
					<div class="disk-title">
						<span class="disk-name">{partition.name || partition.device}</span>
						<Badge>{partition.fsType || "Unknown"}</Badge>
					</div>
					<p class="mount">{partition.mount || "Not mounted"}</p>
					<ProgressBar value={usedRatio(partition)} label="Used space" />
					<p class="size">
						{formatBytes(usedBytes(partition))} of {formatBytes(partition.totalBytes)} used
					</p>
				</div>
				{#if setupLetter}
					<Badge tone="success">In setup as {letterLabel(setupLetter)}</Badge>
				{:else}
					<Button
						size="sm"
						disabled={!partition.mount || !nextLetter}
						onclick={() => onAdd(partition)}
					>
						Add as drive
					</Button>
				{/if}
			</li>
		{/each}
	</ul>
{/if}

<style lang="scss">
	.disk-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.disk-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 16px 18px;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
	}

	.disk-info {
		display: flex;
		flex-direction: column;
		gap: 6px;
		flex: 1;
		min-width: 0;
	}

	.disk-title {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.disk-name {
		color: var(--text-main);
		font-weight: 800;
	}

	.mount,
	.size {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.85rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
