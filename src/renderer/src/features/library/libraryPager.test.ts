import { describe, expect, it, vi } from "vitest";
import type {
	Game,
	LibraryListPage,
	LibraryPageRequest,
	LibraryQuery
} from "../../lib/features/types";
import { DEFAULT_FILTERS } from "./libraryFilters";
import { LibraryPager, MAX_CACHED_PAGES, type ListPage, type PagerSnapshot } from "./libraryPager";
import { PAGE_SIZE } from "./libraryWindow";

const QUERY: LibraryQuery = {
	source: "manual",
	search: "",
	tags: [],
	favorite: false,
	installed: false,
	updates: false,
	sort: DEFAULT_FILTERS.sort
};

function makeGame(id: number): Game {
	return {
		id,
		source: "manual",
		url: "",
		title: `Game ${id}`,
		version: null,
		developer: null,
		engine: null,
		tags: [],
		cover: null,
		coverLocal: null,
		screenshots: [],
		screenshotLocals: [],
		description: null,
		running: false,
		favorite: false,
		notes: null,
		installed: false,
		installedVersion: null,
		installDir: null,
		installSize: null,
		exePath: null,
		exeArgs: null,
		playtimeS: 0,
		lastPlayedAt: null,
		addedAt: "",
		updateAvailable: false,
		origin: "plugin",
		sourceKind: "plugin",
		folderPath: null,
		scanFolder: null,
		localConfig: null,
		pinned: false
	};
}

/** A list of `total` games, cut into the pages the request asks for. */
function fakeList(total: number) {
	return vi.fn(async (_query: LibraryQuery, page: LibraryPageRequest): Promise<LibraryListPage> => {
		const first = page.offset;
		const last = Math.min(total, first + page.limit);
		const items = Array.from({ length: Math.max(0, last - first) }, (_, index) =>
			makeGame(first + index + 1)
		);
		return page.wantTotal ? { items, total } : { items };
	});
}

function deferred<T>() {
	let resolve!: (value: T) => void;
	let reject!: (reason: unknown) => void;
	const promise = new Promise<T>((done, fail) => {
		resolve = done;
		reject = fail;
	});
	return { promise, resolve, reject };
}

function makePager(listPage: ListPage) {
	const snapshots: PagerSnapshot[] = [];
	const pager = new LibraryPager({
		query: QUERY,
		listPage,
		onChange: (snapshot) => snapshots.push(snapshot)
	});
	return { pager, snapshots, last: () => snapshots[snapshots.length - 1] };
}

/** Lets the pending promise callbacks run. */
async function settle() {
	await new Promise((done) => setTimeout(done, 0));
}

