import { describe, expect, it } from "vitest";
import { tileShapeFor } from "./sourceTileShape";

describe("tileShapeFor", () => {
	it("uses wide tiles for f95 covers", () => {
		expect(tileShapeFor("f95")).toEqual({ aspect: "16 / 9", minColumnPx: 240 });
	});

	it("uses portrait tiles for other sources", () => {
		expect(tileShapeFor("itch")).toEqual({ aspect: "3 / 4", minColumnPx: 140 });
	});

	it("uses portrait tiles for unknown names that look like object keys", () => {
		expect(tileShapeFor("constructor").aspect).toBe("3 / 4");
	});
});
