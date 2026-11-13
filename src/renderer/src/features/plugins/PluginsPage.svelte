<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { notifications } from "@stores/notificationStore";
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
	import PluginRow from "./PluginRow.svelte";
	import {
		CHECKING_UPDATES_LABEL,
		CHECK_UPDATES_LABEL,
		COMMUNITY_PLUGINS_URL,
		NOTHING_TO_INSTALL_COPY,
		QUEUE_BUSY_COPY,
		UPLOADING_TEXT,
		installedSummary
	} from "./pluginCopy";
	import { checkOutcome, skippedSummary } from "./pluginNotices";

	let plugins = $state<PluginInfo[]>([]);
	let loading = $state(true);
	let loadFailed = $state(false);
	let selectedId = $state<string | null>(null);
	/** The Install button, which opens the dialog with an empty link box. */
	let installOpen = $state(false);
	const queue = new PluginInstallQueue(uploadPlugin);
	let batch = $state<QueueView>(queue.view);
	let updates = $state<PluginUpdatesStatus>({ checkedAt: null, plugins: [] });
	let checkingUpdates = $state(false);

	const queued = $derived(batch.total > 1);
	const selected = $derived(plugins.find((plugin) => plugin.id === selectedId));
	const updateCount = $derived(updates.plugins.filter((item) => item.available).length);
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
		loadFailed = false;
		try {
			plugins = await pluginsApi.list();
		} catch {
			loadFailed = true;
		} finally {
			loading = false;
		}
	}

	/** Reads the cached update results. It does not ask GitHub. */
	async function loadUpdates() {
		try {
			updates = await pluginsApi.updatesStatus();
		} catch (err) {
			notifications.error(errorMessage(err));
		}
	}

	/** One check run, one notification with the result. */
	async function checkUpdates() {
		checkingUpdates = true;
		try {
			updates = await pluginsApi.checkUpdates();
			const outcome = checkOutcome(updates.plugins);
			notifications[outcome.tone](outcome.message);
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			checkingUpdates = false;
		}
	}

	/** Keeps the list in step with plugin events. Returns a function that stops them. */
	function subscribe(): () => void {
		const stops = [
			queue.subscribe(onQueueView),
			listen<WirePluginInfo | WireRemovedPlugin>(PLUGIN_STATE_EVENT, onState),
			listen<WirePluginNotify>(PLUGIN_NOTIFY_EVENT, (event) => {
				notifications.info(`${event.plugin}: ${event.message}`);
			}),
			listen<WirePluginUpdatesStatus>(PLUGIN_UPDATES_EVENT, (event) => {
				updates = toPluginUpdatesStatus(event);
			})
		];
		return () => stops.forEach((stop) => stop());
	}

	function onQueueView(view: QueueView) {
		batch = view;
		if (view.phase === "uploading") notifications.info(UPLOADING_TEXT);
		if (view.phase === "failed") {
			notifications.error(view.failure);
			void queue.dismissFailure();
		}
		if (view.phase === "done" && view.total > 1) {
			notifications.success(installedSummary(view.installed, view.total));
		}
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
			notifications.info(QUEUE_BUSY_COPY);
			return;
		}
		const plan = planPluginUploads(files);
		if (plan.skipped.length > 0) notifications.error(skippedSummary(plan.skipped));
		if (plan.plugins.length === 0) {
			notifications.error(NOTHING_TO_INSTALL_COPY);
			return;
		}
		void queue.start(plan.plugins);
	}
</script>

<PageHeader title="Plugins" {subtitle} icon="extension">
	{#if updates.checkedAt}
		<span class="checked">Checked {ago(updates.checkedAt)}</span>
	{/if}
	<Button size="sm" disabled={checkingUpdates} onclick={checkUpdates}>
		{checkingUpdates ? CHECKING_UPDATES_LABEL : CHECK_UPDATES_LABEL}
	</Button>
	<Button variant="primary" size="sm" onclick={() => openInstall()}>Install</Button>
</PageHeader>

<PluginDropZone onFiles={uploadDropped}>
	{#if loadFailed}
		<Notice tone="error">
			<p>Plugins could not be loaded.</p>
			<Button size="sm" onclick={load}>Retry</Button>
		</Notice>
	{:else if loading}
		<LoadingState label="Loading plugins..." />
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
		<EmptyState icon="extension" title="No plugins installed">
			<Button variant="primary" size="sm" onclick={() => openInstall()}>Install a plugin</Button>
			<p class="community">
				{#if COMMUNITY_PLUGINS_URL}
					<a href={COMMUNITY_PLUGINS_URL} target="_blank" rel="noreferrer">Community plugins</a>
				{:else}
					<span>Community plugins</span>
				{/if}
			</p>
		</EmptyState>
	{:else}
		<ul class="list" aria-label="Installed plugins">
			{#each plugins as plugin (plugin.id)}
				<li>
					<PluginRow {plugin} update={updateFor(plugin.id)} onOpen={() => (selectedId = plugin.id)} />
				</li>
			{/each}
		</ul>
	{/if}
</PluginDropZone>

{#if batch.phase === "review"}
	{#key batch.position}
		<PluginInstallDialog
			initialSource={batch.source}
			queue={queued ? { position: batch.position, total: batch.total } : null}
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

	.checked {
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.community {
		margin: 0;
		font-size: 0.875rem;
		text-decoration: underline;
	}
</style>
