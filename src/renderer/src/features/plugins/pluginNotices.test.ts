import { describe, expect, it } from "vitest";
import type { PluginUpdateStatus } from "../../lib/features/types";
import {
	checkFailureHint,
	checkOutcome,
	isCheckFailure,
	isNothingToCheck,
	skippedSummary
} from "./pluginNotices";

function status(overrides: Partial<PluginUpdateStatus> = {}): PluginUpdateStatus {
	return {
		id: "plugin-a",
		installedVersion: "1.0.0",
		available: false,
		version: null,
		assetUrl: null,
		sha256Url: null,
		sha256: null,
		notes: null,
		publishedAt: null,
		error: null,
		code: null,
		...overrides
	};
}

const NOT_FOUND = "check github:owner/plugin-a: GitHub returned HTTP 404";
const SERVER_DOWN = "check github:owner/plugin-b: GitHub returned HTTP 503";

describe("isNothingToCheck", () => {
	it("treats a 404 and a repository without releases as nothing to check", () => {
		expect(isNothingToCheck(NOT_FOUND)).toBe(true);
		expect(isNothingToCheck("check github:owner/plugin-a: the repository has no published release")).toBe(true);
	});

	it("does not treat server errors as nothing to check", () => {
		expect(isNothingToCheck(SERVER_DOWN)).toBe(false);
	});

	it("uses the stable code when the error text changes", () => {
		expect(isNothingToCheck(status({ error: "anything", code: "not_found" }))).toBe(true);
	});
});

describe("isCheckFailure", () => {
	it("is false for a plugin with no error", () => {
		expect(isCheckFailure(status())).toBe(false);
	});

	it("is false for a 404 so the page stays quiet", () => {
		expect(isCheckFailure(status({ error: NOT_FOUND }))).toBe(false);
	});

	it("is true for a server error", () => {
		expect(isCheckFailure(status({ error: SERVER_DOWN }))).toBe(true);
	});
});

describe("checkFailureHint", () => {
	it("explains a rate limit without the raw error", () => {
		const hint = checkFailureHint("GitHub rate limit reached, try again at 2026-01-01T00:00:00Z");
		expect(hint).toBe("GitHub limits update checks for a while. Try again later.");
	});

	it("gives a general hint for other failures", () => {
		expect(checkFailureHint(SERVER_DOWN)).toBe("The update source did not answer. Try again later.");
	});

	it("uses the rate_limited code even without the old wording", () => {
		expect(checkFailureHint(status({ error: "slow down", code: "rate_limited" }))).toBe(
			"GitHub limits update checks for a while. Try again later."
		);
	});
});

describe("checkOutcome", () => {
	it("reports up to date when nothing is available or failed", () => {
		expect(checkOutcome([status(), status({ id: "plugin-c" })])).toEqual({
			tone: "success",
			message: "All plugins are up to date"
		});
	});

	it("ignores a 404 when every other plugin is up to date", () => {
		expect(checkOutcome([status({ error: NOT_FOUND })])).toEqual({
			tone: "success",
			message: "All plugins are up to date"
		});
	});

	it("counts available updates in one message", () => {
		expect(checkOutcome([status({ available: true }), status({ id: "plugin-c", available: true })])).toEqual({
			tone: "info",
			message: "2 updates available"
		});
	});

	it("uses the singular for one update", () => {
		expect(checkOutcome([status({ available: true })]).message).toBe("1 update available");
	});

	it("counts real failures once and marks the outcome as an error", () => {
		expect(checkOutcome([status({ error: SERVER_DOWN }), status({ id: "plugin-c", error: SERVER_DOWN })])).toEqual({
			tone: "error",
			message: "2 plugins could not be checked"
		});
	});

	it("joins updates and failures in one message", () => {
		expect(
			checkOutcome([
				status({ available: true }),
				status({ id: "plugin-b", error: SERVER_DOWN }),
				status({ id: "plugin-c", error: NOT_FOUND })
			])
		).toEqual({
			tone: "error",
			message: "1 update available, 1 plugin could not be checked"
		});
	});
});

describe("skippedSummary", () => {
	it("lists every skipped item in one message", () => {
		expect(
			skippedSummary([
				{ name: "notes.txt", reason: "not-a-plugin" },
				{ name: "bundle", reason: "nested-plugins" }
			])
		).toBe(
			"These items were skipped: notes.txt: Not a plugin. Needs a .llplugin or .zip file, or a folder with plugin.json.; bundle: Holds several plugins. Drop each plugin folder on its own."
		);
	});
});
