<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount } from "svelte";
	import { errorMessage, listen } from "../../lib/features/rpc";
	import { PLUGIN_NOTIFY_EVENT, PLUGIN_STATE_EVENT } from "../../lib/features/events";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import { isRemovedPlugin, toPluginInfo } from "../../lib/features/pluginsMap";
	import { planPluginUpload, uploadPlugin, type UploadFile } from "../../lib/features/pluginUpload";
	import type { PluginInfo } from "../../lib/features/types";
	import type { WirePluginInfo, WirePluginNotify, WireRemovedPlugin } from "../../lib/features/wire";
	import PluginDetail from "./PluginDetail.svelte";
	import PluginDropZone from "./PluginDropZone.svelte";
	import PluginInstallDialog from "./PluginInstallDialog.svelte";
	import { COMMUNITY_PLUGINS_URL, UPLOAD_PROBLEM_COPY, UPLOADING_TEXT } from "./pluginCopy";

	let plugins = $state<PluginInfo[]>([]);
	let loading = $state(true);
	let error = $state("");
	let notice = $state("");
	let selectedId = $state<string | null>(null);
	let installOpen = $state(false);
	/** Server path of a dropped plugin, passed to the install dialog to check. */
	let droppedSource = $state("");
	let uploading = $state(false);
	let uploadError = $state("");

	const selected = $derived(plugins.find((plugin) => plugin.id === selectedId));
	const subtitle = $derived(
		loading ? "Loading plugins" : `${plugins.length} ${plugins.length === 1 ? "plugin" : "plugins"} installed`,
	);

	onMount(() => {
		load();
		return subscribe();
	});

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

	/** Keeps the list in step with plugin events. Returns a function that stops them. */
	function subscribe(): () => void {
		const stops = [
			listen<WirePluginInfo | WireRemovedPlugin>(PLUGIN_STATE_EVENT, onState),
			listen<WirePluginNotify>(PLUGIN_NOTIFY_EVENT, (event) => {
				notice = `${event.plugin}: ${event.message}`;
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

	function installed(plugin: PluginInfo) {
		replacePlugin(plugin);
		closeInstall();
		selectedId = plugin.id;
	}

	function removed(id: string) {
		plugins = plugins.filter((plugin) => plugin.id !== id);
		if (selectedId === id) selectedId = null;
	}

	function openInstall(source = "") {
		droppedSource = source;
		installOpen = true;
	}

	function closeInstall() {
		installOpen = false;
		droppedSource = "";
	}

	async function uploadDropped(files: UploadFile[]) {
		const plan = planPluginUpload(files);
		if (plan.ok === false) {
			uploadError = UPLOAD_PROBLEM_COPY[plan.problem];
			return;
		}

		uploading = true;
		uploadError = "";
		try {
			openInstall(await uploadPlugin(plan.files));
		} catch (err) {
			uploadError = errorMessage(err);
		} finally {
			uploading = false;
		}
	}
</script>

<PageHeader title="Plugins" {subtitle} icon="extension">
	<button class="btn sm primary" onclick={() => openInstall()}>Install</button>
</PageHeader>

<PluginDropZone onFiles={uploadDropped}>
	{#if notice}
		<p class="notice" role="status">{notice}</p>
	{/if}

	{#if uploading}
		<p class="notice" role="status">{UPLOADING_TEXT}</p>
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
			<PluginDetail plugin={selected} onChanged={replacePlugin} onRemoved={removed} onBack={() => (selectedId = null)} />
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
							{#if !plugin.signed}<span class="badge warn">Unsigned</span>{/if}
							{#if plugin.status === "error"}<span class="badge warn">Error</span>{/if}
						</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</PluginDropZone>

{#if installOpen}
	<PluginInstallDialog initialSource={droppedSource} onInstalled={installed} onClose={closeInstall} />
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
