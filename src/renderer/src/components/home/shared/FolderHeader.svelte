<script lang="ts">
	export let icon: string = "folder";
	export let name: string = "";
	export let subtitle: string = "";
	export let count: number = 0;
	export let hasMenu: boolean = false;
	export let isMenuOpen: boolean = false;
	export let onToggleMenu: (event: MouseEvent) => void = () => {};
	export let onRescan: () => void = () => {};
	export let onConfigureFolder: () => void = () => {};
	export let onRemoveFolder: () => void = () => {};
</script>

<div class="folder-group-header">
	<div class="folder-title" title={subtitle || name}>
		<span class="material-icons folder-icon-main">{icon}</span>
		<div class="folder-metadata">
			<span class="folder-name">{name}</span>
			{#if subtitle}
				<span class="folder-path">{subtitle}</span>
			{/if}
		</div>
		<span class="badge">{count}</span>
	</div>

	{#if hasMenu}
		<div class="folder-actions-wrapper">
			<button
				class="folder-menu-trigger"
				title="Folder Options"
				on:click|stopPropagation={onToggleMenu}
			>
				<span class="material-icons">more_vert</span>
			</button>
			{#if isMenuOpen}
				<div class="folder-dropdown-menu">
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div class="dropdown-item" on:click={onRescan}>
						<span class="material-icons">refresh</span>
						<span>Rescan Folder</span>
					</div>
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div class="dropdown-item" on:click={onConfigureFolder}>
						<span class="material-icons">settings</span>
						<span>Folder Settings</span>
					</div>
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div class="dropdown-item danger" on:click={onRemoveFolder}>
						<span class="material-icons">delete_outline</span>
						<span>Remove Watch</span>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="scss">
	.folder-group-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin: 24px 12px 16px 12px;
		padding-bottom: 12px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.folder-title {
		display: flex;
		align-items: center;
		gap: 14px;
		min-width: 0;
		flex: 1;
	}

	.folder-icon-main {
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
		font-weight: 800;
		color: var(--text-main);
		letter-spacing: -0.3px;
	}

	.folder-path {
		font-size: 0.75rem;
		color: var(--text-muted);
		opacity: 0.6;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 400px;
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

	.folder-menu-trigger {
		background: transparent;
		border: none;
		color: var(--text-muted);
		width: 32px;
		height: 32px;
		border-radius: var(--radius-md);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-fast);

		&:hover {
			background: rgba(255, 255, 255, 0.08);
			color: var(--text-main);
		}
	}

	.folder-dropdown-menu {
		position: absolute;
		right: 0;
		top: calc(100% + 4px);
		background: var(--bg-surface);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-md);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
		min-width: 160px;
		padding: 4px;
		z-index: 100;
		backdrop-filter: blur(16px);

		.dropdown-item {
			display: flex;
			align-items: center;
			gap: 10px;
			padding: 8px 12px;
			font-size: 0.85rem;
			font-weight: 600;
			color: var(--text-main);
			border-radius: var(--radius-sm);
			cursor: pointer;
			transition: background var(--transition-fast);

			.material-icons {
				font-size: 18px;
				color: var(--text-muted);
			}

			&:hover {
				background: rgba(255, 255, 255, 0.08);

				.material-icons {
					color: var(--text-main);
				}
			}

			&.danger {
				color: var(--danger);

				.material-icons {
					color: var(--danger);
				}

				&:hover {
					background: rgba(255, 59, 48, 0.12);
				}
			}
		}
	}
</style>
