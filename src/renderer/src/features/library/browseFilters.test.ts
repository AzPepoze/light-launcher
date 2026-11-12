import { describe, expect, it } from "vitest";
import type { SidebarGroup, SidebarOption } from "../../lib/features/types";
import {
	browseQuery,
	hasBrowseFilters,
	initialBrowseFilters,
	setBrowseQuery,
	setMulti,
	setOrder,
	setSelect,
	setSort,
	setToggle
} from "./browseFilters";

function option(value: string): SidebarOption {
	return { value, label: value, icon: null, count: null };
}

const GROUPS: SidebarGroup[] = [
	{
		type: "search",
		id: "text",
		label: "",
		placeholder: "Search",
		options: [],
		dynamic: false,
		searchable: false,
		orderable: false,
		default: null
	},
	{
		type: "toggles",
		id: "show",
		label: "Show",
		placeholder: "",
		options: [option("installed"), option("wishlisted")],
		dynamic: false,
		searchable: false,
		orderable: false,
		default: null
	},
	{
		type: "select",
		id: "status",
		label: "Status",
		placeholder: "",
		options: [option("new"), option("done")],
		dynamic: false,
		searchable: false,
		orderable: false,
		default: null
	},
	{
		type: "multi",
		id: "tags",
		label: "Tags",
		placeholder: "",
		options: [],
		dynamic: true,
		searchable: true,
		orderable: false,
		default: null
	},
	{
		type: "sort",
		id: "sort",
		label: "",
		placeholder: "",
		options: [option("title"), option("date")],
		dynamic: false,
		searchable: false,
		orderable: true,
		default: { value: "date", order: "desc" }
	}
];

describe("initialBrowseFilters", () => {
	it("starts with the declared sort default and every filter empty or off", () => {
		expect(initialBrowseFilters(GROUPS)).toEqual({
			query: "",
			values: { installed: false, wishlisted: false, status: "", tags: [] },
			sort: "date",
			order: "desc"
		});
	});

	it("uses the first sort option in ascending order without a default", () => {
		const sort = GROUPS.filter((group) => group.type === "sort").map((group) => ({
			...group,
			default: null
		}));
		expect(initialBrowseFilters(sort)).toMatchObject({ sort: "title", order: "asc" });
	});

	it("leaves the sort empty when no group declares one", () => {
		expect(initialBrowseFilters([])).toEqual({ query: "", values: {}, sort: "", order: "asc" });
	});
});

describe("browseQuery", () => {
	it("sends the search text, active filters, sort and page limit", () => {
		const filters = {
			...initialBrowseFilters(GROUPS),
			query: "  harbor ",
			values: { installed: true, wishlisted: false, status: "", tags: ["rpg", "horror"] }
		};
		expect(browseQuery("shop", filters, "next")).toEqual({
			source: "shop",
			query: "harbor",
			filters: { installed: true, tags: ["rpg", "horror"] },
			sort: "date",
			order: "desc",
			cursor: "next",
			limit: 60
		});
	});

	it("leaves an empty sort and its order out", () => {
		expect(browseQuery("shop", initialBrowseFilters([]))).toMatchObject({
			sort: "",
			order: "",
			cursor: ""
		});
	});

	it("copies a selected list so a later change does not reach the request", () => {
		const filters = setMulti(initialBrowseFilters(GROUPS), "tags", ["rpg"]);
		const request = browseQuery("shop", filters);
		(filters.values.tags as string[]).push("horror");
		expect(request.filters.tags).toEqual(["rpg"]);
	});
});

describe("filter changes", () => {
	const start = initialBrowseFilters(GROUPS);

	it("does not change the state object it is given", () => {
		const before = structuredClone(start);
		setBrowseQuery(start, "dark");
		setToggle(start, "installed", true);
		setSelect(start, "status", "new");
		setMulti(start, "tags", ["rpg"]);
		setSort(start, "title");
		setOrder(start, "asc");
		expect(start).toEqual(before);
	});

	it("sets a toggle, a select, a list, a sort and its direction", () => {
		let filters = setToggle(start, "installed", true);
		filters = setSelect(filters, "status", "new");
		filters = setMulti(filters, "tags", ["rpg"]);
		filters = setSort(filters, "title");
		filters = setOrder(filters, "asc");
		expect(filters.values).toEqual({
			installed: true,
			wishlisted: false,
			status: "new",
			tags: ["rpg"]
		});
		expect(filters).toMatchObject({ sort: "title", order: "asc" });
	});
});

describe("hasBrowseFilters", () => {
	it("is false for the defaults and the sort", () => {
		expect(hasBrowseFilters(initialBrowseFilters(GROUPS))).toBe(false);
	});

	it.each([
		[
			"search",
			(filters: ReturnType<typeof initialBrowseFilters>) => setBrowseQuery(filters, "dark")
		],
		[
			"toggle",
			(filters: ReturnType<typeof initialBrowseFilters>) => setToggle(filters, "installed", true)
		],
		[
			"select",
			(filters: ReturnType<typeof initialBrowseFilters>) => setSelect(filters, "status", "new")
		],
		[
			"list",
			(filters: ReturnType<typeof initialBrowseFilters>) => setMulti(filters, "tags", ["rpg"])
		]
	])("is true when the %s narrows the list", (_name, change) => {
		expect(hasBrowseFilters(change(initialBrowseFilters(GROUPS)))).toBe(true);
	});
});
