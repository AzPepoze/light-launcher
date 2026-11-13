<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import ProgressBar from "@components/shared/ProgressBar.svelte";
	import TransferMeta from "@components/shared/TransferMeta.svelte";
	import TransferStages from "@components/shared/TransferStages.svelte";
	import type { DownloadAction } from "../../lib/features/types";
	import type { TransferRow } from "./downloadGroups";
	import {
		fileIcon,
		progressTone,
		progressValue,
		rowActions,
		rowBadge,
		rowMeta
	} from "./rowView";

	let { row, onAct }: { row: TransferRow; onAct: (action: DownloadAction) => void } = $props();

	const transfer = $derived(row.transfer);
	const actions = $derived(rowActions(row));
	const badge = $derived(rowBadge(transfer));
	const meta = $derived(rowMeta(transfer));
</script>

<li class="row {transfer.state}">
	<span class="kind">
		<Icon name={fileIcon(row.item.filename)} size={24} />
	</span>

	<div class="main">
		<div class="heading">
			<p class="name" title={row.item.filename}>{row.item.filename}</p>
			{#if row.gameTitle}<p class="game">{row.gameTitle}</p>{/if}
			{#if badge}<Badge tone={badge.tone}>{badge.label}</Badge>{/if}
			{#if transfer.state === "done"}
				<span class="done-check">
					<Icon name="check_circle" size={20} filled />
				</span>
			{/if}
		</div>

		<TransferStages stage={transfer.stage} failed={transfer.state === "failed"} />

		<ProgressBar
			value={progressValue(transfer)}
			tone={progressTone(transfer.state)}
			active={transfer.state === "active"}
			label="Progress for {row.item.filename}"
		/>

		<TransferMeta
			speed={meta.speed}
			eta={meta.eta}
			done={meta.done}
			total={meta.total}
		/>

		{#if transfer.error}<p class="error">{transfer.error}</p>{/if}
	</div>

	<div class="actions">
		{#each actions as item (item.action)}
			<IconButton
				icon={item.icon}
				label={item.label}
				tone={item.action === "cancel" ? "danger" : "default"}
				onclick={() => onAct(item.action)}
			/>
		{/each}
	</div>
</li>

<style lang="scss">
	.row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 16px;
		border-radius: var(--radius-md);
		background: var(--bg-surface);
		list-style: none;
		animation: row-in var(--motion-base) var(--ease-out) both;
	}

	.kind {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		flex-shrink: 0;
		border-radius: var(--radius-md);
		background: var(--bg-elevated);
		color: var(--text-muted);
	}

	.main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.heading {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 4px 10px;
		min-width: 0;
	}

	.name {
		margin: 0;
		min-width: 0;
		font-weight: 700;
		font-size: 0.95rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.game {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.done-check {
		display: inline-flex;
		color: var(--success);
		animation: check-in 320ms var(--ease-out) both;
	}

	.error {
		margin: 0;
		color: var(--danger);
		font-size: 0.85rem;
	}

	.actions {
		display: flex;
		gap: 4px;
		flex-shrink: 0;
	}

	@keyframes row-in {
		from {
			opacity: 0;
		}
	}

	@keyframes check-in {
		from {
			opacity: 0;
			transform: scale(0.4);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.row,
		.done-check {
			animation: none;
		}
	}
</style>
