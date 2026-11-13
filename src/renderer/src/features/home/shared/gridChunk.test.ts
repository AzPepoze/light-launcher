import { describe, expect, it } from "vitest";
import { GRID_CHUNK_SIZE, nextRenderedCount } from "./gridChunk";

describe("nextRenderedCount", () => {
	it("adds one chunk", () => {
		expect(nextRenderedCount(GRID_CHUNK_SIZE, 1000)).toBe(GRID_CHUNK_SIZE * 2);
	});

	it("never goes above the total", () => {
		expect(nextRenderedCount(GRID_CHUNK_SIZE, GRID_CHUNK_SIZE + 5)).toBe(GRID_CHUNK_SIZE + 5);
	});

	it("stays put when everything is rendered", () => {
		expect(nextRenderedCount(10, 10)).toBe(10);
		expect(nextRenderedCount(0, 0)).toBe(0);
	});
});
