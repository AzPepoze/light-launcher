/** How many tiles take part in the first reveal. */
export const REVEAL_MAX_TILES = 12;
/** Matches --motion-stagger in style.scss. */
export const REVEAL_STEP_MS = 55;
/** Matches --motion-base in style.scss. */
export const REVEAL_TILE_MS = 240;

/** The reveal position of a tile on screen, or undefined when it does not animate in. */
export function revealIndex(index: number, revealed: boolean): number | undefined {
	if (revealed || index >= REVEAL_MAX_TILES) return undefined;
	return index;
}

/** Time from the first tile starting to the last one finishing. */
export function revealDoneMs(tileCount: number): number {
	if (tileCount === 0) return 0;
	return (tileCount - 1) * REVEAL_STEP_MS + REVEAL_TILE_MS;
}
