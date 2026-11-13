import { notifications } from "@stores/notificationStore";
import { libraryApi } from "../../lib/features/libraryApi";
import type { BrowseQuery, FilterOption, Game, SidebarGroup } from "../../lib/features/types";
import { libraryErrorMessage } from "./libraryError";

/** Typing in a group search box waits this long before asking the source. */
const SEARCH_DELAY_MS = 150;

/**
 * The pages of one plugin source's browse, plus the choices of its dynamic sidebar groups.
 * Results are not saved to the library.
 */
export class BrowseFeed {
	games = $state<Game[]>([]);
	/** The count the source sent, when it sent one. */
	private count = $state<number | null>(null);
	loaded = $state(false);
	nextCursor = $state<string | null>(null);
	loading = $state(false);
	error = $state("");
	/** Choices of each dynamic group, by group id. */
	choices = $state<Record<string, FilterOption[]>>({});
	/** The text in each dynamic group's search box, by group id. */
	choiceSearch = $state<Record<string, string>>({});

	private source = "";
	private lastQuery: BrowseQuery | null = null;
	private generation = 0;
	private choiceRuns: Record<string, number> = {};
	private choiceTimers: Record<string, ReturnType<typeof setTimeout>> = {};

	/** Null until the source says the count, or while more pages may follow. */
	get total(): number | null {
		if (this.count !== null) return this.count;
		return this.loaded && this.nextCursor === null ? this.games.length : null;
	}

	gameAt(index: number): Game | undefined {
		return this.games[index];
	}

	/** Loads the next page when the grid scrolls past the games loaded so far. */
	requestRange(_start: number, end: number): void {
		if (end > this.games.length) this.loadMore();
	}

	/** Starts a source's browse: clears the old choices and loads the dynamic ones. */
	start(source: string, groups: SidebarGroup[]): void {
		this.stop();
		this.source = source;
		this.lastQuery = null;
		this.choices = {};
		this.choiceSearch = {};
		for (const group of groups) {
			if (group.dynamic) void this.loadChoices(group, "");
		}
	}

	/** Drops pending searches. Called when the source changes. */
	stop(): void {
		for (const timer of Object.values(this.choiceTimers)) clearTimeout(timer);
		this.choiceTimers = {};
	}

	/** Loads the first page for a query and forgets the pages before it. */
	reload(query: BrowseQuery): void {
		this.lastQuery = query;
		this.games = [];
		this.count = null;
		this.nextCursor = null;
		this.loaded = false;
		void this.fetchPage("");
	}

	/** Runs the last query again, for a retry or after a plugin action changed the games. */
	reloadCurrent(): void {
		if (this.lastQuery) this.reload(this.lastQuery);
	}

	loadMore(): void {
		if (this.nextCursor === null || this.loading) return;
		void this.fetchPage(this.nextCursor);
	}

	private async fetchPage(cursor: string): Promise<void> {
		if (!this.lastQuery) return;
		const run = ++this.generation;
		this.loading = true;
		this.error = "";
		try {
			const page = await libraryApi.browse({ ...this.lastQuery, cursor });
			if (run !== this.generation) return;
			this.games = cursor ? [...this.games, ...page.games] : page.games;
			if (page.total !== null) this.count = page.total;
			this.nextCursor = page.nextCursor;
			this.loaded = true;
		} catch (err) {
			if (run !== this.generation) return;
			this.error = libraryErrorMessage(err);
		} finally {
			if (run === this.generation) this.loading = false;
		}
	}

	/** Searches inside one dynamic group, after a short pause in typing. */
	searchChoices(group: SidebarGroup, text: string): void {
		this.choiceSearch = { ...this.choiceSearch, [group.id]: text };
		clearTimeout(this.choiceTimers[group.id]);
		this.choiceTimers = {
			...this.choiceTimers,
			[group.id]: setTimeout(() => void this.loadChoices(group, text), SEARCH_DELAY_MS)
		};
	}

	private async loadChoices(group: SidebarGroup, query: string): Promise<void> {
		const run = (this.choiceRuns[group.id] ?? 0) + 1;
		this.choiceRuns[group.id] = run;
		const source = this.source;
		try {
			const options = await libraryApi.filters(source, group.id, query);
			if (this.choiceRuns[group.id] !== run) return;
			this.choices = { ...this.choices, [group.id]: options };
		} catch (err) {
			if (this.choiceRuns[group.id] !== run) return;
			notifications.error(libraryErrorMessage(err));
		}
	}
}
