<script lang="ts">
	import { eta, speed } from "../../lib/features/format";
	import type { DownloadAction, DownloadItem } from "../../lib/features/types";
	import { progressPercent, statusLabel } from "./downloadGroups";

	let {
		item,
		onAct,
	}: { item: DownloadItem; onAct: (action: DownloadAction, id: number) => void } = $props();

	const isRunning = $derived(item.status === "downloading");
	const canResume = $derived(item.status === "paused" || item.status === "queued");
	const isFailed = $derived(item.status === "error" || item.status === "canceled");
	const isFinished = $derived(item.status === "done");
</script>

<li class="row">
	<div class="main">
		<div class="title">
			<p class="name" title={item.filename}>{item.filename}</p>
			<span class="status">{statusLabel(item.status)}</span>
		</div>

		{#if !isFinished}
			<div
				class="bar"
				role="progressbar"
				aria-label="Progress for {item.filename}"
				aria-valuemin="0"
				aria-valuemax="100"
				aria-valuenow={progressPercent(item)}
			>
				<div class="fill" style="width: {progressPercent(item)}%"></div>
			</div>
		{/if}

		<p class="meta">
			{#if isRunning}
				<span>{speed(item.speed)}</span>
				<span>ETA {eta(item.eta)}</span>
			{/if}
			{#if item.error}<span class="warn">{item.error}</span>{/if}
		</p>
	</div>

	<div class="actions">
		{#if isRunning}
			<button class="btn sm" onclick={() => onAct("pause", item.id)}>Pause</button>
		{/if}
		{#if canResume}
			<button class="btn sm" onclick={() => onAct("resume", item.id)}>Resume</button>
		{/if}
		{#if isFailed}
			<button class="btn sm" onclick={() => onAct("retry", item.id)}>Retry</button>
		{/if}
		{#if !isFinished && !isFailed}
			<button class="btn sm" onclick={() => onAct("cancel", item.id)}>Cancel</button>
		{:else}
			<button class="btn sm" aria-label="Remove {item.filename}" onclick={() => onAct("remove", item.id)}>
				Remove
			</button>
		{/if}
	</div>
</li>

<style lang="scss">
	.row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		background: var(--bg-surface);
		list-style: none;
	}

	.main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.title {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.name {
		margin: 0;
		font-weight: 700;
		font-size: 0.9rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		min-width: 0;
	}

	.status {
		flex-shrink: 0;
		font-size: 0.72rem;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.bar {
		height: 6px;
		border-radius: var(--radius-pill);
		background: var(--bg-elevated);
		overflow: hidden;
	}

	.fill {
		height: 100%;
		background: var(--accent-primary);
		transition: width var(--transition-normal);
	}

	.meta {
		margin: 0;
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
		font-size: 0.75rem;
		color: var(--text-muted);

		.warn {
			color: var(--danger);
		}
	}

	.actions {
		display: flex;
		gap: 6px;
		flex-shrink: 0;
		flex-wrap: wrap;
		justify-content: flex-end;
	}
</style>
