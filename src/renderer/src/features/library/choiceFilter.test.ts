import { describe, expect, it } from "vitest";
import { CHOICE_SEARCH_THRESHOLD, filterChoices, showsChoiceSearch } from "./choiceFilter";

const TAGS = [
	{ value: "tag-a", label: "Tag A" },
	{ value: "tag-b", label: "Tag B" },
	{ value: "other", label: "Other" }
];

const none = () => false;

describe("showsChoiceSearch", () => {
	it("shows the field above the threshold or when the group is searchable", () => {
		expect(showsChoiceSearch(CHOICE_SEARCH_THRESHOLD + 1, false)).toBe(true);
		expect(showsChoiceSearch(2, true)).toBe(true);
	});

	it("hides the field for a short, unsearchable group", () => {
		expect(showsChoiceSearch(CHOICE_SEARCH_THRESHOLD, false)).toBe(false);
	});
});

describe("filterChoices", () => {
	it("keeps every choice for an empty or blank query", () => {
		expect(filterChoices(TAGS, "", none).choices).toHaveLength(3);
		expect(filterChoices(TAGS, "   ", none).choices).toHaveLength(3);
	});

	it("matches the label without case, ignoring spaces at the ends", () => {
		const result = filterChoices(TAGS, "  TAG  ", none);
		expect(result.choices.map((choice) => choice.value)).toEqual(["tag-a", "tag-b"]);
		expect(result.noMatches).toBe(false);
	});

	it("keeps selected choices first even when they do not match", () => {
		const result = filterChoices(TAGS, "other", (value) => value === "tag-a");
		expect(result.choices.map((choice) => choice.value)).toEqual(["tag-a", "other"]);
	});

	it("reports no matches when nothing unselected matches the query", () => {
		const result = filterChoices(TAGS, "zzz", none);
		expect(result.choices).toEqual([]);
		expect(result.noMatches).toBe(true);
	});

	it("does not report no matches when a selected choice is shown", () => {
		const result = filterChoices(TAGS, "zzz", (value) => value === "tag-b");
		expect(result.choices.map((choice) => choice.value)).toEqual(["tag-b"]);
		expect(result.noMatches).toBe(true);
	});
});
