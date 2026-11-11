export interface TileShape {
	/** CSS aspect-ratio of the cover. */
	aspect: string;
	/** Smallest width of one grid column, in pixels. */
	minColumnPx: number;
}

const PORTRAIT: TileShape = { aspect: "3 / 4", minColumnPx: 140 };
const LANDSCAPE: TileShape = { aspect: "16 / 9", minColumnPx: 240 };

const SOURCE_SHAPES = new Map<string, TileShape>([["f95", LANDSCAPE]]);

export function tileShapeFor(source: string): TileShape {
	return SOURCE_SHAPES.get(source) ?? PORTRAIT;
}
