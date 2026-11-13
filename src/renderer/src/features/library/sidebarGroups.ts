import type {
	FilterOption,
	LibrarySource,
	SidebarGroup,
	SidebarOption
} from "../../lib/features/types";
import {
	DEFAULT_FILTERS,
	FAVORITE_OPTION,
	INSTALLED_OPTION,
	SORT_GROUP,
	SORT_LABELS,
	STATUS_GROUP,
	TAG_GROUP,
	UPDATES_OPTION
} from "./libraryFilters";
import { LOCAL_SOURCE } from "./librarySections";

const BASE_GROUP: Omit<SidebarGroup, "type" | "id" | "label" | "options"> = {
	placeholder: "",
	dynamic: false,
	searchable: false,
	orderable: false,
	default: null
};

function option(value: string, label: string, count: number | null = null): SidebarOption {
	return { value, label, icon: null, count };
}

/** The groups for local games and sources with no sidebar: sort, status toggles and tags. */
export function builtinGroups(tags: { tag: string; count: number }[]): SidebarGroup[] {
	const sort = Object.entries(SORT_LABELS).map(([value, label]) => option(value, label));
	return [
		{
			...BASE_GROUP,
			type: "sort",
			id: SORT_GROUP,
			label: "Sort",
			options: sort,
			default: { value: DEFAULT_FILTERS.sort, order: "desc" }
		},
		{
			...BASE_GROUP,
			type: "toggles",
			id: STATUS_GROUP,
			label: "Status",
			options: [
				option(INSTALLED_OPTION, "Installed"),
				option(UPDATES_OPTION, "Updates"),
				option(FAVORITE_OPTION, "Favorites")
			]
		},
		{
			...BASE_GROUP,
			type: "multi",
			id: TAG_GROUP,
			label: "Tags",
			options: tags.map((item) => option(item.tag, item.tag, item.count))
		}
	];
}

/** Plugin choices for a dynamic group, in the sidebar's option shape. */
function choicesOf(choices: FilterOption[] | undefined): SidebarOption[] {
	return (choices ?? []).map((choice) => option(choice.value, choice.label, choice.count));
}

/**
 * The groups the sidebar shows for a section. Plugin search groups are left out because the
 * sidebar draws its own search box. Dynamic groups take their choices from the loaded lists.
 */
export function sidebarGroupsFor(
	source: LibrarySource | undefined,
	tags: { tag: string; count: number }[],
	choices: Record<string, FilterOption[]>
): SidebarGroup[] {
	if (!source || source.source === LOCAL_SOURCE || source.sidebar.length === 0) {
		return builtinGroups(tags);
	}
	return source.sidebar
		.filter((group) => group.type !== "search")
		.map((group) =>
			group.dynamic ? { ...group, options: choicesOf(choices[group.id]) } : group
		);
}

/** The placeholder of the sidebar search box for a section. */
export function searchPlaceholder(source: LibrarySource | undefined): string {
	if (!source || source.source === LOCAL_SOURCE) return "Title, developer or tag";
	const search = source.sidebar.find((group) => group.type === "search");
	return search?.placeholder || "Search";
}
