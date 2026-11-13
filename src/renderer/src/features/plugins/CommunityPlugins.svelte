<script lang="ts">
	import { onMount } from "svelte";
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Modal from "@components/shared/Modal.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { communityEntries, pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginRegistry, PluginRegistryEntry } from "../../lib/features/types";
	import {
		COMMUNITY_HOMEPAGE_LABEL,
		COMMUNITY_INSTALL_HINT,
		COMMUNITY_INSTALL_LABEL,
		COMMUNITY_LOAD_FAILED_COPY,
		COMMUNITY_LOADING_LABEL,
		COMMUNITY_NOTHING_AVAILABLE_COPY,
		COMMUNITY_NOT_CONFIGURED_COPY,
		COMMUNITY_NOT_CONFIGURED_HINT,
		COMMUNITY_RETRY_LABEL,
		COMMUNITY_TITLE,
		COMMUNITY_UNKNOWN_AUTHOR
	} from "./pluginCopy";

	let {
		installedIds,
		onInstall,
		onClose
	}: {
		/** Ids already installed. Their entries are hidden so an Install button can never update silently. */
		installedIds: string[];
		onInstall: (entry: PluginRegistryEntry) => void;
		onClose: () => void;
	} = $props();

	let registry = $state<PluginRegistry | null>(null);
	let loading = $state(true);
	let failure = $state("");

	const entries = $derived(registry ? communityEntries(registry.entries, installedIds) : []);

	async function load() {
		loading = true;
		failure = "";
		try {
			registry = await pluginsApi.registry();
		} catch (err) {
			failure = errorMessage(err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void load();
	});
</script>

<Modal show={true} title={COMMUNITY_TITLE} {onClose}>
	{#if loading}
		<LoadingState label={COMMUNITY_LOADING_LABEL} />
	{:else if failure}
		<Notice tone="error" icon="error">
			<p>{COMMUNITY_LOAD_FAILED_COPY}</p>
			<p class="detail">{failure}</p>
		</Notice>
		<div class="actions">
			<Button size="sm" icon="refresh" onclick={load}>{COMMUNITY_RETRY_LABEL}</Button>
		</div>
	{:else if registry}
		{#if !registry.configured}
			<p class="empty">{COMMUNITY_NOT_CONFIGURED_COPY}</p>
			<p class="hint">{COMMUNITY_NOT_CONFIGURED_HINT}</p>
		{:else if registry.error}
			<Notice tone="error" icon="error">
				<p>{COMMUNITY_LOAD_FAILED_COPY}</p>
				<p class="detail">{registry.error}</p>
			</Notice>
			<div class="actions">
				<Button size="sm" icon="refresh" onclick={load}>{COMMUNITY_RETRY_LABEL}</Button>
			</div>
		{:else if entries.length === 0}
			<p class="empty">{COMMUNITY_NOTHING_AVAILABLE_COPY}</p>
		{:else}
			<p class="hint">{COMMUNITY_INSTALL_HINT}</p>
			<ul class="entries" aria-label={COMMUNITY_TITLE}>
				{#each entries as entry (entry.id)}
					<li class="entry">
						<div class="entry-head">
							<span class="name">{entry.name}</span>
							<Badge tone="neutral">v{entry.version}</Badge>
						</div>
						<p class="author">by {entry.author || COMMUNITY_UNKNOWN_AUTHOR}</p>
						{#if entry.description}
							<p class="description">{entry.description}</p>
						{/if}
						<div class="entry-actions">
							{#if entry.homepage}
								<a class="link" href={entry.homepage} target="_blank" rel="noreferrer">
									{COMMUNITY_HOMEPAGE_LABEL}
								</a>
							{/if}
							<Button variant="primary" size="sm" onclick={() => onInstall(entry)}>
								{COMMUNITY_INSTALL_LABEL}
							</Button>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	{/if}
</Modal>

<style lang="scss">
	.entries {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.entry {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 14px 16px;
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: var(--radius-md);
		background: var(--bg-surface);
	}

	.entry-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
	}

	.name {
		font-size: 1rem;
		font-weight: 700;
		color: var(--text-main);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.author,
	.description,
	.empty,
	.hint,
	.detail {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.875rem;
		line-height: 1.45;
	}

	.description {
		color: var(--text-main);
	}

	.entry-actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 14px;
		margin-top: 8px;
	}

	.link {
		color: var(--accent-primary);
		font-size: 0.875rem;
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		margin-top: 12px;
	}

	.empty {
		color: var(--text-main);
		font-size: 0.95rem;
	}

	.hint {
		margin-bottom: 12px;
	}
</style>
