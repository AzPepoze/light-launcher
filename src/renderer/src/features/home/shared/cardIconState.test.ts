import { describe, expect, it } from "vitest";
import { cardIconState } from "./cardIconState";

describe("cardIconState", () => {
	it("shows the icon once it is resolved", () => {
		expect(cardIconState({ icon: "data:image/png;base64,abc", pending: true })).toBe("ready");
	});

	it("shows the placeholder while the icon is pending", () => {
		expect(cardIconState({ icon: "", pending: true })).toBe("pending");
		expect(cardIconState({ icon: null, pending: true })).toBe("pending");
	});

	it("shows the fallback when the game has no icon", () => {
		expect(cardIconState({ icon: "", pending: false })).toBe("missing");
		expect(cardIconState({})).toBe("missing");
	});

	it("shows the fallback after a failed image load", () => {
		expect(cardIconState({ icon: "data:image/png;base64,abc", failed: true })).toBe("missing");
	});

	it("does not stay pending after a failed load", () => {
		expect(cardIconState({ icon: "", pending: false, failed: true })).toBe("missing");
	});
});
