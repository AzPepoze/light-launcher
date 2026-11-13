import { describe, expect, it } from "vitest";
import type { Game, LibrarySource } from "../../lib/features/types";
import {
	ALL_SOURCES,
	gamesOfSource,
	historyOfSource,
	updateSourceTabs,
	type UpdateHistoryItem
} from "./updatesView";

function game(id: number, source: string): Game {
	return { id, source, title: `Game ${id}` } as Game;
}

function librarySource(source: string, label: string): LibrarySource {
	return { source, label } as LibrarySource;
}

function historyRow(gameId: number): UpdateHistoryItem {
	return {
		id: gameId,
		gameId,
		title: `Game ${gameId}`,
		checkedAt: "2026-01-01T00:00:00Z",
		from: null,
		to: null,
		result: "update",
		error: null,
		coverPath: null,
		coverUrl: null,
		source: "f95"
	};
}

describe("gamesOfSource", () => {
	const games = [game(1, "alpha"), game(2, "beta"), game(3, "alpha")];

	it("keeps every game for the all-sources tab", () => {
		expect(gamesOfSource(games, ALL_SOURCES)).toHaveLength(3);
	});

	it("keeps only the games of one source", () => {
		expect(gamesOfSource(games, "alpha").map((item) => item.id)).toEqual([1, 3]);
	});
});

describe("updateSourceTabs", () => {
	const sources = [
		librarySource("beta", "Beta Store"),
		librarySource("local", "Local"),
		librarySource("alpha", "Alpha Hub")
	];
	const games = [game(1, "alpha"), game(2, "beta"), game(3, "alpha")];

	it("starts with All and adds only the sources that can check for updates, by label", () => {
		const tabs = updateSourceTabs(sources, ["alpha", "beta"], games);
		expect(tabs.map((tab) => tab.id)).toEqual([ALL_SOURCES, "alpha", "beta"]);
		expect(tabs.map((tab) => tab.label)).toEqual(["All", "Alpha Hub", "Beta Store"]);
	});

	it("badges each tab with its number of games with updates", () => {
		const tabs = updateSourceTabs(sources, ["alpha", "beta"], games);
		expect(tabs.map((tab) => tab.badge)).toEqual([3, 2, 1]);
	});

	it("shows only All when no source can check for updates", () => {
		expect(updateSourceTabs(sources, [], games)).toEqual([{ id: ALL_SOURCES, label: "All", badge: 3 }]);
	});
});

describe("historyOfSource", () => {
	const rows = [historyRow(1), historyRow(2), historyRow(3)];

	it("keeps every row when no source is chosen", () => {
		expect(historyOfSource(rows, null)).toBe(rows);
	});

	it("keeps the rows of the games in the chosen source", () => {
		expect(historyOfSource(rows, new Set([2, 3])).map((row) => row.gameId)).toEqual([2, 3]);
	});
});
