import { describe, expect, it } from "vitest";
import {
	GAME_FOLDERS_SECTION,
	SETTINGS_NAV_GROUPS,
	sectionScroll,
	settingsNavGroups
} from "./settingsSection";

describe("settingsNavGroups", () => {
	const allShown = { linux: true, prefix: true };

	it("keeps every section on Linux with prefix support", () => {
		const ids = settingsNavGroups(allShown).flatMap((group) => group.items.map((item) => item.id));
		expect(ids).toEqual(SETTINGS_NAV_GROUPS.flatMap((group) => group.items.map((item) => item.id)));
	});

	it("drops the platform-only sections and their empty groups", () => {
		const groups = settingsNavGroups({ linux: false, prefix: false });
		const ids = groups.flatMap((group) => group.items.map((item) => item.id));
		expect(ids).not.toContain("display-platform");
		expect(ids).not.toContain("prefix-storage");
		expect(groups.map((group) => group.label)).toContain("Wine & display");
		const linuxOnly = [{ label: "Only Linux", items: [{ id: "x", label: "X", requires: "linux" as const }] }];
		expect(settingsNavGroups({ linux: false, prefix: false }, linuxOnly)).toEqual([]);
	});
});

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
