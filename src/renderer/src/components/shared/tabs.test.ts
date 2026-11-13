import { describe, expect, it } from "vitest";
import { nextTabId } from "./tabs";

const IDS = ["games", "updates", "downloads"];

describe("nextTabId", () => {
	it("moves right and wraps from the last tab to the first", () => {
		expect(nextTabId(IDS, "games", "ArrowRight")).toBe("updates");
		expect(nextTabId(IDS, "downloads", "ArrowRight")).toBe("games");
	});

	it("moves left and wraps from the first tab to the last", () => {
		expect(nextTabId(IDS, "updates", "ArrowLeft")).toBe("games");
		expect(nextTabId(IDS, "games", "ArrowLeft")).toBe("downloads");
	});

	it("jumps to the first and last tab with Home and End", () => {
		expect(nextTabId(IDS, "updates", "Home")).toBe("games");
		expect(nextTabId(IDS, "updates", "End")).toBe("downloads");
	});

	it("returns null for other keys", () => {
		expect(nextTabId(IDS, "games", "Enter")).toBeNull();
	});

	it("returns null when there are no tabs", () => {
		expect(nextTabId([], "games", "ArrowRight")).toBeNull();
	});

	it("starts from the first tab when the current id is unknown", () => {
		expect(nextTabId(IDS, "missing", "ArrowRight")).toBe("updates");
	});
});
