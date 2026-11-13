import { describe, expect, it } from "vitest";
import type { LibrarySource, SidebarGroup } from "../../lib/features/types";
import { initialBrowseFilters } from "./browseFilters";
import { DEFAULT_FILTERS, toLibraryFilters, withLibraryFilters } from "./libraryFilters";
import { LOCAL_SOURCE } from "./librarySections";
import { builtinGroups, searchPlaceholder, sidebarGroupsFor } from "./sidebarGroups";

const TAGS = [{ tag: "Sample", count: 4 }];

const PLUGIN: LibrarySource = {
	source: "example",
	label: "Example",
	kind: "plugin",
	plugin: "example",
	capabilities: ["listGames"],
	sync: true,
	updateCheck: true,
	count: 2,
	tile: "poster",
	cover: "3:4",
	actions: [],
	card: null,
	sidebar: [
		{
			type: "search",
			id: "query",
			label: "Search",
			placeholder: "Find a game",
			options: [],
			dynamic: false,
			searchable: false,
			orderable: false,
			default: null
		},
		{
			type: "multi",
			id: "genres",
			label: "Genres",
			placeholder: "",
			options: [],
			dynamic: true,
			searchable: true,
			orderable: false,
			default: null
		}
	]
};

describe("builtinGroups", () => {
	it("has sort, status toggles and a tag list in that order", () => {
		expect(builtinGroups(TAGS).map((group) => [group.type, group.id])).toEqual([
			["sort", "sort"],
			["toggles", "status"],
			["multi", "tags"]
		]);
	});

	it("fills the tag list with the counts", () => {
		const tags = builtinGroups(TAGS).find((group) => group.id === "tags");
		expect(tags?.options).toEqual([{ value: "Sample", label: "Sample", icon: null, count: 4 }]);
	});
});

describe("sidebarGroupsFor", () => {
	it("gives local games the built-in groups", () => {
		const local = { ...PLUGIN, source: LOCAL_SOURCE, sidebar: [] };
		expect(sidebarGroupsFor(local, TAGS, {})).toEqual(builtinGroups(TAGS));
	});

	it("gives a plugin with no sidebar the built-in groups", () => {
		expect(sidebarGroupsFor({ ...PLUGIN, sidebar: [] }, TAGS, {})).toEqual(builtinGroups(TAGS));
	});

	it("drops the plugin search group, since the sidebar draws its own search box", () => {
		const groups = sidebarGroupsFor(PLUGIN, TAGS, {});
		expect(groups.map((group) => group.id)).toEqual(["genres"]);
	});

	it("fills a dynamic plugin group with the loaded choices", () => {
		const choices = { genres: [{ value: "rpg", label: "RPG", count: 3 }] };
		const genres: SidebarGroup | undefined = sidebarGroupsFor(PLUGIN, TAGS, choices)[0];
		expect(genres?.options).toEqual([{ value: "rpg", label: "RPG", icon: null, count: 3 }]);
	});
});

describe("searchPlaceholder", () => {
	it("uses the plugin search group placeholder when it has one", () => {
		expect(searchPlaceholder(PLUGIN)).toBe("Find a game");
	});

	it("names the local search in plain words", () => {
		expect(searchPlaceholder(undefined)).toBe("Title, developer or tag");
	});
});

describe("local filter conversion", () => {
	it("reads the built-in sidebar values as a local list query", () => {
		const filters = initialBrowseFilters(builtinGroups(TAGS));
		const changed = withLibraryFilters(filters, {
			...DEFAULT_FILTERS,
			search: "river",
			installed: true,
			tags: ["Sample"]
		});
		expect(toLibraryFilters(changed)).toEqual({
			search: "river",
			tags: ["Sample"],
			installed: true,
			updates: false,
			favorite: false,
			sort: "added"
		});
	});

	it("falls back to the default sort for an unknown sort name", () => {
		const filters = { ...initialBrowseFilters(builtinGroups(TAGS)), sort: "nope" };
		expect(toLibraryFilters(filters).sort).toBe("added");
	});
});
