import { describe, expect, it } from "vitest";
import { tooltipPlacement } from "./tooltip";

const VIEWPORT = { width: 1440, height: 900 };
const BUBBLE = { width: 120, height: 30 };

function anchorAt(left: number, top: number, width = 36, height = 36) {
	return { left, top, right: left + width, bottom: top + height, width };
}

describe("tooltipPlacement", () => {
	it("places the bubble above and centred when there is room", () => {
		const result = tooltipPlacement(anchorAt(500, 300), BUBBLE, VIEWPORT);

		expect(result.placement).toBe("above");
		expect(result.top).toBe(300 - 30 - 8);
		expect(result.left).toBe(500 + 18 - 60);
	});

	it("places the bubble below when there is not room above", () => {
		const result = tooltipPlacement(anchorAt(500, 20), BUBBLE, VIEWPORT);

		expect(result.placement).toBe("below");
		expect(result.top).toBe(20 + 36 + 8);
	});

	it("switches to below when the space above is exactly one short", () => {
		const justShort = tooltipPlacement(anchorAt(500, 37), BUBBLE, VIEWPORT);
		const justEnough = tooltipPlacement(anchorAt(500, 38), BUBBLE, VIEWPORT);

		expect(justShort.placement).toBe("below");
		expect(justEnough.placement).toBe("above");
	});

	it("clamps to 8 px from the left edge", () => {
		const result = tooltipPlacement(anchorAt(2, 300), BUBBLE, VIEWPORT);

		expect(result.left).toBe(8);
	});

	it("clamps to 8 px inside the right edge", () => {
		const result = tooltipPlacement(anchorAt(1430, 300), BUBBLE, VIEWPORT);

		expect(result.left).toBe(1440 - 8 - 120);
	});
});
