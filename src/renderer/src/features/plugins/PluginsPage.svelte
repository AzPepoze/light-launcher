<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount } from "svelte";
	import { ago } from "../../lib/features/format";
	import { errorMessage, listen } from "../../lib/features/rpc";
	import {
		PLUGIN_NOTIFY_EVENT,
		PLUGIN_STATE_EVENT,
		PLUGIN_UPDATES_EVENT
	} from "../../lib/features/events";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import {
		isRemovedPlugin,
		toPluginInfo,
		toPluginUpdatesStatus
	} from "../../lib/features/pluginsMap";
	import { PluginInstallQueue, type QueueView } from "../../lib/features/pluginInstallQueue";
	import {
		planPluginUploads,
		uploadPlugin,
		type SkippedUpload,
		type UploadFile
	} from "../../lib/features/pluginUpload";
	import type { PluginInfo, PluginUpdateStatus, PluginUpdatesStatus } from "../../lib/features/types";
	import type {
		WirePluginInfo,
		WirePluginNotify,
		WirePluginUpdatesStatus,
		WireRemovedPlugin
	} from "../../lib/features/wire";
	import PluginDetail from "./PluginDetail.svelte";
	import PluginDropZone from "./PluginDropZone.svelte";
	import PluginInstallDialog from "./PluginInstallDialog.svelte";
	import {
		CHECKING_UPDATES_LABEL,
		CHECK_UPDATES_LABEL,
		COMMUNITY_PLUGINS_URL,
		DISMISS_LABEL,
		NOTHING_TO_INSTALL_COPY,
		QUEUE_BUSY_COPY,
		SLEEPING_BADGE,
		UPLOADING_TEXT,
		installedSummary,
		pluginProgress,
		skippedText
	} from "./pluginCopy";

	let plugins = $state<PluginInfo[]>([]);
	let loading = $state(true);
	let error = $state("");
	let notice = $state("");
	let selectedId = $state<string | null>(null);
	/** The Install button, which opens the dialog with an empty link box. */
	let installOpen = $state(false);
	const queue = new PluginInstallQueue(uploadPlugin);
	let batch = $state<QueueView>(queue.view);
	/** Dropped items that are not plugins, shown by name. */
	let skipped = $state<SkippedUpload[]>([]);
	let uploadError = $state("");
	let updates = $state<PluginUpdatesStatus>({ checkedAt: null, plugins: [] });
	let checkingUpdates = $state(false);
	let updatesError = $state("");

	const queued = $derived(batch.total > 1);
	const selected = $derived(plugins.find((plugin) => plugin.id === selectedId));
	const updateCount = $derived(updates.plugins.filter((item) => item.available).length);
	const updateProblems = $derived(
		updates.plugins
			.flatMap((item) => (item.error ? [item.error] : []))
			.filter((message, index, all) => all.indexOf(message) === index)
	);
	const subtitle = $derived(
		loading
			? "Loading plugins"
			: `${plugins.length} ${plugins.length === 1 ? "plugin" : "plugins"} installed${
					updateCount ? `, ${updateCount} ${updateCount === 1 ? "update" : "updates"}` : ""
				}`,
	);

	onMount(() => {
		load();
		loadUpdates();
		return subscribe();
	});

	function updateFor(id: string): PluginUpdateStatus | null {
		return updates.plugins.find((item) => item.id === id) ?? null;
	}

	async function load() {
		loading = true;
		error = "";
		try {
			plugins = await pluginsApi.list();
		} catch (err) {
			error = errorMessage(err);
		} finally {
			loading = false;
		}
	}

	async function loadUpdates() {
		try {
			updates = await pluginsApi.updatesStatus();
		} catch (err) {
			updatesError = errorMessage(err);
		}
	}

	async function checkUpdates() {
		checkingUpdates = true;
		updatesError = "";
		try {
			updates = await pluginsApi.checkUpdates();
		} catch (err) {
			updatesError = errorMessage(err);
		} finally {
			checkingUpdates = false;
		}
	}

	/** Keeps the list in step with plugin events. Returns a function that stops them. */
	function subscribe(): () => void {
		const stops = [
			queue.subscribe((view) => (batch = view)),
			listen<WirePluginInfo | WireRemovedPlugin>(PLUGIN_STATE_EVENT, onState),
			listen<WirePluginNotify>(PLUGIN_NOTIFY_EVENT, (event) => {
				notice = `${event.plugin}: ${event.message}`;
			}),
			listen<WirePluginUpdatesStatus>(PLUGIN_UPDATES_EVENT, (event) => {
				updates = toPluginUpdatesStatus(event);
			})
		];
		return () => stops.forEach((stop) => stop());
	}

	function onState(payload: WirePluginInfo | WireRemovedPlugin) {
		if (isRemovedPlugin(payload)) {
			removed(payload.id);
			return;
		}
		replacePlugin(toPluginInfo(payload));
	}

	function replacePlugin(plugin: PluginInfo) {
		const exists = plugins.some((item) => item.id === plugin.id);
		plugins = exists
			? plugins.map((item) => (item.id === plugin.id ? plugin : item))
			: [...plugins, plugin];
	}

	function showInstalled(plugin: PluginInfo) {
		replacePlugin(plugin);
		selectedId = plugin.id;
	}

	function installedManually(plugin: PluginInfo) {
		showInstalled(plugin);
		closeInstall();
	}

	function installedFromQueue(plugin: PluginInfo) {
		showInstalled(plugin);
		void queue.markInstalled();
	}

	function removed(id: string) {
		plugins = plugins.filter((plugin) => plugin.id !== id);
		if (selectedId === id) selectedId = null;
	}

	function openInstall() {
		installOpen = true;
	}

	function closeInstall() {
		installOpen = false;
	}

	function uploadDropped(files: UploadFile[]) {
		if (queue.busy) {
			uploadError = QUEUE_BUSY_COPY;
			return;
		}
		const plan = planPluginUploads(files);
		skipped = plan.skipped;
		uploadError = plan.plugins.length ? "" : NOTHING_TO_INSTALL_COPY;
		void queue.start(plan.plugins);
	}
