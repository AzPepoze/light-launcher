<script lang="ts">
	import { onMount, tick } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import Tabs from "@components/shared/Tabs.svelte";
	import type { TabItem } from "@components/shared/tabs";
	import { libraryTab, type LibraryTab } from "@stores/libraryTabStore";
	import type { Game } from "../../lib/features/types";
	import { libraryApi } from "../../lib/features/libraryApi";
	import { loadAppSettings } from "../../lib/settingsService";
	import {
		LIBRARY_CARD_SIZE_DEFAULT,
		blurCoversUntilHover,
		libraryCardSize
	} from "../settings/libraryAppearance";
	import { buildRows, groupRows } from "../downloads/downloadGroups";
	import AddGameDialog from "./AddGameDialog.svelte";
	import GameDetail from "./GameDetail.svelte";
	import IncomingPanel from "./IncomingPanel.svelte";
	import LibraryGrid from "./LibraryGrid.svelte";
	import LibrarySidebar from "./LibrarySidebar.svelte";
	import SectionHeader from "./SectionHeader.svelte";
	import SourceRail from "./SourceRail.svelte";
	import { LibraryState } from "./libraryState.svelte";
	import { REVEAL_MAX_TILES, revealDoneMs } from "./libraryMotion";
	import { LOCAL_SOURCE, sourceLabel } from "./librarySections";
	import { coverRatioOf } from "./coverRatio";
	import { canSyncGames } from "./syncFlag";
	import { closesDetailOn } from "./detailPanel";
	import { DownloadsState } from "../downloads/downloadsState.svelte";
	import DownloadsPage from "../downloads/DownloadsPage.svelte";
	import UpdatesPage from "../updates/UpdatesPage.svelte";
	import { transferKinds } from "./transferState";

	const library = new LibraryState();
	const downloads = new DownloadsState();
	/** Which game is downloading or installing, for the tile corner states. */
	const transfers = $derived(transferKinds(downloads.items, downloads.installs));
	/** Downloads and installs running now, for the Downloads tab badge. */
	const activeDownloads = $derived(groupRows(buildRows(downloads.items, downloads.installs)).active.length);
	/** True while a download or install is moving, which is when speed samples are taken. */
	const transfersActive = $derived(
		buildRows(downloads.items, downloads.installs).some(
			(row) => row.transfer.state === "active"
		)
	);
	let tab = $state<LibraryTab>("games");
	let updateCount = $state(0);
	let selected = $state<Game | null>(null);
	let addOpen = $state(false);
	let addUrl = $state("");
	let revealed = $state(false);
	let cardSize = $state(LIBRARY_CARD_SIZE_DEFAULT);
	let blurCovers = $state(false);
	/** Under 900px the filter sidebar slides in over the grid. */
	let filtersOpen = $state(false);

	const current = $derived(library.current);
	const isPlugin = $derived(library.section !== LOCAL_SOURCE);
	const ratio = $derived(coverRatioOf(current?.cover));
	const subtitle = $derived(
		`${library.overview.stats.total} games · ${library.overview.stats.installed} installed` +
			(library.overview.stats.updates ? ` · ${library.overview.stats.updates} updates` : "")
	);
	const topTabs = $derived<TabItem[]>([
		{ id: "games", label: "Games", icon: "videogame_asset" },
		{ id: "updates", label: "Updates", icon: "system_update_alt", badge: updateCount },
		{ id: "downloads", label: "Downloads", icon: "downloading", badge: activeDownloads },
		{ id: "incoming", label: "Incoming", icon: "move_to_inbox" }
	]);
	const gameActions = $derived(
		(library.overview.sources.find((source) => source.source === selected?.source)?.actions ?? []).filter(
			(action) => action.scope === "game"
		)
	);
	const hasTiles = $derived((library.total ?? 0) > 0);
	const revealCount = $derived(Math.min(library.total ?? 0, REVEAL_MAX_TILES));

	onMount(() => {
		const stopTab = libraryTab.subscribe((next) => (tab = next));
		void library.start();
		void loadAppearance();
		void loadUpdateCount();
		void downloads.load();
		const stopDownloads = downloads.subscribe();
		const stopLibrary = library.subscribe();
		return () => {
			stopTab();
			downloads.stopSampling();
			stopDownloads();
			stopLibrary();
		};
	});

	/* The Library page owns the speed samples, so they keep coming while another tab is open. */
	$effect(() => {
		downloads.setSamplingActive(transfersActive);
	});

	// Ends the first reveal once its last tile has finished, so later renders never replay it.
	$effect(() => {
		if (!hasTiles || revealed) return;
		const timer = setTimeout(() => (revealed = true), revealDoneMs(revealCount));
		return () => clearTimeout(timer);
	});

	async function loadAppearance() {
		const settings = await loadAppSettings();
		if (!settings) return;
		cardSize = libraryCardSize(settings);
		blurCovers = blurCoversUntilHover(settings);
	}

	/** Reads the unseen update count for the Updates tab badge. Starts no check. */
	async function loadUpdateCount() {
		try {
			updateCount = (await libraryApi.updateStatus()).count;
		} catch {
			// The badge stays empty; the Updates tab shows the error itself.
		}
	}

	function selectTab(id: string) {
		libraryTab.set(id as LibraryTab);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!closesDetailOn(event.key, selected !== null, addOpen)) return;
		void closeGame();
	}

	function openGame(game: Game) {
		selected = game;
	}

	/** A game with an id is in the library. A browse result has no id yet and opens the add dialog. */
	function openTile(game: Game) {
		if (game.id > 0) openGame(game);
		else openBrowseGame(game);
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

<svelte:window onkeydown={handleKeydown} />

<div class="library">
	<PageHeader title="Library" {subtitle} icon="collections_bookmark">
		<Button size="sm" variant="primary" icon="add" onclick={() => (addOpen = true)}>Add game</Button>
	</PageHeader>

	<div class="top-tabs">
		<Tabs tabs={topTabs} value={tab} onChange={selectTab} aria-label="Library sections" />
	</div>

	{#if tab === "updates"}
		<UpdatesPage sources={library.overview.sources} onStatus={(next) => (updateCount = next.count)} />
	{:else if tab === "downloads"}
		<DownloadsPage {downloads} />
	{:else if tab === "incoming"}
		<IncomingPanel onChanged={() => library.refresh()} />
	{:else if !library.overviewLoaded}
		<LoadingState label="Loading library..." />
	{:else if library.overviewError}
		<Notice tone="error">
			{library.overviewError}
			<Button size="sm" onclick={() => library.refresh()}>Retry</Button>
		</Notice>
	{:else if current}
		<div class="layout" class:with-detail={selected !== null}>
			<div class="rail-col">
				<SourceRail
					sections={library.sections}
					selected={library.section}
					onSelect={(source) => library.selectSection(source)}
				/>
			</div>

			<div class="side-col" class:open={filtersOpen}>
				<LibrarySidebar
					{library}
					canSync={canSyncGames(current)}
					syncing={library.isSyncing(library.section)}
					onSync={syncNow}
				/>
			</div>

			<div class="main">
				<div class="filters-toggle">
					<Button size="sm" icon="tune" onclick={() => (filtersOpen = !filtersOpen)}>Filters</Button>
				</div>

				<SectionHeader
					label={current.label}
					count={current.count}
					{isPlugin}
					progress={library.syncs[library.section]}
				/>

				{#if library.error}
					<Notice tone="error">
						{library.error}
						<Button size="sm" onclick={() => library.retry()}>Retry</Button>
					</Notice>
				{/if}

				{#if library.total === null}
					{#if !library.error}<LoadingState label="Loading games..." />{/if}
				{:else if library.total === 0}
					{#if !library.error}
						{#if library.filtersActive}
							<EmptyState icon="search_off" title="No games match these filters.">
								<Button size="sm" onclick={() => library.clearFilters()}>Clear filters</Button>
							</EmptyState>
						{:else if isPlugin}
							<EmptyState
								icon="store"
								title="No games yet."
								hint="Install the plugin, sign in, then press Sync now."
							/>
						{:else}
							<EmptyState
								icon="videogame_asset"
								title="No games yet."
								hint="Add a link or scan your folders."
							>
								<Button size="sm" variant="primary" onclick={() => (addOpen = true)}>Add game</Button>
							</EmptyState>
						{/if}
					{/if}
				{:else}
					{#key library.section}
						<LibraryGrid
							{library}
							{ratio}
							minColumnPx={cardSize}
							card={current.card}
							{transfers}
							blurUntilHover={blurCovers}
							{revealed}
							onOpen={openTile}
						/>
					{/key}
				{/if}
			</div>

			{#if selected}
				<div class="detail-backdrop" role="presentation" onclick={closeGame}></div>
				<aside class="detail-col" aria-label="Game details">
					<!-- Keyed by id so switching tiles loads the new game's files and notes. -->
					{#key selected.id}
						<div class="detail-top">
							<IconButton icon="close" label="Close" onclick={closeGame} />
						</div>
						<GameDetail
							game={selected}
							sourceName={sourceLabel(library.overview.sources, selected.source)}
							actions={gameActions}
							onRemoved={gameRemoved}
							onChanged={gameChanged}
						/>
					{/key}
				</aside>
			{/if}
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
	}

	.top-tabs {
		margin-bottom: 20px;
	}

	/*
	 * One grid: [rail 56px | filters 260px | grid fills the rest]. The rail and filter columns
	 * stretch the full content height and share one panel background, split by a thin divider.
	 * With a game open, the detail panel is a fourth column on the right.
	 */
	.layout {
		display: grid;
		grid-template-columns: 56px 260px minmax(0, 1fr);
		align-items: stretch;
	}

	.layout.with-detail {
		grid-template-columns: 56px 260px minmax(0, 1fr) 420px;
	}

	.rail-col,
	.side-col {
		background: var(--bg-surface);
		border-right: 1px solid var(--glass-border);
	}

	.main {
		min-width: 0;
		padding-left: 24px;
	}

	.filters-toggle {
		display: none;
		margin-bottom: 12px;
	}

	/* Wide: the panel sits beside the grid and stays in view while the grid scrolls. */
	.detail-col {
		position: sticky;
		top: 76px;
		align-self: start;
		max-height: calc(100vh - 116px);
		overflow-y: auto;
		padding: 16px 20px 24px;
		background: var(--bg-surface);
		border-left: 1px solid var(--glass-border);
		animation: detail-slide-in var(--motion-base) var(--ease-out) both;
	}

	.detail-top {
		display: flex;
		justify-content: flex-end;
		margin-bottom: 8px;
	}

	.detail-backdrop {
		display: none;
	}

	@keyframes detail-slide-in {
		from {
			opacity: 0;
			transform: translateX(24px);
		}
	}

	/* Narrow: the rail stays a column. The filters leave the grid and slide over it. */
	@media (max-width: 899px) {
		.layout,
		.layout.with-detail {
			grid-template-columns: 56px minmax(0, 1fr);
		}

		.side-col {
			position: fixed;
			top: 0;
			bottom: 0;
			left: 0;
			z-index: 200;
			width: min(320px, 88vw);
			background: var(--bg-surface);
			padding-top: 24px;
			overflow-y: auto;
			box-shadow: 0 0 24px rgb(0 0 0 / 0.35);
			transform: translateX(-105%);
			transition: transform 180ms var(--ease-out);
		}

		.side-col.open {
			transform: none;
		}

		.main {
			padding-left: 16px;
		}

		.filters-toggle {
			display: flex;
		}

		/* Narrow: the panel overlays the grid from the right, over a backdrop. */
		.detail-col {
			position: fixed;
			top: 0;
			right: 0;
			bottom: 0;
			z-index: 210;
			width: min(420px, 100vw);
			max-height: none;
			padding-top: 24px;
			box-shadow: 0 0 24px rgb(0 0 0 / 0.35);
		}

		.detail-backdrop {
			display: block;
			position: fixed;
			inset: 0;
			z-index: 205;
			background: rgb(0 0 0 / 0.35);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.side-col {
			transition: none;
		}

		.detail-col {
			animation: none;
		}
	}
</style>
