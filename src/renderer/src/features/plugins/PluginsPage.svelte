<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount } from "svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo } from "../../lib/features/types";
	import PluginDetail from "./PluginDetail.svelte";
	import PluginInstallDialog from "./PluginInstallDialog.svelte";
	import { COMMUNITY_PLUGINS_URL } from "./pluginCopy";

	let plugins = $state<PluginInfo[]>([]);
	let loading = $state(true);
	let error = $state("");
	let selectedId = $state<string | null>(null);
	let installOpen = $state(false);

	const selected = $derived(plugins.find((plugin) => plugin.id === selectedId));
	const subtitle = $derived(
		loading ? "Loading plugins" : `${plugins.length} ${plugins.length === 1 ? "plugin" : "plugins"} installed`,
	);

	onMount(() => {
		load();
	});

	async function load() {
		loading = true;
		error = "";
		try {
			plugins = (await pluginsApi.list()).items;
		} catch (err) {
			error = errorMessage(err);
		} finally {
			loading = false;
		}
	}

	function replacePlugin(plugin: PluginInfo) {
		plugins = plugins.map((item) => (item.id === plugin.id ? plugin : item));
	}

	function installed(plugin: PluginInfo) {
		plugins = [...plugins.filter((item) => item.id !== plugin.id), plugin];
		installOpen = false;
		selectedId = plugin.id;
	}

	function removed(id: string) {
		plugins = plugins.filter((plugin) => plugin.id !== id);
		selectedId = null;
	}
</script>

<PageHeader title="Plugins" {subtitle} icon="extension">
	<button class="btn sm primary" onclick={() => (installOpen = true)}>Install</button>
</PageHeader>

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
		<button class="btn sm primary" onclick={() => (installOpen = true)}>Install a plugin</button>
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

{#if installOpen}
	<PluginInstallDialog onInstalled={installed} onClose={() => (installOpen = false)} />
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
