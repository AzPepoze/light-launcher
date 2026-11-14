import { describe, expect, it } from "vitest";
import { isCurrentResponse } from "./browseFeed.svelte";

describe("isCurrentResponse", () => {
	it("accepts an answer for the source and counter the feed still holds", () => {
		expect(isCurrentResponse({ source: "f95", epoch: 2 }, { source: "f95", epoch: 2 })).toBe(true);
	});

	it("drops an answer for a source the user has left", () => {
		expect(isCurrentResponse({ source: "f95", epoch: 2 }, { source: "", epoch: 3 })).toBe(false);
	});

	it("drops an answer for a source that was replaced by another one", () => {
		expect(isCurrentResponse({ source: "f95", epoch: 2 }, { source: "itch", epoch: 3 })).toBe(false);
	});

	it("drops an answer from before a reset of the same source", () => {
		expect(isCurrentResponse({ source: "itch", epoch: 1 }, { source: "itch", epoch: 2 })).toBe(false);
	});

	it("never accepts an answer while the feed holds no source", () => {
		expect(isCurrentResponse({ source: "", epoch: 0 }, { source: "", epoch: 0 })).toBe(false);
	});
});
