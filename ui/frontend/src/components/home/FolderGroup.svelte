<script lang="ts">
	import GameCard from "@components/home/GameCard.svelte";

	export let group: any;
	export let currentView: "grid" | "list-grid" | "sidebar-grid";
	export let gameIcons: Record<string, string> = {};
	export let isGameRunning: (game: any, sessionsList: any[]) => boolean;
	export let sessions: any[] = [];
	export let isSelectionMode = false;
	export let selectedPaths = new Set<string>();

	export let activeFolderMenu: string | null = null;
	export let toggleFolderMenu: (event: MouseEvent, folderPath: string) => void;
	export let handleRescan: () => void;
	export let handleRemoveFolder: (folderPath: string) => void;
	export let handleRightClick: (event: MouseEvent, game: any) => void;
	export let handleQuickLaunch: (game: any) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let toggleGameSelection: (game: any, shiftKey: boolean) => void;
</script>

<div class="folder-group-container">
	<div class="folder-group-header">
		<div class="folder-title" title={group.folderPath}>
			<span class="material-icons folder-icon-main">folder</span>
			<div class="folder-metadata">
				<span class="folder-name">{group.folderName}</span>
				<span class="folder-path">{group.folderPath}</span>
			</div>
			<span class="badge">{group.games.length}</span>
		</div>
		<div class="folder-actions-wrapper">
			<button
				class="folder-menu-trigger"
				title="Folder Options"
				on:click|stopPropagation={(e) => toggleFolderMenu(e, group.folderPath)}
			>
				<span class="material-icons">more_vert</span>
			</button>
			{#if activeFolderMenu === group.folderPath}
				<div class="folder-dropdown-menu">
					<div class="dropdown-item" on:click={handleRescan}>
						<span class="material-icons">refresh</span>
						<span>Rescan Folder</span>
					</div>
					<div class="dropdown-item danger" on:click={() => handleRemoveFolder(group.folderPath)}>
						<span class="material-icons">delete_outline</span>
						<span>Remove Watch</span>
					</div>
				</div>
			{/if}
		</div>
	</div>

	{#if group.games.length > 0}
		<div 
			class="games-grid"
			class:grid-view={currentView === "grid" || currentView === "sidebar-grid"}
			class:list-view={currentView === "list-grid"}
		>
			{#each group.games as game}
				<div on:contextmenu|preventDefault|stopPropagation={(e) => handleRightClick(e, game)}>
					<GameCard
						{game}
						icon={gameIcons[game.path || game.config.LauncherPath]}
						isRunning={isGameRunning(game, sessions)}
						{isSelectionMode}
						isSelected={selectedPaths.has(game.path || game.config.LauncherPath)}
						view={currentView === "sidebar-grid" ? "grid" : currentView}
						onLaunch={() => handleQuickLaunch(game)}
						onConfigure={() => handleConfigure(game)}
						onSelect={toggleGameSelection}
					/>
				</div>
			{/each}
		</div>
	{:else}
		<div class="folder-empty-placeholder">
			<span class="material-icons placeholder-icon">folder_open</span>
			<span class="placeholder-text">No game executables found inside this folder.</span>
			<span class="placeholder-subtext">Add Windows game .exe files to this folder, or remove it from watched scan folders. (Scan depth: 2)</span>
		</div>
	{/if}
</div>

<style lang="scss">
	.games-grid {
		display: grid;
		gap: 28px;
		width: 100%;
		padding: 12px;
		padding-bottom: 40px;

		&.grid-view {
			grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		}

		&.list-view {
			grid-template-columns: 1fr;
			gap: 16px;
		}
	}
	.folder-group-container {
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		padding: 24px;
		margin-bottom: 32px;
		box-shadow: 0 4px 30px rgba(0, 0, 0, 0.2), inset 0 0 20px rgba(255, 255, 255, 0.01);
		transition: border-color var(--transition-fast), box-shadow var(--transition-fast);

		&:hover {
			border-color: rgba(255, 255, 255, 0.08);
			box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3), inset 0 0 20px rgba(255, 255, 255, 0.02);
		}
	}

	.folder-group-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
		padding-bottom: 16px;
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
		font-size: 28px;
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
		font-weight: 900;
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
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		transition: all var(--transition-fast);

		&:hover {
			background: rgba(255, 255, 255, 0.05);
			color: var(--text-main);
		}
	}

	.folder-dropdown-menu {
		position: absolute;
		top: 100%;
		right: 0;
		margin-top: 6px;
		background: var(--bg-elevated);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-md);
		padding: 6px;
		min-width: 160px;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
		z-index: 10;
		animation: dropdown-fade-in var(--transition-fast);
	}

	.dropdown-item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 14px;
		border-radius: var(--radius-sm);
		color: var(--text-muted);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-fast);

		.material-icons {
			font-size: 18px;
		}

		&:hover {
			background: rgba(255, 255, 255, 0.05);
			color: var(--text-main);
		}

		&.danger {
			color: var(--danger);

			&:hover {
				background: rgba(255, 59, 48, 0.15);
				color: #ff453a;
			}
		}
	}

	.folder-empty-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		color: var(--text-muted);
		text-align: center;
		background: rgba(0, 0, 0, 0.15);
		border-radius: var(--radius-md);
		border: 2px dashed rgba(255, 255, 255, 0.04);
		margin: 8px 12px;
		width: calc(100% - 24px);

		.placeholder-icon {
			font-size: 36px;
			color: var(--text-muted);
			opacity: 0.3;
			margin-bottom: 12px;
		}

		.placeholder-text {
			font-size: 0.95rem;
			font-weight: 700;
			color: var(--text-muted);
			margin-bottom: 4px;
		}

		.placeholder-subtext {
			font-size: 0.75rem;
			opacity: 0.6;
			max-width: 400px;
			line-height: 1.4;
		}
	}

	@keyframes dropdown-fade-in {
		from {
			opacity: 0;
			transform: translateY(-4px) scale(0.98);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}
</style>
