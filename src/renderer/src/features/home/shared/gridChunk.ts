/** Number of game cards rendered per chunk. */
export const GRID_CHUNK_SIZE = 48;

/** Rendered count after one more chunk, never above the total. */
export function nextRenderedCount(renderedCount: number, total: number): number {
	if (renderedCount >= total) return renderedCount;
	return Math.min(renderedCount + GRID_CHUNK_SIZE, total);
}
