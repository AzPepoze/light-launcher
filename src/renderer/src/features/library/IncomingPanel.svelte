<script lang="ts">
	import { onMount } from "svelte";
	import { errorMessage, listen } from "../../lib/features/rpc";
	import { LIBRARY_INCOMING_EVENT } from "../../lib/features/events";
	import { libraryApi } from "../../lib/features/libraryApi";
	import { ago, bytes } from "../../lib/features/format";
	import type { IncomingItem, IncomingListResponse } from "../../lib/features/types";

	let { onChanged }: { onChanged: () => void } = $props();

	let incoming = $state<IncomingListResponse | null>(null);
	let message = $state("");
	let busy = $state(false);

	const visibleItems = $derived((incoming?.items ?? []).filter((item) => !item.ignored));

	onMount(() => {
		reload();
		return listen<unknown>(LIBRARY_INCOMING_EVENT, () => reload());
	});

	async function reload() {
		try {
			incoming = await libraryApi.incomingList();
			message = "";
		} catch (err) {
			message = errorMessage(err);
		}
	}

	async function act(action: () => Promise<unknown>, changesLibrary: boolean) {
		busy = true;
		message = "";
		try {
			await action();
			await reload();
			if (changesLibrary) onChanged();
		} catch (err) {
			message = errorMessage(err);
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
		<p class="notice">No incoming folder is set. Files you drop there show up here.</p>
	{:else if incoming?.warning}
		<p class="notice">{incoming.warning}</p>
	{/if}

	{#if message}
		<p class="notice" role="alert">{message}</p>
	{/if}

	{#if incoming && visibleItems.length === 0 && incoming.configured}
		<p class="empty">Nothing waiting. New files appear here.</p>
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
								<button
									class="btn sm"
									disabled={busy || !item.ready}
									onclick={() => link(item, suggestion.gameId)}
								>
									Link to {suggestion.title}
								</button>
							</li>
						{/each}
					</ul>
				{/if}

				<div class="actions">
					<button class="btn sm" disabled={busy || !item.ready} onclick={() => createNew(item)}>
						New game
					</button>
					<button class="btn sm" disabled={busy} onclick={() => ignore(item)}>Ignore</button>
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

	.notice {
		margin: 0;
		padding: 10px 14px;
		border-radius: var(--radius-md);
		background: var(--bg-surface);
		border: 1px solid var(--glass-border);
		font-size: 0.85rem;
	}

	.empty,
	.muted {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.85rem;
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
		font-weight: 700;
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
