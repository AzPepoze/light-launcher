import { describe, expect, it } from "vitest";
import { tileShapeFor } from "./sourceTileShape";

describe("tileShapeFor", () => {
	it("uses 16:9 covers for wide tiles", () => {
		expect(tileShapeFor("wide")).toEqual({ ratio: 16 / 9, minColumnPx: 240 });
	});

	it("uses 3:4 covers for poster tiles", () => {
		expect(tileShapeFor("poster")).toEqual({ ratio: 3 / 4, minColumnPx: 140 });
	});

	it("uses 1:1 covers for square tiles", () => {
		expect(tileShapeFor("square")).toEqual({ ratio: 1, minColumnPx: 140 });
	});
});
