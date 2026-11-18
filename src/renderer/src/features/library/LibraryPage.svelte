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
	import SourceRail from "./SourceRail.svelte";
	import { LibraryState } from "./libraryState.svelte";
	import { sourceLabel } from "./librarySections";
	import { sourceIcon } from "./sourceIcon";
	import { librarySubtitle, updateCountOf } from "./libraryCounts";
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
	let selected = $state<Game | null>(null);
	let addOpen = $state(false);
	let addUrl = $state("");
	let cardSize = $state(LIBRARY_CARD_SIZE_DEFAULT);
	let blurCovers = $state(false);
	/** Under 900px the filter sidebar slides in over the grid. */
	let filtersOpen = $state(false);
	/** Measured height of the sticky header, so the rail and sidebar can stick below it. */
	let headHeight = $state(0);

	const subtitle = $derived(librarySubtitle(library.overview.stats));
	const topTabs = $derived<TabItem[]>([
		{ id: "games", label: "Games", icon: "videogame_asset" },
		{
			id: "updates",
			label: "Updates",
			icon: "system_update_alt",
			badge: updateCountOf(library.overview.stats)
		},
		{ id: "downloads", label: "Downloads", icon: "downloading", badge: activeDownloads },
		{ id: "incoming", label: "Incoming", icon: "move_to_inbox" }
	]);
	const gameActions = $derived(
		(library.overview.sources.find((source) => source.source === selected?.source)?.actions ?? []).filter(
			(action) => action.scope === "game"
		)
	);

	onMount(() => {
		const stopTab = libraryTab.subscribe((next) => {
			tab = next;
			// Opening a tab mounts its page; it stays mounted from then on.
			if (next !== "games") mountedTabs[next] = true;
		});
		void library.start();
		void loadAppearance();
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

	/** A browse result is not saved yet, so it opens the add dialog with its link. */
	function openBrowseGame(game: Game) {
		addUrl = game.url;
		addOpen = true;
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
	<div class="library-head" bind:clientHeight={headHeight}>
		<PageHeader title="Library" {subtitle} icon="collections_bookmark">
			<Button
				class="add-game"
				size="sm"
				variant="primary"
				icon="add"
				aria-label="Add game"
				title="Add game"
				onclick={() => (addOpen = true)}
			>
				<span class="add-label">Add game</span>
			</Button>
		</PageHeader>

		<div class="top-tabs">
			<Tabs tabs={topTabs} value={tab} onChange={selectTab} aria-label="Library sections" />
		</div>
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
			<UpdatesPage sources={library.overview.sources} />
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
				<div class="rail-col">
					<SourceRail
						sections={library.visibleSections}
						selected={library.activeSection?.source ?? ""}
						onSelect={(source) => library.selectSource(source)}
					/>
				</div>

				<div class="side-col" class:open={filtersOpen}>
					<LibrarySidebar {library} />
				</div>

				<div class="main">
					<div class="filters-toggle">
						<Button size="sm" icon="tune" onclick={() => (filtersOpen = !filtersOpen)}>Filters</Button>
					</div>

					{#if library.activeSection}
						{@const source = library.activeSection}
						{@const list = library.listOf(source.source)}
						<section class="group" aria-label={source.label}>
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
									onOpen={openTile}
								/>
							{/if}
						</section>
					{/if}
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
	.library {
		font-size: 1rem;
		line-height: 1.5;
	}

	/*
	 * The title and tabs stay put while the page scrolls, directly below the global top bar
	 * (the page box starts there). The opaque background hides the tiles that pass under them.
	 */
	.library-head {
		position: sticky;
		top: 0;
		z-index: 20;
		box-sizing: border-box;
		padding-bottom: 16px;
		background: var(--bg-base);
	}

	.top-tabs {
		margin-bottom: 0;
	}

	/*
	 * One grid: [rail 56px | filters 260px | grid fills the rest]. The rail and filter columns
	 * sit on the page background; only the divider after the filters separates them from the grid.
	 * The game details open over the page as a drawer, so the grid keeps its full width.
	 */
	.layout {
		display: grid;
		grid-template-columns: 56px 260px minmax(0, 1fr);
		align-items: stretch;
	}

	.side-col {
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

	/* The source on the page: its header, then its sync note, notices and grid. */
	.group {
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

	/* Narrow: the rail stays a column. The filters leave the grid and slide over it. */
	@media (max-width: 899px) {
		.layout {
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

		.library-head :global(.page-header) {
			flex-wrap: wrap;
		}

		.library-head :global(.button.add-game) {
			width: 36px;
			height: 36px;
			padding: 0;
			gap: 0;
		}

		.add-label {
			display: none;
		}

		.detail-col {
			left: 0;
			width: auto;
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
