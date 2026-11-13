import { describe, expect, it, vi } from "vitest";
import type { LibraryQuery } from "../../lib/features/types";
import { DEFAULT_FILTERS } from "./libraryFilters";
import { LibraryPager, type ListPage, type PagerSnapshot } from "./libraryPager";

const QUERY: LibraryQuery = {
	source: "example",
	search: "",
	tags: [],
	favorite: DEFAULT_FILTERS.favorite,
	installed: DEFAULT_FILTERS.installed,
	updates: DEFAULT_FILTERS.updates,
	sort: DEFAULT_FILTERS.sort
};

/** Lets pending promise callbacks run. */
function settle(): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, 0));
}

describe("LibraryPager first page", () => {
	it("loads page 0 with the total when the query changes, without any grid request", async () => {
		const listPage = vi.fn<ListPage>(async () => ({ items: [], total: 0 }));
		const snapshots: PagerSnapshot[] = [];
		const pager = new LibraryPager({
			query: QUERY,
			listPage,
			onChange: (snapshot) => snapshots.push(snapshot)
		});

		pager.setQuery({ ...QUERY, source: "other" });
		await settle();

		expect(listPage).toHaveBeenCalledTimes(1);
		expect(listPage.mock.calls[0]?.[0].source).toBe("other");
		expect(listPage.mock.calls[0]?.[1]).toMatchObject({ offset: 0, wantTotal: true });
		expect(snapshots.at(-1)?.total).toBe(0);
	});

	it("reports the total once the first page arrives", async () => {
		const listPage = vi.fn<ListPage>(async () => ({ items: [], total: 39 }));
		const snapshots: PagerSnapshot[] = [];
		const pager = new LibraryPager({
			query: QUERY,
			listPage,
			onChange: (snapshot) => snapshots.push(snapshot)
		});

		pager.setQuery({ ...QUERY, search: "river" });
		await settle();

		expect(snapshots.at(-1)?.total).toBe(39);
	});

	it("does not load again for the same query", async () => {
		const listPage = vi.fn<ListPage>(async () => ({ items: [], total: 0 }));
		const pager = new LibraryPager({ query: QUERY, listPage, onChange: () => {} });

		pager.setQuery({ ...QUERY });
		await settle();

		expect(listPage).not.toHaveBeenCalled();
	});
});
