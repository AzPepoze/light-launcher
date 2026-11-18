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
	LibraryQuery,
	LibrarySource,
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
import {
	anyPicked,
	libraryQuery,
	sourceValues,
	toLibraryFilters,
	withLibraryFilters,
	withoutPluginFilters
} from "./libraryFilters";
import { libraryErrorMessage } from "./libraryError";
import { LibraryPager, type PagerSnapshot } from "./libraryPager";
import { LOCAL_SOURCE, librarySections } from "./librarySections";
import { PAGE_SIZE, pagesToLoad } from "./libraryWindow";
import { loadSavedFilters, storeSavedFilters, type SavedFilter } from "./savedFilters";
import { builtinGroups, choiceGroupsOf, sidebarGroupsFor } from "./sidebarGroups";

/** Library events come often during an import. Reloading waits this long, then runs once. */
const REFRESH_DELAY_MS = 1000;
const SEARCH_DELAY_MS = 150;

const EMPTY_OVERVIEW: LibraryOverview = {
	sources: [],
	tags: [],
	stats: { total: 0, installed: 0, updates: 0, bytes: 0, playtime: 0 }
};
/** The list of a source that has not loaded yet. */
const NO_LIST: SectionList = { total: null, generation: 0, error: "" };
const log = createLogger("Library");

/** What one source's list shows on the page. */
export interface SectionList {
	/** Games that match the filters, or null until the first page arrives. */
	total: number | null;
	/** Changes whenever the source's pages are dropped. The grid asks again when it does. */
	generation: number;
	/** The message of the last failed load, or an empty string. */
	error: string;
}

/**
 * What the library page shows. Every source has its own list (one pager and one grid), and all
 * lists share one filter set. A plugin group shows its browse instead of its list while a choice is picked.
 */
export class LibraryState {
	overview = $state<LibraryOverview>(EMPTY_OVERVIEW);
	overviewLoaded = $state(false);
	overviewError = $state("");
	/** The sidebar filters. Every source uses this one object. */
	filters = $state<BrowseFilters>(initialBrowseFilters(builtinGroups([])));
	savedFilters = $state<SavedFilter[]>(loadSavedFilters());
	/** The list of each source on the page, by source id. */
	lists = $state<Record<string, SectionList>>({});
	syncs = $state<Record<string, LibrarySyncProgress>>({});
	/** The id of the plugin action that is running, or an empty string. */
	runningAction = $state("");

	/** One pager per source. Created on first use, so each source loads its own pages. */
	private readonly pagers = new Map<string, LibraryPager>();
	/** Plugin choices for the sidebar. Filled only when the user asks for them. */
	readonly browse = new BrowseFeed();
	sections = $derived(librarySections(this.overview.sources));
	/** The source the rail picked last. The sidebar shows its groups. */
	activeSource = $state(LOCAL_SOURCE);
	activeSection = $derived(this.sections.find((item) => item.source === this.activeSource));
	sidebarGroups = $derived(
		sidebarGroupsFor(this.activeSection, this.overview.tags, this.browse.choices)
	);
	filtersActive = $derived(hasBrowseFilters(this.filters));

	private refreshTimer: ReturnType<typeof setTimeout> | undefined;
	private searchTimer: ReturnType<typeof setTimeout> | undefined;

	/** The list of a source. Reads as empty until its first page arrives. */
	listOf(source: string): SectionList {
		if (this.browse.shows(source)) {
			return { total: this.browse.total, generation: this.browse.version, error: this.browse.error };
		}
		return this.lists[source] ?? NO_LIST;
	}

	/** Asks again for the rows that failed: the plugin browse when it shows, then the overview. */
	retry(source: string): void {
		if (this.browse.shows(source)) this.browse.reloadCurrent();
		void this.refresh();
	}

	/** Loads the counts and tags, then the list of each source. Pages load when the grid asks. */
	async start(): Promise<void> {
		await this.loadOverview();
		this.syncPagers();
	}

	/** Drops every loaded list page and reloads the overview. */
	async refresh(): Promise<void> {
		for (const pager of this.pagers.values()) pager.reset();
		await this.loadOverview();
		this.syncPagers();
	}

	/** Reloads soon, once per burst of library events. */
	scheduleRefresh(): void {
		if (this.refreshTimer !== undefined) return;
		this.refreshTimer = setTimeout(() => {
			this.refreshTimer = undefined;
			void this.refresh();
		}, REFRESH_DELAY_MS);
	}

	/** Asks for the pages of one source behind a visible range. Only missing pages are loaded. */
	requestRange(source: string, start: number, end: number): void {
		// Reading the generation here makes the caller run again after that source's pages were dropped.
		const generation = this.listOf(source).generation;
		untrack(() => {
			if (this.browse.shows(source)) {
				this.browse.requestRange(start, end);
				return;
			}
			const total = this.listOf(source).total ?? PAGE_SIZE;
			this.pagers.get(source)?.requestPages(pagesToLoad({ start, end }, total), generation);
		});
	}

