<script lang="ts">
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo } from "../../lib/features/types";
	import PluginAccounts from "./PluginAccounts.svelte";
	import PluginLogs from "./PluginLogs.svelte";
	import PluginSettings from "./PluginSettings.svelte";
	import { UNSIGNED_WARNING } from "./pluginCopy";

	let {
		plugin,
		onChanged,
		onRemoved,
		onBack,
	}: {
		plugin: PluginInfo;
		onChanged: (plugin: PluginInfo) => void;
		onRemoved: (id: string) => void;
		onBack: () => void;
	} = $props();

	let busy = $state(false);
	let message = $state("");
	let confirmingRemove = $state(false);

	async function run(action: () => Promise<void>) {
		busy = true;
		message = "";
		try {
			await action();
		} catch (err) {
			message = errorMessage(err);
		} finally {
			busy = false;
		}
	}

	function toggleEnabled() {
		run(async () => {
			onChanged(await pluginsApi.setEnabled(plugin.id, !plugin.enabled));
		});
	}

	function update() {
		run(async () => {
			onChanged(await pluginsApi.update(plugin.id, !plugin.signed));
		});
	}

	function rollback() {
		run(async () => {
			onChanged(await pluginsApi.rollback(plugin.id));
		});
	}

	function remove() {
		run(async () => {
			await pluginsApi.remove(plugin.id);
			onRemoved(plugin.id);
		});
	}
</script>

<div class="detail">
	<button class="btn sm back" onclick={onBack}>
		<span class="material-icons" aria-hidden="true">arrow_back</span> Plugins
	</button>

	<header>
		<div>
			<h2>{plugin.name}</h2>
			<p class="muted">
				v{plugin.version}
			</p>
		</div>
		<div class="badges">
			<span class="badge">{plugin.enabled ? "Enabled" : "Disabled"}</span>
			{#if !plugin.signed}
				<span class="badge warn">Unsigned</span>
			{/if}
			{#if plugin.status === "error"}
				<span class="badge warn">Error</span>
			{/if}
		</div>
	</header>

	{#if plugin.error}
		<p class="error" role="alert">{plugin.error}</p>
	{/if}
	{#if !plugin.signed}
		<p class="notice">{UNSIGNED_WARNING}</p>
	{/if}

	<div class="actions">
		<button class="btn sm" aria-pressed={plugin.enabled} disabled={busy} onclick={toggleEnabled}>
			{plugin.enabled ? "Disable" : "Enable"}
		</button>
		<button class="btn sm" disabled={busy} onclick={update}>Update</button>
		<button class="btn sm" disabled={busy} onclick={rollback}>Roll back</button>
		{#if confirmingRemove}
			<button class="btn sm danger" disabled={busy} onclick={remove}>Confirm remove</button>
			<button class="btn sm" onclick={() => (confirmingRemove = false)}>Cancel</button>
		{:else}
			<button class="btn sm" disabled={busy} onclick={() => (confirmingRemove = true)}>Remove</button>
		{/if}
	</div>

	{#if message}
		<p class="error" role="alert">{message}</p>
	{/if}

	<PluginSettings pluginId={plugin.id} />

	{#if plugin.sources.length}
		<PluginAccounts {plugin} />
	{/if}

	<PluginLogs pluginId={plugin.id} />
</div>

<style lang="scss">
	.detail {
		display: flex;
		flex-direction: column;
		gap: 22px;
		padding-bottom: 24px;
	}

	.back {
		align-self: flex-start;
	}

	header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 12px;
		flex-wrap: wrap;

		h2 {
			margin: 0;
			font-size: 1.4rem;
			font-weight: 800;
		}
	}

	.muted {
		margin: 4px 0 0;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.badges {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}

	.badge {
		font-size: 0.72rem;
		padding: 3px 9px;
		border-radius: var(--radius-pill);
		border: 1px solid var(--glass-border-bright);

		&.warn {
			border-color: var(--danger);
			color: var(--danger);
		}
	}

	.notice {
		margin: 0;
		padding: 10px 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--danger);
		font-size: 0.85rem;
	}

	.error {
		margin: 0;
		font-size: 0.85rem;
		color: var(--danger);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
</style>
