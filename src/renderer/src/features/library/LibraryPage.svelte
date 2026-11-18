<script lang="ts">
	import { onMount, tick } from "svelte";
	import { get } from "svelte/store";
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import Tabs from "@components/shared/Tabs.svelte";
	import type { TabItem } from "@components/shared/tabs";
	import { libraryTab, type LibraryTab } from "@stores/libraryTabStore";
	import type { Game, LibrarySource, SourceAction } from "../../lib/features/types";
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
	import SourceChips from "./SourceChips.svelte";
	import { LibraryState } from "./libraryState.svelte";
	import { REVEAL_MAX_TILES, revealDoneMs } from "./libraryMotion";
	import { librarySubtitle, sectionElementId, showsSection, sourceLabel } from "./librarySections";
	import { sourceIcon } from "./sourceIcon";
	import { syncMessage } from "./librarySync";
	import { coverRatioOf } from "./coverRatio";
	import { canSyncGames } from "./syncFlag";
	import { closesDetailOn } from "./detailPanel";
	import { DownloadsState } from "../downloads/downloadsState.svelte";
	import DownloadsPage from "../downloads/DownloadsPage.svelte";
	import UpdatesPage from "../updates/UpdatesPage.svelte";
	import { transferKinds } from "./transferState";

	const SYNC_HINT = "Adds the games from your account to the Library";

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
	/** Tabs whose page stays mounted once opened. The Games page is cheap to rebuild from the lists. */
	type PanelTab = Exclude<LibraryTab, "games">;

	/** The tab the store opens on, used to mount its panel on the first render. */
	const startingTab = get(libraryTab);
	let tab = $state<LibraryTab>(startingTab);
	/**
	 * Panels mount on their first visit and are only hidden afterwards, never thrown away, so
	 * scroll position, expanded rows and running checks survive a switch to another tab.
	 */
	let mountedTabs = $state<Record<PanelTab, boolean>>({
		updates: startingTab === "updates",
		downloads: startingTab === "downloads",
		incoming: startingTab === "incoming"
	});
	let updateCount = $state(0);
	let selected = $state<Game | null>(null);
	let addOpen = $state(false);
	let addUrl = $state("");
	let revealed = $state(false);
	let cardSize = $state(LIBRARY_CARD_SIZE_DEFAULT);
	let blurCovers = $state(false);
	/** The filter panel slides in over the grid, at every width. */
	let filtersOpen = $state(false);
	/** Measured height of the sticky toolbar, so the page anchors below it. */
	let headHeight = $state(0);
	/** True once the page has scrolled, when the header needs its glass backing. */
	let headStuck = $state(false);

	/** The sections that show on the page, in rail order. Empty plugin sources are left out. */
	const visibleSections = $derived(library.sections.filter(isShown));
	/** Games in the sections on the page, for the reveal animation. */
	const shownTotal = $derived(
		visibleSections.reduce((sum, source) => sum + (library.listOf(source.source).total ?? 0), 0)
	);
	const subtitle = $derived(librarySubtitle(library.overview.stats));
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
	const hasTiles = $derived(shownTotal > 0);
	const revealCount = $derived(Math.min(shownTotal, REVEAL_MAX_TILES));

	onMount(() => {
		const stopTab = libraryTab.subscribe((next) => {
			tab = next;
			// Opening a tab mounts its page; it stays mounted from then on.
			if (next !== "games") mountedTabs[next] = true;
		});
		void library.start();
		void loadAppearance();
		// The Updates page reads its own status and reports it back, so start it only off that tab.
		if (tab !== "updates") void loadUpdateCount();
		void downloads.load();
		const stopDownloads = downloads.subscribe();
		const stopLibrary = library.subscribe();

		// The header only needs its glass backing while the page is scrolled under it.
		const scroller = document.querySelector<HTMLElement>(".page-wrapper");
		const onScroll = () => {
			headStuck = (scroller?.scrollTop ?? 0) > 0;
		};
		scroller?.addEventListener("scroll", onScroll, { passive: true });
		onScroll();

		return () => {
			stopTab();
			downloads.stopSampling();
			stopDownloads();
			stopLibrary();
			scroller?.removeEventListener("scroll", onScroll);
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

	// The detail drawer is modal: stop the page behind its blurred backdrop from scrolling.
	$effect(() => {
		if (!selected) return;
		const scroller = document.querySelector<HTMLElement>(".page-wrapper");
		if (!scroller) return;
		const previous = scroller.style.overflowY;
		scroller.style.overflowY = "hidden";
		return () => {
			scroller.style.overflowY = previous;
		};
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
		if (closesDetailOn(event.key, selected !== null, addOpen)) {
			void closeGame();
		} else if (event.key === "Escape" && filtersOpen && !addOpen && selected === null) {
			filtersOpen = false;
		}
	}

	function openGame(game: Game) {
		selected = game;
	}

	/** A game with an id is in the library. A browse result has no id yet and opens the add dialog. */
	function openTile(game: Game) {
		if (game.id > 0) openGame(game);
		else openBrowseGame(game);
	}

	/** A browse result is not saved yet, so it opens the add dialog with its link. */
	function openBrowseGame(game: Game) {
		addUrl = game.url;
		addOpen = true;
	}

	/** A section shows when it is local, or a plugin source with games, an error or a live browse. */
	function isShown(source: LibrarySource): boolean {
		if (library.browse.shows(source.source)) return true;
		const list = library.listOf(source.source);
		return showsSection(source, list.total, list.error !== "");
	}

	/** Scrolls the page to a source's group and shows its filter groups. The source chips call this. */
	function showSection(source: string) {
		library.selectSource(source);
		document
			.getElementById(sectionElementId(source))
			?.scrollIntoView({ behavior: "smooth", block: "start" });
	}

	/** Scrolls back to the first group without changing the filter source. The All chip calls this. */
	function showAll() {
		const first = visibleSections[0];
		if (!first) return;
		document
			.getElementById(sectionElementId(first.source))
			?.scrollIntoView({ behavior: "smooth", block: "start" });
	}

	function libraryActionsOf(source: LibrarySource): SourceAction[] {
		return source.actions.filter((action) => action.scope === "library");
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

	function closeAdd() {
		addOpen = false;
		addUrl = "";
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="library" style:--library-head-height="{headHeight}px">
	<div class="library-head" class:stuck={headStuck} bind:clientHeight={headHeight}>
		<PageHeader title="Library" {subtitle} icon="collections_bookmark">
			<Button size="sm" icon="tune" onclick={() => (filtersOpen = true)}>Filters</Button>
			<Button size="sm" variant="primary" icon="add" onclick={() => (addOpen = true)}>Add game</Button>
		</PageHeader>

		<div class="top-tabs">
			<Tabs tabs={topTabs} value={tab} onChange={selectTab} aria-label="Library sections" />
		</div>

		{#if tab === "games" && library.overviewLoaded && !library.overviewError}
			<SourceChips
				sections={visibleSections}
				selected={library.activeSource}
				onSelect={showSection}
				onAll={showAll}
			/>
		{/if}
	</div>

	<!--
		Updates, Downloads and Incoming mount on their first visit, then stay mounted and hidden,
		so their scroll position, expanded rows and running transfers survive a tab switch.
	-->
	{#if mountedTabs.updates}
		<div
			role="tabpanel"
			aria-label="Updates"
			hidden={tab !== "updates"}
			aria-hidden={tab !== "updates"}
		>
			<UpdatesPage sources={library.overview.sources} onStatus={(next) => (updateCount = next.count)} />
		</div>
	{/if}

	{#if mountedTabs.downloads}
		<div
			role="tabpanel"
			aria-label="Downloads"
			hidden={tab !== "downloads"}
			aria-hidden={tab !== "downloads"}
		>
			<DownloadsPage {downloads} />
		</div>
	{/if}

	{#if mountedTabs.incoming}
		<div
			role="tabpanel"
			aria-label="Incoming"
			hidden={tab !== "incoming"}
			aria-hidden={tab !== "incoming"}
		>
			<IncomingPanel onChanged={() => library.refresh()} />
		</div>
	{/if}

	{#if tab === "games"}
		{#if !library.overviewLoaded}
			<LoadingState label="Loading library..." />
		{:else if library.overviewError}
			<Notice tone="error">
				{library.overviewError}
				<Button size="sm" onclick={() => library.refresh()}>Retry</Button>
			</Notice>
		{:else}
			<div class="layout">
				{#if filtersOpen}
					<div class="scrim" role="presentation" onclick={() => (filtersOpen = false)}></div>
				{/if}
				<aside class="side-col" class:open={filtersOpen} aria-label="Filters" inert={!filtersOpen}>
					<div class="side-chrome">
						<span class="side-title">Filters</span>
						<IconButton icon="close" label="Close filters" onclick={() => (filtersOpen = false)} />
					</div>
					<div class="side-scroll">
						<LibrarySidebar {library} />
					</div>
				</aside>

				<div class="main">
					{#each visibleSections as source (source.source)}
						{@const list = library.listOf(source.source)}
						<section class="group" id={sectionElementId(source.source)} aria-label={source.label}>
							<header class="group-header">
								<span class="group-icon"><Icon name={sourceIcon(source.source)} size={22} /></span>
								<h2>{source.label}</h2>
								<Badge>{list.total ?? 0}</Badge>
								<div class="group-actions">
									{#if canSyncGames(source)}
										<Button
											size="sm"
											icon="sync"
											title={SYNC_HINT}
											disabled={library.isSyncing(source.source)}
											onclick={() => library.startSync(source.source)}
										>
											{library.isSyncing(source.source) ? "Syncing..." : "Sync now"}
										</Button>
									{/if}
									{#each libraryActionsOf(source) as action (action.id)}
										<Button
											size="sm"
											icon={action.icon}
											loading={library.runningAction === action.id}
											disabled={library.runningAction !== ""}
											onclick={() => library.runSourceAction(source.source, action)}
										>
											{action.label}
										</Button>
									{/each}
								</div>
							</header>

							{#if library.syncs[source.source]}
								{@const progress = library.syncs[source.source]}
								<p class="sync" class:failed={progress.phase === "error"} role="status">
									{syncMessage(progress)}
								</p>
							{/if}

							{#if list.error}
								<Notice tone="error">
									{list.error}
									<Button size="sm" onclick={() => library.retry(source.source)}>Retry</Button>
								</Notice>
							{/if}

							{#if list.total === null}
								{#if !list.error}<LoadingState label="Loading games..." />{/if}
							{:else if list.total === 0}
								{#if !list.error}
									{#if library.filtersActive}
										<EmptyState icon="search_off" title="No games match these filters.">
											<Button size="sm" onclick={() => library.clearFilters()}>Clear filters</Button>
										</EmptyState>
									{:else}
										<EmptyState icon="videogame_asset" title="No games yet." hint="Add a link or scan your folders.">
											<Button size="sm" variant="primary" onclick={() => (addOpen = true)}>Add game</Button>
										</EmptyState>
									{/if}
								{/if}
							{:else}
								<LibraryGrid
									{library}
									source={source.source}
									ratio={coverRatioOf(source.cover)}
									minColumnPx={cardSize}
									card={source.card}
									variant={source.kind === "local" ? "home" : "cover"}
									{transfers}
									blurUntilHover={blurCovers}
									{revealed}
									onOpen={openTile}
								/>
							{/if}
						</section>
					{/each}
				</div>

				{#if selected}
					<div
						class="detail-backdrop"
						role="presentation"
						title="Click outside to close"
						onclick={closeGame}
					></div>
					<aside
						class="detail-col"
						role="dialog"
						aria-modal="true"
						aria-label="Game details"
					>
						<div class="detail-chrome">
							<span class="detail-eyebrow">Game details</span>
							<IconButton icon="close" label="Close" onclick={closeGame} />
						</div>
						<!-- Keyed by id so switching tiles resets the scroll and loads the new game's files and notes. -->
						{#key selected.id}
							<div class="detail-scroll">
								<GameDetail
									game={selected}
									sourceName={sourceLabel(library.overview.sources, selected.source)}
									actions={gameActions}
									onRemoved={gameRemoved}
									onChanged={gameChanged}
								/>
							</div>
						{/key}
					</aside>
				{/if}
			</div>
		{/if}
	{/if}

	{#if addOpen}
		<AddGameDialog onAdded={addedGame} onClose={closeAdd} initialUrl={addUrl} />
	{/if}
</div>

<style lang="scss">
	/*
	 * Library-only tokens. --library-top is the search bar's height above the page; the toolbar,
	 * the chips and the filter panel all anchor to it.
	 */
	.library {
		--library-top: 76px;
		--radius-lg: 20px;
		--library-tile-radius: 20px;

		font-size: 1rem;
		line-height: 1.5;
	}

	/*
	 * One sticky toolbar: title, tabs and source chips. Its solid background hides covers that
	 * scroll under it, and the band above it hides them under the search bar too.
	 */
	.library-head {
		position: sticky;
		top: var(--library-top);
		z-index: 20;
		display: flex;
		flex-direction: column;
		gap: 12px;
		box-sizing: border-box;
		padding-bottom: 12px;
		background: var(--bg-base);

		&::before {
			content: "";
			position: absolute;
			left: 0;
			right: 0;
			bottom: 100%;
			height: var(--library-top);
			background: inherit;
			pointer-events: none;
		}

		&.stuck {
			box-shadow: 0 8px 24px rgb(0 0 0 / 0.35);
		}
	}

	.top-tabs {
		margin-bottom: 0;
	}

	/* The grid takes the full width. The filters and the game details both slide over it. */
	.layout {
		min-width: 0;
	}

	.main {
		min-width: 0;
	}

	.scrim {
		position: fixed;
		inset: 0;
		z-index: 190;
		background: rgb(0 0 0 / 0.45);
		cursor: pointer;
		animation: detail-backdrop-in var(--motion-base) var(--ease-out) both;
	}

	.side-col {
		position: fixed;
		top: 0;
		bottom: 0;
		left: 0;
		z-index: 200;
		display: flex;
		flex-direction: column;
		width: min(340px, 88vw);
		background: var(--bg-surface);
		border-right: 1px solid var(--glass-border);
		box-shadow: 0 0 24px rgb(0 0 0 / 0.35);
		transform: translateX(-105%);
		transition: transform var(--motion-base) var(--ease-out);
	}

	.side-col.open {
		transform: none;
	}

	.side-chrome {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-shrink: 0;
		padding: 12px 12px 12px 16px;
		border-bottom: 1px solid var(--glass-border);
	}

	.side-title {
		font-weight: 600;
	}

	.side-scroll {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-gutter: stable;
		padding: 16px;
	}

	/* Each source is a group: its header, then its sync note, notices and grid. */
	.group {
		/* Keeps the group title clear of the sticky toolbar when a chip scrolls to it. */
		scroll-margin-top: calc(var(--library-top) + var(--library-head-height, 0px) + 16px);
		margin-bottom: 32px;
	}

	.group-header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
		margin-bottom: 12px;
	}

	.group-icon {
		display: inline-flex;
		color: var(--text-muted);
	}

	.group-header h2 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.group-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-left: auto;
	}

	.sync {
		min-height: 1.5em;
		margin: 0 0 12px;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.sync.failed {
		color: var(--danger);
	}

	/* The backdrop blurs and dims the page, and a click anywhere on it closes the drawer. */
	.detail-backdrop {
		position: fixed;
		inset: 0;
		z-index: 205;
		background: rgb(0 0 0 / 0.45);
		-webkit-backdrop-filter: blur(10px);
		backdrop-filter: blur(10px);
		cursor: pointer;
		animation: detail-backdrop-in var(--motion-base) var(--ease-out) both;
	}

	/* The drawer itself is sharp and opaque: the blur belongs to the excluded backdrop only. */
	.detail-col {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		z-index: 210;
		display: flex;
		flex-direction: column;
		width: min(460px, 100vw);
		background: var(--bg-surface);
		border-left: 1px solid var(--glass-border);
		box-shadow: -18px 0 60px rgb(0 0 0 / 0.4);
		animation: detail-slide-in var(--motion-base) var(--ease-out) both;
		overflow: hidden;
	}

	.detail-chrome {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-shrink: 0;
		padding: 12px 16px;
		border-bottom: 1px solid var(--glass-border);
		background: var(--bg-surface);
	}

	.detail-eyebrow {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 800;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	.detail-scroll {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-gutter: stable;
		padding: 16px 20px 28px;
	}

	@keyframes detail-backdrop-in {
		from {
			opacity: 0;
		}
	}

	@keyframes detail-slide-in {
		from {
			opacity: 0;
			transform: translateX(28px);
		}
	}

	/* Narrow screens: the game details take the whole width. */
	@media (max-width: 899px) {
		.detail-col {
			width: 100vw;
			border-left: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.side-col {
			transition: none;
		}

		.detail-col,
		.detail-backdrop {
			animation: none;
		}
	}
</style>
