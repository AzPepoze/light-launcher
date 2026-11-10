<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount } from "svelte";
	import type { Game } from "../../lib/features/types";
	import AddGameDialog from "./AddGameDialog.svelte";
	import FilterBar from "./FilterBar.svelte";
	import GameDetail from "./GameDetail.svelte";
	import GameTile from "./GameTile.svelte";
	import IncomingPanel from "./IncomingPanel.svelte";
	import { LibraryState } from "./libraryState.svelte";

	type Tab = "games" | "incoming";

	const library = new LibraryState();
	let tab = $state<Tab>("games");
	let selectedId = $state<number | null>(null);
	let addOpen = $state(false);

	const selected = $derived(selectedId === null ? undefined : library.gameById(selectedId));
	const subtitle = $derived(
		`${library.stats.total} games · ${library.stats.installed} installed` +
			(library.stats.updates ? ` · ${library.stats.updates} updates` : ""),
	);

	onMount(() => {
		library.load();
		return library.subscribe();
	});

	function openGame(game: Game) {
		selectedId = game.id;
	}

	function addedGame(game: Game) {
		library.addGame(game);
		addOpen = false;
		selectedId = game.id;
	}

	function gameRemoved(id: number) {
		library.removeGame(id);
		selectedId = null;
	}
</script>

<PageHeader title="Library" {subtitle} icon="collections_bookmark">
	<div class="tabs" role="tablist" aria-label="Library views">
		<button
			role="tab"
			class="btn sm"
			class:primary={tab === "games"}
			aria-selected={tab === "games"}
			onclick={() => (tab = "games")}
		>
			Games
		</button>
		<button
			role="tab"
			class="btn sm"
			class:primary={tab === "incoming"}
			aria-selected={tab === "incoming"}
			onclick={() => (tab = "incoming")}
		>
			Incoming
		</button>
	</div>
	<button class="btn sm primary" onclick={() => (addOpen = true)}>
		<span class="material-icons" aria-hidden="true">add</span> Add game
	</button>
</PageHeader>

{#if tab === "incoming"}
	<IncomingPanel onChanged={() => library.load()} />
{:else if selected}
	<GameDetail game={selected} {library} onBack={() => (selectedId = null)} onRemoved={gameRemoved} />
{:else if library.loading}
	<p class="status">Loading library...</p>
{:else if library.error}
	<div class="status">
		<p>{library.error}</p>
		<button class="btn sm" onclick={() => library.load()}>Retry</button>
	</div>
{:else if library.stats.total === 0}
	<div class="status">
		<p>No games yet. Add a link or scan your folders.</p>
		<button class="btn sm primary" onclick={() => (addOpen = true)}>Add game</button>
	</div>
{:else}
	<FilterBar {library} />

	{#if library.visible.length === 0}
		<div class="status">
			<p>No games match these filters.</p>
			<button class="btn sm" onclick={() => library.resetFilters()}>Clear filters</button>
		</div>
	{:else}
		<ul class="grid" aria-label="Games">
			{#each library.visible as game (game.id)}
				<li>
					<GameTile {game} onOpen={openGame} />
				</li>
			{/each}
		</ul>
	{/if}
{/if}

{#if addOpen}
	<AddGameDialog onAdded={addedGame} onClose={() => (addOpen = false)} />
{/if}

<style lang="scss">
	.tabs {
		display: flex;
		gap: 6px;
	}

	.grid {
		list-style: none;
		margin: 0;
		padding: 0 0 24px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
		gap: 18px 14px;
	}

	.status {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 48px 16px;
		color: var(--text-muted);
		text-align: center;

		p {
			margin: 0;
		}
	}
</style>
