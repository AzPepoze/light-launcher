import { describe, expect, it } from "vitest";
import type { LibrarySource } from "../../lib/features/types";
import {
	LOCAL_SOURCE,
	librarySections,
	pickSection,
	showsSection,
	sourceLabel
} from "./librarySections";

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

describe("pickSection", () => {
	const [f95, local] = SERVER_SOURCES;

	it("selects the first visible source when none is chosen", () => {
		expect(pickSection([local, f95], "")?.source).toBe(LOCAL_SOURCE);
	});

	it("selects the chosen source when it is visible", () => {
		expect(pickSection([local, f95], "f95")?.source).toBe("f95");
	});

	it("falls back to the first visible source when the chosen one is gone", () => {
		expect(pickSection([local], "f95")?.source).toBe(LOCAL_SOURCE);
	});

	it("returns nothing when no source is visible", () => {
		expect(pickSection([], "f95")).toBeUndefined();
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
