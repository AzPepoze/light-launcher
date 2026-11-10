<script lang="ts">
	import BrowseInput from "@components/shared/BrowseInput.svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type { Game } from "../../lib/features/types";
	import type { LibraryState } from "./libraryState.svelte";

	let { game, library }: { game: Game; library: LibraryState } = $props();

	let folder = $state("");
	let message = $state("");
	let linking = $state(false);

	async function linkFolder() {
		if (!folder) return;
		linking = true;
		message = "";
		try {
			library.replaceGame(await libraryApi.linkLocal(game.id, folder));
			message = "Folder linked.";
			folder = "";
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
		<p class="muted">Not installed. Download a build or link a folder you already have.</p>
	{/if}

	<div class="link">
		<BrowseInput bind:value={folder} label="Link an existing folder" id="link-folder" placeholder="Pick the game folder" />
		<button class="btn sm" disabled={!folder || linking} onclick={linkFolder}>Link folder</button>
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

		:global(.browse-input-container) {
			flex: 1 1 260px;
			min-width: 0;
		}
	}
</style>
