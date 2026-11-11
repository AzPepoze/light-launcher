import type { GameGroup } from "./libraryGroups";

/** How many tiles take part in the first reveal. */
export const REVEAL_MAX_TILES = 12;
/** Matches --motion-stagger in style.scss. */
export const REVEAL_STEP_MS = 55;
/** Matches --motion-base in style.scss. */
export const REVEAL_TILE_MS = 240;

/** Maps game id to reveal position for the first tiles on screen, in display order. */
export function revealOrder(groups: GameGroup[]): Map<number, number> {
	const order = new Map<number, number>();
	for (const group of groups) {
		for (const game of group.games) {
			if (order.size >= REVEAL_MAX_TILES) return order;
			order.set(game.id, order.size);
		}
	}
	return order;
}

/** Time from the first tile starting to the last one finishing. */
export function revealDoneMs(tileCount: number): number {
	if (tileCount === 0) return 0;
	return (tileCount - 1) * REVEAL_STEP_MS + REVEAL_TILE_MS;
}
