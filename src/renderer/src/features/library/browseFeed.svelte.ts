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

/** A dynamic group whose unfiltered choices are being asked for, or whose last ask failed. */
export type ChoiceStatus = "loading" | "failed";

/**
 * The unfiltered choices of each source's dynamic groups, by source and group id. Kept for the whole
 * app run, so a source is not asked again on each visit.
 */
const savedChoices: Record<string, Record<string, FilterOption[]>> = {};

function saveChoices(source: string, groupId: string, options: FilterOption[]): void {
	savedChoices[source] = { ...savedChoices[source], [groupId]: options };
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
	/** Dynamic groups of the source shown that are loading or failed, by group id. */
	choiceStatus = $state<Record<string, ChoiceStatus>>({});
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

	/**
	 * Shows a source's dynamic choices. Saved choices are kept, and each dynamic group with none yet is asked once.
	 */
	showChoices(source: string, groups: SidebarGroup[]): void {
		const dynamic = groups.filter((group) => group.dynamic);
		if (dynamic.length === 0) return;
		if (source !== this.source) this.switchTo(source);
		for (const group of dynamic) {
			const askable = this.choices[group.id] === undefined && this.choiceStatus[group.id] === undefined;
			if (askable) void this.loadChoices(group, "");
		}
	}

	/** Asks the source again for one dynamic group's unfiltered choices. Only a press gets here. */
	refreshChoices(group: SidebarGroup): void {
		clearTimeout(this.choiceTimers[group.id]);
		this.choiceSearch = { ...this.choiceSearch, [group.id]: "" };
		void this.loadChoices(group, "");
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

	/** Moves the feed to another source. The saved choices of that source come back at once. */
	private switchTo(source: string): void {
		this.epoch++;
		this.source = source;
		for (const timer of Object.values(this.choiceTimers)) clearTimeout(timer);
		this.choiceTimers = {};
		this.choiceSearch = {};
		this.choiceStatus = {};
		this.choices = { ...(savedChoices[source] ?? {}) };
		this.clearPages();
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

	/** Asks for one dynamic group's choices. Only an unfiltered answer is saved, so a search never replaces them. */
	private async loadChoices(group: SidebarGroup, query: string): Promise<void> {
		if (!this.source) return;
		const run = (this.choiceRuns[group.id] ?? 0) + 1;
		this.choiceRuns[group.id] = run;
		const asked = this.choiceStamp();
		const unfiltered = query === "";
		if (unfiltered) this.setChoiceStatus(group.id, "loading");
		try {
			const options = await libraryApi.filters(asked.source, group.id, query);
			if (unfiltered && this.choiceRuns[group.id] === run) saveChoices(asked.source, group.id, options);
			if (!this.choiceIsLive(group.id, run, asked)) return;
			this.choices = { ...this.choices, [group.id]: options };
			if (unfiltered) this.setChoiceStatus(group.id, null);
		} catch (err) {
			if (!this.choiceIsLive(group.id, run, asked)) return;
			if (unfiltered) this.setChoiceStatus(group.id, "failed");
			notifications.error(libraryErrorMessage(err));
		}
	}

	private setChoiceStatus(groupId: string, status: ChoiceStatus | null): void {
		const next = { ...this.choiceStatus };
		if (status) next[groupId] = status;
		else delete next[groupId];
		this.choiceStatus = next;
	}

	/** A choice answer lands only if its group was not asked again and its source is still shown. */
	private choiceIsLive(groupId: string, run: number, asked: RequestStamp): boolean {
		return this.choiceRuns[groupId] === run && isCurrentResponse(asked, this.choiceStamp());
	}
}
