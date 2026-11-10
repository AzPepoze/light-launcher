<script lang="ts">
	import { errorMessage } from "../../lib/features/rpc";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type { Game } from "../../lib/features/types";
	import type { LibraryState } from "./libraryState.svelte";

	let { game, library }: { game: Game; library: LibraryState } = $props();

	let sourceUrl = $state("");
	let message = $state("");
	let linking = $state(false);

	async function linkSource() {
		const url = sourceUrl.trim();
		if (!url) return;
		linking = true;
		message = "";
		try {
			library.replaceGame(await libraryApi.linkSource(game.id, url));
			message = "Linked to the source page.";
			sourceUrl = "";
		} catch (err) {
			message = errorMessage(err);
		} finally {
			linking = false;
		}
	}
</script>

<section>
	<h3>Install</h3>

	{#if game.installDir}
		<p class="path" title={game.installDir}>{game.installDir}</p>
	{:else}
		<p class="muted">Not installed. Download a file above, or link a source page.</p>
	{/if}

	<div class="link">
		<label class="source-field">
			<span>Link a source page</span>
			<input class="input" type="url" bind:value={sourceUrl} placeholder="https://..." />
		</label>
		<button class="btn sm" disabled={!sourceUrl.trim() || linking} onclick={linkSource}>Link</button>
	</div>

	{#if message}
		<p class="muted" role="status">{message}</p>
	{/if}
</section>

<style lang="scss">
	section {
		display: flex;
		flex-direction: column;
		gap: 8px;

		h3 {
			margin: 0;
			font-size: 0.8rem;
			font-weight: 800;
			letter-spacing: 1px;
			text-transform: uppercase;
			color: var(--text-muted);
		}

		p {
			margin: 0;
		}
	}

	.path {
		font-family: monospace;
		font-size: 0.85rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.link {
		display: flex;
		align-items: flex-end;
		gap: 8px;
		flex-wrap: wrap;
	}

	.source-field {
		display: flex;
		flex-direction: column;
		gap: 4px;
		flex: 1 1 260px;
		min-width: 0;
		font-size: 0.85rem;
		color: var(--text-muted);
	}
</style>
