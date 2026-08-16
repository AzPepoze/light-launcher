<script lang="ts">
	import FolderHeader from "@components/home/shared/FolderHeader.svelte";
	import GameCardGrid from "@components/home/shared/GameCardGrid.svelte";
	import EmptySectionPlaceholder from "@components/home/shared/EmptySectionPlaceholder.svelte";

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
	export let handleConfigureFolder: (folderPath: string) => void;
	export let handleRightClick: (event: MouseEvent, game: any) => void;
	export let handleQuickLaunch: (game: any) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let toggleGameSelection: (game: any, shiftKey: boolean) => void;
	export let loadIcon: (path: string) => void = () => {};
</script>

<div class="folder-group-container">
	<FolderHeader
		icon="folder"
		name={group.folderName}
		subtitle={group.folderPath}
		count={group.games.length}
		hasMenu={true}
		isMenuOpen={activeFolderMenu === group.folderPath}
		onToggleMenu={(e) => toggleFolderMenu(e, group.folderPath)}
		onRescan={handleRescan}
		onConfigureFolder={() => handleConfigureFolder(group.folderPath)}
		onRemoveFolder={() => handleRemoveFolder(group.folderPath)}
	/>

	{#if group.games.length > 0}
		<GameCardGrid
			games={group.games}
			view={currentView}
			{gameIcons}
			{isGameRunning}
			{sessions}
			{isSelectionMode}
			{selectedPaths}
			{handleRightClick}
			{handleQuickLaunch}
			{handleConfigure}
			{toggleGameSelection}
			{loadIcon}
		/>
	{:else}
		<EmptySectionPlaceholder
			icon="folder_open"
			title="No game executables found inside this folder."
			subtitle="Add Windows game .exe files to this folder, or remove it from watched scan folders. (Scan depth: 2)"
		/>
	{/if}
</div>

<style lang="scss">
	.folder-group-container {
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		padding: 8px 16px 20px 16px;
		margin-bottom: 32px;
		box-shadow: 0 4px 30px rgba(0, 0, 0, 0.2), inset 0 0 20px rgba(255, 255, 255, 0.01);
		transition: border-color var(--transition-fast), box-shadow var(--transition-fast);

		&:hover {
			border-color: rgba(255, 255, 255, 0.08);
			box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3), inset 0 0 20px rgba(255, 255, 255, 0.02);
		}
	}
</style>
