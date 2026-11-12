import type { BrowseFilterValue, BrowseQuery, SidebarGroup } from "../../lib/features/types";

/** The host caps a browse page at 100; 60 fills a wide grid without a huge answer. */
export const BROWSE_PAGE_SIZE = 60;

export interface BrowseFilters {
	/** The search group's text. */
	query: string;
	/** Values sent under the plugin's filter keys. */
	values: Record<string, BrowseFilterValue>;
	sort: string;
	order: "asc" | "desc";
}

/** Starts empty, except a sort group's declared default. Toggle options start off. */
export function initialBrowseFilters(groups: SidebarGroup[]): BrowseFilters {
	const values: Record<string, BrowseFilterValue> = {};
	for (const group of groups) {
		if (group.type === "toggles") {
			for (const option of group.options) values[option.value] = false;
		} else if (group.type === "select") {
			values[group.id] = "";
		} else if (group.type === "multi") {
			values[group.id] = [];
		}
	}
	const sortGroup = groups.find((group) => group.type === "sort");
	const sort = sortGroup?.default?.value || sortGroup?.options[0]?.value || "";
	return { query: "", values, sort, order: sortGroup?.default?.order ?? "asc" };
}

function withValue(filters: BrowseFilters, id: string, value: BrowseFilterValue): BrowseFilters {
	return { ...filters, values: { ...filters.values, [id]: value } };
}

export function setBrowseQuery(filters: BrowseFilters, query: string): BrowseFilters {
	return { ...filters, query };
}

export function setToggle(filters: BrowseFilters, id: string, on: boolean): BrowseFilters {
	return withValue(filters, id, on);
}

export function setSelect(filters: BrowseFilters, id: string, value: string): BrowseFilters {
	return withValue(filters, id, value);
}

export function setMulti(filters: BrowseFilters, id: string, values: string[]): BrowseFilters {
	return withValue(filters, id, [...values]);
}

export function setSort(filters: BrowseFilters, sort: string): BrowseFilters {
	return { ...filters, sort };
}

export function setOrder(filters: BrowseFilters, order: "asc" | "desc"): BrowseFilters {
	return { ...filters, order };
}

/** Sends only the values that narrow the list; empty ones fall back to the source's own default. */
export function browseQuery(source: string, filters: BrowseFilters, cursor = ""): BrowseQuery {
	const active: Record<string, BrowseFilterValue> = {};
	for (const [id, value] of Object.entries(filters.values)) {
		if (value === true) active[id] = true;
		else if (typeof value === "string" && value !== "") active[id] = value;
		else if (Array.isArray(value) && value.length > 0) active[id] = [...value];
	}
	return {
		source,
		query: filters.query.trim(),
		filters: active,
		sort: filters.sort,
		order: filters.sort ? filters.order : "",
		cursor,
		limit: BROWSE_PAGE_SIZE
	};
}

export function hasBrowseFilters(filters: BrowseFilters): boolean {
	if (filters.query.trim() !== "") return true;
	return Object.values(filters.values).some(
		(value) =>
			value === true ||
			(typeof value === "string" && value !== "") ||
			(Array.isArray(value) && value.length > 0)
	);
}
