import { describe, expect, it } from "vitest";
import { sparkPoints, sparkShape } from "./sparkline";

describe("sparkPoints", () => {
	it("scales to the largest sample and puts the newest on the right", () => {
		const points = sparkPoints([1, 4], 100, 50);
		expect(points[0].x).toBe(0);
		expect(points[1].x).toBe(100);
		expect(points[1].y).toBe(1);
		expect(points[0].y).toBeGreaterThan(points[1].y);
	});

	it("draws a single sample at the right edge", () => {
		expect(sparkPoints([7], 100, 50)[0].x).toBe(100);
	});

	it("keeps zero samples on the baseline", () => {
		expect(sparkPoints([0, 0], 100, 50).map((point) => point.y)).toEqual([49, 49]);
	});
});

describe("sparkShape", () => {
	it("returns a flat baseline with no last point when empty", () => {
		const shape = sparkShape([], 100, 56);
		expect(shape.last).toBeNull();
		expect(shape.line).toBe("M0 55 H100");
	});

	it("closes the area down to the bottom edge", () => {
		const shape = sparkShape([2, 4, 3], 100, 56);
		expect(shape.line.startsWith("M0 ")).toBe(true);
		expect(shape.area.endsWith("V56 H0 Z")).toBe(true);
		expect(shape.last?.x).toBe(100);
	});
});
