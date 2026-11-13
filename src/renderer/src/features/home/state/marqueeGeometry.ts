export interface MarqueeRect {
	x: number;
	y: number;
	w: number;
	h: number;
}

export interface RectBounds {
	left: number;
	top: number;
	right: number;
	bottom: number;
}

export interface ElementBounds {
	left: number;
	top: number;
	right: number;
	bottom: number;
}

/** Builds the visible marquee rectangle, clamped to the scroller viewport. */
export function marqueeRectFromPoints(
	originX: number,
	originY: number,
	pointerX: number,
	pointerY: number,
	bounds: RectBounds
): MarqueeRect {
	const x1 = Math.max(Math.min(originX, pointerX), bounds.left);
	const x2 = Math.min(Math.max(originX, pointerX), bounds.right);
	const y1 = Math.max(Math.min(originY, pointerY), bounds.top);
	const y2 = Math.min(Math.max(originY, pointerY), bounds.bottom);
	return { x: x1, y: y1, w: Math.max(0, x2 - x1), h: Math.max(0, y2 - y1) };
}

/** True when the element rectangle intersects any part of the marquee. */
export function rectsIntersect(element: ElementBounds, marquee: MarqueeRect): boolean {
	return (
		element.left < marquee.x + marquee.w &&
		element.right > marquee.x &&
		element.top < marquee.y + marquee.h &&
		element.bottom > marquee.y
	);
}

/**
 * How far a one-line label must slide to reveal its end. The content width is the label's
 * scroll width and the viewport width is the clipping box's client width. Zero when it fits.
 */
export function marqueeTextShift(viewportWidth: number, contentWidth: number): number {
	return Math.max(0, contentWidth - viewportWidth);
}