	/** The game at one index of a source's list. Undefined while its page loads. */
	gameAt(source: string, index: number): Game | undefined {
		if (this.browse.shows(source)) return this.browse.gameAt(index);
		return this.pagers.get(source)?.gameAt(index);
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

	/**
	 * Shows another source's groups in the sidebar. The choices and the plugin browse of the source
	 * left are dropped, and no plugin is asked.
	 */
	selectSource(source: string): void {
		if (source === this.activeSource) return;
		this.activeSource = source;
		this.browse.stop();
		this.filters = withoutPluginFilters(this.filters);
	}

	/** Whether a dynamic group's choices have arrived for the sidebar. */
	choicesLoaded(groupId: string): boolean {
		return this.browse.choices[groupId] !== undefined;
	}

	/** Asks the plugin for the choices of the active source's dynamic groups. Runs only on a press. */
	loadChoices(): void {
		const source = this.activeSection;
		if (!source || choiceGroupsOf(source).length === 0) return;
		this.browse.start(source.source, source.sidebar);
	}

	searchChoices(group: SidebarGroup, text: string): void {
		this.browse.searchChoices(group, text);
	}

	/** Clears every filter but the sort, which the user keeps. */
	clearFilters(): void {
		clearTimeout(this.searchTimer);
		const cleared = initialBrowseFilters(this.sidebarGroups);
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
		for (const pager of this.pagers.values()) pager.replaceGame(game);
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

	/** Runs one action a source declares. Refreshes the lists when the plugin asks. */
	async runSourceAction(source: string, action: SourceAction): Promise<void> {
		if (this.runningAction) return;
		this.runningAction = action.id;
		try {
			const result = await libraryApi.runAction(source, action.id, null);
			if (result.message) notifications.success(result.message);
			if (result.refresh) {
				await this.refresh();
				if (this.browse.shows(source)) this.browse.reloadCurrent();
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

	/** Makes sure every section has a pager. A new pager loads its first page at once. */
	private syncPagers(): void {
		for (const section of this.sections) this.pagerFor(section.source);
	}

	private pagerFor(source: string): LibraryPager {
		const existing = this.pagers.get(source);
		if (existing) return existing;
		const pager = new LibraryPager({
			query: this.queryFor(source),
			listPage: libraryApi.list,
			onChange: (snapshot) => this.applyPagerSnapshot(source, snapshot)
		});
		this.pagers.set(source, pager);
		pager.reset();
		return pager;
	}

	private queryFor(source: string): LibraryQuery {
		return libraryQuery(source, toLibraryFilters(this.filters));
	}

	/** Reads every list again with the current filters. Only a user filter change gets here. */
	private applyFilters(): void {
		for (const [source, pager] of this.pagers) pager.setQuery(this.queryFor(source));
		this.applyBrowse();
	}

	/**
	 * Shows the plugin browse of the active source while it has a picked choice, and asks the plugin
	 * for it. With no choice picked, the grid goes back to the local list and the plugin is not asked.
	 */
	private applyBrowse(): void {
		const section = this.activeSection;
		if (!section || section.source === LOCAL_SOURCE) return;
		const values = sourceValues(this.filters.values, section.sidebar);
		if (!anyPicked(values)) {
			this.browse.clearResults();
			return;
		}
		this.browse.reload(browseQuery(section.source, { ...this.filters, values }));
	}

	private applyPagerSnapshot(source: string, snapshot: PagerSnapshot): void {
		this.lists = {
			...this.lists,
			[source]: {
				total: snapshot.total,
				generation: snapshot.generation,
				error: snapshot.error ? libraryErrorMessage(snapshot.error) : ""
			}
		};
	}

	private applySyncEvent(event: LibrarySyncProgress): void {
		this.setSyncProgress(event);
		if (event.phase !== "running") this.scheduleRefresh();
	}

	private setSyncProgress(progress: LibrarySyncProgress): void {
		this.syncs = { ...this.syncs, [progress.source]: progress };
	}

	private hasGame(id: number): boolean {
		return [...this.pagers.values()].some((pager) => pager.hasGame(id));
	}

	/** Reloads one game when its tile is on screen. Other games wait for the next reload. */
	private async refreshGame(id: number): Promise<void> {
		if (!this.hasGame(id)) return;
		try {
			const { game } = await libraryApi.get(id);
			this.replaceGame(game);
		} catch (err) {
			log.warn(`could not refresh game ${id}`, errorMessage(err));
		}
	}
}
