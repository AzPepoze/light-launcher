import { describe, expect, it } from "vitest";
import type { TileShape } from "./sourceTileShape";
import { tileShapeFor } from "./sourceTileShape";
import {
	MAX_RENDERED_TILES,
	PAGE_SIZE,
	gridLayout,
	indexesIn,
	gridWindow,
	pagesToLoad,
	rowCount,
	type GridLayout
} from "./libraryWindow";

const POSTER: TileShape = tileShapeFor("poster");

/** A 1000 px wide poster grid: six columns of 155 px. */
const WIDE_POSTER = gridLayout(1000, POSTER);

function windowAt(total: number, layout: GridLayout, scrollTop: number, viewportHeight: number) {
	return gridWindow({ total, layout, scrollTop, viewportHeight });
}

describe("gridLayout", () => {
	it("fits as many columns as the minimum width allows", () => {
		expect(WIDE_POSTER.columns).toBe(6);
		expect(WIDE_POSTER.tileWidth).toBeCloseTo(155);
	});

	it("uses one column on a narrow grid", () => {
		expect(gridLayout(100, POSTER).columns).toBe(1);
	});

	it("uses the tile ratio for the row height", () => {
		const wide = gridLayout(1000, tileShapeFor("wide"));
		expect(wide.columns).toBe(3);
		// Cover 324 px wide at 16:9 is 182.25 px high; 56 px of text and 18 px of gap follow.
		expect(wide.rowHeight).toBeCloseTo(256.25);
	});

	it("changes the layout when the width changes", () => {
		expect(gridLayout(400, POSTER).columns).toBe(2);
		expect(gridLayout(1400, POSTER).columns).toBe(9);
	});
});

describe("rowCount", () => {
	it("rounds a partly filled last row up", () => {
		expect(rowCount(7, 3)).toBe(3);
	});

	it("is zero for an empty list", () => {
		expect(rowCount(0, 3)).toBe(0);
	});
});

describe("gridWindow", () => {
	it("renders nothing for an empty list and has no spacers", () => {
		expect(windowAt(0, WIDE_POSTER, 0, 900)).toEqual({
			start: 0,
			end: 0,
			topSpacer: 0,
			bottomSpacer: 0
		});
	});

	it("renders a single page in full when it fits the view", () => {
		const result = windowAt(40, WIDE_POSTER, 0, 2000);
		expect(result.start).toBe(0);
		expect(result.end).toBe(40);
		expect(result.bottomSpacer).toBe(0);
	});

	it("renders only the rows near the top of the view", () => {
		const result = windowAt(50_000, WIDE_POSTER, 0, 900);
		expect(result.start).toBe(0);
		expect(result.end).toBeLessThan(60);
		expect(result.topSpacer).toBe(0);
	});

	it("renders more rows when the view is taller, from the same first row", () => {
		const short = windowAt(50_000, WIDE_POSTER, 3000, 400);
		const tall = windowAt(50_000, WIDE_POSTER, 3000, 1600);
		expect(tall.end - tall.start).toBeGreaterThan(short.end - short.start);
		expect(tall.start).toBe(short.start);
	});

	it("re-lays out a narrower grid with fewer columns and starts on a row edge", () => {
		const narrow = gridLayout(400, POSTER);
		const result = windowAt(50_000, narrow, 2000, 900);
		expect(result.start % narrow.columns).toBe(0);
	});

	it("jumps straight to a far scroll position without rendering the rows in between", () => {
		const result = windowAt(50_000, WIDE_POSTER, 1_000_000, 900);
		expect(result.start).toBeGreaterThan(0);
		expect(result.end - result.start).toBeLessThanOrEqual(MAX_RENDERED_TILES);
		expect(result.topSpacer).toBeGreaterThan(900_000);
	});

	it("ends a partly filled last row at the last item, not at the row edge", () => {
		const layout = gridLayout(1000, POSTER);
		const result = windowAt(7, { ...layout, columns: 3 }, 0, 2000);
		expect(result.end).toBe(7);
	});

	it("keeps the spacers and the rendered rows equal to the full height", () => {
		const total = 1234;
		const result = windowAt(total, WIDE_POSTER, 9000, 800);
		const rendered = rowCount(result.end - result.start, WIDE_POSTER.columns);
		const fullHeight = rowCount(total, WIDE_POSTER.columns) * WIDE_POSTER.rowHeight;
		const drawn = result.topSpacer + rendered * WIDE_POSTER.rowHeight + result.bottomSpacer;
		expect(drawn).toBeCloseTo(fullHeight);
	});

	it("never renders more than the tile cap for a 50,000 item section", () => {
		const widths = [200, 800, 1400, 2400];
		const viewports = [0, 900, 3000];
		const scrolls = [0, 1000, 250_000, 2_000_000, 9_000_000];
		for (const width of widths) {
			for (const shape of ["poster", "wide", "square"] as const) {
				const layout = gridLayout(width, tileShapeFor(shape));
				for (const viewportHeight of viewports) {
					for (const scrollTop of scrolls) {
						const result = gridWindow({
							total: 50_000,
							layout,
							scrollTop,
							viewportHeight
						});
						expect(result.end - result.start).toBeLessThanOrEqual(MAX_RENDERED_TILES);
					}
				}
			}
		}
	});
});

describe("indexesIn", () => {
	it("lists the indexes from start up to end, not including end", () => {
		expect(indexesIn({ start: 6, end: 9 })).toEqual([6, 7, 8]);
	});

	it("is empty for an empty window", () => {
		expect(indexesIn({ start: 0, end: 0 })).toEqual([]);
	});
});

describe("pagesToLoad", () => {
	it("asks for nothing when the window is empty", () => {
		expect(pagesToLoad({ start: 0, end: 0 }, 0)).toEqual([]);
	});

	it("asks for the first page and no more for a short list", () => {
		expect(pagesToLoad({ start: 0, end: 40 }, 40)).toEqual([0]);
	});

	it("asks for the page after the window, when there is one", () => {
		expect(pagesToLoad({ start: 0, end: 60 }, 500)).toEqual([0, 1]);
	});

	it("does not ask past the last page on an exact multiple", () => {
		expect(pagesToLoad({ start: 0, end: 120 }, 120)).toEqual([0, 1]);
	});

	it("asks for two pages only when the window jumps far down a long list", () => {
		const result = windowAt(50_000, WIDE_POSTER, 1_000_000, 900);
		const pages = pagesToLoad(result, 50_000);
		expect(pages.length).toBeLessThanOrEqual(3);
		expect(pages[0]).toBe(Math.floor(result.start / PAGE_SIZE));
	});

	it("covers a window that crosses a page boundary", () => {
		expect(pagesToLoad({ start: 50, end: 130 }, 300, PAGE_SIZE, 0)).toEqual([0, 1, 2]);
	});
});
