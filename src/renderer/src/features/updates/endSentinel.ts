import { observeOnce } from "../../lib/lazyObserver";

/** Calls onNear when the element comes near the viewport, if active. Returns the function that stops watching. */
export function armEndSentinel(element: Element | undefined, active: boolean, onNear: () => void): () => void {
	if (!element || !active) return () => {};
	return observeOnce(element, onNear);
}
