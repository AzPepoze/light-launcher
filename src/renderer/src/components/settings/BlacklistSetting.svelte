<script lang="ts">
	import * as service from "@lib/settingsService";
	import type { AppSettings } from "@shared";

	let { appSettings, onRefresh } = $props<{
		appSettings: AppSettings;
		onRefresh: () => Promise<void>;
	}>();

	async function handleRestoreGame(path: string) {
		await service.unblacklistGame(path);
		await onRefresh();
	}

	function getBaseName(path: string): string {
		return path.split("/").pop()?.replace(".exe", "") || path;
	}
</script>

<div class="zone-card">
	<div class="zone-header">
		<span class="material-icons">block</span>
		<h2>Hidden / Blacklisted Games</h2>
	</div>

	<div class="settings-section">
		<p class="desc">
			Manage games that you have hidden/blacklisted from dynamic scans. Restoring them will allow them to appear in your library again.
		</p>

		{#if appSettings.Blacklist && appSettings.Blacklist.length > 0}
			<div class="blacklist-list">
				{#each appSettings.Blacklist as path}
					<div class="blacklist-item">
						<div class="game-info-col">
							<span class="game-name">{getBaseName(path)}</span>
							<span class="game-path" title={path}>{path}</span>
						</div>
						<button class="restore-btn btn secondary" onclick={() => handleRestoreGame(path)} title="Restore game">
							<span class="material-icons mini-icon">settings_backup_restore</span>
							Restore
						</button>
					</div>
				{/each}
			</div>
		{:else}
			<div class="empty-blacklist glass">
				<span class="material-icons empty-icon">done_all</span>
				<span class="empty-text">No games are currently hidden.</span>
			</div>
		{/if}
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
	.blacklist-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		max-height: 250px;
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

	.blacklist-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 18px;
		background: var(--bg-base);
		border: 1px solid rgba(255, 255, 255, 0.03);
		border-radius: var(--radius-md);
		gap: 16px;
		transition: background var(--transition-fast);

		&:hover {
			background: var(--bg-elevated);
		}

		.game-info-col {
			display: flex;
			flex-direction: column;
			gap: 4px;
			min-width: 0;
			flex: 1;
		}

		.game-name {
			font-size: 0.95rem;
			font-weight: 800;
			color: var(--text-main);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.game-path {
			font-size: 0.8rem;
			font-weight: 600;
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.restore-btn {
			padding: 8px 16px;
			font-size: 0.85rem;
			display: inline-flex;
			align-items: center;
			gap: 6px;
			flex-shrink: 0;
		}
	}

	.empty-blacklist {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px;
		background: rgba(255, 255, 255, 0.01);
		border: 2px dashed rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		color: var(--text-muted);

		.empty-icon {
			font-size: 24px;
			color: var(--success, #44ffaa);
			opacity: 0.8;
		}

		.empty-text {
			font-size: 0.9rem;
			font-weight: 600;
		}
	}

	.mini-icon {
		font-size: 16px;
	}

	.desc {
		margin: 0 0 16px 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text-muted);
		line-height: 1.5;
	}
</style>
