<script lang="ts">
	export let filteredGames: any[] = [];
	export let scannedFolderGroups: any[] = [];
	export let selectedGroupKey = "no-folder";
</script>

<div class="sidebar-panel">
	<!-- No Folder Navigation Item -->
	<button 
		class="sidebar-item" 
		class:active={selectedGroupKey === "no-folder"}
		on:click={() => selectedGroupKey = "no-folder"}
	>
		<span class="material-icons item-icon">sports_esports</span>
		<div class="item-meta">
			<span class="item-name">No Folder</span>
			<span class="item-path">Manually registered profiles</span>
		</div>
		<span class="item-badge">{filteredGames.length}</span>
	</button>

	<!-- Monitored Folder Navigation Items -->
	{#each scannedFolderGroups as group}
		<button 
			class="sidebar-item" 
			class:active={selectedGroupKey === group.folderPath}
			on:click={() => selectedGroupKey = group.folderPath}
		>
			<span class="material-icons item-icon">folder</span>
			<div class="item-meta">
				<span class="item-name">{group.folderName}</span>
				<span class="item-path" title={group.folderPath}>{group.folderPath}</span>
			</div>
			<span class="item-badge">{group.games.length}</span>
		</button>
	{/each}
</div>

<style lang="scss">
	.sidebar-panel {
		width: 280px;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
		overflow-y: auto;
		padding-right: 8px;
	}

	.sidebar-item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 16px;
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		color: var(--text-muted);
		cursor: pointer;
		text-align: left;
		transition: all var(--transition-fast);
		width: 100%;

		.item-icon {
			font-size: 20px;
			color: var(--text-dim);
			transition: color var(--transition-fast);
		}

		.item-meta {
			flex: 1;
			min-width: 0;
			display: flex;
			flex-direction: column;
			gap: 2px;
		}

		.item-name {
			font-size: 0.9rem;
			font-weight: 700;
			color: var(--text-main);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.item-path {
			font-size: 0.7rem;
			color: var(--text-dim);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.item-badge {
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: var(--radius-sm);
			padding: 2px 6px;
			font-size: 0.7rem;
			font-weight: 700;
			color: var(--text-muted);
		}

		&:hover {
			background: rgba(255, 255, 255, 0.05);
			border-color: rgba(255, 255, 255, 0.1);
			color: var(--text-main);

			.item-icon {
				color: var(--text-muted);
			}
		}

		&.active {
			background: rgba(255, 255, 255, 0.07);
			border-color: var(--accent-primary);
			color: var(--text-main);

			.item-icon {
				color: var(--text-main);
			}

			.item-badge {
				background: var(--accent-primary);
				color: var(--bg-base);
				border-color: var(--accent-primary);
			}
		}
	}
</style>
