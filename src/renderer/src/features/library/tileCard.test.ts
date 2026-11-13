import { describe, expect, it } from "vitest";
import type { Game } from "../../lib/features/types";
import { cardView, tileStates } from "./tileCard";

const GAME: Game = {
	id: 1,
	source: "example",
	url: "https://example.test/games/1",
	title: "Example Game",
	version: "1.2",
	developer: "Sample Studio",
	engine: null,
	tags: [],
	cover: null,
	coverLocal: null,
	screenshots: [],
	screenshotLocals: [],
	description: null,
	running: false,
	favorite: true,
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
	updateAvailable: true,
	origin: "plugin",
	sourceKind: "plugin",
	folderPath: null,
	scanFolder: null,
	localConfig: null,
	pinned: false
};

const INSTALLED: Game = { ...GAME, installed: true, updateAvailable: false, favorite: false };

describe("cardView", () => {
	it("shows the title, version and details when there is no card", () => {
		const view = cardView(GAME, null, null);
		expect(view.title).toBe("Example Game");
		expect(view.subtitle).toBe("1.2");
		expect(view.details).toEqual(["Sample Studio"]);
	});

	it("reads the fields the card names", () => {
		const view = cardView(GAME, { title: "developer", subtitle: "engine" }, null);
		expect(view.title).toBe("Sample Studio");
		expect(view.subtitle).toBe("");
	});

	it("shows nothing for an unknown or non-text title or subtitle field", () => {
		const view = cardView(GAME, { title: "nope", subtitle: "tags" }, null);
		expect(view.title).toBe("");
		expect(view.subtitle).toBe("");
	});
});

describe("tileStates", () => {
	it("shows update and favorite by default, with not installed for a library game", () => {
		expect(tileStates(GAME, null, null)).toEqual(["updateAvailable", "notInstalled", "favorite"]);
	});

	it("shows no state for a normal installed game", () => {
		expect(tileStates(INSTALLED, null, null)).toEqual([]);
	});

	it("never shows installed, even when a card asks for it", () => {
		expect(tileStates(INSTALLED, { badges: ["installed"] }, null)).toEqual([]);
	});

	it("turns favorite and update off when the card does not name them", () => {
		expect(tileStates(GAME, { badges: [] }, null)).toEqual(["notInstalled"]);
	});

	it("shows the transfer in place of not installed", () => {
		expect(tileStates(GAME, null, "downloading")).toEqual([
			"updateAvailable",
			"downloading",
			"favorite"
		]);
		expect(tileStates(GAME, null, "installing")).toContain("installing");
	});

	it("gives a browse result no not-installed state", () => {
		expect(tileStates({ ...GAME, id: 0 }, null, null)).toEqual(["updateAvailable", "favorite"]);
	});
});
