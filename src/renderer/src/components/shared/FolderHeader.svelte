<script lang="ts">
	import AppMenu from "@components/shared/AppMenu.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import { tooltip } from "@components/shared/tooltip";
	import MenuTrigger from "@components/shared/MenuTrigger.svelte";
	import ProgressBar from "@components/shared/ProgressBar.svelte";
	import { GAME_FOLDERS_SECTION } from "@lib/features/settingsSection";
	import { navigateToSettingsSection } from "@stores/navigationStore";

	export let icon: string = "folder";
	export let name: string = "";
	export let subtitle: string = "";
	export let count: number | string = 0;
	/** Scan line shown under the title. Null hides it. */
	export let scanText: string | null = null;
	/** Share done, from 0 to 1. Null shows a sliding bar. */
	export let scanValue: number | null = null;
	export let hasMenu: boolean = false;
	export let isMenuOpen: boolean = false;
	export let onToggleMenu: (event: MouseEvent) => void = () => {};
	export let onCloseMenu: () => void = () => {};
	export let onRescan: () => void = () => {};
	export let onRemoveFolder: () => void = () => {};
	export let stuckBleed = 0;
	export let isStuck: boolean = false;
	export let stuckKey: string = "";
</script>

<div
	class="folder-group-header"
	class:is-stuck={isStuck}
	class:is-scanning={scanText !== null}
	style:--stuck-bleed="{stuckBleed}px"
	data-sticky-header=""
	data-stuck-key={stuckKey}
>
	<div class="folder-title" use:tooltip={subtitle || name}>
		<Icon name={icon} size={26} class="folder-icon-main" />
		<div class="folder-metadata">
			<span class="folder-name">{name}</span>
			{#if subtitle}
				<span class="folder-path">{subtitle}</span>
			{/if}
			{#if scanText !== null}
				<div class="folder-scan">
					<span class="folder-scan-text">{scanText}</span>
					<div class="folder-scan-bar">
						<ProgressBar value={scanValue} label={scanText} active height={4} />
					</div>
				</div>
			{/if}
		</div>
		<span class="badge">{count}</span>
	</div>

	{#if hasMenu}
		<div class="folder-actions-wrapper">
			<MenuTrigger
				active={isMenuOpen}
				label="Folder Options"
				on:click={(event) => onToggleMenu(event.detail)}
			/>
			<AppMenu visible={isMenuOpen} mode="anchor" onClose={onCloseMenu}>
				<button class="menu-item" on:click={onRescan}>
					<Icon name="refresh" size={18} />
					<span>Rescan Folder</span>
				</button>
				<button class="menu-item" on:click={() => navigateToSettingsSection(GAME_FOLDERS_SECTION)}>
					<Icon name="folder_managed" size={18} />
					<span>Game folders</span>
				</button>
				<button class="menu-item danger" on:click={onRemoveFolder}>
					<Icon name="delete_outline" size={18} />
					<span>Remove Watch</span>
				</button>
			</AppMenu>
		</div>
	{/if}
</div>

<style lang="scss">
	.folder-group-header {
		position: sticky;
		top: 0;
		z-index: 7;
		display: flex;
		justify-content: space-between;
		align-items: center;
		box-sizing: border-box;
		height: 64px;
		margin: 12px;
		padding: 12px 16px;
		border: 1px solid transparent;
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);
		background: transparent;
		border-radius: 0;
		transition: background var(--transition-fast), border-color var(--transition-fast), border-radius var(--transition-fast);

		&.is-scanning {
			height: auto;
			min-height: 64px;
		}

		&.is-stuck {
			background: var(--bg-base);
			border: 1px solid rgba(255, 255, 255, 0.06);
			border-radius: var(--radius-md);
			margin-left: calc(12px - var(--stuck-bleed, 0px));
			margin-right: calc(12px - var(--stuck-bleed, 0px));
		}
	}

	.folder-title {
		display: flex;
		align-items: center;
		gap: 14px;
		min-width: 0;
		flex: 1;
	}

	:global(.folder-icon-main) {
		font-size: 26px;
		color: var(--accent-secondary);
		filter: drop-shadow(0 0 8px rgba(255, 102, 171, 0.3));
	}

	.folder-metadata {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	.folder-name {
		font-size: 1.15rem;
		line-height: 22px;
		font-weight: 800;
		color: var(--text-main);
		letter-spacing: -0.3px;
	}

	.folder-path {
		font-size: 0.75rem;
		line-height: 14px;
		color: var(--text-muted);
		opacity: 0.6;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 400px;
	}

	.folder-scan {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-top: 4px;
		min-width: 0;
	}

	.folder-scan-text {
		flex: 0 1 auto;
		font-size: 0.75rem;
		line-height: 14px;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.folder-scan-bar {
		flex: 0 1 160px;
		min-width: 60px;
	}

	.badge {
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-sm);
		padding: 2px 8px;
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--text-muted);
	}

	.folder-actions-wrapper {
		position: relative;
	}
</style>
