<script lang="ts">
	import { onMount } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import { gamesInFolder } from "@lib/features/scanFolders";
	import { GAME_FOLDERS_SECTION } from "@lib/features/settingsSection";
	import Notice from "@components/shared/Notice.svelte";
	import ScanFolderRow from "@components/settings/ScanFolderRow.svelte";
	import { pluginFolderMatches, pluginFolderWarning } from "./pluginFolders";
	import type { SettingsState } from "./settingsState.svelte";

	let { settings } = $props<{ settings: SettingsState }>();

	onMount(() => {
		// The rows warn when a folder is one a plugin uses. UpdatesSettings usually loads the same
		// overview first; this covers a page that renders this section on its own.
		if (settings.sources.length === 0) void settings.loadSources();
	});
</script>

<div class="zone-card" id={GAME_FOLDERS_SECTION}>
	<div class="zone-header">
		<Icon name="folder_shared" size={24} class="zone-icon" />
		<h2>Game folders</h2>
	</div>

	<div class="settings-section">
		<p class="desc">
			The launcher looks for game programs (<code>.exe</code> files) in these folders. Each folder
			shows how many games it found.
		</p>

		<Notice tone="info" icon="info">
			Games installed by plugins (F95, itch.io) are already in your Library. You do not need to add
			their download folder here.
		</Notice>

		{#if settings.watched.length > 0}
			<ul class="folder-list">
				{#each settings.watched as folder (folder)}
					{@const config = settings.configFor(folder)}
					{@const pluginWarnings = [
						...new Set(pluginFolderMatches(folder, settings.sources).map(pluginFolderWarning))
					]}
					<ScanFolderRow
						{folder}
						depth={config.Depth}
						excludeNames={config.ExcludeNames ?? []}
						gamesFound={gamesInFolder(settings.groups, folder)}
						{pluginWarnings}
						rescanNote={settings.rescanNotes[folder] ?? ""}
						rescanning={settings.rescanningFolder === folder}
						onRescan={() => settings.rescanFolder(folder)}
						onRemove={() => settings.removeFolder(folder)}
						onSaved={() => settings.refresh()}
					/>
				{/each}
			</ul>
		{:else}
			<div class="empty-folders">
				<Icon name="folder_off" size={24} class="empty-icon" />
				<span>No game folders yet. Add a folder to find your games.</span>
			</div>
		{/if}

		<div class="add-folder">
			<h3>Add a folder</h3>
			{#if settings.canPickFolder}
				<div class="add-row">
					<Button
						variant="primary"
						icon="add"
						loading={settings.adding}
						onclick={settings.pickAndAddFolder}
					>
						{settings.adding ? "Adding..." : "Choose folder"}
					</Button>
				</div>
			{:else}
				<label class="field-label" for="new-folder-path">Folder path</label>
				<div class="add-row">
					<input
						id="new-folder-path"
						type="text"
						bind:value={settings.pathInput}
						placeholder="Full path to the folder"
						aria-invalid={settings.addError !== ""}
						aria-describedby={settings.addError ? "add-folder-error" : undefined}
						onkeydown={(event) => {
							if (event.key === "Enter") settings.addFolder(settings.pathInput);
						}}
					/>
					<Button
						variant="primary"
						icon="add"
						loading={settings.adding}
						onclick={() => settings.addFolder(settings.pathInput)}
					>
						{settings.adding ? "Adding..." : "Add folder"}
					</Button>
				</div>
			{/if}
			{#if settings.addError}
				<p class="error" id="add-folder-error" role="alert">{settings.addError}</p>
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

	:global(.zone-icon) {
		color: var(--accent-primary);
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

		:global(.empty-icon) {
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
	}

	.error {
		margin: 0;
		font-size: 0.875rem;
		color: var(--danger);
	}
</style>
