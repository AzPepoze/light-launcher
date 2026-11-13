import { untrack } from "svelte";
import { notifications } from "@stores/notificationStore";
import { createLogger } from "../../lib/logger";
import {
	LIBRARY_CHANGED_EVENT,
	LIBRARY_GAME_EVENT,
	LIBRARY_REMOVED_EVENT,
	LIBRARY_SYNC_EVENT
} from "../../lib/features/events";
import { errorMessage, listen } from "../../lib/features/rpc";
import { libraryApi } from "../../lib/features/libraryApi";
import type {
	Game,
	LibraryOverview,
	LibrarySyncProgress,
	SidebarGroup,
	SourceAction
} from "../../lib/features/types";
import { BrowseFeed } from "./browseFeed.svelte";
import {
	browseQuery,
	hasBrowseFilters,
	initialBrowseFilters,
	setBrowseQuery,
	setMulti,
	setOrder,
	setSelect,
	setSort,
	setToggle,
	type BrowseFilters
} from "./browseFilters";
import { libraryQuery, toLibraryFilters, withLibraryFilters } from "./libraryFilters";
import { libraryErrorMessage } from "./libraryError";
import { LibraryPager, type PagerSnapshot } from "./libraryPager";
import { librarySections, LOCAL_SOURCE } from "./librarySections";
import { PAGE_SIZE, pagesToLoad } from "./libraryWindow";
import { loadSavedFilters, storeSavedFilters, type SavedFilter } from "./savedFilters";
import { builtinGroups, sidebarGroupsFor } from "./sidebarGroups";

/** Library events come often during an import. Reloading waits this long, then runs once. */
const REFRESH_DELAY_MS = 1000;
const SEARCH_DELAY_MS = 150;

const EMPTY_OVERVIEW: LibraryOverview = {
	sources: [],
	tags: [],
	stats: { total: 0, installed: 0, updates: 0, bytes: 0, playtime: 0 }
};
const log = createLogger("Library");

/**
 * What the library page shows. One sidebar state drives every section. Local games and plugin
 * sources without a sidebar read the library list. Plugin sources with a sidebar read their browse.
 */
export class LibraryState {
	overview = $state<LibraryOverview>(EMPTY_OVERVIEW);
	overviewLoaded = $state(false);
	overviewError = $state("");
	section = $state(LOCAL_SOURCE);
	/** The sidebar filters. Every section uses this one object. */
	filters = $state<BrowseFilters>(initialBrowseFilters(builtinGroups([])));
	savedFilters = $state<SavedFilter[]>(loadSavedFilters());
	/** Games in the library list for the current section and filters. Null until the first page. */
	listTotal = $state<number | null>(null);
	/** Changes whenever the list pages are dropped. The grid asks again when it does. */
	generation = $state(0);
	pageError = $state("");
	syncs = $state<Record<string, LibrarySyncProgress>>({});
	/** The id of the plugin action that is running, or an empty string. */
	runningAction = $state("");

	/** Each section's filters while the page is open, so switching back restores them. */
	private sectionFilters = new Map<string, BrowseFilters>();
	readonly pager: LibraryPager;
	readonly browse = new BrowseFeed();
	sections = $derived(librarySections(this.overview.sources));
	current = $derived(this.sections.find((item) => item.source === this.section));
	usesBrowse = $derived(this.section !== LOCAL_SOURCE && (this.current?.sidebar.length ?? 0) > 0);
	sidebarGroups = $derived(
		sidebarGroupsFor(this.current, this.overview.tags, this.browse.choices)
	);
	filtersActive = $derived(hasBrowseFilters(this.filters));

	private refreshTimer: ReturnType<typeof setTimeout> | undefined;
	private searchTimer: ReturnType<typeof setTimeout> | undefined;

	constructor() {
		this.pager = new LibraryPager({
			query: libraryQuery(this.section, toLibraryFilters(this.filters)),
			listPage: libraryApi.list,
			onChange: (snapshot) => this.applyPagerSnapshot(snapshot)
		});
	}

	/** The games in the section and filters. Null while the first page is loading. */
	get total(): number | null {
		return this.usesBrowse ? this.browse.total : this.listTotal;
	}

