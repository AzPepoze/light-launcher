export type TooltipAnchor = {
	top: number;
	bottom: number;
	left: number;
	right: number;
	width: number;
};

export type TooltipBubbleSize = { width: number; height: number };

export type TooltipViewport = { width: number; height: number };

export type TooltipPlacement = {
	top: number;
	left: number;
	placement: "above" | "below";
};

const GAP = 8;
const EDGE_MARGIN = 8;
const SHOW_DELAY_MS = 350;
const FADE_MS = 120;
const BUBBLE_ID = "app-tooltip";

/** Places the bubble centred above the anchor, or below when there is no room above. */
export function tooltipPlacement(
	anchor: TooltipAnchor,
	bubble: TooltipBubbleSize,
	viewport: TooltipViewport,
): TooltipPlacement {
	const spaceNeeded = bubble.height + GAP;
	const above = anchor.top >= spaceNeeded;
	const top = above ? anchor.top - spaceNeeded : anchor.bottom + GAP;

	const centredLeft = anchor.left + anchor.width / 2 - bubble.width / 2;
	const rightmostLeft = viewport.width - EDGE_MARGIN - bubble.width;
	const left = Math.max(EDGE_MARGIN, Math.min(centredLeft, rightmostLeft));

	return { top, left, placement: above ? "above" : "below" };
}

let bubble: HTMLElement | null = null;
/** The node whose bubble is on screen, or null when no bubble is shown. */
let shownFor: HTMLElement | null = null;

function prefersReducedMotion(): boolean {
	return window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? false;
}

function createBubble(): HTMLElement {
	const element = document.createElement("div");
	element.id = BUBBLE_ID;
	element.setAttribute("role", "tooltip");
	Object.assign(element.style, {
		position: "fixed",
		top: "0px",
		left: "0px",
		zIndex: "1000",
		pointerEvents: "none",
		background: "var(--bg-elevated)",
		color: "var(--text-main)",
		border: "1px solid var(--glass-border)",
		borderRadius: "var(--radius-md)",
		fontSize: "0.8rem",
		padding: "6px 10px",
		maxWidth: "260px",
		whiteSpace: "normal",
		opacity: "0",
		visibility: "hidden",
		transition: prefersReducedMotion()
			? "none"
			: `opacity ${FADE_MS}ms ease, visibility ${FADE_MS}ms ease`,
	});
	document.body.append(element);
	return element;
}

function getBubble(): HTMLElement {
	bubble ??= createBubble();
	return bubble;
}

function placeBubble(node: HTMLElement, element: HTMLElement) {
	const placed = tooltipPlacement(
		node.getBoundingClientRect(),
		{ width: element.offsetWidth, height: element.offsetHeight },
		{ width: window.innerWidth, height: window.innerHeight },
	);
	element.style.top = `${placed.top}px`;
	element.style.left = `${placed.left}px`;
}

function showBubble(node: HTMLElement, text: string) {
	const element = getBubble();
	if (shownFor && shownFor !== node) {
		shownFor.removeAttribute("aria-describedby");
	}
	shownFor = node;
	element.textContent = text;
	node.setAttribute("aria-describedby", element.id);
	element.style.visibility = "visible";
	element.style.opacity = "1";
	placeBubble(node, element);
	window.addEventListener("scroll", hideBubble, { capture: true, passive: true });
	window.addEventListener("keydown", onWindowKeydown);
}

function hideBubble() {
	if (!shownFor || !bubble) return;
	shownFor.removeAttribute("aria-describedby");
	shownFor = null;
	bubble.style.opacity = "0";
	bubble.style.visibility = "hidden";
	window.removeEventListener("scroll", hideBubble, { capture: true });
	window.removeEventListener("keydown", onWindowKeydown);
}

function hideBubbleFor(node: HTMLElement) {
	if (shownFor === node) hideBubble();
}

function onWindowKeydown(event: KeyboardEvent) {
	if (event.key === "Escape") hideBubble();
}

/** Shows a tooltip bubble after a short delay on hover or keyboard focus. */
export function tooltip(node: HTMLElement, text: string) {
	let label = text;
	let timer: ReturnType<typeof setTimeout> | undefined;
	/** Set by a click, so the focus that follows it does not reopen the bubble. */
	let suppressed = false;

	function cancelTimer() {
		clearTimeout(timer);
		timer = undefined;
	}

	function scheduleShow() {
		cancelTimer();
		if (!label) return;
		timer = setTimeout(() => {
			timer = undefined;
			showBubble(node, label);
		}, SHOW_DELAY_MS);
	}

	function hide() {
		cancelTimer();
		hideBubbleFor(node);
	}

	function onPointerEnter() {
		suppressed = false;
		scheduleShow();
	}

	function onPointerLeave() {
		suppressed = false;
		hide();
	}

	function onFocus() {
		if (!suppressed) scheduleShow();
	}

	function onBlur() {
		suppressed = false;
		hide();
	}

	function onClick() {
		suppressed = true;
		hide();
	}

	node.addEventListener("pointerenter", onPointerEnter);
	node.addEventListener("pointerleave", onPointerLeave);
	node.addEventListener("focus", onFocus);
	node.addEventListener("blur", onBlur);
	node.addEventListener("click", onClick);
	node.removeAttribute("title");

	return {
		update(nextText: string) {
			label = nextText;
			if (shownFor === node) {
				showBubble(node, label);
			}
		},
		destroy() {
			cancelTimer();
			hideBubbleFor(node);
			node.removeEventListener("pointerenter", onPointerEnter);
			node.removeEventListener("pointerleave", onPointerLeave);
			node.removeEventListener("focus", onFocus);
			node.removeEventListener("blur", onBlur);
			node.removeEventListener("click", onClick);
		},
	};
}