describe("LibraryPager", () => {
	it("loads only the pages it is asked for, never the whole list", async () => {
		const list = fakeList(50_000);
		const { pager } = makePager(list);

		pager.requestPages([400], 0);
		await settle();

		expect(list).toHaveBeenCalledTimes(1);
		expect(list.mock.calls[0][1]).toEqual({
			offset: 400 * PAGE_SIZE,
			limit: PAGE_SIZE,
			wantTotal: false
		});
	});

	it("asks for the total on the first page and reports it", async () => {
		const list = fakeList(130);
		const { pager, last } = makePager(list);

		pager.requestPages([0], 0);
		await settle();

		expect(list.mock.calls[0][1].wantTotal).toBe(true);
		expect(last().total).toBe(130);
		expect(pager.gameAt(0)?.id).toBe(1);
	});

	it("uses the number of items when the first page sends no total", async () => {
		const pager = new LibraryPager({
			query: QUERY,
			listPage: async () => ({ items: [makeGame(1), makeGame(2)] }),
			onChange: () => {}
		});
		pager.requestPages([0], 0);
		await settle();
		expect(pager.gameAt(1)?.id).toBe(2);
	});

	it("gives undefined for a game on a page that is not loaded", () => {
		const { pager } = makePager(fakeList(500));
		expect(pager.gameAt(250)).toBeUndefined();
	});

	it("does not ask twice for a page that is still loading", async () => {
		const answer = deferred<LibraryListPage>();
		const list = vi.fn(() => answer.promise);
		const { pager } = makePager(list);

		pager.requestPages([2], 0);
		pager.requestPages([2], 0);
		expect(list).toHaveBeenCalledTimes(1);

		answer.resolve({ items: [] });
		await settle();
	});

	it("keeps at most the cache size of pages and drops the oldest", async () => {
		const list = fakeList(PAGE_SIZE * 30);
		const { pager } = makePager(list);

		for (let page = 0; page < 25; page++) pager.requestPages([page], 0);
		await settle();

		expect(pager.pages.size).toBe(MAX_CACHED_PAGES);
		expect(pager.pages.has(0)).toBe(false);
		expect(pager.pages.has(24)).toBe(true);
	});

	it("does not reload a cached page that is asked for again", async () => {
		const list = fakeList(500);
		const { pager } = makePager(list);

		pager.requestPages([0, 1], 0);
		await settle();
		pager.requestPages([0, 1], 0);
		await settle();

		expect(list).toHaveBeenCalledTimes(2);
	});

	it("drops every page and reports no total when the query changes", async () => {
		const list = fakeList(500);
		const { pager, last } = makePager(list);
		pager.requestPages([0], 0);
		await settle();

		pager.setQuery({ ...QUERY, search: "dark" });

		expect(pager.pages.size).toBe(0);
		expect(last().total).toBeNull();
		expect(last().generation).toBe(1);
	});

	it("does nothing when the query is the same", async () => {
		const list = fakeList(500);
		const { pager, snapshots } = makePager(list);
		pager.requestPages([0], 0);
		await settle();
		const before = snapshots.length;

		pager.setQuery({ ...QUERY });

		expect(snapshots.length).toBe(before);
	});

	it("ignores an answer that arrives after the query changed", async () => {
		const answer = deferred<LibraryListPage>();
		// Only the first request waits for the answer; later loads never finish in this test.
		let calls = 0;
		const { pager, last } = makePager(() =>
			calls++ === 0 ? answer.promise : new Promise<LibraryListPage>(() => {})
		);
		pager.requestPages([0], 0);

		pager.setQuery({ ...QUERY, favorite: true });
		answer.resolve({ items: [makeGame(1)], total: 1 });
		await settle();

		expect(pager.gameAt(0)).toBeUndefined();
		expect(last().total).toBeNull();
	});

	it("ignores requests made with an older generation", async () => {
		const list = fakeList(500);
		const { pager } = makePager(list);
		pager.reset();

		pager.requestPages([0], 0);
		await settle();

		// Only the reset's own first-page load runs; the stale request adds nothing.
		expect(list).toHaveBeenCalledTimes(1);
	});

	it("reports a failed page and loads it again after a reset", async () => {
		let fail = true;
		const list = vi.fn(async (): Promise<LibraryListPage> => {
			if (fail) throw new Error("offline");
			return { items: [makeGame(1)], total: 1 };
		});
		const { pager, last } = makePager(list);

		pager.requestPages([0], 0);
		await settle();
		expect(last().error).toBeInstanceOf(Error);

		fail = false;
		pager.reset();
		pager.requestPages([0], last().generation);
		await settle();

		expect(last().error).toBeNull();
		expect(pager.gameAt(0)?.id).toBe(1);
	});

	it("replaces a changed game in the cached page that holds it", async () => {
		const { pager } = makePager(fakeList(10));
		pager.requestPages([0], 0);
		await settle();

		pager.replaceGame({ ...makeGame(3), title: "Renamed" });

		expect(pager.gameAt(2)?.title).toBe("Renamed");
		expect(pager.gameAt(0)?.title).toBe("Game 1");
	});

	it("reports whether a game is in a cached page", async () => {
		const { pager } = makePager(fakeList(10));
		pager.requestPages([0], 0);
		await settle();

		expect(pager.hasGame(4)).toBe(true);
		expect(pager.hasGame(40)).toBe(false);
	});
});
