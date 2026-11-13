import { describe, expect, it } from "vitest";
import {
	DROPDOWN_SEARCH_THRESHOLD,
	filterDropdownOptions,
	showsDropdownSearch
} from "./dropdownSearch";

const OPTIONS = [
	{ value: "new", label: "New" },
	{ value: "done", label: "Done" },
	{ value: "hidden-key", label: "Mystery" }
];

describe("showsDropdownSearch", () => {
	it("shows the box above the threshold when the dropdown is searchable", () => {
		expect(showsDropdownSearch(DROPDOWN_SEARCH_THRESHOLD + 1, true)).toBe(true);
	});

	it("hides the box at the threshold", () => {
		expect(showsDropdownSearch(DROPDOWN_SEARCH_THRESHOLD, true)).toBe(false);
	});

	it("hides the box when the dropdown is not searchable", () => {
		expect(showsDropdownSearch(DROPDOWN_SEARCH_THRESHOLD + 1, false)).toBe(false);
	});
});

describe("filterDropdownOptions", () => {
	it("keeps every option for an empty or blank query", () => {
		expect(filterDropdownOptions(OPTIONS, "")).toBe(OPTIONS);
		expect(filterDropdownOptions(OPTIONS, "   ")).toBe(OPTIONS);
	});

	it("matches labels without case, ignoring spaces at the ends", () => {
		expect(filterDropdownOptions(OPTIONS, "  MYST  ")).toEqual([
			{ value: "hidden-key", label: "Mystery" }
		]);
	});

	it("matches values too, without case", () => {
		expect(filterDropdownOptions(OPTIONS, "HIDDEN")).toEqual([
			{ value: "hidden-key", label: "Mystery" }
		]);
	});

	it("returns no options when nothing matches", () => {
		expect(filterDropdownOptions(OPTIONS, "zzz")).toEqual([]);
	});
});
