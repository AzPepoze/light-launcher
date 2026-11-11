<script lang="ts">
	import { onMount } from "svelte";
	import * as service from "@lib/settingsService";
	import { GetAutoScannedGames, IsDir } from "@lib/api";
	import {
		folderInputError,
		folderKey,
		gamesFoundText,
		gamesInFolder
	} from "@lib/features/scanFolders";
	import { GAME_FOLDERS_SECTION } from "@lib/features/settingsSection";
	import { createLogger } from "@lib/logger";
	import type { AppSettings, ScanFolderConfig, ScannedFolderGroup } from "@shared";
	import ScanFolderRow from "./ScanFolderRow.svelte";

	const log = createLogger("GameFolders");

	let { appSettings, onRefresh } = $props<{
		appSettings: AppSettings;
		onRefresh: () => Promise<void>;
	}>();

	let groups = $state<ScannedFolderGroup[]>([]);
	let rescanNotes = $state<Record<string, string>>({});
	let rescanningFolder = $state<string | null>(null);
	let scanError = $state("");
	let pathInput = $state("");
	let addError = $state("");
	let adding = $state(false);

	const watched = $derived(appSettings.ScanFolders ?? []);
	const configs = $derived(appSettings.ScanFolderConfigs ?? []);
	const canPickFolder = typeof window !== "undefined" && typeof window.llShell?.pickFolder === "function";

	onMount(() => {
		loadGroups(false);
	});

	/** Loads the scan results. Returns false when the scan failed. */
	async function loadGroups(force: boolean): Promise<boolean> {
		try {
			groups = await GetAutoScannedGames(force);
			scanError = "";
			return true;
		} catch (err) {
			log.error("Failed to scan game folders", err);
			scanError = "Could not scan the game folders. Try again.";
			return false;
		}
	}

	function configFor(folder: string): ScanFolderConfig {
		const key = folderKey(folder);
		const config = configs.find((item) => folderKey(item.Path) === key);
		return config ?? { Path: folder, Depth: 2, ExcludeNames: [] };
	}

	async function rescanFolder(folder: string) {
		rescanningFolder = folder;
		try {
			if (!(await loadGroups(true))) return;
			const count = gamesInFolder(groups, folder) ?? 0;
			rescanNotes = { ...rescanNotes, [folder]: gamesFoundText(count) };
		} finally {
			rescanningFolder = null;
		}
	}

	async function addFolder(path: string) {
		const folder = path.trim();
		addError = folderInputError(folder, watched) ?? "";
		if (addError) return;

		adding = true;
		try {
			if (!(await IsDir(folder))) {
				addError = "That folder does not exist.";
				return;
			}
			await service.addScanFolder(folder);
			pathInput = "";
			await onRefresh();
			await loadGroups(false);
		} catch (err) {
			log.error("Failed to add game folder", err);
			addError = "Could not add this folder. Try again.";
		} finally {
			adding = false;
		}
	}

	async function pickAndAddFolder() {
		try {
			const picked = await window.llShell?.pickFolder?.();
			if (picked) await addFolder(picked);
		} catch (err) {
			log.error("Folder picker failed", err);
			addError = "Could not open the folder picker. Type the path instead.";
		}
	}

	async function removeFolder(folder: string) {
		try {
			await service.removeScanFolder(folder);
			await onRefresh();
			await loadGroups(false);
		} catch (err) {
			log.error("Failed to remove game folder", err);
			scanError = "Could not remove this folder. Try again.";
		}
	}
</script>

