import { describe, expect, it } from "vitest";
import type { LibrarySource } from "../../lib/features/types";
import {
	LOCAL_SOURCE,
	gameCount,
	librarySections,
	librarySubtitle,
	sectionElementId,
	showsSection,
	sourceLabel
} from "./librarySections";

describe("gameCount", () => {
	it("uses the singular for one game and the plural otherwise", () => {
		expect(gameCount(0)).toBe("0 games");
		expect(gameCount(1)).toBe("1 game");
		expect(gameCount(2)).toBe("2 games");
	});
});

describe("librarySubtitle", () => {
	it("joins the counts with commas, not middle dots", () => {
		expect(librarySubtitle({ total: 12, installed: 4, updates: 2 })).toBe(
			"12 games, 4 installed, 2 updates"
		);
	});

	it("leaves out the updates when none are waiting and uses the singular for one game", () => {
		expect(librarySubtitle({ total: 1, installed: 0, updates: 0 })).toBe("1 game, 0 installed");
	});
});

const SERVER_SOURCES: LibrarySource[] = [
	{ source: "f95", label: "F95 Zone", kind: "plugin", plugin: "f95", capabilities: ["browse", "updateCheck"], sync: false, updateCheck: true, count: 12, tile: "wide", cover: "16:9", actions: [], card: null, sidebar: [] },
	{ source: LOCAL_SOURCE, label: "Local games", kind: "local", plugin: null, capabilities: ["launch", "scan"], sync: false, updateCheck: false, count: 3, tile: "square", cover: "1:1", actions: [], card: null, sidebar: [] },
	{ source: "itch", label: "itch.io", kind: "plugin", plugin: "itch", capabilities: ["browse", "auth"], sync: false, updateCheck: false, count: 0, tile: "poster", cover: "3:4", actions: [], card: null, sidebar: [] }
];

describe("librarySections", () => {
	it("puts local games first, then the plugin sources in the server order", () => {
		expect(librarySections(SERVER_SOURCES).map((item) => item.source)).toEqual([
			LOCAL_SOURCE,
			"f95",
			"itch"
		]);
	});

	it("keeps a plugin source with no games so its Sync button can be used", () => {
		const itch = librarySections(SERVER_SOURCES).find((item) => item.source === "itch");
		expect(itch?.count).toBe(0);
	});

	it("adds an empty local section when the overview has none", () => {
		const sections = librarySections([
			{ source: "f95", label: "F95 Zone", kind: "plugin", plugin: "f95", capabilities: [], sync: false, updateCheck: false, count: 1, tile: "wide", cover: "16:9", actions: [], card: null, sidebar: [] }
		]);
		expect(sections[0]).toEqual({
			source: LOCAL_SOURCE,
			label: "Local games",
			kind: "local",
			plugin: null,
			capabilities: ["launch", "scan"],
			sync: false,
			updateCheck: false,
			count: 0,
			tile: "square",
			cover: "1:1",
			actions: [],
			card: null,
			sidebar: []
		});
	});

	it("gives a single section for an empty overview", () => {
		expect(librarySections([]).map((item) => item.source)).toEqual([LOCAL_SOURCE]);
	});
});

describe("showsSection", () => {
	const [f95, local, itch] = SERVER_SOURCES;

	it("always shows the local section, even when empty", () => {
		expect(showsSection(local, 0, false)).toBe(true);
	});

	it("hides a plugin source with no games", () => {
		expect(showsSection(itch, 0, false)).toBe(false);
	});

	it("shows a plugin source with games", () => {
		expect(showsSection(f95, 4, false)).toBe(true);
	});

	it("shows a plugin source that failed to load, so the error is seen", () => {
		expect(showsSection(itch, 0, true)).toBe(true);
	});

	it("hides a plugin source that is still loading", () => {
		expect(showsSection(f95, null, false)).toBe(false);
	});
});

describe("sectionElementId", () => {
	it("builds a stable id from the source id", () => {
		expect(sectionElementId("f95")).toBe("library-section-f95");
	});
});

describe("sourceLabel", () => {
	it("uses the label the overview sent", () => {
		expect(sourceLabel(SERVER_SOURCES, "f95")).toBe("F95 Zone");
	});

	it("names the local source in plain words when the overview has no entry", () => {
		expect(sourceLabel([], LOCAL_SOURCE)).toBe("Local games");
	});

	it("falls back to the id for an unknown source", () => {
		expect(sourceLabel(SERVER_SOURCES, "gog")).toBe("gog");
	});
});
