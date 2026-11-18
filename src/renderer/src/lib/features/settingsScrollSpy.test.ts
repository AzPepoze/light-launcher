import { describe, expect, it } from "vitest";
import { activeSectionId, navSectionIds } from "./settingsScrollSpy";
import type { SettingsNavGroup } from "./settingsSection";

const groups: SettingsNavGroup[] = [
	{ label: "A", items: [{ id: "one", label: "One" }] },
	{
		label: "B",
		items: [
			{ id: "two", label: "Two" },
			{ id: "three", label: "Three" }
		]
	}
];

describe("navSectionIds", () => {
	it("lists every item id in group order", () => {
		expect(navSectionIds(groups)).toEqual(["one", "two", "three"]);
	});
});

describe("activeSectionId", () => {
	const order = ["one", "two", "three"];

	it("picks the last section that has reached the line", () => {
		const tops = { one: -300, two: -10, three: 400 };
		expect(activeSectionId(order, tops, 24, false)).toBe("two");
	});

	it("picks the first section when none has reached the line yet", () => {
		const tops = { one: 80, two: 500, three: 900 };
		expect(activeSectionId(order, tops, 24, false)).toBe("one");
	});

	it("picks the last section at the bottom of the page", () => {
		const tops = { one: -900, two: -400, three: 300 };
		expect(activeSectionId(order, tops, 24, true)).toBe("three");
	});

	it("skips sections that are not on the page", () => {
		const tops = { one: -300, three: 100 };
		expect(activeSectionId(order, tops, 24, false)).toBe("one");
	});

	it("returns nothing when no section is measured", () => {
		expect(activeSectionId(order, {}, 24, false)).toBeNull();
	});
});
