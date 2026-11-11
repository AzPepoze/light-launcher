import { describe, expect, it } from "vitest";
import type { LibrarySyncProgress } from "../../lib/features/types";
import { syncMessage } from "./librarySync";

const BASE: LibrarySyncProgress = {
	source: "f95",
	phase: "running",
	imported: 120,
	skipped: 3,
	failed: 0
};

describe("syncMessage", () => {
	it("shows the counts while the sync runs", () => {
		expect(syncMessage(BASE)).toBe("Syncing. 120 added, 3 unchanged.");
	});

	it("says the sync finished with the same counts", () => {
		expect(syncMessage({ ...BASE, phase: "done" })).toBe("Sync finished. 120 added, 3 unchanged.");
	});

	it("adds the failed count only when some games failed", () => {
		expect(syncMessage({ ...BASE, phase: "done", failed: 2 })).toBe(
			"Sync finished. 120 added, 3 unchanged, 2 failed."
		);
	});

	it("shows the error message the server sent", () => {
		expect(syncMessage({ ...BASE, phase: "error", message: "Sign in first." })).toBe(
			"Sign in first."
		);
	});

	it("gives a plain fallback when the error has no message", () => {
		expect(syncMessage({ ...BASE, phase: "error" })).toBe(
			"The sync failed. Check the plugin, then try again."
		);
	});
});
