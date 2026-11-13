import { describe, expect, it } from "vitest";
import { DEFAULT_COVER_RATIO, coverRatioOf, parseCoverRatio } from "./coverRatio";

describe("parseCoverRatio", () => {
	it("turns W:H into width divided by height", () => {
		expect(parseCoverRatio("4:3")).toBeCloseTo(4 / 3);
		expect(parseCoverRatio("16:9")).toBeCloseTo(16 / 9);
	});

	it("accepts spaces around the numbers", () => {
		expect(parseCoverRatio(" 1:1 ")).toBe(1);
	});

	it("rejects text that is not W:H", () => {
		expect(parseCoverRatio("poster")).toBeNull();
		expect(parseCoverRatio("4x3")).toBeNull();
		expect(parseCoverRatio("")).toBeNull();
		expect(parseCoverRatio(null)).toBeNull();
	});

	it("rejects terms outside 1..32", () => {
		expect(parseCoverRatio("0:3")).toBeNull();
		expect(parseCoverRatio("33:20")).toBeNull();
	});

	it("rejects ratios outside 0.25..4", () => {
		expect(parseCoverRatio("1:5")).toBeNull();
		expect(parseCoverRatio("5:1")).toBeNull();
		expect(parseCoverRatio("1:4")).toBeCloseTo(0.25);
	});
});

describe("coverRatioOf", () => {
	it("uses the parsed ratio when valid", () => {
		expect(coverRatioOf("4:3")).toBeCloseTo(4 / 3);
	});

	it("falls back to 3:4 when the text is missing or invalid", () => {
		expect(coverRatioOf(undefined)).toBe(DEFAULT_COVER_RATIO);
		expect(coverRatioOf("bad")).toBe(DEFAULT_COVER_RATIO);
	});
});
