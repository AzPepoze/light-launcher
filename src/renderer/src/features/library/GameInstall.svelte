<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import { notifications } from "@stores/notificationStore";
	import { libraryApi } from "../../lib/features/libraryApi";
	import type { Game } from "../../lib/features/types";
	import { libraryErrorMessage } from "./libraryError";

	let { game, onChanged }: { game: Game; onChanged: (game: Game) => void } = $props();

	let sourceUrl = $state("");
	let linking = $state(false);
	let folderPath = $state("");
	let showFolderField = $state(false);
	let linkingFolder = $state(false);

	async function linkFolder(path: string) {
		const folder = path.trim();
		if (!folder) {
			notifications.error("Enter a folder path.");
			return;
		}
		linkingFolder = true;
		try {
			onChanged(await libraryApi.linkFolder(game.id, folder));
			notifications.success("Linked folder.");
			folderPath = "";
			showFolderField = false;
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
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
		try {
			const picked = await shell.pickFolder();
			if (picked) await linkFolder(picked);
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		}
	}

	async function linkSource() {
		const url = sourceUrl.trim();
		if (!url) return;
		linking = true;
		try {
			onChanged(await libraryApi.linkSource(game.id, url));
			notifications.success("Linked to the source page.");
			sourceUrl = "";
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
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
		<Button size="sm" disabled={linkingFolder} onclick={pickFolder}>Link folder</Button>
		{#if showFolderField}
			<input
				class="input"
				bind:value={folderPath}
				placeholder="/path/to/game"
				aria-label="Game folder"
				onkeydown={(event) => event.key === "Enter" && linkFolder(folderPath)}
			/>
			<Button size="sm" disabled={linkingFolder} onclick={() => linkFolder(folderPath)}>Link</Button>
			<Button size="sm" onclick={() => (showFolderField = false)}>Cancel</Button>
		{/if}
	</div>

	<div class="link">
		<label class="source-field">
			<span>Link a source page</span>
			<input class="input" type="url" bind:value={sourceUrl} placeholder="https://..." />
		</label>
		<Button size="sm" disabled={!sourceUrl.trim() || linking} onclick={linkSource}>Link</Button>
	</div>
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
