/** What a game card should show in its icon slot. */
export type CardIconState = "pending" | "ready" | "missing";

export interface CardIconStateInput {
	/** Icon URL from the icon cache; empty when none has been resolved yet. */
	icon?: string | null;
	/** True while the icon request has not settled. */
	pending?: boolean;
	/** True when the icon URL itself failed to load. */
	failed?: boolean;
}

/**
 * Picks the icon slot content:
 * - `ready` renders the image,
 * - `pending` renders a neutral placeholder so the rocket fallback does not flash,
 * - `missing` renders the rocket fallback for a game without an icon or a failed load.
 */
export function cardIconState({
	icon,
	pending = false,
	failed = false
}: CardIconStateInput): CardIconState {
	if (failed) return "missing";
	if (icon) return "ready";
	return pending ? "pending" : "missing";
}
