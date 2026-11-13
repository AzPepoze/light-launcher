import { describe, expect, it } from "vitest";
import { closesDetailOn } from "./detailPanel";

describe("closesDetailOn", () => {
	it("closes on Escape while the panel is open", () => {
		expect(closesDetailOn("Escape", true, false)).toBe(true);
	});

	it("ignores other keys", () => {
		expect(closesDetailOn("Enter", true, false)).toBe(false);
	});

	it("leaves an open dialog to handle Escape", () => {
		expect(closesDetailOn("Escape", true, true)).toBe(false);
	});

	it("does nothing when no panel is open", () => {
		expect(closesDetailOn("Escape", false, false)).toBe(false);
	});
});
