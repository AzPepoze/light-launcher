import type { LibraryQuery, LibrarySort } from "../../lib/features/types";

export interface LibraryFilters {
	search: string;
	tags: string[];
	installed: boolean;
	updates: boolean;
	favorite: boolean;
	sort: LibrarySort;
}

export const DEFAULT_FILTERS: LibraryFilters = {
	search: "",
	tags: [],
	installed: false,
	updates: false,
	favorite: false,
	sort: "added"
};

export const SORT_LABELS: Record<LibrarySort, string> = {
	title: "Title",
	added: "Recently added",
	lastPlayed: "Recently played"
};

const SORTS = Object.keys(SORT_LABELS) as LibrarySort[];

/** Keeps a sort name only when the library knows it. Saved filters can hold older names. */
export function sortOf(value: unknown): LibrarySort {
	return SORTS.find((sort) => sort === value) ?? DEFAULT_FILTERS.sort;
}

export function hasActiveFilters(filters: LibraryFilters): boolean {
	return (
		filters.search !== "" ||
		filters.tags.length > 0 ||
		filters.installed ||
		filters.updates ||
		filters.favorite
	);
}

/** The list request for one source with the current filters. An empty source means every source. */
export function libraryQuery(source: string, filters: LibraryFilters): LibraryQuery {
	return {
		source,
		search: filters.search.trim(),
		tags: [...filters.tags],
		favorite: filters.favorite,
		installed: filters.installed,
		updates: filters.updates,
		sort: filters.sort
	};
}
