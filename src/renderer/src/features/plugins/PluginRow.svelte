<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import { pluginStatusKind } from "../../lib/features/pluginStatus";
	import type { PluginInfo, PluginUpdateStatus } from "../../lib/features/types";
	import PluginIdentity from "./PluginIdentity.svelte";
	import UnsignedBadge from "./UnsignedBadge.svelte";
	import { STATUS_COPY } from "./pluginCopy";
	import { UNAVAILABLE_CHECK_TEXT, checkFailureHint, isCheckFailure } from "./pluginNotices";

	let {
		plugin,
		update,
		onOpen,
	}: {
		plugin: PluginInfo;
		update: PluginUpdateStatus | null;
		onOpen: () => void;
	} = $props();

	const status = $derived(pluginStatusKind(plugin));
	const dimmed = $derived(status === "sleeping" || status === "disabled");
	const checkFailed = $derived(update !== null && isCheckFailure(update));
</script>

<button class="row" onclick={onOpen}>
	<PluginIdentity id={plugin.id} name={plugin.name} />
	<span class="text">
		<span class="line">
			<span class="name">{plugin.name}</span>
			<span class="version">v{plugin.version}</span>
		</span>
		<span class="status" class:dimmed>
			{#if status === "unsigned"}
				<UnsignedBadge />
			{:else if status === "error"}
				<Badge tone="danger" icon="error">{STATUS_COPY[status]}</Badge>
			{:else}
				{#if status === "sleeping"}
					<Icon name="bedtime" size={16} />
				{/if}
				{STATUS_COPY[status]}
			{/if}
		</span>
		{#if checkFailed && update?.error}
			<span class="unavailable" title={checkFailureHint(update.error)}>{UNAVAILABLE_CHECK_TEXT}</span>
		{/if}
	</span>
	{#if update?.available}
		<span class="update">Update {update.version}</span>
	{/if}
</button>

<style lang="scss">
	.row {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 12px 14px;
		border: 1px solid transparent;
		border-radius: var(--radius-md);
		background: var(--bg-surface);
		color: var(--text-main);
		text-align: left;
		cursor: pointer;
		font: inherit;

		&:hover {
			border-color: var(--glass-border-bright);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	.text {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 1;
		min-width: 0;
	}

	.line {
		display: flex;
		align-items: baseline;
		gap: 10px;
		min-width: 0;
	}

	.name {
		font-size: 1.05rem;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.version {
		font-size: 0.875rem;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.status {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-size: 0.875rem;
		color: var(--text-muted);

		&.dimmed {
			opacity: 0.8;
		}
	}

	.unavailable {
		font-size: 0.8125rem;
		color: var(--text-dim);
	}

	.update {
		flex-shrink: 0;
		padding: 3px 10px;
		border-radius: var(--radius-pill);
		border: 1px solid var(--accent-primary);
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--accent-primary);
		animation: badge-in var(--motion-base) var(--ease-snap) both;
	}

	@keyframes badge-in {
		from {
			opacity: 0;
			transform: scale(0.6);
		}
	}
</style>
