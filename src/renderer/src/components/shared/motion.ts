import { cubicIn, cubicOut } from "svelte/easing";
import type { TransitionConfig } from "svelte/transition";

export type MotionKind = "enter" | "exit";
export type MenuPlacement = "bottom" | "top";

export const ENTER_MS = 200;
export const EXIT_MS = 150;

const MENU_OFFSET_PX = 6;
const MENU_START_SCALE = 0.98;
const DIALOG_START_SCALE = 0.96;
const TOAST_OFFSET_PX = 24;

const REDUCED_MOTION_QUERY = "(prefers-reduced-motion: reduce)";

export function prefersReducedMotion(): boolean {
	return (
		typeof window !== "undefined" &&
		typeof window.matchMedia === "function" &&
		window.matchMedia(REDUCED_MOTION_QUERY).matches
	);
}

export function motionMs(kind: MotionKind, reduced = prefersReducedMotion()): number {
	if (reduced) return 0;
	return kind === "enter" ? ENTER_MS : EXIT_MS;
}

function motionEasing(kind: MotionKind) {
	return kind === "enter" ? cubicOut : cubicIn;
}

/** Fade only, for backdrops. */
export function fadeParams(kind: MotionKind, reduced = prefersReducedMotion()) {
	return { duration: motionMs(kind, reduced), easing: motionEasing(kind) };
}

/** Dialog panel: fades and scales from a slightly smaller size. */
export function fadeScale(kind: MotionKind, reduced = prefersReducedMotion()) {
	return {
		...fadeParams(kind, reduced),
		start: reduced ? 1 : DIALOG_START_SCALE
	};
}

/** Toast: slides in from the right edge with a fade. */
export function slideIn(kind: MotionKind, reduced = prefersReducedMotion()) {
	return {
		...fadeParams(kind, reduced),
		x: reduced ? 0 : TOAST_OFFSET_PX,
		y: 0
	};
}

/** Menu panel config: fades, moves 6 px into place and scales from 0.98. */
export function menuPanelConfig(
	kind: MotionKind,
	placement: MenuPlacement = "bottom",
	reduced = prefersReducedMotion()
): TransitionConfig {
	const offset = reduced ? 0 : MENU_OFFSET_PX * (placement === "top" ? 1 : -1);
	const startScale = reduced ? 1 : MENU_START_SCALE;
	return {
		duration: motionMs(kind, reduced),
		easing: motionEasing(kind),
		css: (t, u) =>
			`opacity: ${t}; transform: translateY(${u * offset}px) scale(${startScale + (1 - startScale) * t});`
	};
}

/** Transition function for menu panels, used with in: and out:. */
export function menuPanel(
	_node: Element,
	params: { placement?: MenuPlacement } = {},
	options: { direction: "in" | "out" | "both" }
): TransitionConfig {
	const kind: MotionKind = options.direction === "out" ? "exit" : "enter";
	return menuPanelConfig(kind, params.placement);
}
