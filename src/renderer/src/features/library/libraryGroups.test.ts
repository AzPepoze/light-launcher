import { describe, expect, it } from "vitest";
import type { Game } from "../../lib/features/types";
import { groupGames, LOCAL_SOURCE } from "./libraryGroups";

function makeGame(id: number, source: string): Game {
	return {
		id,
		source,
		url: `https://example.test/${id}`,
		title: `Game ${id}`,
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
		updateAvailable: false
	};
}

const SOURCES = [
	{ source: "f95", label: "F95zone", count: 2 },
	{ source: LOCAL_SOURCE, label: LOCAL_SOURCE, count: 1 },
	{ source: "itch", label: "itch.io", count: 1 }
];

describe("groupGames", () => {
	it("puts local games first, then plugin sources in server order", () => {
		const games = [makeGame(1, "itch"), makeGame(2, "f95"), makeGame(3, LOCAL_SOURCE)];

		const groups = groupGames(games, SOURCES);

		expect(groups.map((group) => group.label)).toEqual(["Local games", "F95zone", "itch.io"]);
		expect(groups.map((group) => group.games.map((game) => game.id))).toEqual([[3], [2], [1]]);
	});

	it("hides groups without games", () => {
		const groups = groupGames([makeGame(1, "itch")], SOURCES);

		expect(groups.map((group) => group.source)).toEqual(["itch"]);
	});

	it("returns no groups for no games", () => {
		expect(groupGames([], SOURCES)).toEqual([]);
	});

	it("keeps games from a source the server did not list, labelled by its id", () => {
		const groups = groupGames([makeGame(1, "gog")], SOURCES);

		expect(groups).toHaveLength(1);
		expect(groups[0].label).toBe("gog");
	});

	it("keeps the order of games inside a group", () => {
		const games = [makeGame(5, "itch"), makeGame(2, "itch"), makeGame(9, "itch")];

		const [group] = groupGames(games, SOURCES);

		expect(group.games.map((game) => game.id)).toEqual([5, 2, 9]);
	});

	it("gives f95 games wide tiles and other groups portrait tiles", () => {
		const groups = groupGames([makeGame(1, "f95"), makeGame(2, "itch")], SOURCES);

		expect(groups.map((group) => group.shape.aspect)).toEqual(["16 / 9", "3 / 4"]);
	});
});
