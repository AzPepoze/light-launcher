<script lang="ts">
	import FolderHeader from "@components/shared/FolderHeader.svelte";
	import type { HomeFolderGroup } from "../state/scanGroups";
	import type { FolderScanStatus } from "../state/homeLoadingStatus";
	import GameCardGrid from "./GameCardGrid.svelte";
	import EmptySectionPlaceholder from "./EmptySectionPlaceholder.svelte";
	import HomeSkeleton from "../HomeSkeleton.svelte";

	export let group: HomeFolderGroup;
	/** Set only while this folder is the one being scanned. */
	export let scanStatus: FolderScanStatus | null = null;
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
	export let handleQuickLaunch: (game: any, showLogs?: boolean) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let handleChooseLauncher: (game: any, launcherPath: string | null) => void = () => {};
	export let toggleGameSelection: (game: any, shiftKey: boolean, ctrlKey?: boolean) => void;
	export let loadIcon: (path: string, customIconPath?: string | null, hasCover?: boolean) => void = () => {};
	export let isStuck: boolean = false;
</script>

<div class="folder-group-container">
	<FolderHeader
		icon="folder"
		name={group.folderName}
		subtitle={group.folderPath}
		count={group.loading ? "..." : group.games.length}
		scanText={scanStatus?.text ?? null}
		scanValue={scanStatus?.percent ?? null}
		hasMenu={true}
		stuckBleed={17}
		stuckKey={group.folderPath}
		{isStuck}
		isMenuOpen={activeFolderMenu === group.folderPath}
		onToggleMenu={(e) => toggleFolderMenu(e, group.folderPath)}
		onCloseMenu={() => activeFolderMenu = null}
		onRescan={handleRescan}
		onRemoveFolder={() => handleRemoveFolder(group.folderPath)}
	/>

	{#if group.loading}
		<HomeSkeleton view={currentView} />
	{:else if group.games.length > 0}
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
			{handleChooseLauncher}
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
		overflow: visible;

		&:hover {
			border-color: rgba(255, 255, 255, 0.08);
			box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3), inset 0 0 20px rgba(255, 255, 255, 0.02);
		}
	}
</style>
