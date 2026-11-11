import { describe, expect, it } from "vitest";
import {
	DEFAULT_FILTERS,
	hasActiveFilters,
	libraryQuery,
	sortOf,
	type LibraryFilters
} from "./libraryFilters";

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

describe("libraryQuery", () => {
	it("uses the source the section gives, and the filters", () => {
		const query = libraryQuery("f95", { ...DEFAULT_FILTERS, favorite: true, updates: true });
		expect(query).toEqual({
			source: "f95",
			search: "",
			tags: [],
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
