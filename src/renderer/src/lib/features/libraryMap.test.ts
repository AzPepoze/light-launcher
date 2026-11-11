import { describe, expect, it } from "vitest";
import {
	hasNewVersion,
	htmlToText,
	toGame,
	toGameDetail,
	toGameUpdate,
	toIncomingList,
	toLibraryListing,
	toPreview,
	toSearchResponse,
	toUpdatesSummary
} from "./libraryMap";
import type { Game } from "./types";
import type { WireGame, WireGameDetail, WireLibraryPage } from "./wire";

function makeWire(overrides: Partial<WireGame> = {}): WireGame {
	return {
		id: 7,
		source: "itch",
		sourceId: "abc",
		url: "https://example.test/7",
		title: "Dust Runner",
		version: "1.2",
		developer: "Studio",
		engine: null,
		tags: ["platformer"],
		coverPath: "/data/covers/7.jpg",
		coverUrl: "https://example.test/7.jpg",
		coverLocal: "covers/7.jpg",
		screenshots: null,
		description: "<p>First line</p><p>Second &amp; last</p>",
		descriptionText: "First line\n\nSecond & last",
		changelog: null,
		rating: null,
		status: "installed",
		running: false,
		favorite: 1,
		hidden: 0,
		notes: "note",
		installedVersion: "1.2",
		installDir: "/games/dust",
		installSize: 500,
		exePath: "/games/dust/run.exe",
		exeArgs: null,
		lastPlayedAt: null,
		playtimeS: 120,
		addedAt: "2026-01-01T00:00:00Z",
		updatedAt: "2026-01-02T00:00:00Z",
		...overrides
	};
}

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

describe("htmlToText", () => {
	it("turns paragraphs into line breaks and decodes entities", () => {
		expect(htmlToText("<p>First</p><p>Tom &amp; Jerry</p>")).toBe("First\n\nTom & Jerry");
	});

	it("returns null for empty input", () => {
		expect(htmlToText(null)).toBeNull();
		expect(htmlToText("   ")).toBeNull();
	});
});

describe("hasNewVersion", () => {
	it("matches the Go updates filter", () => {
		expect(hasNewVersion("2.0", "1.0")).toBe(true);
		expect(hasNewVersion("1.0", "1.0")).toBe(false);
		expect(hasNewVersion("2.0", null)).toBe(false);
		expect(hasNewVersion(null, "1.0")).toBe(false);
	});
});

describe("toGame", () => {
	it("maps the Go game to the view type", () => {
		const game = toGame(makeWire());
		expect(game).toMatchObject({
			id: 7,
			cover: "https://example.test/7.jpg",
			coverLocal: "covers/7.jpg",
			description: "First line\n\nSecond & last",
			running: false,
			favorite: true,
			installed: true,
			installDir: "/games/dust",
			updateAvailable: false
		});
	});

	it("takes the description text from the API, not from the HTML", () => {
		const game = toGame(makeWire({ descriptionText: "Plain only", description: "<b>ignored</b>" }));
		expect(game.description).toBe("Plain only");
	});

	it("reports a running game", () => {
		expect(toGame(makeWire({ running: true })).running).toBe(true);
	});

	it("uses empty lists for null tags and screenshots", () => {
		const game = toGame(makeWire({ tags: null, screenshots: null }));
		expect(game.tags).toEqual([]);
		expect(game.screenshots).toEqual([]);
	});

	it("is not installed without an install folder", () => {
		const game = toGame(makeWire({ installDir: null, installedVersion: null }));
		expect(game.installed).toBe(false);
	});

	it("flags a game whose stored version differs from the installed one", () => {
		const game = toGame(makeWire({ version: "1.3", installedVersion: "1.2" }));
		expect(game.updateAvailable).toBe(true);
	});

	it("does not use the local cover path, which has no URL", () => {
		expect(toGame(makeWire({ coverUrl: null })).cover).toBeNull();
	});
});

function makePage(overrides: Partial<WireLibraryPage> = {}): WireLibraryPage {
	return {
		items: [],
		total: 0,
		sources: [],
		tags: [],
		stats: { total: 0, installed: 0, updates: 0, bytes: 0, playtime: 0 },
		...overrides
	};
}

describe("toLibraryListing", () => {
	it("uses the counts the API sent for the whole library", () => {
		const games = [makeGame({ id: 1 })];
		const listing = toLibraryListing(
			games,
			makePage({
				stats: { total: 3, installed: 1, updates: 1, bytes: 150, playtime: 15 },
				sources: [
					{ id: "gog", label: "GOG", count: 2 },
					{ id: "itch", label: "itch.io", count: 1 }
				],
				tags: [{ tag: "a", count: 2 }]
			})
		);
		expect(listing.items).toBe(games);
		expect(listing.stats).toEqual({ total: 3, installed: 1, updates: 1, bytes: 150, playtime: 15 });
		expect(listing.sources).toEqual([
			{ source: "gog", label: "GOG", count: 2 },
			{ source: "itch", label: "itch.io", count: 1 }
		]);
		expect(listing.tags).toEqual([{ tag: "a", count: 2 }]);
	});

	it("returns empty lists when the API sends none", () => {
		expect(toLibraryListing([], makePage({ sources: null, tags: null }))).toEqual({
			items: [],
			stats: { total: 0, installed: 0, updates: 0, bytes: 0, playtime: 0 },
			sources: [],
			tags: []
		});
	});
});