</script>

<PageHeader title="Plugins" {subtitle} icon="extension">
	{#if updates.checkedAt}
		<span class="checked">Checked {ago(updates.checkedAt)}</span>
	{/if}
	<button class="btn sm" disabled={checkingUpdates} onclick={checkUpdates}>
		{checkingUpdates ? CHECKING_UPDATES_LABEL : CHECK_UPDATES_LABEL}
	</button>
	<button class="btn sm primary" onclick={() => openInstall()}>Install</button>
</PageHeader>

<PluginDropZone onFiles={uploadDropped}>
	{#if notice}
		<p class="notice" role="status">{notice}</p>
	{/if}

	{#if updatesError}
		<p class="upload-error" role="alert">{updatesError}</p>
	{/if}

	{#each updateProblems as problem (problem)}
		<p class="upload-error" role="alert">{problem}</p>
	{/each}

	{#if batch.phase === "uploading"}
		<p class="notice" role="status">{UPLOADING_TEXT}</p>
	{/if}

	{#if batch.phase === "failed"}
		<div class="failure">
			<p class="upload-error" role="alert">{batch.failure}</p>
			<button class="btn sm" onclick={() => queue.dismissFailure()}>{DISMISS_LABEL}</button>
		</div>
	{/if}

	{#if batch.phase === "done" && queued}
		<p class="notice" role="status">{installedSummary(batch.installed, batch.total)}</p>
	{/if}

	{#if skipped.length}
		<ul class="skipped" aria-label="Skipped items">
			{#each skipped as item (item.name)}
				<li>{skippedText(item)}</li>
			{/each}
		</ul>
	{/if}

	{#if uploadError}
		<p class="upload-error" role="alert">{uploadError}</p>
	{/if}

	{#if error}
		<div class="status">
			<p role="alert">{error}</p>
			<button class="btn sm" onclick={load}>Retry</button>
		</div>
	{:else if loading}
		<p class="status">Loading plugins...</p>
	{:else if selected}
		{#key selected.id}
			<PluginDetail
				plugin={selected}
				update={updateFor(selected.id)}
				onChanged={replacePlugin}
				onRemoved={removed}
				onBack={() => (selectedId = null)}
			/>
		{/key}
	{:else if plugins.length === 0}
		<div class="status">
			<p class="title">No plugins installed</p>
			<button class="btn sm primary" onclick={() => openInstall()}>Install a plugin</button>
			<p class="community">
				{#if COMMUNITY_PLUGINS_URL}
					<a href={COMMUNITY_PLUGINS_URL} target="_blank" rel="noreferrer">Community plugins</a>
				{:else}
					<span>Community plugins</span>
				{/if}
			</p>
		</div>
	{:else}
		<ul class="list" aria-label="Installed plugins">
			{#each plugins as plugin (plugin.id)}
				<li>
					<button class="row" onclick={() => (selectedId = plugin.id)}>
						<span class="name">{plugin.name}</span>
						<span class="muted">v{plugin.version}</span>
						<span class="badges">
							<span class="badge">{plugin.enabled ? "Enabled" : "Disabled"}</span>
							{#if plugin.idle}<span class="badge">{SLEEPING_BADGE}</span>{/if}
							{#if !plugin.signed}<span class="badge warn">Unsigned</span>{/if}
							{#if plugin.status === "error"}<span class="badge warn">Error</span>{/if}
							{#if updateFor(plugin.id)?.available}
								<span class="badge update">Update available ({updateFor(plugin.id)?.version})</span>
							{/if}
						</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</PluginDropZone>

{#if batch.phase === "review"}
	{#key batch.position}
		<PluginInstallDialog
			initialSource={batch.source}
			progress={queued ? pluginProgress(batch.position, batch.total) : ""}
			onInstalled={installedFromQueue}
			onClose={() => queue.skip()}
		/>
	{/key}
{:else if installOpen}
	<PluginInstallDialog onInstalled={installedManually} onClose={closeInstall} />
{/if}

<style lang="scss">
	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.row {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border: 1px solid transparent;
		border-radius: var(--radius-md);
		background: var(--bg-surface);
		color: var(--text-main);
		text-align: left;
		cursor: pointer;
		font: inherit;

		&:hover,
		&:focus-visible {
			border-color: var(--glass-border-bright);
			outline: none;
		}
	}

	.name {
		font-weight: 700;
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.muted {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.upload-error {
		margin: 0 0 12px;
		font-size: 0.85rem;
		color: var(--danger);
	}

	.failure {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;

		.upload-error {
			margin: 0;
		}
	}

	.skipped {
		margin: 0 0 12px;
		padding-left: 18px;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.badges {
		display: flex;
		gap: 6px;
	}

	.badge {
		font-size: 0.7rem;
		padding: 2px 8px;
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

	.checked {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.status {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 48px 16px;
		color: var(--text-muted);
		text-align: center;

		p {
			margin: 0;
		}

		.title {
			font-size: 1rem;
			font-weight: 700;
			color: var(--text-main);
		}

		.community {
			font-size: 0.85rem;
			text-decoration: underline;
		}
	}
</style>
