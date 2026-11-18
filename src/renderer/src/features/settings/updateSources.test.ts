import { describe, expect, it } from "vitest";
import type { LibrarySource } from "@lib/features/types";
import { sourceIcon } from "../library/sourceIcon";
import { updateCheckSources } from "./updateSources";

function makeSource(id: string, overrides: Partial<LibrarySource> = {}): LibrarySource {
	return {
		source: id,
		label: id,
		kind: "plugin",
		plugin: id,
		capabilities: ["updateCheck"],
		sync: false,
		updateCheck: true,
		count: 0,
		tile: "poster",
		cover: "3:4",
		actions: [],
		card: null,
		sidebar: [],
		engines: [],
		tags: [],
		...overrides
	};
}

describe("updateCheckSources", () => {
	it("keeps only plugin sources that can check for updates", () => {
		const rows = updateCheckSources([
			makeSource("f95", { label: "F95 Zone" }),
			makeSource("itch", { label: "itch.io", updateCheck: false, capabilities: ["browse"] })
		]);
		expect(rows.map((row) => row.id)).toEqual(["f95"]);
	});

	it("leaves local and manual sources out even when they claim update checks", () => {
		const rows = updateCheckSources([
			makeSource("manual", {
				label: "Local games",
				kind: "local",
				plugin: null,
				capabilities: ["launch", "scan"],
				updateCheck: true
			}),
			makeSource("manual", { label: "Manual import", kind: "plugin", updateCheck: true }),
			makeSource("f95", { label: "F95 Zone" })
		]);
		expect(rows.map((row) => row.id)).toEqual(["f95"]);
	});

	it("shows the overview label with the shared source icon", () => {
		expect(updateCheckSources([makeSource("f95", { label: "F95 Zone" })])).toEqual([
			{ id: "f95", label: "F95 Zone", icon: sourceIcon("f95") }
		]);
		expect(sourceIcon("f95")).toBe("store");
	});

	it("lists each source once and sorts by label", () => {
		const rows = updateCheckSources([
			makeSource("zeta", { label: "Zeta" }),
			makeSource("alpha", { label: "Alpha" }),
			makeSource("zeta", { label: "Zeta" })
		]);
		expect(rows.map((row) => row.label)).toEqual(["Alpha", "Zeta"]);
	});

	it("falls back to the source id when the overview has no label", () => {
		expect(updateCheckSources([makeSource("f95", { label: "  " })])[0]?.label).toBe("f95");
	});

	it("returns no rows for an empty overview", () => {
		expect(updateCheckSources([])).toEqual([]);
	});
});
