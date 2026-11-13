<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Modal from "@components/shared/Modal.svelte";
	import { onMount } from "svelte";
	import { notifications } from "@stores/notificationStore";
	import { libraryErrorMessage } from "./libraryError";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type { Game, GamePreview, SearchHit } from "../../lib/features/types";

	let {
		onAdded,
		onClose,
		initialUrl = ""
	}: { onAdded: (game: Game) => void; onClose: () => void; initialUrl?: string } = $props();

	let query = $state("");
	let hits = $state<SearchHit[]>([]);
	let searchErrors = $state<string[]>([]);
	let preview = $state<GamePreview | null>(null);
	let busy = $state(false);
	let message = $state("");

	const isLink = (text: string) => /^https?:\/\//i.test(text.trim());

	// A browse tile opens the dialog with its link ready to preview.
	onMount(() => {
		query = initialUrl;
		if (initialUrl) void search();
	});

	async function search() {
		const text = query.trim();
		if (!text) return;
		await run(async () => {
			preview = null;
			searchErrors = [];
			if (isLink(text)) {
				preview = await libraryApi.preview(text);
				return;
			}
			const result = await libraryApi.search(text);
			hits = result.hits;
			searchErrors = result.errors;
			if (hits.length === 0) message = "No matches. Try another name or paste a link.";
		});
	}

	function choose(hit: SearchHit) {
		run(async () => {
			preview = await libraryApi.preview(hit.url);
		});
	}

	function add() {
		if (!preview) return;
		const url = preview.url;
		run(async () => {
			onAdded(await libraryApi.addFromUrl(url));
		});
	}

	async function run(action: () => Promise<void>) {
		busy = true;
		message = "";
		try {
			await action();
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		} finally {
			busy = false;
		}
	}
</script>

<Modal show={true} title="Add game" showDone={false} {onClose}>
	<form class="form" onsubmit={(event) => { event.preventDefault(); search(); }}>
		<label class="field">
			<span>Search by name or paste a link</span>
			<div class="row">
				<input class="input" type="text" bind:value={query} placeholder="Game name or https://..." />
				<Button variant="primary" size="sm" type="submit" disabled={busy || !query.trim()}>Search</Button>
			</div>
		</label>
	</form>

	{#if message}
		<p class="message" role="status">{message}</p>
	{/if}

	{#each searchErrors as error (error)}
		<p class="message" role="alert">{error}</p>
	{/each}

	{#if preview}
		<article class="preview" aria-label="Preview">
			<h3>{preview.title}</h3>
			<p class="muted">
				{[preview.source, preview.developer, preview.version].filter(Boolean).join(" · ")}
			</p>
			{#if preview.description}
				<p class="description">{preview.description}</p>
			{/if}
			{#if preview.existing}
				<p class="message">Already in your library.</p>
			{:else}
				<Button variant="primary" size="sm" disabled={busy} onclick={add}>Add to library</Button>
			{/if}
		</article>
	{:else if hits.length}
		<ul class="hits" aria-label="Search results">
			{#each hits as hit (hit.url)}
				<li>
					<Button disabled={busy} onclick={() => choose(hit)}>{hit.title}</Button>
				</li>
			{/each}
		</ul>
	{/if}
</Modal>

<style lang="scss">
	.form {
		margin: 0;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.row {
		display: flex;
		gap: 8px;

		.input {
			flex: 1;
			min-width: 0;
		}
	}

	.message {
		margin: 12px 0 0;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.hits {
		list-style: none;
		margin: 12px 0 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
		max-height: 340px;
		overflow-y: auto;
	}

	.hits :global(.button) {
		width: 100%;
		justify-content: flex-start;
	}

	.preview {
		margin-top: 16px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: flex-start;

		h3 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 600;
		}

		p {
			margin: 0;
		}
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.description {
		max-width: 68ch;
		font-size: 1rem;
		line-height: 1.5;
		display: -webkit-box;
		line-clamp: 4;
		-webkit-line-clamp: 4;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
