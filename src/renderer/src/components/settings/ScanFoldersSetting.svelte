<script lang="ts">
	import * as service from "@lib/settingsService";
	import type { AppSettings } from "@shared";

	let { appSettings, onRefresh } = $props<{
		appSettings: AppSettings;
		onRefresh: () => Promise<void>;
	}>();

	async function handleAddFolder() {
		const folder = await service.browseScanFolder();
		if (folder) {
			await service.addScanFolder(folder);
			await onRefresh();
		}
	}

	async function handleRemoveFolder(folder: string) {
		await service.removeScanFolder(folder);
		await onRefresh();
	}
</script>

<div class="zone-card">
	<div class="zone-header">
		<span class="material-icons">folder_shared</span>
		<h2>Automatic Scan Folders</h2>
	</div>

	<div class="settings-section">
		<p class="desc">
			Add folders that will be monitored. The launcher dynamically scans these folders for game executables (`.exe`) and lists them in your library.
		</p>

		{#if appSettings.ScanFolders && appSettings.ScanFolders.length > 0}
			<div class="folder-list">
				{#each appSettings.ScanFolders as folder}
					<div class="folder-item">
						<span class="material-icons folder-icon">folder</span>
						<span class="folder-path" title={folder}>{folder}</span>
						<button class="remove-btn" onclick={() => handleRemoveFolder(folder)} title="Remove folder">
							<span class="material-icons">delete</span>
						</button>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty-folders glass">
				<span class="material-icons empty-icon">folder_off</span>
				<span class="empty-text">No monitored folders added yet.</span>
			</div>
		{/if}

		<div class="actions-row">
			<button class="btn primary" onclick={handleAddFolder}>
				<span class="material-icons mini-icon">add</span>
				Add Monitored Folder
			</button>
		</div>
	</div>
</div>

<style lang="scss">
	.zone-card {
		background: var(--bg-surface);
		border-radius: var(--radius-lg);
		border: 2px solid rgba(255, 255, 255, 0.05);
		padding: 32px;
		display: flex;
		flex-direction: column;
		gap: 28px;
		transition: border-color var(--transition-fast), box-shadow var(--transition-fast);

		&:hover {
			border-color: rgba(255, 255, 255, 0.08);
			box-shadow: 0 12px 40px rgba(0, 0, 0, 0.25);
		}

		.zone-header {
			display: flex;
			align-items: center;
			gap: 12px;
			border-bottom: 2px solid rgba(255, 255, 255, 0.05);
			padding-bottom: 16px;
			margin-bottom: 4px;

			.material-icons {
				font-size: 24px;
				color: var(--accent-primary);
			}

			h2 {
				margin: 0;
				font-size: 1.15rem;
				font-weight: 800;
				color: var(--text-main);
				text-transform: uppercase;
				letter-spacing: 1px;
			}
		}
	}

	.settings-section {
		display: flex;
		flex-direction: column;
		gap: 12px;

		.desc {
			margin: 0 0 16px 0;
			font-size: 0.9rem;
			font-weight: 600;
			color: var(--text-muted);
			line-height: 1.5;
		}
	}
	.folder-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin-bottom: 16px;
		max-height: 200px;
		overflow-y: auto;
		padding-right: 4px;

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 102, 171, 0.2);
			border-radius: 4px;
		}
	}

	.folder-item {
		display: flex;
		align-items: center;
		padding: 10px 16px;
		background: var(--bg-base);
		border: 1px solid rgba(255, 255, 255, 0.03);
		border-radius: var(--radius-md);
		color: var(--text-main);
		gap: 12px;
		transition: background var(--transition-fast);

		&:hover {
			background: var(--bg-elevated);
		}

		.folder-icon {
			color: var(--accent-secondary);
			font-size: 20px;
		}

		.folder-path {
			font-size: 0.9rem;
			font-weight: 600;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			flex: 1;
		}

		.remove-btn {
			background: none;
			border: none;
			color: var(--text-muted);
			cursor: pointer;
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 4px;
			border-radius: 50%;
			transition: all var(--transition-fast);

			&:hover {
				color: var(--danger, #ff4444);
				background: rgba(255, 68, 68, 0.1);
				transform: scale(1.1);
			}

			.material-icons {
				font-size: 18px;
			}
		}
	}

	.empty-folders {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px;
		background: rgba(255, 255, 255, 0.01);
		border: 2px dashed rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		color: var(--text-muted);
		margin-bottom: 16px;

		.empty-icon {
			font-size: 24px;
			opacity: 0.5;
		}

		.empty-text {
			font-size: 0.9rem;
			font-weight: 600;
		}
	}

	.actions-row {
		display: flex;
		gap: 12px;

		button {
			display: inline-flex;
			align-items: center;
			justify-content: center;
		}
	}

	.mini-icon {
		font-size: 18px;
		margin-right: 6px;
	}

	.desc {
		margin: 0 0 16px 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text-muted);
		line-height: 1.5;
	}
</style>
