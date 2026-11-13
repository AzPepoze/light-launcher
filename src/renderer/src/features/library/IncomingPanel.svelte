<script lang="ts">
	import { onMount } from "svelte";
	import { listen } from "../../lib/features/rpc";
	import { LIBRARY_INCOMING_EVENT } from "../../lib/features/events";
	import { libraryApi } from "../../lib/features/libraryApi";
	import { ago, bytes } from "../../lib/features/format";
	import type { IncomingItem, IncomingListResponse } from "../../lib/features/types";
	import { notifications } from "@stores/notificationStore";
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import { libraryErrorMessage } from "./libraryError";

	let { onChanged }: { onChanged: () => void } = $props();

	let incoming = $state<IncomingListResponse | null>(null);
	let busy = $state(false);

	const visibleItems = $derived((incoming?.items ?? []).filter((item) => !item.ignored));

	onMount(() => {
		reload();
		return listen<unknown>(LIBRARY_INCOMING_EVENT, () => reload());
	});

	async function reload() {
		try {
			incoming = await libraryApi.incomingList();
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		}
	}

	async function act(action: () => Promise<unknown>, changesLibrary: boolean) {
		busy = true;
		try {
			await action();
			await reload();
			if (changesLibrary) onChanged();
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		} finally {
			busy = false;
		}
	}

	function link(item: IncomingItem, gameId: number) {
		act(() => libraryApi.incomingLink(item.name, gameId), true);
	}

	function createNew(item: IncomingItem) {
		act(() => libraryApi.incomingCreateLocal(item.name), true);
	}

	function ignore(item: IncomingItem) {
		act(() => libraryApi.incomingIgnore(item.name), false);
	}
</script>

<section class="panel" aria-label="Incoming files">
	{#if incoming && !incoming.configured}
		<Notice>No incoming folder is set. Files you drop there show up here.</Notice>
	{:else if incoming?.warning}
		<Notice tone="warning">{incoming.warning}</Notice>
	{/if}

	{#if incoming && visibleItems.length === 0 && incoming.configured}
		<EmptyState icon="inbox" title="Nothing waiting. New files appear here." />
	{/if}

	<ul class="list">
		{#each visibleItems as item (item.name)}
			<li class="item">
				<div class="head">
					<p class="name" title={item.name}>{item.name}</p>
					<p class="muted">
						{bytes(item.size)} · {ago(item.mtime)}
						{#if item.auto}· Will import automatically{/if}
						{#if !item.ready}· Still copying{/if}
					</p>
				</div>

				{#if item.suggestions.length}
					<ul class="suggestions" aria-label="Possible matches">
						{#each item.suggestions as suggestion (suggestion.gameId)}
							<li>
								<Button
									size="sm"
									disabled={busy || !item.ready}
									onclick={() => link(item, suggestion.gameId)}
								>
									Link to {suggestion.title}
								</Button>
							</li>
						{/each}
					</ul>
				{/if}

				<div class="actions">
					<Button size="sm" disabled={busy || !item.ready} onclick={() => createNew(item)}>
						New game
					</Button>
					<Button size="sm" disabled={busy} onclick={() => ignore(item)}>Ignore</Button>
				</div>
			</li>
		{/each}
	</ul>
</section>

<style lang="scss">
	.panel {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding-bottom: 24px;
	}

	.muted {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.item {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 12px 14px;
		border-radius: var(--radius-md);
		background: var(--bg-surface);
	}

	.head {
		min-width: 0;

		p {
			margin: 0;
		}
	}

	.name {
		font-size: 1rem;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.suggestions,
	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		list-style: none;
		margin: 0;
		padding: 0;
	}
</style>
