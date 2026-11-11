<script lang="ts">
	import { errorMessage } from "../../lib/features/rpc";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type { Game } from "../../lib/features/types";
	import type { LibraryState } from "./libraryState.svelte";

	let { game, library }: { game: Game; library: LibraryState } = $props();

	let sourceUrl = $state("");
	let message = $state("");
	let linking = $state(false);
	let folderPath = $state("");
	let showFolderField = $state(false);
	let linkingFolder = $state(false);
	let folderMessage = $state("");
	let folderError = $state("");

	async function linkFolder(path: string) {
		const folder = path.trim();
		if (!folder) {
			folderError = "Enter a folder path.";
			return;
		}
		linkingFolder = true;
		folderError = "";
		folderMessage = "";
		try {
			library.replaceGame(await libraryApi.linkFolder(game.id, folder));
			folderMessage = "Linked folder.";
			folderPath = "";
			showFolderField = false;
		} catch (err) {
			folderError = errorMessage(err);
		} finally {
			linkingFolder = false;
		}
	}

	async function pickFolder() {
		const shell = window.llShell;
		if (!shell?.pickFolder) {
			showFolderField = true;
			return;
		}
		folderError = "";
		folderMessage = "";
		try {
			const picked = await shell.pickFolder();
			if (picked) await linkFolder(picked);
		} catch (err) {
			folderError = errorMessage(err);
		}
	}

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
		<button class="btn sm" disabled={linkingFolder} onclick={pickFolder}>Link folder</button>
		{#if showFolderField}
			<input
				class="input"
				bind:value={folderPath}
				placeholder="/path/to/game"
				aria-label="Game folder"
				onkeydown={(event) => event.key === "Enter" && linkFolder(folderPath)}
			/>
			<button class="btn sm" disabled={linkingFolder} onclick={() => linkFolder(folderPath)}>
				Link
			</button>
			<button class="btn sm" onclick={() => (showFolderField = false)}>Cancel</button>
		{/if}
	</div>

	{#if folderError}
		<p class="error" role="alert">{folderError}</p>
	{:else if folderMessage}
		<p class="muted" role="status">{folderMessage}</p>
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
		gap: 10px;

		h3 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 600;
		}

		p {
			margin: 0;
		}
	}

	.path {
		font-family: monospace;
		font-size: 0.875rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.muted {
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.error {
		color: var(--danger);
		font-size: 0.875rem;
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
		font-size: 0.875rem;
		color: var(--text-muted);
	}
</style>
