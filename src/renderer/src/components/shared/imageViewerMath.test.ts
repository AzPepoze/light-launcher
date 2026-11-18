import { describe, expect, it } from "vitest";
import {
	clampPan,
	clampZoom,
	fitScale,
	minimapToPan,
	minimapViewRect,
	zoomAround,
} from "./imageViewerMath";

const STAGE = { width: 800, height: 600 };
const MINIMAP = { width: 160, height: 90 };

describe("clampZoom", () => {
	it("keeps the zoom between the limits", () => {
		expect(clampZoom(0.5, 1, 4)).toBe(1);
		expect(clampZoom(2, 1, 4)).toBe(2);
		expect(clampZoom(9, 1, 4)).toBe(4);
	});
});

describe("fitScale", () => {
	it("shrinks an image bigger than the stage", () => {
		expect(fitScale(1600, 1200, 800, 600)).toBe(0.5);
	});

	it("scales up an image smaller than the stage", () => {
		expect(fitScale(400, 300, 800, 600)).toBe(2);
	});

	it("uses the tighter axis so the whole image fits", () => {
		expect(fitScale(1920, 1080, 800, 600)).toBeCloseTo(800 / 1920);
	});
});

describe("clampPan", () => {
	it("centres an image that fits the stage", () => {
		expect(clampPan({ x: 50, y: -30 }, 600, 300, 800, 600)).toEqual({ x: 0, y: 0 });
	});

	it("clamps at both horizontal edges", () => {
		// Scaled width 1200 on an 800 stage: the pan can move 200px either way.
		expect(clampPan({ x: 500, y: 0 }, 1200, 600, 800, 600)).toEqual({ x: 200, y: 0 });
		expect(clampPan({ x: -500, y: 0 }, 1200, 600, 800, 600)).toEqual({ x: -200, y: 0 });
	});

	it("leaves a pan inside the limits unchanged", () => {
		expect(clampPan({ x: 120, y: -40 }, 1200, 900, 800, 600)).toEqual({ x: 120, y: -40 });
	});
});

describe("zoomAround", () => {
	it("keeps the image point under the pointer fixed", () => {
		const pointer = { x: 100, y: -50 };
		const pan = { x: 10, y: 20 };
		const before = { x: (pointer.x - pan.x) / 1, y: (pointer.y - pan.y) / 1 };
		const nextPan = zoomAround(pointer, pan, 1, 2.5);
		const after = { x: (pointer.x - nextPan.x) / 2.5, y: (pointer.y - nextPan.y) / 2.5 };
		expect(after.x).toBeCloseTo(before.x);
		expect(after.y).toBeCloseTo(before.y);
	});

	it("does not move the pan when the pointer is at the centre", () => {
		expect(zoomAround({ x: 0, y: 0 }, { x: 0, y: 0 }, 1, 3)).toEqual({ x: 0, y: 0 });
	});
});

describe("minimapViewRect", () => {
	it("is the whole minimap at fit", () => {
		const image = { width: 1600, height: 1200 };
		const scale = fitScale(image.width, image.height, STAGE.width, STAGE.height);
		const rect = minimapViewRect({ x: 0, y: 0 }, scale, STAGE, image, MINIMAP);
		expect(rect).toEqual({ x: 0, y: 0, width: 160, height: 90 });
	});

	it("is a smaller rectangle when zoomed in and centred", () => {
		const image = { width: 1600, height: 1200 };
		// At scale 1 the image is 1600x1200 and the stage shows half of it in each axis.
		const rect = minimapViewRect({ x: 0, y: 0 }, 1, STAGE, image, MINIMAP);
		expect(rect.width).toBeCloseTo(80);
		expect(rect.height).toBeCloseTo(45);
		expect(rect.x).toBeCloseTo(40);
		expect(rect.y).toBeCloseTo(22.5);
	});
});

describe("minimapToPan", () => {
	it("centres the view on the clicked minimap point", () => {
		const image = { width: 1600, height: 1200 };
		// Click the middle of the top-left quarter of the minimap.
		const pan = minimapToPan({ x: 40, y: 22.5 }, MINIMAP, image, 1, STAGE);
		// The view centre should now sit at image point (400, 300).
		const viewCentreX = 1600 / 2 - pan.x;
		const viewCentreY = 1200 / 2 - pan.y;
		expect(viewCentreX).toBeCloseTo(400);
		expect(viewCentreY).toBeCloseTo(300);
	});

	it("clamps so the view stays inside the image", () => {
		const image = { width: 1600, height: 1200 };
		const pan = minimapToPan({ x: 0, y: 0 }, MINIMAP, image, 1, STAGE);
		expect(pan).toEqual({ x: 400, y: 300 });
	});
});
