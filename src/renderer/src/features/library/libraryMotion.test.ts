import { describe, expect, it } from "vitest";
import type { Game } from "../../lib/features/types";
import type { GameGroup } from "./libraryGroups";
import { REVEAL_MAX_TILES, REVEAL_STEP_MS, REVEAL_TILE_MS, revealDoneMs, revealOrder } from "./libraryMotion";
import { tileShapeFor } from "./sourceTileShape";

function gameWithId(id: number): Game {
	return { id } as Game;
}

function groupOf(source: string, ids: number[]): GameGroup {
	return {
		source,
		label: source,
		shape: tileShapeFor(source),
		games: ids.map(gameWithId)
	};
}

describe("revealOrder", () => {
	it("numbers tiles in display order across groups", () => {
		const order = revealOrder([groupOf("manual", [4, 2]), groupOf("f95", [9])]);
		expect([...order.entries()]).toEqual([
			[4, 0],
			[2, 1],
			[9, 2]
		]);
	});

	it("stops at the reveal limit", () => {
		const ids = Array.from({ length: REVEAL_MAX_TILES + 5 }, (_, index) => index + 1);
		const order = revealOrder([groupOf("manual", ids)]);
		expect(order.size).toBe(REVEAL_MAX_TILES);
		expect(order.has(REVEAL_MAX_TILES + 1)).toBe(false);
	});

	it("is empty when there are no groups", () => {
		expect(revealOrder([]).size).toBe(0);
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
