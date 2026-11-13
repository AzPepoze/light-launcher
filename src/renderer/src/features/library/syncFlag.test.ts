import { describe, expect, it } from "vitest";
import type { LibrarySource } from "../../lib/features/types";
import { LOCAL_SOURCE } from "./librarySections";
import { canSyncGames } from "./syncFlag";

const PLUGIN: LibrarySource = {
	source: "example",
	label: "Example",
	kind: "plugin",
	plugin: "example",
	capabilities: ["listGames"],
	sync: true,
	updateCheck: true,
	count: 0,
	tile: "poster",
	cover: "3:4",
	actions: [],
	card: null,
	sidebar: []
};

describe("canSyncGames", () => {
	it("never offers sync for local games", () => {
		expect(canSyncGames({ ...PLUGIN, source: LOCAL_SOURCE })).toBe(false);
	});

	it("follows the source capability", () => {
		expect(canSyncGames({ ...PLUGIN, sync: true })).toBe(true);
		expect(canSyncGames({ ...PLUGIN, sync: false })).toBe(false);
	});

	it("is false without a source", () => {
		expect(canSyncGames(undefined)).toBe(false);
	});
});
