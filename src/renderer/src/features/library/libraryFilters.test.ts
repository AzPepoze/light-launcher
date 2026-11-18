import { describe, expect, it } from "vitest";
import { initialBrowseFilters } from "./browseFilters";
import {
	DEFAULT_FILTERS,
	ENGINE_GROUP,
	anyPicked,
	hasActiveFilters,
	libraryQuery,
	pluginKeysOf,
	sortOf,
	sourceValues,
	toLibraryFilters,
	withLibraryFilters,
	withoutPluginFilters,
	type LibraryFilters
} from "./libraryFilters";
import type { SidebarGroup } from "../../lib/features/types";

const PLUGIN_GROUPS: SidebarGroup[] = [
	{
		type: "sort",
		id: "sort",
		label: "Sort",
		placeholder: "",
		options: [],
		dynamic: false,
		searchable: false,
		orderable: true,
		default: null
	},
	{
		type: "select",
		id: "genre",
		label: "Genre",
		placeholder: "",
		options: [],
		dynamic: true,
		searchable: true,
		orderable: false,
		default: null
	},
	{
		type: "toggles",
		id: "adult",
		label: "Adult",
		placeholder: "",
		options: [{ value: "hide_adult", label: "Hide adult", icon: null, count: null }],
		dynamic: false,
		searchable: false,
		orderable: false,
		default: null
	}
];

describe("plugin filter keys", () => {
	it("lists the select id and the toggle option values, and skips the sort", () => {
		expect(pluginKeysOf(PLUGIN_GROUPS)).toEqual(["genre", "hide_adult"]);
	});

	it("keeps only the values of the given groups", () => {
		const values = { genre: "rpg", hide_adult: true, installed: true, other: "x" };
		expect(sourceValues(values, PLUGIN_GROUPS)).toEqual({ genre: "rpg", hide_adult: true });
	});

	it("counts a value as picked only when it narrows the list", () => {
		expect(anyPicked({ genre: "" })).toBe(false);
		expect(anyPicked({ tags: [], hide_adult: false })).toBe(false);
		expect(anyPicked({ genre: "rpg" })).toBe(true);
	});
});

describe("plugin choices in the saved filters", () => {
	const filters = initialBrowseFilters(PLUGIN_GROUPS);

	it("passes the picked plugin choices through toLibraryFilters", () => {
		const picked = { ...filters, values: { ...filters.values, genre: "rpg", hide_adult: true } };
		expect(toLibraryFilters(picked).plugin).toEqual({ genre: "rpg", hide_adult: true });
	});

	it("leaves out plugin keys that are not picked", () => {
		expect(toLibraryFilters(filters).plugin).toEqual({});
	});

	it("restores the saved plugin choices and drops the ones it does not hold", () => {
		const current = { ...filters, values: { ...filters.values, genre: "rpg" } };
		const saved: LibraryFilters = { ...DEFAULT_FILTERS, plugin: { hide_adult: true } };
		const restored = withLibraryFilters(current, saved);
		expect(restored.values.hide_adult).toBe(true);
		expect(restored.values.genre).toBeUndefined();
	});

	it("drops the plugin choices and keeps the local values", () => {
		const picked = {
			...filters,
			values: { ...filters.values, genre: "rpg", installed: true }
		};
		expect(withoutPluginFilters(picked).values).toEqual({ installed: true });
	});
});

describe("hasActiveFilters", () => {
	it("is false for the defaults", () => {
		expect(hasActiveFilters(DEFAULT_FILTERS)).toBe(false);
	});

	it("ignores a search that is only spaces", () => {
		expect(hasActiveFilters({ ...DEFAULT_FILTERS, search: "" })).toBe(false);
	});

	it.each<Partial<LibraryFilters>>([
		{ search: "dark" },
		{ tags: ["rpg"] },
		{ installed: true },
		{ updates: true },
		{ favorite: true }
	])("is true when %o is set", (change) => {
		expect(hasActiveFilters({ ...DEFAULT_FILTERS, ...change })).toBe(true);
	});

	it("ignores the sort, which does not narrow the list", () => {
		expect(hasActiveFilters({ ...DEFAULT_FILTERS, sort: "title" })).toBe(false);
	});
});

describe("engine filter", () => {
	const filters = initialBrowseFilters(PLUGIN_GROUPS);

	it("reads the picked engine into the local filters, not the plugin choices", () => {
		const picked = { ...filters, values: { ...filters.values, [ENGINE_GROUP]: "Unity" } };
		expect(toLibraryFilters(picked).engine).toBe("Unity");
		expect(toLibraryFilters(picked).plugin).toEqual({});
	});

	it("reads an empty engine when none is picked", () => {
		expect(toLibraryFilters(filters).engine).toBe("");
	});

	it("keeps the engine out of the plugin browse values", () => {
		const values = { [ENGINE_GROUP]: "Unity", hide_adult: true };
		expect(sourceValues(values, PLUGIN_GROUPS)).toEqual({ hide_adult: true });
	});

	it("drops the engine when the user leaves the source", () => {
		const picked = { ...filters, values: { ...filters.values, [ENGINE_GROUP]: "Unity", tags: ["rpg"] } };
		expect(withoutPluginFilters(picked).values).toEqual({ tags: ["rpg"] });
	});

	it("counts a picked engine as an active filter", () => {
		expect(hasActiveFilters({ ...DEFAULT_FILTERS, engine: "Unity" })).toBe(true);
	});

	it("sends the engine in the list query", () => {
		expect(libraryQuery("f95", { ...DEFAULT_FILTERS, engine: "Unity" }).engine).toBe("Unity");
	});
});

describe("libraryQuery", () => {
	it("uses the source the section gives, and the filters", () => {
		const query = libraryQuery("f95", { ...DEFAULT_FILTERS, favorite: true, updates: true });
		expect(query).toEqual({
			source: "f95",
			search: "",
			tags: [],
			engine: "",
			favorite: true,
			installed: false,
			updates: true,
			sort: "added"
		});
	});

	it("trims the search text", () => {
		expect(libraryQuery("manual", { ...DEFAULT_FILTERS, search: "  dark  " }).search).toBe("dark");
	});

	it("copies the tags so a later change does not reach the query", () => {
		const filters = { ...DEFAULT_FILTERS, tags: ["rpg"] };
		const query = libraryQuery("manual", filters);
		filters.tags.push("visual novel");
		expect(query.tags).toEqual(["rpg"]);
	});

	it("uses an empty source for every source", () => {
		expect(libraryQuery("", DEFAULT_FILTERS).source).toBe("");
	});
});

describe("sortOf", () => {
	it("keeps a known sort name", () => {
		expect(sortOf("lastPlayed")).toBe("lastPlayed");
	});

	it("falls back to recently added for an unknown or missing name", () => {
		expect(sortOf("played")).toBe("added");
		expect(sortOf(undefined)).toBe("added");
		expect(sortOf("constructor")).toBe("added");
	});
});
