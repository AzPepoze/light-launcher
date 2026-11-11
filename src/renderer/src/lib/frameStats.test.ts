import { describe, expect, it } from "vitest";
import { FrameHistory, percentile, snapToRefreshRate, summarize } from "./frameStats";

const FRAME_60_MS = 1000 / 60;
const FRAME_180_MS = 1000 / 180;

function steady(frameMs: number, count: number): number[] {
	return Array.from({ length: count }, () => frameMs);
}

describe("percentile", () => {
	it("picks the nearest-rank value", () => {
		expect(percentile([1, 2, 3, 4], 0.5)).toBe(2);
		expect(percentile([1, 2, 3, 4], 0.95)).toBe(4);
		expect(percentile([1, 2, 3, 4], 0)).toBe(1);
	});

	it("returns 0 for empty input", () => {
		expect(percentile([], 0.5)).toBe(0);
	});
});

describe("snapToRefreshRate", () => {
	it("snaps a frame time to the nearest common rate", () => {
		expect(snapToRefreshRate(1000 / 60)).toBe(60);
		expect(snapToRefreshRate(1000 / 180)).toBe(180);
		expect(snapToRefreshRate(1000 / 144)).toBe(144);
		expect(snapToRefreshRate(1000 / 59)).toBe(60);
	});

	it("returns the raw rate when no common rate is within 4 percent", () => {
		expect(snapToRefreshRate(1000 / 36)).toBe(36);
	});

	it("returns 0 for an unknown frame time", () => {
		expect(snapToRefreshRate(0)).toBe(0);
		expect(snapToRefreshRate(-5)).toBe(0);
		expect(snapToRefreshRate(Number.NaN)).toBe(0);
	});
});

describe("summarize", () => {
	it("reports 60 fps on a steady 60 Hz display", () => {
		const stats = summarize(steady(FRAME_60_MS, 300));
		expect(stats.fps).toBe(60);
		expect(stats.p50).toBeCloseTo(FRAME_60_MS);
		expect(stats.p95).toBeCloseTo(FRAME_60_MS);
		expect(stats.worst).toBeCloseTo(FRAME_60_MS);
		expect(stats.longFrames).toBe(0);
		expect(stats.displayHz).toBe(60);
	});

	it("reports 180 fps on a steady 180 Hz display", () => {
		const stats = summarize(steady(FRAME_180_MS, 600));
		expect(stats.fps).toBe(180);
		expect(stats.p50).toBeCloseTo(FRAME_180_MS);
		expect(stats.longFrames).toBe(0);
		expect(stats.displayHz).toBe(180);
	});

	it("counts stutters as long frames and shows them in the worst frame", () => {
		// Every 10th frame takes 50 ms, the rest run at 60 fps.
		const intervals = Array.from({ length: 300 }, (_, i) => (i % 10 === 9 ? 50 : FRAME_60_MS));
		const stats = summarize(intervals);
		expect(stats.fps).toBe(50);
		expect(stats.p50).toBeCloseTo(FRAME_60_MS);
		expect(stats.p95).toBe(50);
		expect(stats.worst).toBe(50);
		expect(stats.longFrames).toBe(30);
		expect(stats.displayHz).toBe(60);
	});

	it("returns zeros for empty input", () => {
		expect(summarize([])).toEqual({
			fps: 0,
			p50: 0,
			p95: 0,
			worst: 0,
			longFrames: 0,
			displayHz: 0
		});
	});

	it("handles a single frame", () => {
		const stats = summarize([16.7]);
		expect(stats.fps).toBe(1);
		expect(stats.p50).toBe(16.7);
		expect(stats.p95).toBe(16.7);
		expect(stats.worst).toBe(16.7);
		expect(stats.longFrames).toBe(0);
		expect(stats.displayHz).toBe(60);
	});
});

describe("FrameHistory", () => {
	it("keeps only the newest values when full", () => {
		const history = new FrameHistory(4);
		[1, 2, 3, 4, 5].forEach((value) => history.push(value));
		expect(Array.from(history.recentWithin(Infinity))).toEqual([2, 3, 4, 5]);
	});

	it("returns the newest values that fit in the window, oldest first", () => {
		const history = new FrameHistory(4);
		[1, 2, 3, 4, 5].forEach((value) => history.push(value));
		expect(Array.from(history.recentWithin(9))).toEqual([4, 5]);
	});

	it("returns nothing before any push", () => {
		expect(new FrameHistory(4).recentWithin(1000).length).toBe(0);
	});
});
