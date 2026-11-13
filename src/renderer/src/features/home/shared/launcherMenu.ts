import { launcherKindLabel } from "@lib/features/launcher";
import type { Launcher } from "@shared";

export interface LauncherMenuChoice {
	/** The picker value, or null when automatic selection is chosen. */
	path: string | null;
	icon: string;
	label: string;
	detail: string;
}

export type LauncherMenuSide = "left" | "right";

/** Width the open menu needs, from the anchor menu's min-width. */
export const LAUNCHER_MENU_WIDTH = 180;

/** Gap between the card and the open menu. */
const MENU_GAP = 8;

/** Right corner by default; left only when the card sits too close to the right edge of the window. */
export function pickLauncherSide(
	cardRight: number,
	viewportWidth: number,
	menuWidth: number = LAUNCHER_MENU_WIDTH
): LauncherMenuSide {
	return viewportWidth - cardRight >= menuWidth ? "right" : "left";
}

/** Fixed position for the open menu: beside the anchor box, outside it, top aligned. */
export function launcherMenuPosition(
	cardRect: Pick<DOMRect, "left" | "right" | "top">,
	side: LauncherMenuSide,
	menuWidth: number = LAUNCHER_MENU_WIDTH
): { x: number; y: number } {
	const x =
		side === "right" ? cardRect.right + MENU_GAP : cardRect.left - MENU_GAP - menuWidth;
	return { x, y: cardRect.top };
}

export type Bounds = Pick<DOMRect, "left" | "top" | "right" | "bottom">;

/** Nearest ancestor that scrolls vertically, or the document's scrolling element. */
export function findScrollParent(element: Element): Element | null {
	for (let parent = element.parentElement; parent; parent = parent.parentElement) {
		if (isVerticallyScrollable(parent)) return parent;
	}
	return document.scrollingElement;
}

function isVerticallyScrollable(element: Element): boolean {
	const { overflowY } = getComputedStyle(element);
	const canScroll = overflowY === "auto" || overflowY === "scroll";
	return canScroll && element.scrollHeight > element.clientHeight;
}

/** The part of an element's box that is inside the window. */
export function visibleArea(element: Element): Bounds {
	const rect = element.getBoundingClientRect();
	return {
		left: Math.max(rect.left, 0),
		top: Math.max(rect.top, 0),
		right: Math.min(rect.right, window.innerWidth),
		bottom: Math.min(rect.bottom, window.innerHeight)
	};
}

/** True when any part of the rect overlaps the container. */
export function isRectVisible(rect: Bounds, container: Bounds): boolean {
	return (
		rect.right > container.left &&
		rect.left < container.right &&
		rect.bottom > container.top &&
		rect.top < container.bottom
	);
}

/** Menu rows: automatic first, then every launcher in the order the backend ranks them. */
export function launcherChoices(launchers: Launcher[]): LauncherMenuChoice[] {
	return [
		{ path: null, icon: "autorenew", label: "Use automatic", detail: "" },
		...launchers.map((launcher) => ({
			path: launcher.path,
			icon: launcher.primary ? "check" : "play_arrow",
			label: launcher.name,
			detail: launcherKindLabel(launcher.kind)
		}))
	];
}
