<script lang="ts">
	import FolderHeader from "@components/home/shared/FolderHeader.svelte";
	import GameCardGrid from "@components/home/shared/GameCardGrid.svelte";
	import EmptySectionPlaceholder from "@components/home/shared/EmptySectionPlaceholder.svelte";

	export let filteredGames: any[] = [];
	export let gameIcons: Record<string, string> = {};
	export let isGameRunning: (game: any, sessions: any[]) => boolean;
	export let sessions: any[] = [];
	export let isSelectionMode: boolean = false;
	export let selectedPaths = new Set<string>();
	export let handleRightClick: (event: MouseEvent, game: any) => void;
	export let handleQuickLaunch: (game: any) => Promise<void>;
	export let handleConfigure: (game: any) => void;
	export let toggleGameSelection: (game: any, shiftKey: boolean) => void;
	export let loadIcon: (path: string) => void = () => {};
</script>

<FolderHeader
	icon="sports_esports"
	name="No Folder"
	subtitle="Manually registered custom profiles"
	count={filteredGames.length}
/>

{#if filteredGames.length > 0}
	<GameCardGrid
		games={filteredGames}
		view="grid"
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
		icon="sports_esports"
		title="No custom profiles in this section."
		subtitle="Add specific games or configure monitored watched directories to list games."
	/>
{/if}
