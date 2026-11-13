import { describe, expect, it } from "vitest";
import {
	hasNewVersion,
	htmlToText,
	toBrowseGame,
	toBrowsePage,
	toFilterOptions,
	toGame,
	toGameDetail,
	toGameUpdate,
	toIncomingList,
	toLibraryListPage,
	toLibraryOverview,
	toPreview,
	toSearchResponse,
	toSourceTile,
	toSidebarGroups,
	toUpdateCheckEvent,
	toUpdateHistory,
	toUpdateStatus,
	toUpdatesSummary,
	toWireBrowseRequest,
	toWireListRequest
} from "./libraryMap";
import type { BrowseQuery, Game, LibraryQuery } from "./types";
import type {
	WireBrowsePage,
	WireGame,
	WireGameDetail,
	WireGameDTO,
	WireLibraryList,
	WireLibraryOverview,
	WireSidebarGroup
} from "./wire";

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
		screenshotsLocal: null,
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
		sourceUpdatedAt: null,
		lastCheckedAt: null,
		addedAt: "2026-01-01T00:00:00Z",
		updatedAt: "2026-01-02T00:00:00Z",
		extra: null,
		origin: "plugin",
		sourceKind: "plugin",
		folderPath: null,
		scanFolder: null,
		localConfig: null,
		pinned: 0,
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
		origin: "plugin",
		sourceKind: "plugin",
		folderPath: null,
		scanFolder: null,
		localConfig: null,
		pinned: false,
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
		const game = toGame(makeWire({ tags: null, screenshots: null, screenshotsLocal: null }));
		expect(game.tags).toEqual([]);
		expect(game.screenshots).toEqual([]);
		expect(game.screenshotLocals).toEqual([]);
	});

	it("maps cached screenshot paths in screenshot order", () => {
		const game = toGame(
			makeWire({
				screenshots: ["https://x/1.png", "https://x/2.png"],
				screenshotsLocal: ["screenshots/a.png", null]
			})
		);
		expect(game.screenshotLocals).toEqual(["screenshots/a.png", null]);
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

const EMPTY_STATS = { total: 0, installed: 0, updates: 0, bytes: 0, playtime: 0 };

function makeOverview(overrides: Partial<WireLibraryOverview> = {}): WireLibraryOverview {
	return { sources: [], tags: [], stats: EMPTY_STATS, ...overrides };
}

describe("toLibraryOverview", () => {
	it("maps the sources with their tile shape, and the tags and stats", () => {
		const overview = toLibraryOverview(
			makeOverview({
				stats: { total: 3, installed: 1, updates: 1, bytes: 150, playtime: 15 },
				sources: [
					{ id: "manual", label: "Local games", kind: "local", capabilities: ["launch", "scan"], sync: false, updateCheck: false, count: 1, tile: "square", cover: "1:1", actions: null, card: null },
					{
						id: "f95",
						label: "F95 Zone",
						kind: "plugin",
						plugin: "f95",
						capabilities: ["browse", "updateCheck"],
						sync: true,
						updateCheck: true,
						count: 2,
						tile: "wide",
						cover: "4:3",
						actions: [{ id: "sync", label: "Sync", icon: "sync", method: "f95.sync" }],
						card: { title: "title", badges: ["installed"] }
					}
				],
				tags: [{ tag: "a", count: 2 }]
			})
		);
		expect(overview.sources).toEqual([
			{ source: "manual", label: "Local games", kind: "local", plugin: null, capabilities: ["launch", "scan"], sync: false, updateCheck: false, count: 1, tile: "square", cover: "1:1", actions: [], card: null, sidebar: [] },
			{
				source: "f95",
				label: "F95 Zone",
				kind: "plugin",
				plugin: "f95",
				capabilities: ["browse", "updateCheck"],
				sync: true,
				updateCheck: true,
				count: 2,
				tile: "wide",
				cover: "4:3",
				actions: [{ id: "sync", label: "Sync", icon: "sync", method: "f95.sync", scope: "library" }],
				card: { title: "title", badges: ["installed"] },
				sidebar: []
			}
		]);
		expect(overview.tags).toEqual([{ tag: "a", count: 2 }]);
		expect(overview.stats).toEqual({
			total: 3,
			installed: 1,
			updates: 1,
			bytes: 150,
			playtime: 15
		});
	});

	it("returns empty lists when the API sends none", () => {
		expect(toLibraryOverview(makeOverview({ sources: null, tags: null }))).toEqual({
			sources: [],
			tags: [],
			stats: EMPTY_STATS
		});
	});
});

describe("toSourceTile", () => {
	it("keeps the three known shapes", () => {
		expect(toSourceTile("poster")).toBe("poster");
		expect(toSourceTile("wide")).toBe("wide");
		expect(toSourceTile("square")).toBe("square");
	});

	it("falls back to the poster shape for an unknown or missing value", () => {
		expect(toSourceTile("round")).toBe("poster");
		expect(toSourceTile(undefined)).toBe("poster");
	});
});

const SIDEBAR: WireSidebarGroup[] = [
	{ type: "search", id: "text", placeholder: "Search titles" },
	{ type: "multi", id: "tags", label: "Tags", dynamic: true, searchable: true },
	{
		type: "sort",
		id: "sort",
		orderable: true,
		options: [{ value: "title", label: "Title", icon: "sort_by_alpha" }],
		default: { value: "title", order: "asc" }
	}
];

describe("toSidebarGroups", () => {
	it("maps every known group with its optional fields", () => {
		const groups = toSidebarGroups(SIDEBAR);
		expect(groups).toHaveLength(3);
		expect(groups[0]).toMatchObject({ type: "search", id: "text", placeholder: "Search titles" });
		expect(groups[1]).toMatchObject({
			type: "multi",
			label: "Tags",
			dynamic: true,
			searchable: true
		});
		expect(groups[2]).toMatchObject({
			type: "sort",
			orderable: true,
			default: { value: "title", order: "asc" },
			options: [{ value: "title", label: "Title", icon: "sort_by_alpha", count: null }]
		});
	});

	it("fills the missing optional fields and drops an unknown group type", () => {
		const groups = toSidebarGroups([
			{ type: "select", id: "status" },
			{ type: "slider", id: "size" } as WireSidebarGroup
		]);
		expect(groups).toEqual([
			{
				type: "select",
				id: "status",
				label: "",
				placeholder: "",
				options: [],
				dynamic: false,
				searchable: false,
				orderable: false,
				default: null
			}
		]);
	});

	it("treats null as no groups", () => {
		expect(toSidebarGroups(null)).toEqual([]);
	});

	it("drops a sort default with an unknown order", () => {
		const group = toSidebarGroups([
			{
				type: "sort",
				id: "sort",
				options: [{ value: "title", label: "Title" }],
				default: { value: "title", order: "up" }
			}
		])[0];
		expect(group.default).toBeNull();
	});
});

const BROWSE_QUERY: BrowseQuery = {
	source: "shop",
	query: "harbor",
	filters: { tags: ["rpg"], featured: true },
	sort: "title",
	order: "desc",
	cursor: "next",
	limit: 20
};

describe("toWireBrowseRequest", () => {
	it("sends the query, filters, sort and cursor", () => {
		expect(toWireBrowseRequest(BROWSE_QUERY)).toEqual({
			source: "shop",
			query: "harbor",
			filters: { tags: ["rpg"], featured: true },
			sort: "title",
			order: "desc",
			cursor: "next",
			limit: 20
		});
	});

	it("leaves the optional fields out when they are empty", () => {
		expect(
			toWireBrowseRequest({
				...BROWSE_QUERY,
				query: "",
				filters: {},
				sort: "",
				order: "",
				cursor: ""
			})
		).toEqual({ source: "shop", limit: 20 });
	});
});

describe("toBrowsePage", () => {
	const dto: WireGameDTO = {
		source: "shop",
		sourceId: "b1",
		url: "https://shop/b1",
		title: "Alpha Quest",
		version: "1.0",
		developer: "Studio",
		tags: ["rpg"],
		coverUrl: "https://shop/b1.jpg",
		descriptionHtml: "<p>Short &amp; sweet</p>"
	};

	it("maps a browse game that is not in the library yet", () => {
		const game = toBrowseGame(dto);
		expect(game).toMatchObject({
			id: 0,
			source: "shop",
			url: "https://shop/b1",
			title: "Alpha Quest",
			version: "1.0",
			tags: ["rpg"],
			cover: "https://shop/b1.jpg",
			description: "Short & sweet",
			installed: false,
			favorite: false
		});
	});

	it("keeps the total and cursor when the source sends them", () => {
		const page = toBrowsePage({
			games: [dto],
			nextCursor: "next",
			total: 8
		} satisfies WireBrowsePage);
		expect(page.games).toHaveLength(1);
		expect(page.nextCursor).toBe("next");
		expect(page.total).toBe(8);
	});

	it("uses null for a missing total, cursor and game list", () => {
		expect(toBrowsePage({ games: null } as WireBrowsePage)).toEqual({
			games: [],
			nextCursor: null,
			total: null
		});
	});
});

describe("toFilterOptions", () => {
	it("keeps the counts the source sends and null where it does not", () => {
		expect(
			toFilterOptions([
				{ value: "rpg", label: "RPG", count: 4 },
				{ value: "horror", label: "Horror" }
			])
		).toEqual([
			{ value: "rpg", label: "RPG", count: 4 },
			{ value: "horror", label: "Horror", count: null }
		]);
	});

	it("treats a missing list as empty", () => {
		expect(toFilterOptions(null)).toEqual([]);
	});
});

describe("toLibraryListPage", () => {
	it("maps the games and keeps the total when the API sends it", () => {
		const page = toLibraryListPage({ items: [makeWire({ id: 4 })], total: 90 });
		expect(page.items.map((game) => game.id)).toEqual([4]);
		expect(page.total).toBe(90);
	});

	it("leaves the total out when the API omits it", () => {
		const page = toLibraryListPage({ items: [] } as WireLibraryList);
		expect(page).toEqual({ items: [] });
	});

	it("treats null items as an empty page", () => {
		expect(toLibraryListPage({ items: null } as WireLibraryList).items).toEqual([]);
	});
});

const QUERY: LibraryQuery = {
	source: "f95",
	search: "dark",
	tags: ["rpg"],
	favorite: true,
	installed: false,
	updates: false,
	sort: "added"
};

describe("toWireListRequest", () => {
	it("sends the flat filter fields, the page and the total flag", () => {
		expect(toWireListRequest(QUERY, { offset: 120, limit: 60, wantTotal: true })).toEqual({
			search: "dark",
			source: "f95",
			tags: ["rpg"],
			status: null,
			favorite: true,
			installed: null,
			updates: false,
			hidden: false,
			sort: "added",
			order: "desc",
			offset: 120,
			limit: 60,
			wantTotal: true
		});
	});

	it("sends installed as null unless the installed filter is on", () => {
		const request = toWireListRequest(QUERY, { offset: 0, limit: 60, wantTotal: false });
		expect(request.installed).toBeNull();
	});

	it("sends installed as true when the filter is on", () => {
		const request = toWireListRequest(
			{ ...QUERY, installed: true },
			{ offset: 0, limit: 60, wantTotal: false }
		);
		expect(request.installed).toBe(true);
	});

	it.each([
		["title", "title", "asc"],
		["added", "added", "desc"],
		["lastPlayed", "lastPlayed", "desc"]
	] as const)("maps the %s sort to the server sort %s, %s", (sort, serverSort, order) => {
		const request = toWireListRequest(
			{ ...QUERY, sort },
			{ offset: 0, limit: 60, wantTotal: false }
		);
		expect(request.sort).toBe(serverSort);
		expect(request.order).toBe(order);
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
		expect(response.errors).toEqual(["C: timeout"]);
		expect(response.lastCheckedAt).toBe("2026-03-01T10:00:00Z");
	});

	it("has no last check before the first run", () => {
		const response = toUpdatesSummary({ unseen: 0, last: null }, []);
		expect(response).toEqual({ count: 0, games: [], errors: [], lastCheckedAt: null });
	});
});

describe("toUpdateHistory", () => {
	it("maps checks and treats unknown results as current", () => {
		const history = toUpdateHistory([
			{
				id: 9,
				gameId: 1,
				title: "Dust",
				checkedAt: "2026-03-01T10:00:00Z",
				from: "1.0",
				to: "2.0",
				result: "update",
				error: null,
				coverPath: "covers/dust.jpg",
				coverUrl: "https://example.test/dust.jpg",
				source: "f95"
			},
			{
				id: 8,
				gameId: 2,
				title: "Aurora",
				checkedAt: "2026-03-01T09:00:00Z",
				from: null,
				to: null,
				result: "surprise",
				error: null,
				coverPath: null,
				coverUrl: null,
				source: "itch"
			}
		]);
		expect(history[0]).toMatchObject({ id: 9, title: "Dust", result: "update" });
		expect(history[0]).toMatchObject({ coverPath: "covers/dust.jpg", coverUrl: "https://example.test/dust.jpg" });
		expect(history[1].result).toBe("current");
	});
});

describe("toUpdateCheckEvent", () => {
	it("maps a running event with its outcome and no issues", () => {
		const event = toUpdateCheckEvent({
			state: "running",
			done: 1,
			total: 4,
			skipped: 0,
			current: { gameId: 7, title: "Dust" },
			result: {
				gameId: 3,
				title: "Aurora",
				from: "1.0",
				to: "2.0",
				result: "update",
				error: null,
				coverPath: null,
				coverUrl: "https://example.test/aurora.jpg"
			},
			issues: []
		});
		expect(event).toEqual({
			state: "running",
			done: 1,
			total: 4,
			skipped: 0,
			current: { gameId: 7, title: "Dust" },
			result: {
				gameId: 3,
				title: "Aurora",
				from: "1.0",
				to: "2.0",
				result: "update",
				error: null,
				coverPath: null,
				coverUrl: "https://example.test/aurora.jpg"
			},
			issues: []
		});
	});

	it("keeps skipped outcomes and maps the issues of the final event", () => {
		const event = toUpdateCheckEvent({
			state: "done",
			done: 5,
			total: 5,
			skipped: 2,
			current: null,
			result: null,
			issues: [{ source: "itch", label: "itch.io", kind: "auth", message: "itch.io needs a sign-in", games: 3 }]
		});
		expect(event.state).toBe("done");
		expect(event.skipped).toBe(2);
		expect(event.issues).toEqual([
			{ source: "itch", label: "itch.io", kind: "auth", message: "itch.io needs a sign-in", games: 3 }
		]);
	});

	it("treats unknown outcome results as current and tolerates a missing issue list", () => {
		const event = toUpdateCheckEvent({
			state: "running",
			done: 1,
			total: 2,
			skipped: 0,
			current: null,
			result: {
				gameId: 3,
				title: "Aurora",
				from: null,
				to: null,
				result: "surprise",
				error: null,
				coverPath: null,
				coverUrl: null
			},
			issues: null
		});
		expect(event.result?.result).toBe("current");
		expect(event.issues).toEqual([]);
	});
});

describe("toUpdateStatus", () => {
	it("maps the saved run, the running progress and the games with updates", () => {
		const status = toUpdateStatus({
			lastCheckedAt: "2026-03-01T10:00:00Z",
			running: true,
			progress: { done: 2, total: 9 },
			count: 1,
			nextRunAt: "2026-03-02T10:00:00Z",
			games: [{ id: 4, title: "Dust", version: "2", installedVersion: "1", favorite: 0 } as unknown as WireGame]
		});
		expect(status).toMatchObject({
			lastCheckedAt: "2026-03-01T10:00:00Z",
			running: true,
			progress: { done: 2, total: 9 },
			count: 1,
			nextRunAt: "2026-03-02T10:00:00Z"
		});
		expect(status.games.map((game) => game.id)).toEqual([4]);
	});

	it("reads an empty status when no run was saved", () => {
		const status = toUpdateStatus({
			lastCheckedAt: null,
			running: false,
			progress: null,
			count: 0,
			nextRunAt: null,
			games: null
		});
		expect(status).toEqual({
			lastCheckedAt: null,
			running: false,
			progress: null,
			count: 0,
			nextRunAt: null,
			games: []
		});
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
