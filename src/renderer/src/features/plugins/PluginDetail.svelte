<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import { notifications } from "@stores/notificationStore";
	import { errorMessage } from "../../lib/features/rpc";
	import { ago } from "../../lib/features/format";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo, PluginUpdatePreview, PluginUpdateStatus } from "../../lib/features/types";
	import PluginAccounts from "./PluginAccounts.svelte";
	import PluginConsent from "./PluginConsent.svelte";
	import PluginIdentity from "./PluginIdentity.svelte";
	import PluginLogs from "./PluginLogs.svelte";
	import PluginSettings from "./PluginSettings.svelte";
	import UnsignedBadge from "./UnsignedBadge.svelte";
	import {
		NO_NEW_PERMISSIONS,
		SLEEPING_BADGE,
		UNSIGNED_WARNING,
		UPDATE_PERMISSION_HINT
	} from "./pluginCopy";
	import { UNAVAILABLE_CHECK_TEXT, checkFailureHint, isCheckFailure } from "./pluginNotices";

	const NO_PACKAGE_COPY = "This release has no plugin package.";

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
	let confirmingRemove = $state(false);
	let pendingUpdate = $state<PluginUpdatePreview | null>(null);
	const checkFailed = $derived(update !== null && isCheckFailure(update));

	async function run(action: () => Promise<void>) {
		busy = true;
		try {
			await action();
		} catch (err) {
			notifications.error(errorMessage(err));
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
				notifications.error(NO_PACKAGE_COPY);
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
	<Button size="sm" icon="arrow_back" class="back" onclick={onBack}>Plugins</Button>

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
			<Badge>{plugin.enabled ? "Enabled" : "Disabled"}</Badge>
			{#if plugin.idle}<Badge icon="bedtime">{SLEEPING_BADGE}</Badge>{/if}
			{#if !plugin.signed}
				<UnsignedBadge />
			{/if}
			{#if plugin.status === "broken"}
				<Badge tone="danger" icon="error">Error</Badge>
			{/if}
			{#if update?.available}
				<Badge tone="accent">Update available ({update.version})</Badge>
			{/if}
		</div>
	</header>

	{#if plugin.error}
		<Notice tone="error">{plugin.error}</Notice>
	{/if}
	{#if !plugin.signed}
		<Notice tone="warning" icon="warning">{UNSIGNED_WARNING}</Notice>
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
			<Button variant="primary" size="sm" disabled={busy} onclick={() => previewUpdate(update)}>
				Update to v{update.version}
			</Button>
		</section>
	{/if}
	{#if checkFailed && update?.error}
		<p class="muted" title={checkFailureHint(update)}>{UNAVAILABLE_CHECK_TEXT}</p>
	{/if}

	<div class="actions">
		<Button size="sm" aria-pressed={plugin.enabled} disabled={busy} onclick={toggleEnabled}>
			{plugin.enabled ? "Disable" : "Enable"}
		</Button>
		<Button size="sm" disabled={busy} onclick={rollback}>Roll back</Button>
		{#if confirmingRemove}
			<Button size="sm" variant="danger" disabled={busy} onclick={remove}>Confirm remove</Button>
			<Button size="sm" onclick={() => (confirmingRemove = false)}>Cancel</Button>
		{:else}
			<Button size="sm" disabled={busy} onclick={() => (confirmingRemove = true)}>Remove</Button>
		{/if}
	</div>

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

	:global(.detail .back) {
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

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
</style>
