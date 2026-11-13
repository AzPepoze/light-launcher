import { describe, expect, it } from "vitest";
import { marqueeTextShift } from "./marqueeGeometry";

describe("marqueeTextShift", () => {
	it("keeps a label that fits still", () => {
		expect(marqueeTextShift(180, 180)).toBe(0);
	});

	it("returns the hidden distance of an overflowing label", () => {
		expect(marqueeTextShift(180, 260)).toBe(80);
	});

	it("never shifts a label narrower than its viewport", () => {
		expect(marqueeTextShift(180, 120)).toBe(0);
	});
});
