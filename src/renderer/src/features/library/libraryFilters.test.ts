import { describe, expect, it } from "vitest";
import type { Game } from "../../lib/features/types";
import {
	DEFAULT_FILTERS,
	filterGames,
	hasActiveFilters,
	sortGames,
	visibleGames
} from "./libraryFilters";
import { parseSavedFilters } from "./savedFilters";

function makeGame(overrides: Partial<Game>): Game {
	return {
		id: 1,
		source: "itch",
		url: "https://example.test/1",
		title: "Untitled",
		version: null,
		developer: null,
		engine: null,
		tags: [],
		cover: null,
		coverLocal: null,
		screenshots: [],
		screenshotLocals: [],
		description: null,
		running: false,
		favorite: false,
		notes: null,
		installed: false,
		installedVersion: null,
		installDir: null,
		installSize: null,
		exePath: null,
		exeArgs: null,
		playtimeS: 0,
		lastPlayedAt: null,
		addedAt: "2026-01-01T00:00:00Z",
		updateAvailable: false,
		...overrides
	};
}

const games = [
	makeGame({
		id: 1,
		title: "Dust Runner",
		developer: "Acme",
		tags: ["platformer", "pixel"],
		installed: true,
		addedAt: "2026-01-02T00:00:00Z",
		lastPlayedAt: "2026-03-01T00:00:00Z"
	}),
	makeGame({
		id: 2,
		title: "Aurora",
		source: "steam",
		tags: ["puzzle"],
		favorite: true,
		updateAvailable: true,
		addedAt: "2026-01-05T00:00:00Z"
	}),
	makeGame({
		id: 3,
		title: "Bramble",
		developer: "Pixel Co",
		tags: ["platformer"],
		installed: true,
		updateAvailable: true,
		addedAt: "2026-01-03T00:00:00Z",
		lastPlayedAt: "2026-02-01T00:00:00Z"
	})
];

describe("filterGames", () => {
	it("returns every game with default filters", () => {
		expect(filterGames(games, DEFAULT_FILTERS)).toHaveLength(3);
	});

	it("matches every search word against title, developer and tags", () => {
		const result = filterGames(games, { ...DEFAULT_FILTERS, search: "  PIXEL  platformer " });
		expect(result.map((g) => g.id)).toEqual([1, 3]);
	});

	it("filters by source", () => {
		const result = filterGames(games, { ...DEFAULT_FILTERS, source: "steam" });
		expect(result.map((g) => g.id)).toEqual([2]);
	});

	it("requires all selected tags", () => {
		const result = filterGames(games, { ...DEFAULT_FILTERS, tags: ["platformer", "pixel"] });
		expect(result.map((g) => g.id)).toEqual([1]);
	});

	it("applies the installed, updates and favorite toggles", () => {
		expect(filterGames(games, { ...DEFAULT_FILTERS, installed: true }).map((g) => g.id)).toEqual([
			1, 3
		]);
		expect(filterGames(games, { ...DEFAULT_FILTERS, updates: true }).map((g) => g.id)).toEqual([
			2, 3
		]);
		expect(filterGames(games, { ...DEFAULT_FILTERS, favorite: true }).map((g) => g.id)).toEqual([
			2
		]);
	});
});

describe("sortGames", () => {
	it("sorts by title without changing the input", () => {
		const sorted = sortGames(games, "title");
		expect(sorted.map((g) => g.title)).toEqual(["Aurora", "Bramble", "Dust Runner"]);
		expect(games[0].title).toBe("Dust Runner");
	});

	it("sorts by newest added first", () => {
		expect(sortGames(games, "added").map((g) => g.id)).toEqual([2, 3, 1]);
	});

	it("sorts by last played and puts never-played games last", () => {
		expect(sortGames(games, "played").map((g) => g.id)).toEqual([1, 3, 2]);
	});
});

describe("visibleGames", () => {
	it("filters then sorts", () => {
		const result = visibleGames(games, { ...DEFAULT_FILTERS, sort: "title", installed: true });
		expect(result.map((g) => g.title)).toEqual(["Bramble", "Dust Runner"]);
	});
});

describe("hasActiveFilters", () => {
	it("ignores the sort order", () => {
		expect(hasActiveFilters({ ...DEFAULT_FILTERS, sort: "title" })).toBe(false);
	});

	it("is true when any filter narrows the list", () => {
		expect(hasActiveFilters({ ...DEFAULT_FILTERS, search: "  " })).toBe(false);
		expect(hasActiveFilters({ ...DEFAULT_FILTERS, tags: ["puzzle"] })).toBe(true);
	});
});

describe("parseSavedFilters", () => {
	it("returns an empty list for missing or broken data", () => {
		expect(parseSavedFilters(null)).toEqual([]);
		expect(parseSavedFilters("{not json")).toEqual([]);
		expect(parseSavedFilters('{"id":"x"}')).toEqual([]);
	});

	it("keeps valid entries and fills missing filter fields", () => {
		const raw = JSON.stringify([
			{ id: "a", name: "Installed", filters: { installed: true } },
			{ id: 5, name: "Bad id" }
		]);
		const parsed = parseSavedFilters(raw);
		expect(parsed).toHaveLength(1);
		expect(parsed[0].filters).toEqual({ ...DEFAULT_FILTERS, installed: true });
	});
});