<div class="zone-card" id={GAME_FOLDERS_SECTION}>
	<div class="zone-header">
		<span class="material-icons" aria-hidden="true">folder_shared</span>
		<h2>Game folders</h2>
	</div>

	<div class="settings-section">
		<p class="desc">
			The launcher looks for game programs (<code>.exe</code> files) in these folders. Each folder
			shows how many games it found.
		</p>

		{#if scanError}
			<p class="error" role="alert">{scanError}</p>
		{/if}

		{#if watched.length > 0}
			<ul class="folder-list">
				{#each watched as folder (folder)}
					{@const config = configFor(folder)}
					<ScanFolderRow
						{folder}
						depth={config.Depth}
						excludeNames={config.ExcludeNames ?? []}
						gamesFound={gamesInFolder(groups, folder)}
						rescanNote={rescanNotes[folder] ?? ""}
						rescanning={rescanningFolder === folder}
						onRescan={() => rescanFolder(folder)}
						onRemove={() => removeFolder(folder)}
						onSaved={async () => {
							await onRefresh();
							await loadGroups(false);
						}}
					/>
				{/each}
			</ul>
		{:else}
			<div class="empty-folders">
				<span class="material-icons empty-icon" aria-hidden="true">folder_off</span>
				<span>No game folders yet. Add a folder to find your games.</span>
			</div>
		{/if}

		<div class="add-folder">
			<h3>Add a folder</h3>
			{#if canPickFolder}
				<div class="add-row">
					<button class="btn primary" onclick={pickAndAddFolder} disabled={adding}>
						<span class="material-icons mini-icon" aria-hidden="true">add</span>
						{adding ? "Adding..." : "Choose folder"}
					</button>
				</div>
			{:else}
				<label class="field-label" for="new-folder-path">Folder path</label>
				<div class="add-row">
					<input
						id="new-folder-path"
						type="text"
						bind:value={pathInput}
						placeholder="Full path to the folder"
						aria-invalid={addError !== ""}
						aria-describedby={addError ? "add-folder-error" : undefined}
						onkeydown={(event) => {
							if (event.key === "Enter") addFolder(pathInput);
						}}
					/>
					<button class="btn primary" onclick={() => addFolder(pathInput)} disabled={adding}>
						<span class="material-icons mini-icon" aria-hidden="true">add</span>
						{adding ? "Adding..." : "Add folder"}
					</button>
				</div>
			{/if}
			{#if addError}
				<p class="error" id="add-folder-error" role="alert">{addError}</p>
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	.zone-card {
		scroll-margin-top: 24px;
	}

	.zone-header h2 {
		font-size: 1.25rem;
		font-weight: 600;
		text-transform: none;
		letter-spacing: normal;
	}

	.settings-section {
		display: flex;
		flex-direction: column;
		gap: 16px;

		.desc {
			margin: 0;
			font-size: 1rem;
			color: var(--text-muted);
			line-height: 1.5;
		}

		code {
			font-size: 0.875rem;
		}
	}

	.folder-list {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin: 0;
		padding: 0;
	}

	.empty-folders {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px;
		border: 2px dashed rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-md);
		color: var(--text-muted);
		font-size: 1rem;

		.empty-icon {
			font-size: 24px;
		}
	}

	.add-folder {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding-top: 8px;

		h3 {
			margin: 0;
			font-size: 1rem;
			font-weight: 600;
			color: var(--text-main);
		}

		.field-label {
			font-size: 0.875rem;
			font-weight: 600;
			color: var(--text-main);
		}
	}

	.add-row {
		display: flex;
		gap: 12px;
		align-items: center;

		input {
			flex: 1;
			min-width: 0;
			font: inherit;
			font-size: 1rem;
			color: var(--text-main);
			background: var(--bg-input);
			border: 1px solid rgba(255, 255, 255, 0.1);
			border-radius: var(--radius-md);
			padding: 8px 12px;

			&[aria-invalid="true"] {
				border-color: var(--danger);
			}

			&:focus-visible {
				outline: 2px solid var(--accent-primary);
				outline-offset: 2px;
			}
		}

		button {
			display: inline-flex;
			align-items: center;
			justify-content: center;
		}
	}

	.btn:focus-visible {
		outline: 2px solid var(--accent-primary);
		outline-offset: 2px;
	}

	.error {
		margin: 0;
		font-size: 0.875rem;
		color: var(--danger);
	}

	.mini-icon {
		font-size: 18px;
		margin-right: 6px;
	}
</style>
