/*
 * Geometry for the image viewer. No DOM here.
 * Pan and pointers are offsets from the stage centre, in stage pixels.
 */

export interface Point {
	x: number;
	y: number;
}

export interface Size {
	width: number;
	height: number;
}

export interface Rect extends Point, Size {}

export function clampZoom(zoom: number, min: number, max: number): number {
	return Math.min(max, Math.max(min, zoom));
}

/** The scale at which the whole image fits the stage. Smaller images scale up to fill it. */
export function fitScale(imageW: number, imageH: number, stageW: number, stageH: number): number {
	return Math.min(stageW / imageW, stageH / imageH);
}

function clampAxis(offset: number, scaled: number, stage: number): number {
	if (scaled <= stage) return 0;
	const limit = (scaled - stage) / 2;
	return Math.min(limit, Math.max(-limit, offset));
}

/** Keeps the image edges inside the stage. An axis that fits the stage is centred. */
export function clampPan(
	pan: Point,
	scaledW: number,
	scaledH: number,
	stageW: number,
	stageH: number,
): Point {
	return {
		x: clampAxis(pan.x, scaledW, stageW),
		y: clampAxis(pan.y, scaledH, stageH),
	};
}

/** The pan that keeps the image point under the pointer still while the scale changes. */
export function zoomAround(
	pointer: Point,
	pan: Point,
	oldScale: number,
	newScale: number,
): Point {
	const ratio = newScale / oldScale;
	return {
		x: pointer.x - (pointer.x - pan.x) * ratio,
		y: pointer.y - (pointer.y - pan.y) * ratio,
	};
}

/**
 * The part of the image the stage shows, drawn inside the minimap.
 * At fit the rectangle is the whole minimap.
 */
export function minimapViewRect(
	pan: Point,
	scale: number,
	stage: Size,
	image: Size,
	minimap: Size,
): Rect {
	const scaledW = image.width * scale;
	const scaledH = image.height * scale;
	const widthFraction = Math.min(1, stage.width / scaledW);
	const heightFraction = Math.min(1, stage.height / scaledH);
	// The stage centre sits at this fraction of the scaled image, measured from its left and top.
	const leftFraction = (scaledW / 2 - pan.x - stage.width / 2) / scaledW;
	const topFraction = (scaledH / 2 - pan.y - stage.height / 2) / scaledH;
	const left = Math.min(1 - widthFraction, Math.max(0, leftFraction));
	const top = Math.min(1 - heightFraction, Math.max(0, topFraction));
	return {
		x: left * minimap.width,
		y: top * minimap.height,
		width: widthFraction * minimap.width,
		height: heightFraction * minimap.height,
	};
}

/** The pan that centres the view on a point clicked in the minimap, kept inside the image. */
export function minimapToPan(
	point: Point,
	minimap: Size,
	image: Size,
	scale: number,
	stage: Size,
): Point {
	const scaledW = image.width * scale;
	const scaledH = image.height * scale;
	const fractionX = point.x / minimap.width;
	const fractionY = point.y / minimap.height;
	const pan = {
		x: scaledW * (0.5 - fractionX),
		y: scaledH * (0.5 - fractionY),
	};
	return clampPan(pan, scaledW, scaledH, stage.width, stage.height);
}
