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

/**
 * How far a one-line label must slide to reveal its end. The content width is the label's
 * scroll width and the viewport width is the clipping box's client width. Zero when it fits.
 */
export function marqueeTextShift(viewportWidth: number, contentWidth: number): number {
	return Math.max(0, contentWidth - viewportWidth);
}
