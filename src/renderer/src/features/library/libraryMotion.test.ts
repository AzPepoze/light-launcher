import { describe, expect, it } from "vitest";
import {
	REVEAL_MAX_TILES,
	REVEAL_STEP_MS,
	REVEAL_TILE_MS,
	revealDoneMs,
	revealIndex
} from "./libraryMotion";

describe("revealIndex", () => {
	it("gives the first tiles their position in the reveal", () => {
		expect(revealIndex(0, false)).toBe(0);
		expect(revealIndex(5, false)).toBe(5);
	});

	it("does not animate tiles past the reveal limit", () => {
		expect(revealIndex(REVEAL_MAX_TILES, false)).toBeUndefined();
	});

	it("does not animate again once the first reveal has run", () => {
		expect(revealIndex(0, true)).toBeUndefined();
	});
});

describe("revealDoneMs", () => {
	it("is zero with no tiles", () => {
		expect(revealDoneMs(0)).toBe(0);
	});

	it("is one tile duration for a single tile", () => {
		expect(revealDoneMs(1)).toBe(REVEAL_TILE_MS);
	});

	it("adds one stagger step per extra tile", () => {
		expect(revealDoneMs(REVEAL_MAX_TILES)).toBe(
			(REVEAL_MAX_TILES - 1) * REVEAL_STEP_MS + REVEAL_TILE_MS
		);
	});
});