	get loading(): boolean {
		return this.usesBrowse && this.browse.loading;
	}

	get error(): string {
		if (this.overviewError) return this.overviewError;
		return this.usesBrowse ? this.browse.error : this.pageError;
	}

	/** Loads the counts and tags. Pages load when the grid asks for them. */
	async start(): Promise<void> {
		await this.loadOverview();
	}

	/** Drops every loaded list page and reloads the overview. */
	async refresh(): Promise<void> {
		this.pager.reset();
		await this.loadOverview();
	}

	/** Runs the failed load again: the overview, and the plugin browse when that is the section. */
	retry(): void {
		if (this.usesBrowse) this.browse.reloadCurrent();
		void this.refresh();
	}

	/** Reloads soon, once per burst of library events. */
	scheduleRefresh(): void {
		if (this.refreshTimer !== undefined) return;
		this.refreshTimer = setTimeout(() => {
			this.refreshTimer = undefined;
			void this.refresh();
		}, REFRESH_DELAY_MS);
	}

	/** Asks for the pages behind a visible range. Only missing pages are loaded. */
	requestRange(start: number, end: number): void {
		// Reading generation here makes the caller run again after the pages were dropped.
		const generation = this.generation;
		untrack(() => {
			if (this.usesBrowse) {
				this.browse.requestRange(start, end);
			} else {
				this.pager.requestPages(pagesToLoad({ start, end }, this.listTotal ?? PAGE_SIZE), generation);
			}
		});
	}

	gameAt(index: number): Game | undefined {
		return this.usesBrowse ? this.browse.gameAt(index) : this.pager.gameAt(index);
	}

	selectSection(source: string): void {
		if (source === this.section) return;
		clearTimeout(this.searchTimer);
		this.sectionFilters.set(this.section, this.filters);
		this.section = source;
		const groups = sidebarGroupsFor(this.current, this.overview.tags, {});
		this.filters = this.sectionFilters.get(source) ?? initialBrowseFilters(groups);
		this.browse.stop();
		const current = this.current;
		if (this.usesBrowse && current) this.browse.start(source, current.sidebar);
		this.applyFilters();
	}

	setQuery(text: string): void {
		this.filters = setBrowseQuery(this.filters, text);
		clearTimeout(this.searchTimer);
		this.searchTimer = setTimeout(() => this.applyFilters(), SEARCH_DELAY_MS);
	}

	setToggle(id: string, on: boolean): void {
		this.filters = setToggle(this.filters, id, on);
		this.applyFilters();
	}

	setSelect(id: string, value: string): void {
		this.filters = setSelect(this.filters, id, value);
		this.applyFilters();
	}

	toggleMulti(id: string, value: string): void {
		const current = this.filters.values[id];
		const selected = Array.isArray(current) ? current : [];
		const next = selected.includes(value)
			? selected.filter((item) => item !== value)
			: [...selected, value];
		this.filters = setMulti(this.filters, id, next);
		this.applyFilters();
	}

	setSort(value: string): void {
		this.filters = setSort(this.filters, value);
		this.applyFilters();
	}

	toggleOrder(): void {
		this.filters = setOrder(this.filters, this.filters.order === "asc" ? "desc" : "asc");
		this.applyFilters();
	}

	searchChoices(group: SidebarGroup, text: string): void {
		this.browse.searchChoices(group, text);
	}

	/** Clears every filter but the sort, which the user keeps. */
	clearFilters(): void {
		clearTimeout(this.searchTimer);
		const groups = sidebarGroupsFor(this.current, this.overview.tags, this.browse.choices);
		const cleared = initialBrowseFilters(groups);
		this.filters = { ...cleared, sort: this.filters.sort, order: this.filters.order };
		this.applyFilters();
	}

	saveCurrentFilter(name: string): void {
		const trimmed = name.trim();
		if (!trimmed) return;
		const entry: SavedFilter = {
			id: `${Date.now()}`,
			name: trimmed,
			filters: toLibraryFilters(this.filters)
		};
		this.savedFilters = [...this.savedFilters, entry];
		storeSavedFilters(this.savedFilters);
	}

