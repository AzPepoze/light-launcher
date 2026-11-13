import type { LibraryQuery, LibrarySort } from "../../lib/features/types";
import type { BrowseFilters } from "./browseFilters";

/** Ids of the built-in sidebar groups, used for local games and sources with no sidebar. */
export const SORT_GROUP = "sort";
export const STATUS_GROUP = "status";
export const TAG_GROUP = "tags";
export const INSTALLED_OPTION = "installed";
export const UPDATES_OPTION = "updates";
export const FAVORITE_OPTION = "favorite";

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

function tagsOf(filters: BrowseFilters): string[] {
	const tags = filters.values[TAG_GROUP];
	return Array.isArray(tags) ? [...tags] : [];
}

/** The local list filters that the shared sidebar state holds. */
export function toLibraryFilters(filters: BrowseFilters): LibraryFilters {
	return {
		search: filters.query,
		tags: tagsOf(filters),
		installed: filters.values[INSTALLED_OPTION] === true,
		updates: filters.values[UPDATES_OPTION] === true,
		favorite: filters.values[FAVORITE_OPTION] === true,
		sort: sortOf(filters.sort)
	};
}

/** Writes saved local filters into the shared sidebar state. */
export function withLibraryFilters(filters: BrowseFilters, saved: LibraryFilters): BrowseFilters {
	return {
		...filters,
		query: saved.search,
		sort: saved.sort,
		values: {
			...filters.values,
			[INSTALLED_OPTION]: saved.installed,
			[UPDATES_OPTION]: saved.updates,
			[FAVORITE_OPTION]: saved.favorite,
			[TAG_GROUP]: [...saved.tags]
		}
	};
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
