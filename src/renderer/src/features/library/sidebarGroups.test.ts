import { describe, expect, it } from "vitest";
import type { LibrarySource, SidebarGroup } from "../../lib/features/types";
import { initialBrowseFilters } from "./browseFilters";
import {
	DEFAULT_FILTERS,
	ENGINE_GROUP,
	TAG_GROUP,
	toLibraryFilters,
	withLibraryFilters
} from "./libraryFilters";
import { LOCAL_SOURCE } from "./librarySections";
import {
	builtinGroups,
	choiceGroupsOf,
	ownedKindOf,
	pluginGroupsOf,
	searchPlaceholder,
	sidebarGroupsFor
} from "./sidebarGroups";

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
	engines: [],
	tags: [],
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

	it("sorts ascending by default, like the plugin sort groups", () => {
		const sort = builtinGroups(TAGS).find((group) => group.type === "sort");
		expect(sort?.default?.order).toBe("asc");
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

const OWNED_PLUGIN: LibrarySource = {
	...PLUGIN,
	engines: [
		{ name: "Unity", count: 9 },
		{ name: "RPG Maker", count: 2 }
	],
	tags: [{ name: "Fantasy", count: 5 }],
	sidebar: [
		{
			type: "select",
			id: "engine",
			label: "Engine",
			placeholder: "",
			options: [],
			dynamic: true,
			searchable: true,
			orderable: false,
			default: null
		},
		{
			type: "multi",
			id: "tag-list",
			label: "Tags",
			placeholder: "",
			options: [],
			dynamic: true,
			searchable: true,
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

describe("owned engine and tags groups", () => {
	it("matches the engine and tags groups by id or label", () => {
		const kinds = OWNED_PLUGIN.sidebar.map(ownedKindOf);
		expect(kinds).toEqual(["engine", "tags", null]);
	});

	it("gives the engine group one choice per owned engine, with its count", () => {
		const engine = sidebarGroupsFor(OWNED_PLUGIN, TAGS, {}).find((group) => group.id === ENGINE_GROUP);
		expect(engine?.type).toBe("single");
		expect(engine?.dynamic).toBe(false);
		expect(engine?.options).toEqual([
			{ value: "Unity", label: "Unity", icon: null, count: 9 },
			{ value: "RPG Maker", label: "RPG Maker", icon: null, count: 2 }
		]);
	});

	it("gives the tags group the owned tags, with their counts", () => {
		const tags = sidebarGroupsFor(OWNED_PLUGIN, TAGS, {}).find((group) => group.id === TAG_GROUP);
		expect(tags?.type).toBe("multi");
		expect(tags?.options).toEqual([{ value: "Fantasy", label: "Fantasy", icon: null, count: 5 }]);
	});

	it("leaves the other plugin groups on their loaded choices", () => {
		const choices = { genres: [{ value: "rpg", label: "RPG", count: 3 }] };
		const genres = sidebarGroupsFor(OWNED_PLUGIN, TAGS, choices).find((group) => group.id === "genres");
		expect(genres?.options).toEqual([{ value: "rpg", label: "RPG", icon: null, count: 3 }]);
	});

	it("asks the plugin only for the groups it owns none of", () => {
		expect(choiceGroupsOf(OWNED_PLUGIN).map((group) => group.id)).toEqual(["genres"]);
		expect(pluginGroupsOf(OWNED_PLUGIN).map((group) => group.id)).toEqual(["genres"]);
	});

	it("has no options when the source owns no engines or tags", () => {
		const empty = { ...OWNED_PLUGIN, engines: [], tags: [] };
		const groups = sidebarGroupsFor(empty, TAGS, {});
		expect(groups.find((group) => group.id === ENGINE_GROUP)?.options).toEqual([]);
		expect(groups.find((group) => group.id === TAG_GROUP)?.options).toEqual([]);
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
			engine: "",
			installed: true,
			updates: false,
			favorite: false,
			sort: "added",
			plugin: {}
		});
	});

	it("falls back to the default sort for an unknown sort name", () => {
		const filters = { ...initialBrowseFilters(builtinGroups(TAGS)), sort: "nope" };
		expect(toLibraryFilters(filters).sort).toBe("added");
	});
});

describe("choiceGroupsOf", () => {
	it("asks for nothing on the local list or a source with no sidebar", () => {
		expect(choiceGroupsOf(undefined)).toEqual([]);
		expect(choiceGroupsOf({ ...PLUGIN, source: LOCAL_SOURCE })).toEqual([]);
		expect(choiceGroupsOf({ ...PLUGIN, sidebar: [] })).toEqual([]);
	});

	it("asks only for the dynamic groups of a plugin source", () => {
		expect(choiceGroupsOf(PLUGIN).map((group) => group.id)).toEqual(["genres"]);
	});
});
