import { describe, expect, it } from "vitest";
import {
	INTERVAL_OPTIONS,
	RECHECK_OPTIONS,
	clampHours,
	globalUpdateChecks,
	hoursFromOption,
	hoursLabel,
	optionsWithHours,
	sourceHelp,
	sourceModeFromOption,
	sourceView,
	updateCheckSources,
	withGlobalUpdateChecks,
	withSourceMode,
	withSourceRecheck
} from "./updateChecks";

describe("clampHours", () => {
	it("keeps 0 to 720 and rounds", () => {
		expect(clampHours(-4, 24)).toBe(0);
		expect(clampHours(900, 24)).toBe(720);
		expect(clampHours(6.4, 24)).toBe(6);
	});

	it("uses the fallback for values that are not finite", () => {
		expect(clampHours(Number.NaN, 24)).toBe(24);
	});
});

describe("hoursLabel and optionsWithHours", () => {
	it("reads whole days as days", () => {
		expect(hoursLabel(48)).toBe("2 days");
		expect(hoursLabel(24)).toBe("1 day");
		expect(hoursLabel(5)).toBe("5 hours");
		expect(hoursLabel(1)).toBe("1 hour");
	});

	it("keeps a list that already has the value", () => {
		expect(optionsWithHours(INTERVAL_OPTIONS, 24)).toBe(INTERVAL_OPTIONS);
	});

	it("adds a hand-edited value in order", () => {
		const values = optionsWithHours(RECHECK_OPTIONS, 5).map((option) => option.value);
		expect(values).toEqual(["0", "1", "5", "6", "24", "72"]);
	});
});

describe("hoursFromOption and sourceModeFromOption", () => {
	it("reads and clamps a dropdown value", () => {
		expect(hoursFromOption("168", 24)).toBe(168);
		expect(hoursFromOption("9999", 24)).toBe(720);
	});

	it("maps unknown modes to inherit", () => {
		expect(sourceModeFromOption("custom")).toBe("custom");
		expect(sourceModeFromOption("off")).toBe("off");
		expect(sourceModeFromOption("bogus")).toBe("inherit");
	});
});

describe("globalUpdateChecks", () => {
	it("fills the defaults when nothing is saved", () => {
		expect(globalUpdateChecks(undefined)).toEqual({
			enabled: true,
			onStartup: true,
			intervalHours: 24,
			recheckHours: 24
		});
	});

	it("keeps saved values and clamps out-of-range hours", () => {
		expect(
			globalUpdateChecks({ enabled: false, intervalHours: 1000, recheckHours: 0, onStartup: false })
		).toEqual({ enabled: false, onStartup: false, intervalHours: 720, recheckHours: 0 });
	});
});

describe("sourceView", () => {
	it("inherits the global skip window by default", () => {
		const settings = { recheckHours: 6 };
		expect(sourceView(settings, "f95")).toEqual({ mode: "inherit", recheckHours: 6 });
	});

	it("uses its own skip window in custom mode", () => {
		const settings = { recheckHours: 6, sources: { f95: { mode: "custom", recheckHours: 72 } } };
		expect(sourceView(settings, "f95")).toEqual({ mode: "custom", recheckHours: 72 });
	});

	it("falls back to the global window when custom has none", () => {
		const settings = { recheckHours: 6, sources: { f95: { mode: "custom" } } };
		expect(sourceView(settings, "f95")).toEqual({ mode: "custom", recheckHours: 6 });
	});

	it("keeps off as its mode", () => {
		expect(sourceView({ sources: { itch: { mode: "off" } } }, "itch").mode).toBe("off");
	});
});

describe("sourceHelp", () => {
	it("explains each mode in one line", () => {
		expect(sourceHelp("inherit")).toBe("Uses the global settings.");
		expect(sourceHelp("custom")).toBe("Uses its own skip window.");
		expect(sourceHelp("off")).toBe("No background checks. Check now still works.");
	});
});

describe("settings patches", () => {
	it("fills the four top-level values when one changes", () => {
		const next = withGlobalUpdateChecks(undefined, { enabled: false });
		expect(next).toEqual({
			enabled: false,
			onStartup: true,
			intervalHours: 24,
			recheckHours: 24
		});
	});

	it("keeps the other sources when one mode changes", () => {
		const settings = { sources: { itch: { mode: "off", recheckHours: 3 } } };
		const next = withSourceMode(settings, "f95", "custom");
		expect(next.sources).toEqual({ itch: { mode: "off", recheckHours: 3 }, f95: { mode: "custom" } });
	});

	it("sets a source skip window without changing its mode", () => {
		const settings = { sources: { f95: { mode: "custom" as const } } };
		const next = withSourceRecheck(settings, "f95", 1000);
		expect(next.sources?.f95).toEqual({ mode: "custom", recheckHours: 720 });
	});
});

describe("updateCheckSources", () => {
	it("keeps only enabled plugins whose sources can check for updates", () => {
		const sources = updateCheckSources([
			{
				enabled: true,
				sources: [
					{ sourceId: "sample-a", label: "Sample A", capabilities: ["browse", "updateCheck"] },
					{ sourceId: "sample-b", label: "Sample B", capabilities: ["browse"] }
				]
			},
			{
				enabled: false,
				sources: [{ sourceId: "sample-c", label: "Sample C", capabilities: ["updateCheck"] }]
			}
		]);
		expect(sources).toEqual([{ id: "sample-a", label: "Sample A" }]);
	});

	it("lists each source once, sorted by label", () => {
		const plugin = {
			enabled: true,
			sources: [
				{ sourceId: "sample-z", label: "Zeta", capabilities: ["updateCheck"] },
				{ sourceId: "sample-a", label: "Alpha", capabilities: ["updateCheck"] }
			]
		};
		expect(updateCheckSources([plugin, plugin]).map((source) => source.label)).toEqual(["Alpha", "Zeta"]);
	});

	it("handles plugins with no sources list", () => {
		expect(updateCheckSources([{ enabled: true, sources: null }])).toEqual([]);
	});
});
