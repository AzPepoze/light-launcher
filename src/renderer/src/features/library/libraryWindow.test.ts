import { describe, expect, it } from "vitest";
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

/** A 3:4 cover with 140 px columns, the same as the default card size before the setting. */
const POSTER = { ratio: 3 / 4, minColumnPx: 140 };
const WIDE = { ratio: 16 / 9, minColumnPx: 240 };
const SQUARE = { ratio: 1, minColumnPx: 140 };

/** A 1000 px wide poster grid: six columns of 155 px. */
const WIDE_POSTER = gridLayout(1000, POSTER.ratio, POSTER.minColumnPx);

function windowAt(total: number, layout: GridLayout, scrollTop: number, viewportHeight: number) {
	return gridWindow({ total, layout, scrollTop, viewportHeight });
}

describe("gridLayout", () => {
	it("fits as many columns as the minimum width allows", () => {
		expect(WIDE_POSTER.columns).toBe(6);
		// Six columns share the width minus five 12 px gaps.
		expect(WIDE_POSTER.tileWidth).toBeCloseTo(940 / 6);
	});

	it("uses one column on a narrow grid", () => {
		expect(gridLayout(100, POSTER.ratio, POSTER.minColumnPx).columns).toBe(1);
	});

	it("uses the cover ratio for the row height", () => {
		const wide = gridLayout(1000, WIDE.ratio, WIDE.minColumnPx);
		expect(wide.columns).toBe(4);
		// Four 241 px covers at 16:9 are 135.6 px high; the tile is the cover, then 12 px of gap.
		expect(wide.rowHeight).toBeCloseTo(147.5625);
	});

	it("uses the minimum column width from the card size setting", () => {
		expect(gridLayout(1000, SQUARE.ratio, 260).columns).toBe(3);
		expect(gridLayout(1000, SQUARE.ratio, 100).columns).toBe(9);
	});

	it("changes the layout when the width changes", () => {
		expect(gridLayout(400, POSTER.ratio, POSTER.minColumnPx).columns).toBe(2);
		expect(gridLayout(1400, POSTER.ratio, POSTER.minColumnPx).columns).toBe(9);
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
		const narrow = gridLayout(400, POSTER.ratio, POSTER.minColumnPx);
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
		const layout = gridLayout(1000, POSTER.ratio, POSTER.minColumnPx);
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
			for (const card of [POSTER, WIDE, SQUARE]) {
				const layout = gridLayout(width, card.ratio, card.minColumnPx);
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