	applySavedFilter(entry: SavedFilter): void {
		clearTimeout(this.searchTimer);
		this.filters = withLibraryFilters(this.filters, entry.filters);
		this.applyFilters();
	}

	deleteSavedFilter(id: string): void {
		this.savedFilters = this.savedFilters.filter((entry) => entry.id !== id);
		storeSavedFilters(this.savedFilters);
	}

	/** Puts a changed game into the cached pages, so its tile updates at once. */
	replaceGame(game: Game): void {
		this.pager.replaceGame(game);
	}

	isSyncing(source: string): boolean {
		return this.syncs[source]?.phase === "running";
	}

	/** Starts an import from a plugin source. Progress arrives as library.sync events. */
	async startSync(source: string): Promise<void> {
		this.setSyncProgress({ source, phase: "running", imported: 0, skipped: 0, failed: 0 });
		try {
			await libraryApi.syncSource(source);
		} catch (err) {
			this.setSyncProgress({
				source,
				phase: "error",
				imported: 0,
				skipped: 0,
				failed: 0,
				message: errorMessage(err)
			});
		}
	}

	/** Runs one action the current source declares. Refreshes the list when the plugin asks. */
	async runSourceAction(action: SourceAction): Promise<void> {
		if (this.runningAction) return;
		this.runningAction = action.id;
		try {
			const result = await libraryApi.runAction(this.section, action.id, null);
			if (result.message) notifications.success(result.message);
			if (result.refresh) {
				await this.refresh();
				if (this.usesBrowse) this.browse.reloadCurrent();
			}
		} catch (err) {
			notifications.error(libraryErrorMessage(err));
		} finally {
			this.runningAction = "";
		}
	}

	/** Keeps the lists in step with library events. Returns a function that stops them. */
	subscribe(): () => void {
		const stops = [
			listen<unknown>(LIBRARY_CHANGED_EVENT, () => this.scheduleRefresh()),
			listen<unknown>(LIBRARY_REMOVED_EVENT, () => this.scheduleRefresh()),
			listen<LibrarySyncProgress>(LIBRARY_SYNC_EVENT, (event) => this.applySyncEvent(event)),
			listen<{ id: number }>(LIBRARY_GAME_EVENT, (event) => this.refreshGame(event.id))
		];
		return () => {
			stops.forEach((stop) => stop());
			clearTimeout(this.refreshTimer);
			clearTimeout(this.searchTimer);
			this.browse.stop();
			this.refreshTimer = undefined;
		};
	}

	private async loadOverview(): Promise<void> {
		try {
			this.overview = await libraryApi.overview();
			this.overviewError = "";
		} catch (err) {
			this.overviewError = errorMessage(err);
		} finally {
			this.overviewLoaded = true;
		}
	}

	/** Reads the list or the browse again with the current section and filters. */
	private applyFilters(): void {
		if (this.usesBrowse) {
			this.browse.reload(browseQuery(this.section, this.filters));
		} else {
			this.pager.setQuery(libraryQuery(this.section, toLibraryFilters(this.filters)));
		}
	}

	private applyPagerSnapshot(snapshot: PagerSnapshot): void {
		this.listTotal = snapshot.total;
		this.generation = snapshot.generation;
		this.pageError = snapshot.error ? libraryErrorMessage(snapshot.error) : "";
	}

	private applySyncEvent(event: LibrarySyncProgress): void {
		this.setSyncProgress(event);
		if (event.phase !== "running") this.scheduleRefresh();
	}

	private setSyncProgress(progress: LibrarySyncProgress): void {
		this.syncs = { ...this.syncs, [progress.source]: progress };
	}

	/** Reloads one game when its tile is on screen. Other games wait for the next reload. */
	private async refreshGame(id: number): Promise<void> {
		if (!this.pager.hasGame(id)) return;
		try {
			const { game } = await libraryApi.get(id);
			this.pager.replaceGame(game);
		} catch (err) {
			log.warn(`could not refresh game ${id}`, errorMessage(err));
		}
	}
}
