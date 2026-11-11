import { untrack } from "svelte";
import { createLogger } from "../../lib/logger";
import {
	LIBRARY_CHANGED_EVENT,
	LIBRARY_GAME_EVENT,
	LIBRARY_REMOVED_EVENT,
	LIBRARY_SYNC_EVENT
} from "../../lib/features/events";
import { errorMessage, listen } from "../../lib/features/rpc";
import { libraryApi } from "../../lib/features/libraryApi";
import type { Game, LibraryOverview, LibrarySyncProgress } from "../../lib/features/types";
import {
	DEFAULT_FILTERS,
	hasActiveFilters,
	libraryQuery,
	type LibraryFilters
} from "./libraryFilters";
import { LibraryPager, type PagerSnapshot } from "./libraryPager";
import { librarySections, LOCAL_SOURCE } from "./librarySections";
import { PAGE_SIZE, pagesToLoad } from "./libraryWindow";
import { loadSavedFilters, storeSavedFilters, type SavedFilter } from "./savedFilters";

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
 * What the library page shows. It holds the overview (counts, tags), the current section's pages
 * and the filters. It never holds the whole library.
 */
export class LibraryState {
	overview = $state<LibraryOverview>(EMPTY_OVERVIEW);
	overviewLoaded = $state(false);
	overviewError = $state("");
	section = $state(LOCAL_SOURCE);
	filters = $state<LibraryFilters>({ ...DEFAULT_FILTERS, tags: [] });
	/** What the search box shows. The server gets it after a short pause in typing. */
	searchText = $state("");
	savedFilters = $state<SavedFilter[]>(loadSavedFilters());
	/** Games in the current section and filters. Null until the first page arrives. */
	total = $state<number | null>(null);
	/** Changes whenever the pages are dropped. The grid asks again when it does. */
	generation = $state(0);
	pageError = $state("");
	syncs = $state<Record<string, LibrarySyncProgress>>({});

	readonly pager: LibraryPager;
	sections = $derived(librarySections(this.overview.sources));
	filtersActive = $derived(hasActiveFilters(this.filters));

	private refreshTimer: ReturnType<typeof setTimeout> | undefined;
	private searchTimer: ReturnType<typeof setTimeout> | undefined;

	constructor() {
		this.pager = new LibraryPager({
			query: libraryQuery(this.section, this.filters),
			listPage: libraryApi.list,
			onChange: (snapshot) => this.applyPagerSnapshot(snapshot)
		});
	}

	get error(): string {
		return this.overviewError || this.pageError;
	}

	/** Loads the counts and tags. Pages load when the grid asks for them. */
	async start(): Promise<void> {
		await this.loadOverview();
	}

	/** Drops every loaded page and reloads the overview. */
	async refresh(): Promise<void> {
		this.pager.reset();
		await this.loadOverview();
	}

	/** Reloads soon, once per burst of library events. */
	scheduleRefresh(): void {
		if (this.refreshTimer !== undefined) return;
		this.refreshTimer = setTimeout(() => {
			this.refreshTimer = undefined;
			void this.refresh();
		}, REFRESH_DELAY_MS);
	}

	/** Asks the pager for the pages behind a visible range. Only missing pages are loaded. */
	requestRange(start: number, end: number): void {
		// Reading generation here makes the caller run again after the pages were dropped.
		const generation = this.generation;
		const pages = pagesToLoad({ start, end }, this.total ?? PAGE_SIZE);
		untrack(() => this.pager.requestPages(pages, generation));
	}

	gameAt(index: number): Game | undefined {
		return this.pager.gameAt(index);
	}

	selectSection(source: string): void {
		if (source === this.section) return;
		this.section = source;
		this.applyQuery();
	}

	setSearchText(text: string): void {
		this.searchText = text;
		clearTimeout(this.searchTimer);
		this.searchTimer = setTimeout(
			() => this.patchFilters({ search: text.trim() }),
			SEARCH_DELAY_MS
		);
	}

	patchFilters(patch: Partial<LibraryFilters>): void {
		this.filters = { ...this.filters, ...patch };
		this.applyQuery();
	}

	toggleTag(tag: string): void {
		const tags = this.filters.tags.includes(tag)
			? this.filters.tags.filter((item) => item !== tag)
			: [...this.filters.tags, tag];
		this.patchFilters({ tags });
	}

	resetFilters(): void {
		clearTimeout(this.searchTimer);
		this.searchText = "";
		this.patchFilters({ ...DEFAULT_FILTERS, tags: [], sort: this.filters.sort, search: "" });
	}

	saveCurrentFilter(name: string): void {
		const trimmed = name.trim();
		if (!trimmed) return;
		const entry: SavedFilter = {
			id: `${Date.now()}`,
			name: trimmed,
			filters: { ...this.filters, tags: [...this.filters.tags] }
		};
		this.savedFilters = [...this.savedFilters, entry];
		storeSavedFilters(this.savedFilters);
	}

	applySavedFilter(entry: SavedFilter): void {
		this.searchText = entry.filters.search;
		this.filters = { ...entry.filters, tags: [...entry.filters.tags] };
		this.applyQuery();
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

	private applyQuery(): void {
		this.pager.setQuery(libraryQuery(this.section, this.filters));
	}

	private applyPagerSnapshot(snapshot: PagerSnapshot): void {
		this.total = snapshot.total;
		this.generation = snapshot.generation;
		this.pageError = snapshot.error ? errorMessage(snapshot.error) : "";
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
