const DEAD_ZONE_PX = 12;
const SPEED_GAIN_PER_PX = 6;
const MAX_SPEED_PX_PER_SECOND = 2400;
const DRAG_CANCEL_PX = 6;
const DRAG_CANCEL_MS = 250;
const MAX_FRAME_MS = 50;
const SCROLL_OVERFLOW_VALUES = new Set(["auto", "scroll", "overlay"]);
const BLOCKED_TAGS = new Set(["INPUT", "TEXTAREA"]);

/** Signed scroll speed in px per second for a pointer `offsetPx` from the anchor. */
export function autoscrollSpeed(offsetPx: number): number {
	if (!Number.isFinite(offsetPx)) return 0;
	const distance = Math.abs(offsetPx);
	if (distance <= DEAD_ZONE_PX) return 0;
	const speed = Math.min((distance - DEAD_ZONE_PX) * SPEED_GAIN_PER_PX, MAX_SPEED_PX_PER_SECOND);
	return offsetPx < 0 ? -speed : speed;
}

/** True for links, form fields and editable regions, where middle-click has its own meaning. */
export function isAutoscrollBlocked(target: Element | null): boolean {
	for (let node = target; node; node = node.parentElement) {
		if (BLOCKED_TAGS.has(node.tagName)) return true;
		if (node.tagName === "A" && node.getAttribute("href") !== null) return true;
		const editable = node.getAttribute("contenteditable");
		if (editable !== null && editable !== "false") return true;
	}
	return false;
}

/** Nearest ancestor that can scroll on either axis, or null when there is none. */
export function findScrollableAncestor(start: Element | null): HTMLElement | null {
	for (let node = start; node; node = node.parentElement) {
		if (isScrollable(node as HTMLElement)) return node as HTMLElement;
	}
	const root = start?.ownerDocument?.scrollingElement;
	return root && isRootScrollable(root as HTMLElement) ? (root as HTMLElement) : null;
}

function isScrollable(element: HTMLElement): boolean {
	const style = element.ownerDocument?.defaultView?.getComputedStyle(element);
	if (!style) return false;
	const vertical =
		element.scrollHeight > element.clientHeight && SCROLL_OVERFLOW_VALUES.has(style.overflowY);
	const horizontal =
		element.scrollWidth > element.clientWidth && SCROLL_OVERFLOW_VALUES.has(style.overflowX);
	return vertical || horizontal;
}

function isRootScrollable(root: HTMLElement): boolean {
	return root.scrollHeight > root.clientHeight || root.scrollWidth > root.clientWidth;
}

interface AutoscrollSession {
	scroller: HTMLElement;
	marker: HTMLElement;
	anchorX: number;
	anchorY: number;
	pointerX: number;
	pointerY: number;
	startX: number;
	startY: number;
	startedAt: number;
	pressed: boolean;
	dragged: boolean;
	lastFrame: number;
	frame: number;
}

/**
 * Adds Chrome-style middle-button autoscroll to any engine.
 * Returns a cleanup function that removes the listeners.
 */
export function installAutoscroll(doc: Document = document): () => void {
	const win = doc.defaultView;
	if (!win) return () => {};

	let session: AutoscrollSession | null = null;

	function now(): number {
		return win.performance?.now() ?? Date.now();
	}

	function stop(): void {
		if (!session) return;
		win.cancelAnimationFrame(session.frame);
		session.marker.remove();
		session = null;
	}

	function tick(time: number): void {
		if (!session) return;
		const elapsed = Math.min(Math.max(time - session.lastFrame, 0), MAX_FRAME_MS) / 1000;
		session.lastFrame = time;
		const left = autoscrollSpeed(session.pointerX - session.anchorX) * elapsed;
		const top = autoscrollSpeed(session.pointerY - session.anchorY) * elapsed;
		if (left !== 0 || top !== 0) {
			session.scroller.scrollBy(left, top);
		}
		session.frame = win.requestAnimationFrame(tick);
	}

	function start(scroller: HTMLElement, event: MouseEvent): void {
		const marker = doc.createElement("div");
		marker.className = "autoscroll-marker";
		marker.setAttribute("aria-hidden", "true");
		marker.style.left = `${event.clientX}px`;
		marker.style.top = `${event.clientY}px`;
		(doc.body ?? doc.documentElement).append(marker);

		session = {
			scroller,
			marker,
			anchorX: event.clientX,
			anchorY: event.clientY,
			pointerX: event.clientX,
			pointerY: event.clientY,
			startX: event.clientX,
			startY: event.clientY,
			startedAt: now(),
			pressed: true,
			dragged: false,
			lastFrame: now(),
			frame: 0
		};
		session.frame = win.requestAnimationFrame(tick);
	}

	function handleMouseDown(event: MouseEvent): void {
		stop();
		if (event.button !== 1) return;
		const target = event.target instanceof Element ? event.target : null;
		if (!target || isAutoscrollBlocked(target)) return;
		const scroller = findScrollableAncestor(target);
		if (!scroller) return;
		event.preventDefault();
		start(scroller, event);
	}

	function handleMouseMove(event: MouseEvent): void {
		if (!session) return;
		session.pointerX = event.clientX;
		session.pointerY = event.clientY;
		if (!session.pressed) return;
		const dx = event.clientX - session.startX;
		const dy = event.clientY - session.startY;
		if (Math.hypot(dx, dy) > DRAG_CANCEL_PX) session.dragged = true;
	}

	function handleMouseUp(event: MouseEvent): void {
		if (!session || event.button !== 1) return;
		if (session.dragged || now() - session.startedAt > DRAG_CANCEL_MS) {
			stop();
			return;
		}
		session.pressed = false;
	}

	function handleAuxClick(event: MouseEvent): void {
		if (session && event.button === 1) event.preventDefault();
	}

	function handleWheel(): void {
		stop();
	}

	function handleKeyDown(event: KeyboardEvent): void {
		if (event.key === "Escape") stop();
	}

	doc.addEventListener("mousedown", handleMouseDown, true);
	doc.addEventListener("mousemove", handleMouseMove, true);
	doc.addEventListener("mouseup", handleMouseUp, true);
	doc.addEventListener("auxclick", handleAuxClick, true);
	doc.addEventListener("wheel", handleWheel, { capture: true, passive: true });
	doc.addEventListener("keydown", handleKeyDown, true);
	win.addEventListener("blur", stop);

	return () => {
		stop();
		doc.removeEventListener("mousedown", handleMouseDown, true);
		doc.removeEventListener("mousemove", handleMouseMove, true);
		doc.removeEventListener("mouseup", handleMouseUp, true);
		doc.removeEventListener("auxclick", handleAuxClick, true);
		doc.removeEventListener("wheel", handleWheel, true);
		doc.removeEventListener("keydown", handleKeyDown, true);
		win.removeEventListener("blur", stop);
	};
}
