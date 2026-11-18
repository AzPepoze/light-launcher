import type {
	BrowseFilterValue,
	LibraryQuery,
	LibrarySort,
	SidebarGroup
} from "../../lib/features/types";
import type { BrowseFilters } from "./browseFilters";

/** Ids of the built-in sidebar groups, used for local games and sources with no sidebar. */
export const SORT_GROUP = "sort";
export const STATUS_GROUP = "status";
export const TAG_GROUP = "tags";
/** The owned-games engine group of a plugin source. The launcher builds it, not the plugin. */
export const ENGINE_GROUP = "engine";
export const INSTALLED_OPTION = "installed";
export const UPDATES_OPTION = "updates";
export const FAVORITE_OPTION = "favorite";

/** The sidebar values the local list reads. Dropped when the user leaves a source. */
const BUILTIN_VALUE_KEYS = [INSTALLED_OPTION, UPDATES_OPTION, FAVORITE_OPTION, TAG_GROUP];
/** Every value the launcher reads itself. These never go to a plugin's browse. */
const LAUNCHER_VALUE_KEYS = [...BUILTIN_VALUE_KEYS, ENGINE_GROUP];

export interface LibraryFilters {
	search: string;
	tags: string[];
	/** An engine of the owned games, or an empty string for none. */
	engine: string;
	installed: boolean;
	updates: boolean;
	favorite: boolean;
	sort: LibrarySort;
	/** The plugin choices picked, by the plugin's filter key. The local list does not read them. */
	plugin: Record<string, BrowseFilterValue>;
}

export const DEFAULT_FILTERS: LibraryFilters = {
	search: "",
	tags: [],
	engine: "",
	installed: false,
	updates: false,
	favorite: false,
	sort: "added",
	plugin: {}
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
		filters.engine !== "" ||
		filters.installed ||
		filters.updates ||
		filters.favorite
	);
}

/** True when a sidebar value narrows the list: a toggle on, a choice picked, or some tags. */
export function isPicked(value: BrowseFilterValue): boolean {
	if (value === true) return true;
	if (typeof value === "string") return value !== "";
	return Array.isArray(value) && value.length > 0;
}

/** True when any of the values is picked. */
export function anyPicked(values: Record<string, BrowseFilterValue>): boolean {
	return Object.values(values).some(isPicked);
}

function tagsOf(filters: BrowseFilters): string[] {
	const tags = filters.values[TAG_GROUP];
	return Array.isArray(tags) ? [...tags] : [];
}

function engineOf(filters: BrowseFilters): string {
	const engine = filters.values[ENGINE_GROUP];
	return typeof engine === "string" ? engine : "";
}

function builtinValuesOf(values: Record<string, BrowseFilterValue>): Record<string, BrowseFilterValue> {
	return Object.fromEntries(Object.entries(values).filter(([key]) => BUILTIN_VALUE_KEYS.includes(key)));
}

function pluginValuesOf(values: Record<string, BrowseFilterValue>): Record<string, BrowseFilterValue> {
	return Object.fromEntries(
		Object.entries(values).filter(([key, value]) => !LAUNCHER_VALUE_KEYS.includes(key) && isPicked(value))
	);
}

/** The filter keys a source's sidebar sends to its plugin. Toggles send their option values. */
export function pluginKeysOf(groups: SidebarGroup[]): string[] {
	return groups.flatMap((group) => {
		if (group.type === "search" || group.type === "sort") return [];
		if (group.type === "toggles") return group.options.map((option) => option.value);
		return [group.id];
	});
}

/** The values that belong to one source's sidebar groups. Other sources' values are left out. */
export function sourceValues(
	values: Record<string, BrowseFilterValue>,
	groups: SidebarGroup[]
): Record<string, BrowseFilterValue> {
	const keys = pluginKeysOf(groups);
	return Object.fromEntries(Object.entries(values).filter(([key]) => keys.includes(key)));
}

/** Drops the plugin choices, which belong to the source they were picked on. Keeps the local values. */
export function withoutPluginFilters(filters: BrowseFilters): BrowseFilters {
	return { ...filters, values: builtinValuesOf(filters.values) };
}

/** The local list filters that the shared sidebar state holds, plus the plugin choices picked. */
export function toLibraryFilters(filters: BrowseFilters): LibraryFilters {
	return {
		search: filters.query,
		tags: tagsOf(filters),
		engine: engineOf(filters),
		installed: filters.values[INSTALLED_OPTION] === true,
		updates: filters.values[UPDATES_OPTION] === true,
		favorite: filters.values[FAVORITE_OPTION] === true,
		sort: sortOf(filters.sort),
		plugin: pluginValuesOf(filters.values)
	};
}

/** Writes saved filters into the shared sidebar state, replacing the values it had. */
export function withLibraryFilters(filters: BrowseFilters, saved: LibraryFilters): BrowseFilters {
	return {
		...filters,
		query: saved.search,
		sort: saved.sort,
		values: {
			...builtinValuesOf(filters.values),
			[INSTALLED_OPTION]: saved.installed,
			[UPDATES_OPTION]: saved.updates,
			[FAVORITE_OPTION]: saved.favorite,
			[TAG_GROUP]: [...saved.tags],
			[ENGINE_GROUP]: saved.engine,
			...saved.plugin
		}
	};
}

/** The list request for one source with the current filters. An empty source means every source. */
export function libraryQuery(source: string, filters: LibraryFilters): LibraryQuery {
	return {
		source,
		search: filters.search.trim(),
		tags: [...filters.tags],
		engine: filters.engine,
		favorite: filters.favorite,
		installed: filters.installed,
		updates: filters.updates,
		sort: filters.sort
	};
}
