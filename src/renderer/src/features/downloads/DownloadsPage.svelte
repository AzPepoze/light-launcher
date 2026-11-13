<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import DownloadControls from "./DownloadControls.svelte";
	import DownloadRow from "./DownloadRow.svelte";
	import GroupHeader from "./GroupHeader.svelte";
	import SpeedPanel from "./SpeedPanel.svelte";
	import type { DownloadsState } from "./downloadsState.svelte";
	import { buildRows, groupRows, listEntries, toastTone, totalSpeed } from "./downloadGroups";

	/** Owned by the Library page, which keeps it subscribed and sampling while another tab is open. */
	let { downloads }: { downloads: DownloadsState } = $props();

	let finishedOpen = $state(false);
	let controls: { focusInput: () => void } | undefined = $state();

	const rows = $derived(buildRows(downloads.items, downloads.installs));
	const groups = $derived(groupRows(rows));
	const entries = $derived(listEntries(groups, finishedOpen));
	const speedNow = $derived(totalSpeed(downloads.items));

</script>

{#if downloads.toast}
	<Notice tone={toastTone(downloads.toast.level)}>{downloads.toast.message}</Notice>
{/if}

{#if downloads.error}
	<Notice tone="error">
		{downloads.error}
		<Button variant="text" size="sm" icon="refresh" onclick={() => downloads.load()}>Retry</Button>
	</Notice>
{/if}

<DownloadControls {downloads} bind:this={controls} />

{#if downloads.loading}
	<LoadingState label="Loading downloads" />
{:else}
	<SpeedPanel
		speed={speedNow}
		samples={downloads.speedSamples}
		active={groups.active.length}
		queued={groups.queued.length}
		limitKBps={downloads.speedLimitKBps}
	/>

	{#if rows.length === 0}
		<EmptyState icon="download" title="Nothing downloading" hint="Paste a direct link above to start.">
			<Button variant="secondary" size="sm" icon="add_link" onclick={() => controls?.focusInput()}>
				Add a link
			</Button>
		</EmptyState>
	{:else}
		<ul class="list">
			{#each entries as entry (entry.key)}
				{#if entry.kind === "header"}
					<GroupHeader
						group={entry.group}
						count={entry.count}
						open={finishedOpen}
						onToggle={entry.group === "finished" ? () => (finishedOpen = !finishedOpen) : undefined}
						onClear={entry.group === "finished" ? () => downloads.clearFinished() : undefined}
					/>
				{:else}
					<DownloadRow row={entry.row} onAct={(action) => downloads.act(action, entry.row.item.id)} />
				{/if}
			{/each}
		</ul>
	{/if}
{/if}

<style lang="scss">
	.list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0;
		padding: 0;
		list-style: none;
	}
</style>
