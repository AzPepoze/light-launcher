<script lang="ts">
	import { errorMessage } from "../../lib/features/rpc";
	import { ago } from "../../lib/features/format";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo, PluginUpdatePreview, PluginUpdateStatus } from "../../lib/features/types";
	import Notice from "./Notice.svelte";
	import PluginAccounts from "./PluginAccounts.svelte";
	import PluginConsent from "./PluginConsent.svelte";
	import PluginIdentity from "./PluginIdentity.svelte";
	import PluginLogs from "./PluginLogs.svelte";
	import PluginSettings from "./PluginSettings.svelte";
	import {
		NO_NEW_PERMISSIONS,
		SLEEPING_BADGE,
		UNSIGNED_WARNING,
		UPDATE_PERMISSION_HINT
	} from "./pluginCopy";

	let {
		plugin,
		update,
		onChanged,
		onRemoved,
		onBack,
	}: {
		plugin: PluginInfo;
		update: PluginUpdateStatus | null;
		onChanged: (plugin: PluginInfo) => void;
		onRemoved: (id: string) => void;
		onBack: () => void;
	} = $props();

	let busy = $state(false);
	let message = $state("");
	let confirmingRemove = $state(false);
	let pendingUpdate = $state<PluginUpdatePreview | null>(null);

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

	/** Shows the consent screen for the release the last check found. */
	function previewUpdate(release: PluginUpdateStatus) {
		run(async () => {
			pendingUpdate = null;
			if (!release.assetUrl) {
				message = "This release has no plugin package.";
				return;
			}
			pendingUpdate = await pluginsApi.previewUpdate(plugin, release.assetUrl, release.sha256 ?? "");
		});
	}

	function confirmUpdate(shown: PluginUpdatePreview) {
		run(async () => {
			const next = shown.preview;
			onChanged(
				await pluginsApi.update(
					plugin.id,
					!next.signed,
					next.consent,
					shown.source,
					shown.sha256 ?? ""
				)
			);
			pendingUpdate = null;
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
		<div class="title">
			<PluginIdentity id={plugin.id} name={plugin.name} large />
			<div>
				<h2>{plugin.name}</h2>
				<p class="muted">
					v{plugin.version}
					{#if plugin.repository}
						<span>Updates from {plugin.repository}</span>
					{/if}
				</p>
			</div>
		</div>
		<div class="badges">
			<span class="badge">{plugin.enabled ? "Enabled" : "Disabled"}</span>
			{#if plugin.idle}<span class="badge">{SLEEPING_BADGE}</span>{/if}
			{#if !plugin.signed}
				<span class="badge warn">Unsigned</span>
			{/if}
			{#if plugin.status === "error"}
				<span class="badge warn">Error</span>
			{/if}
			{#if update?.available}
				<span class="badge update">Update available ({update.version})</span>
			{/if}
		</div>
	</header>

	{#if plugin.error}
		<p class="error" role="alert">{plugin.error}</p>
	{/if}
	{#if !plugin.signed}
		<Notice icon="warning">
			<p>{UNSIGNED_WARNING}</p>
		</Notice>
	{/if}

	{#if update?.available}
		<section class="release" aria-label="Update available">
			<div class="release-head">
				<p class="release-title">Update available: v{update.version}</p>
				{#if update.publishedAt}
					<span class="muted">{ago(update.publishedAt)}</span>
				{/if}
			</div>
			{#if update.notes}
				<p class="notes">{update.notes}</p>
			{/if}
			<button class="btn sm primary" disabled={busy} onclick={() => previewUpdate(update)}>
				Update to v{update.version}
			</button>
		</section>
	{:else if update?.error}
		<p class="error" role="alert">Update check failed: {update.error}</p>
	{/if}

	<div class="actions">
		<button class="btn sm" aria-pressed={plugin.enabled} disabled={busy} onclick={toggleEnabled}>
			{plugin.enabled ? "Disable" : "Enable"}
		</button>
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

	{#if pendingUpdate}
		{@const pending = pendingUpdate}
		<PluginConsent
			preview={pending.preview}
			hint={UPDATE_PERMISSION_HINT}
			lines={pending.newPermissions}
			emptyText={NO_NEW_PERMISSIONS}
			confirmLabel="Update plugin"
			{busy}
			onConfirm={() => confirmUpdate(pending)}
			onCancel={() => (pendingUpdate = null)}
		/>
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
			font-size: 1.25rem;
			font-weight: 600;
		}
	}

	.title {
		display: flex;
		align-items: center;
		gap: 14px;
	}

	.muted {
		margin: 4px 0 0;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.badges {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}

	.badge {
		font-size: 0.75rem;
		padding: 3px 10px;
		border-radius: var(--radius-pill);
		border: 1px solid var(--glass-border-bright);

		&.warn {
			border-color: var(--danger);
			color: var(--danger);
		}

		&.update {
			border-color: var(--accent-primary);
			color: var(--accent-primary);
		}
	}

	.release {
		display: flex;
		flex-direction: column;
		gap: 10px;
		align-items: flex-start;
		padding: 14px 16px;
		border-radius: var(--radius-md);
		border: 1px solid var(--glass-border-bright);
		background: var(--bg-surface);
	}

	.release-head {
		display: flex;
		gap: 12px;
		align-items: baseline;
		flex-wrap: wrap;
	}

	.release-title {
		margin: 0;
		font-size: 1.05rem;
		font-weight: 600;
	}

	.notes {
		margin: 0;
		font-size: 0.9375rem;
		line-height: 1.5;
		color: var(--text-muted);
		white-space: pre-wrap;
	}

	.error {
		margin: 0;
		font-size: 0.9375rem;
		color: var(--danger);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
</style>
