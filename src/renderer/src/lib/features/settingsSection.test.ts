import { describe, expect, it } from "vitest";
import { GAME_FOLDERS_SECTION, sectionScroll } from "./settingsSection";

describe("sectionScroll", () => {
	it("returns nothing when no section is pending", () => {
		expect(sectionScroll(null, false)).toBeNull();
		expect(sectionScroll("", false)).toBeNull();
	});

	it("scrolls smoothly to the section", () => {
		expect(sectionScroll(GAME_FOLDERS_SECTION, false)).toEqual({
			id: "game-folders",
			behavior: "smooth"
		});
	});

	it("jumps instead of animating when reduced motion is on", () => {
		expect(sectionScroll(GAME_FOLDERS_SECTION, true)).toEqual({
			id: "game-folders",
			behavior: "auto"
		});
	});
});
