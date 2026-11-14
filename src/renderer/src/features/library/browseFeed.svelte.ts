import { notifications } from "@stores/notificationStore";
import { libraryApi } from "../../lib/features/libraryApi";
import type {
	BrowseQuery,
	FilterOption,
	Game,
	SidebarGroup
} from "../../lib/features/types";
import { libraryErrorMessage } from "./libraryError";
import { PAGE_SIZE } from "./libraryWindow";

/** Typing in a group search box waits this long before asking the source. */
const SEARCH_DELAY_MS = 150;

/** The source and counter a request was sent under. */
export interface RequestStamp {
	source: string;
	epoch: number;
}

/**
 * True when an answer may still land: the feed holds the same source and counter as the request.
 * An empty source never matches, so a feed the user left takes no answers.
 */
export function isCurrentResponse(asked: RequestStamp, now: RequestStamp): boolean {
	return asked.source !== "" && asked.source === now.source && asked.epoch === now.epoch;
}

/**
 * The pages of one plugin source's browse, plus the choices of its dynamic sidebar groups.
 * Results are not saved to the library.
 */
export class BrowseFeed {
	/** The source the pages and choices belong to. Empty once the user has left it. */
	source = $state("");
	/** The query of the grid's pages, or null when the grid shows nothing from the feed. */
	query = $state<BrowseQuery | null>(null);
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
	/** Changes whenever the grid's pages are dropped, so the list asks again. */
	version = $state(0);

	/** Changes whenever the feed changes source. Choice answers from the old source are dropped. */
	private epoch = 0;
	/** Changes whenever the pages are dropped. Page answers from before are dropped. */
	private pageEpoch = 0;
	private choiceRuns: Record<string, number> = {};
	private choiceTimers: Record<string, ReturnType<typeof setTimeout>> = {};

	/** Null until the source says the count, or while more pages may follow. */
	get total(): number | null {
		if (this.count !== null) return this.count;
		if (!this.loaded) return null;
		// Without a count, one more page of placeholders keeps the grid asking while the cursor says more follow.
		return this.nextCursor === null ? this.games.length : this.games.length + PAGE_SIZE;
	}

	/** True when the grid shows this source's browse pages. */
	shows(source: string): boolean {
		return this.query !== null && this.source === source;
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
		if (source !== this.source) this.switchTo(source);
		else this.clearChoices();
		for (const group of groups) {
			if (group.dynamic) void this.loadChoices(group, "");
		}
	}

	/** Lets go of the source: drops its grid, its choices and every answer still on its way. */
	stop(): void {
		this.switchTo("");
	}

	/** Loads the first page for a query and forgets the pages before it. */
	reload(query: BrowseQuery): void {
		if (query.source !== this.source) this.switchTo(query.source);
		this.clearPages();
		this.query = query;
		void this.fetchPage("");
	}

	/** Drops the grid's pages, keeping the choices. Used when no plugin choice is picked any more. */
	clearResults(): void {
		if (this.query !== null) this.clearPages();
	}

	/** Runs the last query again, for a retry or after a plugin action changed the games. */
	reloadCurrent(): void {
		if (this.query) this.reload(this.query);
	}

	loadMore(): void {
		if (!this.query || this.nextCursor === null || this.loading) return;
		void this.fetchPage(this.nextCursor);
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

	private switchTo(source: string): void {
		this.epoch++;
		this.source = source;
		this.clearChoices();
		this.clearPages();
	}

	private clearChoices(): void {
		for (const timer of Object.values(this.choiceTimers)) clearTimeout(timer);
		this.choiceTimers = {};
		this.choices = {};
		this.choiceSearch = {};
	}

	private clearPages(): void {
		this.pageEpoch++;
		this.query = null;
		this.games = [];
		this.count = null;
		this.nextCursor = null;
		this.loaded = false;
		this.loading = false;
		this.error = "";
		this.version++;
	}

	private pageStamp(): RequestStamp {
		return { source: this.source, epoch: this.pageEpoch };
	}

	private choiceStamp(): RequestStamp {
		return { source: this.source, epoch: this.epoch };
	}

	private async fetchPage(cursor: string): Promise<void> {
		const query = this.query;
		if (!query) return;
		const asked = this.pageStamp();
		this.loading = true;
		this.error = "";
		try {
			const page = await libraryApi.browse({ ...query, cursor });
			if (!this.pageIsLive(asked)) return;
			this.games = cursor ? [...this.games, ...page.games] : page.games;
			if (page.total !== null) this.count = page.total;
			this.nextCursor = page.nextCursor;
			this.loaded = true;
		} catch (err) {
			if (!this.pageIsLive(asked)) return;
			this.error = libraryErrorMessage(err);
		} finally {
			if (this.pageIsLive(asked)) this.loading = false;
		}
	}

	private pageIsLive(asked: RequestStamp): boolean {
		return isCurrentResponse(asked, this.pageStamp());
	}

	private async loadChoices(group: SidebarGroup, query: string): Promise<void> {
		if (!this.source) return;
		const run = (this.choiceRuns[group.id] ?? 0) + 1;
		this.choiceRuns[group.id] = run;
		const asked = this.choiceStamp();
		try {
			const options = await libraryApi.filters(asked.source, group.id, query);
			if (!this.choiceIsLive(group.id, run, asked)) return;
			this.choices = { ...this.choices, [group.id]: options };
		} catch (err) {
			if (!this.choiceIsLive(group.id, run, asked)) return;
			notifications.error(libraryErrorMessage(err));
		}
	}

	/** A choice answer lands only if its group was not asked again and its source is still shown. */
	private choiceIsLive(groupId: string, run: number, asked: RequestStamp): boolean {
		return this.choiceRuns[groupId] === run && isCurrentResponse(asked, this.choiceStamp());
	}
}
