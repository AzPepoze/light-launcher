<script lang="ts">
	import GridCard from "@components/home/gamecards/GridCard.svelte";
	import ListGridCard from "@components/home/gamecards/ListGridCard.svelte";
	import PerspectiveCard from "@components/home/gamecards/PerspectiveCard.svelte";

	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let view: "grid" | "list-grid" | "perspective" = "grid";
	export let active: boolean = false; // For perspective view
	export let isSelectionMode: boolean = false;
	export let isSelected: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};
	export let onSelect: (game: any) => void = () => {};
</script>

{#if view === "grid"}
	<GridCard
		{game}
		{icon}
		{isRunning}
		{isSelectionMode}
		{isSelected}
		{onLaunch}
		{onConfigure}
		{onSelect}
	/>
{:else if view === "list-grid"}
	<ListGridCard
		{game}
		{icon}
		{isRunning}
		{isSelectionMode}
		{isSelected}
		{onLaunch}
		{onConfigure}
		{onSelect}
	/>
{:else}
	<PerspectiveCard
		{game}
		{icon}
		{isRunning}
		{active}
		{onLaunch}
		{onConfigure}
	/>
{/if}
