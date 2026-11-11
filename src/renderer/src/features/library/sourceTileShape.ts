import type { SourceTile } from "../../lib/features/types";

export interface TileShape {
	/** Width divided by height of the cover. Used as CSS aspect-ratio. */
	ratio: number;
	/** Smallest width of one grid column, in pixels. */
	minColumnPx: number;
}

const SHAPES: Record<SourceTile, TileShape> = {
	poster: { ratio: 3 / 4, minColumnPx: 140 },
	wide: { ratio: 16 / 9, minColumnPx: 240 },
	square: { ratio: 1, minColumnPx: 140 }
};

/** The cover shape a source declares in its plugin manifest. */
export function tileShapeFor(tile: SourceTile): TileShape {
	return SHAPES[tile];
}
