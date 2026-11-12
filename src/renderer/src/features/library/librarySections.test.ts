import { describe, expect, it } from "vitest";
import type { LibrarySource } from "../../lib/features/types";
import { LOCAL_SOURCE, librarySections, sourceLabel } from "./librarySections";

const SERVER_SOURCES: LibrarySource[] = [
	{ source: "f95", label: "F95 Zone", count: 12, tile: "wide", sidebar: [] },
	{ source: LOCAL_SOURCE, label: "Local games", count: 3, tile: "square", sidebar: [] },
	{ source: "itch", label: "itch.io", count: 0, tile: "poster", sidebar: [] }
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
			{ source: "f95", label: "F95 Zone", count: 1, tile: "wide", sidebar: [] }
		]);
		expect(sections[0]).toEqual({
			source: LOCAL_SOURCE,
			label: "Local games",
			count: 0,
			tile: "square",
			sidebar: []
		});
	});

	it("gives a single section for an empty overview", () => {
		expect(librarySections([]).map((item) => item.source)).toEqual([LOCAL_SOURCE]);
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
