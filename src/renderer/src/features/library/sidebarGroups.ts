import type {
	FilterOption,
	LibrarySource,
	NameCount,
	SidebarGroup,
	SidebarOption
} from "../../lib/features/types";
import {
	DEFAULT_FILTERS,
	ENGINE_GROUP,
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
			default: { value: DEFAULT_FILTERS.sort, order: "asc" }
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

/** A plugin group the launcher fills from the owned games: the engine list or the tags. */
type OwnedKind = "engine" | "tags";

/**
 * The plugin's engine or tags group, matched by id or by label. Null for every other group.
 * Only select and multi groups can be owned.
 */
export function ownedKindOf(group: SidebarGroup): OwnedKind | null {
	if (group.type !== "select" && group.type !== "multi") return null;
	const label = group.label.toLowerCase();
	if (group.id === ENGINE_GROUP || label === "engine") return "engine";
	if (group.id === TAG_GROUP || label === "tags") return "tags";
	return null;
}

/** The plugin's sidebar groups the plugin answers for. The owned engine and tags groups are left out. */
export function pluginGroupsOf(source: LibrarySource): SidebarGroup[] {
	return source.sidebar.filter((group) => ownedKindOf(group) === null);
}

function ownedOptions(items: NameCount[]): SidebarOption[] {
	return items.map((item) => option(item.name, item.name, item.count));
}

/** The owned group: one engine choice, or several tags. Its options come from the source's owned games. */
function ownedGroup(group: SidebarGroup, kind: OwnedKind, source: LibrarySource): SidebarGroup {
	if (kind === "engine") {
		return {
			...group,
			type: "single",
			id: ENGINE_GROUP,
			dynamic: false,
			searchable: true,
			options: ownedOptions(source.engines)
		};
	}
	return {
		...group,
		type: "multi",
		id: TAG_GROUP,
		dynamic: false,
		searchable: true,
		options: ownedOptions(source.tags)
	};
}

function sidebarGroupOf(
	group: SidebarGroup,
	source: LibrarySource,
	choices: Record<string, FilterOption[]>
): SidebarGroup {
	const kind = ownedKindOf(group);
	if (kind) return ownedGroup(group, kind, source);
	return group.dynamic ? { ...group, options: choicesOf(choices[group.id]) } : group;
}

/**
 * The groups the sidebar shows for a section. Plugin search groups are left out because the
 * sidebar draws its own search box. Dynamic groups take their choices from the loaded lists.
 * The engine and tags groups of a plugin come from its owned games.
 */
export function sidebarGroupsFor(
	source: LibrarySource | undefined,
	tags: { tag: string; count: number }[],
	choices: Record<string, FilterOption[]>
): SidebarGroup[] {
	if (!source || source.source === LOCAL_SOURCE || source.sidebar.length === 0) {
		return builtinGroups(tags);
	}
	const groups = source.sidebar
		.filter((group) => group.type !== "search")
		.map((group) => sidebarGroupOf(group, source, choices));
	return hasTagsGroup(source) ? groups : [...groups, addedTagsGroup(source)];
}

function hasTagsGroup(source: LibrarySource): boolean {
	return source.sidebar.some((group) => ownedKindOf(group) === "tags");
}

/** The tags group the launcher adds to the end of a plugin sidebar that declares none. */
function addedTagsGroup(source: LibrarySource): SidebarGroup {
	return {
		...BASE_GROUP,
		type: "multi",
		id: TAG_GROUP,
		label: "Tags",
		searchable: true,
		options: ownedOptions(source.tags)
	};
}

/**
 * The dynamic groups whose choices the plugin must be asked for. Empty for local games and for
 * sources with no sidebar, so nothing is requested for them.
 */
export function choiceGroupsOf(source: LibrarySource | undefined): SidebarGroup[] {
	if (!source || source.source === LOCAL_SOURCE || source.sidebar.length === 0) return [];
	return pluginGroupsOf(source).filter((group) => group.dynamic);
}

/** The placeholder of the sidebar search box for a section. */
export function searchPlaceholder(source: LibrarySource | undefined): string {
	if (!source || source.source === LOCAL_SOURCE) return "Title, developer or tag";
	const search = source.sidebar.find((group) => group.type === "search");
	return search?.placeholder || "Search";
}