describe("toGameDetail", () => {
	it("maps the game with its files and version changes", () => {
		const wire: WireGameDetail = {
			...makeWire(),
			files: [
				{
					id: 4,
					gameId: 7,
					label: "Windows build",
					url: "https://example.test/win.zip",
					filename: "win.zip",
					size: 1024,
					host: null,
					kind: "direct",
					platform: "windows",
					rangeHint: null,
					requiresAuth: 1
				}
			],
			changes: [
				{
					id: 9,
					gameId: 7,
					checkedAt: "2026-02-01T00:00:00Z",
					oldVersion: "1.1",
					newVersion: "1.2",
					changelog: "fixes",
					seen: 1,
					applied: 0
				}
			]
		};
		const detail = toGameDetail(wire);
		expect(detail.game.title).toBe("Dust Runner");
		expect(detail.files).toEqual([
			{
				id: 4,
				label: "Windows build",
				url: "https://example.test/win.zip",
				filename: "win.zip",
				size: 1024,
				host: null,
				kind: "direct",
				platform: "windows",
				requiresAuth: true
			}
		]);
		expect(detail.changes).toEqual([
			{
				id: 9,
				checkedAt: "2026-02-01T00:00:00Z",
				from: "1.1",
				to: "1.2",
				changelog: "fixes",
				seen: true,
				applied: false
			}
		]);
	});

	it("uses empty lists for null files and changes", () => {
		const detail = toGameDetail({ ...makeWire(), files: null, changes: null });
		expect(detail.files).toEqual([]);
		expect(detail.changes).toEqual([]);
	});
});

describe("toGameUpdate", () => {
	it("sends the id and the changed fields flat, as the Go side expects", () => {
		expect(toGameUpdate(4, { notes: "new", favorite: true })).toEqual({
			id: 4,
			notes: "new",
			favorite: true
		});
	});
});

describe("toSearchResponse", () => {
	it("flattens the games of every source and keeps the errors", () => {
		const response = toSearchResponse([
			{
				plugin: "p1",
				source: "itch",
				games: [{ source: "itch", sourceId: "1", url: "u1", title: "One", score: 2.5 }]
			},
			{ plugin: "p2", source: "gog", games: null, error: "offline" }
		]);
		expect(response).toEqual({
			hits: [{ url: "u1", title: "One", source: "itch", score: 2.5 }],
			errors: ["gog: offline"]
		});
	});
});

describe("toPreview", () => {
	it("maps the source game, its files and the source label, and drops the HTML description", () => {
		const preview = toPreview({
			game: {
				source: "itch",
				sourceId: "1",
				url: "https://example.test/1",
				title: "One",
				descriptionHtml: "<p>Hello</p>",
				coverUrl: "https://example.test/1.jpg"
			},
			existing: null,
			files: [
				{
					id: "dl",
					label: "Windows",
					url: "https://example.test/dl.zip",
					kind: "direct",
					platform: "windows"
				}
			],
			sourceLabel: "itch.io"
		});
		expect(preview).toMatchObject({
			source: "itch",
			title: "One",
			version: null,
			tags: [],
			description: "Hello",
			cover: "https://example.test/1.jpg",
			url: "https://example.test/1",
			existing: null,
			sourceLabel: "itch.io"
		});
		expect(preview.files).toEqual([
			{
				id: "dl",
				label: "Windows",
				url: "https://example.test/dl.zip",
				filename: null,
				size: null,
				host: null,
				kind: "direct",
				platform: "windows",
				requiresAuth: false
			}
		]);
	});

	it("uses empty files when the source cannot list them", () => {
		const preview = toPreview({
			game: { source: "itch", sourceId: "1", url: "u", title: "One" },
			existing: null,
			files: null,
			sourceLabel: "itch.io"
		});
		expect(preview.files).toEqual([]);
	});
});

describe("toUpdatesSummary", () => {
	it("keeps only games with a newer version, the unseen count and the failed checks of the last run", () => {
		const games = [
			makeGame({ id: 1, title: "A", updateAvailable: true }),
			makeGame({ id: 2, title: "B", updateAvailable: false })
		];
		const response = toUpdatesSummary(
			{
				unseen: 4,
				last: {
					checkedAt: "2026-03-01T10:00:00Z",
					results: [
						{ id: 1, title: "A", from: "1", to: "2", hasUpdate: true },
						{ id: 3, title: "C", from: null, to: null, hasUpdate: false, error: "timeout" }
					]
				}
			},
			games
		);
		expect(response.count).toBe(4);
		expect(response.games.map((game) => game.id)).toEqual([1]);
		expect(response.history).toEqual([]);
		expect(response.errors).toEqual(["C: timeout"]);
		expect(response.lastCheckedAt).toBe("2026-03-01T10:00:00Z");
	});

	it("has no last check before the first run", () => {
		const response = toUpdatesSummary({ unseen: 0, last: null }, []);
		expect(response).toEqual({ count: 0, games: [], history: [], errors: [], lastCheckedAt: null });
	});
});

describe("toIncomingList", () => {
	it("maps items and treats a null list as empty", () => {
		const list = toIncomingList({
			configured: true,
			folder: "/incoming",
			items: null,
			count: 0
		});
		expect(list).toEqual({
			configured: true,
			folder: "/incoming",
			items: [],
			count: 0,
			warning: undefined
		});
	});

	it("maps suggestions", () => {
		const list = toIncomingList({
			configured: true,
			folder: "/incoming",
			count: 1,
			items: [
				{
					name: "setup.zip",
					size: 10,
					mtime: "2026-01-01T00:00:00Z",
					ready: true,
					ignored: false,
					auto: false,
					suggestions: [
						{
							gameId: 7,
							title: "Dust",
							cover: null,
							version: "1",
							installed: false,
							confidence: 0.9
						}
					]
				}
			]
		});
		expect(list.items[0].suggestions).toEqual([
			{ gameId: 7, title: "Dust", cover: null, version: "1", installed: false, confidence: 0.9 }
		]);
	});
});
