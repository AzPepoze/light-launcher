<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import Section from "@components/shared/Section.svelte";
	import { tooltip } from "@components/shared/tooltip";
	import type { SettingsState } from "./settingsState.svelte";

	let { settings } = $props<{ settings: SettingsState }>();
	const appSettings = $derived(settings.appSettings);

	function handleRestoreGame(path: string) {
		return settings.restoreGame(path);
	}

	function getBaseName(path: string): string {
		return path.split("/").pop()?.replace(".exe", "") || path;
	}
</script>

<Section
	title="Hidden / Blacklisted Games"
	icon="block"
	help="Games you hid from dynamic scans. Restore one to show it in your Library again."
>
	{#if appSettings.Blacklist && appSettings.Blacklist.length > 0}
		<div class="blacklist-list">
			{#each appSettings.Blacklist as path}
				<div class="blacklist-item">
					<div class="game-info-col">
						<span class="game-name">{getBaseName(path)}</span>
						<span class="game-path" use:tooltip={path}>{path}</span>
					</div>
					<Button
						variant="secondary"
						size="sm"
						icon="settings_backup_restore"
						class="restore-btn"
						onclick={() => handleRestoreGame(path)}
						title="Restore game"
					>
						Restore
					</Button>
				</div>
			{/each}
		</div>
	{:else}
		<div class="empty-blacklist">
			<Icon name="done_all" size={24} class="empty-icon" />
			<span class="empty-text">No games are currently hidden.</span>
		</div>
	{/if}
</Section>

<style lang="scss">
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

		:global(.restore-btn) {
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

		:global(.empty-icon) {
			font-size: 24px;
			color: var(--success, #44ffaa);
			opacity: 0.8;
		}

		.empty-text {
			font-size: 0.9rem;
			font-weight: 600;
		}
	}
</style>
