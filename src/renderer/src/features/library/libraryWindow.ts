import type { TileShape } from "./sourceTileShape";

/** Games per list page. Keep in step with the page size the library asks for. */
export const PAGE_SIZE = 60;
/** Most tiles in the DOM at once. The rest of the list is spacer height. */
export const MAX_RENDERED_TILES = 80;

const COLUMN_GAP_PX = 14;
const ROW_GAP_PX = 18;
/** Room under the cover for the title and the status lines. */
const TEXT_BLOCK_PX = 56;
const OVERSCAN_ROWS = 1;
const PREFETCH_PAGES = 1;

export interface GridLayout {
	columns: number;
	tileWidth: number;
	/** Height of one row, including the gap under it. */
	rowHeight: number;
}

export interface GridWindow {
	/** First item index to render. */
	start: number;
	/** Index after the last item to render. */
	end: number;
	topSpacer: number;
	bottomSpacer: number;
}

export interface WindowInput {
	total: number;
	layout: GridLayout;
	/** Top of the visible area, measured from the top of the grid. */
	scrollTop: number;
	viewportHeight: number;
}

const NOTHING: GridWindow = { start: 0, end: 0, topSpacer: 0, bottomSpacer: 0 };

function clamp(value: number, low: number, high: number): number {
	return Math.min(Math.max(value, low), high);
}

/** Columns and row height for a grid of this width and tile shape. */
export function gridLayout(width: number, shape: TileShape): GridLayout {
	const columns = Math.max(
		1,
		Math.floor((width + COLUMN_GAP_PX) / (shape.minColumnPx + COLUMN_GAP_PX))
	);
	const tileWidth = Math.max(0, (width - (columns - 1) * COLUMN_GAP_PX) / columns);
	const rowHeight = tileWidth / shape.ratio + TEXT_BLOCK_PX + ROW_GAP_PX;
	return { columns, tileWidth, rowHeight };
}

/** Rows needed for the items. The last row may be partly filled. */
export function rowCount(total: number, columns: number): number {
	return Math.ceil(total / columns);
}

/** The items to render for the visible area, plus the spacers that keep the scroll height right. */
export function gridWindow(input: WindowInput): GridWindow {
	const { total, layout } = input;
	const rows = rowCount(total, layout.columns);
	if (rows === 0) return NOTHING;

	const firstVisibleRow = clamp(Math.floor(input.scrollTop / layout.rowHeight), 0, rows - 1);
	const visibleRows = Math.ceil(input.viewportHeight / layout.rowHeight) + 1;
	const maxRows = Math.max(1, Math.floor(MAX_RENDERED_TILES / layout.columns));
	const renderedRows = Math.min(visibleRows + 2 * OVERSCAN_ROWS, maxRows);

	const firstRow = clamp(firstVisibleRow - OVERSCAN_ROWS, 0, rows - 1);
	const endRow = Math.min(rows, firstRow + renderedRows);
	return {
		start: firstRow * layout.columns,
		end: Math.min(total, endRow * layout.columns),
		topSpacer: firstRow * layout.rowHeight,
		bottomSpacer: (rows - endRow) * layout.rowHeight
	};
}

/** The item indexes inside a window, in order. */
export function indexesIn(window: Pick<GridWindow, "start" | "end">): number[] {
	return Array.from(
		{ length: Math.max(0, window.end - window.start) },
		(_, offset) => window.start + offset
	);
}

/**
 * The page numbers to load for a window: the pages it touches, plus the pages after it.
 * Pages past the end of the list are left out.
 */
export function pagesToLoad(
	window: Pick<GridWindow, "start" | "end">,
	total: number,
	pageSize = PAGE_SIZE,
	prefetch = PREFETCH_PAGES
): number[] {
	if (window.end <= window.start || total <= 0) return [];
	const first = Math.floor(window.start / pageSize);
	const lastPage = Math.floor((total - 1) / pageSize);
	const last = Math.min(Math.floor((window.end - 1) / pageSize) + prefetch, lastPage);
	return Array.from({ length: Math.max(0, last - first + 1) }, (_, index) => first + index);
}
