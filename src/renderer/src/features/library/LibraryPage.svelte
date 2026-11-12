<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { onMount, tick } from "svelte";
	import type { Game } from "../../lib/features/types";
	import AddGameDialog from "./AddGameDialog.svelte";
	import FilterBar from "./FilterBar.svelte";
	import GameDetail from "./GameDetail.svelte";
	import IncomingPanel from "./IncomingPanel.svelte";
	import LibraryGrid from "./LibraryGrid.svelte";
	import PluginBrowse from "./PluginBrowse.svelte";
	import SectionHeader from "./SectionHeader.svelte";
	import SectionList from "./SectionList.svelte";
	import { LibraryState } from "./libraryState.svelte";
	import { REVEAL_MAX_TILES, revealDoneMs } from "./libraryMotion";
	import { LOCAL_SOURCE, sourceLabel } from "./librarySections";
	import { tileShapeFor } from "./sourceTileShape";

	type Tab = "games" | "incoming";

	const library = new LibraryState();
	let tab = $state<Tab>("games");
	let selected = $state<Game | null>(null);
	let addOpen = $state(false);
	let addUrl = $state("");
	let revealed = $state(false);

	const current = $derived(library.sections.find((item) => item.source === library.section));
	const isPlugin = $derived(library.section !== LOCAL_SOURCE);
	const canBrowse = $derived((current?.sidebar.length ?? 0) > 0);
	const subtitle = $derived(
		`${library.overview.stats.total} games · ${library.overview.stats.installed} installed` +
			(library.overview.stats.updates ? ` · ${library.overview.stats.updates} updates` : "")
	);
	const hasTiles = $derived((library.total ?? 0) > 0);
	const revealCount = $derived(Math.min(library.total ?? 0, REVEAL_MAX_TILES));

	onMount(() => {
		void library.start();
		return library.subscribe();
	});

	// Ends the first reveal once its last tile has finished, so later renders never replay it.
	$effect(() => {
		if (!hasTiles || revealed) return;
		const timer = setTimeout(() => (revealed = true), revealDoneMs(revealCount));
		return () => clearTimeout(timer);
	});

	function openGame(game: Game) {
		selected = game;
	}

	async function closeGame() {
		const id = selected?.id;
		selected = null;
		await tick();
		if (id !== undefined) {
			document.querySelector<HTMLElement>(`[data-game-id="${id}"]`)?.focus();
		}
	}

	function gameChanged(game: Game) {
		selected = game;
		library.replaceGame(game);
	}

	function gameRemoved() {
		selected = null;
		library.scheduleRefresh();
	}

	function addedGame(game: Game) {
		addUrl = "";
		addOpen = false;
		selected = game;
		library.scheduleRefresh();
	}

	function syncNow() {
		void library.startSync(library.section);
	}

	/** A browse result is not saved yet, so it opens the add dialog with its link. */
	function openBrowseGame(game: Game) {
		addUrl = game.url;
		addOpen = true;
	}

	function closeAdd() {
		addOpen = false;
		addUrl = "";
	}
</script>

<div class="library">
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
		<IncomingPanel onChanged={() => library.refresh()} />
	{:else if selected}
		<GameDetail
			game={selected}
			sourceName={sourceLabel(library.overview.sources, selected.source)}
			onBack={closeGame}
			onRemoved={gameRemoved}
			onChanged={gameChanged}
		/>
	{:else if !library.overviewLoaded}
		<p class="status">Loading library...</p>
	{:else if library.overviewError}
		<div class="status">
			<p>{library.overviewError}</p>
			<button class="btn sm" onclick={() => library.refresh()}>Retry</button>
		</div>
	{:else if current}
		<div class="layout">
			<SectionList
				sections={library.sections}
				selected={library.section}
				onSelect={(source) => library.selectSection(source)}
			/>

			<div class="main">
				<SectionHeader
					label={current.label}
					count={current.count}
					{isPlugin}
					syncing={library.isSyncing(library.section)}
					progress={library.syncs[library.section]}
					onSync={syncNow}
				/>

				{#if canBrowse && current}
					{#key library.section}
						<PluginBrowse source={current} onOpenGame={openBrowseGame} />
					{/key}
				{:else}
					<FilterBar {library} />

					{#if library.error}
						<div class="status">
							<p>{library.error}</p>
							<button class="btn sm" onclick={() => library.refresh()}>Retry</button>
						</div>
					{:else if library.total === 0 && library.filtersActive}
						<div class="status">
							<p>No games match these filters.</p>
							<button class="btn sm" onclick={() => library.resetFilters()}>Clear filters</button>
						</div>
					{:else if library.total === 0 && isPlugin}
						<div class="status">
							<p>No games yet. Install the plugin, sign in, then press Sync now.</p>
						</div>
					{:else if library.total === 0}
						<div class="status">
							<p>No games yet. Add a link or scan your folders.</p>
							<button class="btn sm primary" onclick={() => (addOpen = true)}>Add game</button>
						</div>
					{:else}
						{#key library.section}
							<LibraryGrid {library} shape={tileShapeFor(current.tile)} {revealed} onOpen={openGame} />
						{/key}
					{/if}
				{/if}
			</div>
		</div>
	{/if}

	{#if addOpen}
		<AddGameDialog onAdded={addedGame} onClose={closeAdd} initialUrl={addUrl} />
	{/if}
</div>

<style lang="scss">
	.library {
		font-size: 1rem;
		line-height: 1.5;

		/* Shared buttons read at the same size as the rest of the page. */
		:global(.btn.sm) {
			font-size: 0.875rem;
		}

		:global(.btn:focus-visible) {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	.tabs {
		display: flex;
		gap: 6px;
	}

	.layout {
		display: grid;
		grid-template-columns: 220px minmax(0, 1fr);
		align-items: start;
		gap: 32px;
	}

	.main {
		min-width: 0;
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

	@media (max-width: 760px) {
		.layout {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
