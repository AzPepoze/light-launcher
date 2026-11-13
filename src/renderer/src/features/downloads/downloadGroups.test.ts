import { describe, expect, it } from "vitest";
import type { DownloadItem, InstallProgress } from "../../lib/features/types";
import {
	buildRows,
	gameTitleOf,
	groupRows,
	groupTitle,
	listEntries,
	toastTone,
	totalSpeed
} from "./downloadGroups";

function makeItem(overrides: Partial<DownloadItem>): DownloadItem {
	return {
		id: 1,
		gameId: null,
		url: "https://example.test/example.zip",
		filename: "example.zip",
		size: 100,
		bytesDone: 0,
		status: "queued",
		error: null,
		speed: 0,
		eta: null,
		connections: 1,
		rateLimited: false,
		createdAt: "2026-01-01T00:00:00Z",
		finishedAt: null,
		...overrides
	};
}

function makeInstall(overrides: Partial<InstallProgress>): InstallProgress {
	return {
		downloadId: 1,
		gameId: 7,
		title: "Sample Game",
		phase: "extracting",
		done: 0,
		total: 100,
		progress: 0,
		speed: 0,
		eta: null,
		engine: null,
		state: "active",
		error: null,
		...overrides
	};
}

const items = [
	makeItem({ id: 1, status: "downloading", speed: 300 }),
	makeItem({ id: 2, status: "paused" }),
	makeItem({ id: 3, status: "queued" }),
	makeItem({ id: 4, status: "done" }),
	makeItem({ id: 5, status: "error" }),
	makeItem({ id: 6, status: "canceled" })
];

describe("buildRows and groupRows", () => {
	it("puts paused rows in active and canceled rows in failed", () => {
		const groups = groupRows(buildRows(items, []));
		expect(groups.active.map((row) => row.item.id)).toEqual([1, 2]);
		expect(groups.queued.map((row) => row.item.id)).toEqual([3]);
		expect(groups.failed.map((row) => row.item.id)).toEqual([5, 6]);
		expect(groups.finished.map((row) => row.item.id)).toEqual([4]);
	});

	it("keeps a download in active while its install runs", () => {
		const [row] = buildRows(
			[makeItem({ id: 1, status: "done", size: 200, bytesDone: 200 })],
			[makeInstall({ downloadId: 1, phase: "extracting", progress: 0.5 })]
		);
		expect(row.group).toBe("active");
		expect(row.transfer).toMatchObject({ stage: "extract", percent: 50 });
	});

	it("moves the row to finished once its install is done", () => {
		const [row] = buildRows(
			[makeItem({ id: 1, status: "done", size: 200, bytesDone: 200 })],
			[makeInstall({ downloadId: 1, phase: "done", state: "done", progress: null })]
		);
		expect(row.group).toBe("finished");
	});

	it("matches an install to its download by id", () => {
		const rows = buildRows(
			[makeItem({ id: 1 }), makeItem({ id: 2, filename: "other.zip" })],
			[makeInstall({ downloadId: 2, title: "Sample Game" })]
		);
		expect(rows[0].install).toBeNull();
		expect(rows[1].install?.title).toBe("Sample Game");
	});

	it("marks a failed install as failed even when the download is done", () => {
		const [row] = buildRows(
			[makeItem({ id: 1, status: "done" })],
			[makeInstall({ downloadId: 1, state: "error", error: "Extraction failed" })]
		);
		expect(row.group).toBe("failed");
		expect(row.transfer.error).toBe("Extraction failed");
	});
});

describe("gameTitleOf", () => {
	it("returns the install title only when it differs from the file name", () => {
		const item = makeItem({ filename: "example.zip" });
		expect(gameTitleOf(item, makeInstall({ title: "Sample Game" }))).toBe("Sample Game");
		expect(gameTitleOf(item, makeInstall({ title: "example.zip" }))).toBeNull();
		expect(gameTitleOf(item, null)).toBeNull();
	});
});

describe("listEntries", () => {
	it("orders groups Active, Queued, Failed, Finished and hides empty groups", () => {
		const groups = groupRows(buildRows(items, []));
		const entries = listEntries(groups, true);
		const headers = entries.flatMap((entry) => (entry.kind === "header" ? [entry.group] : []));
		expect(headers).toEqual(["active", "queued", "failed", "finished"]);
	});

	it("leaves finished rows out while the group is closed, keeping its header with the count", () => {
		const groups = groupRows(buildRows(items, []));
		const closed = listEntries(groups, false);
		expect(closed).toContainEqual(
			expect.objectContaining({ kind: "header", group: "finished", count: 1 })
		);
		expect(closed.some((entry) => entry.kind === "row" && entry.row.group === "finished")).toBe(
			false
		);
		const open = listEntries(groups, true);
		expect(open.some((entry) => entry.kind === "row" && entry.row.group === "finished")).toBe(true);
	});

	it("gives every entry a stable key", () => {
		const groups = groupRows(buildRows(items, []));
		const keys = listEntries(groups, true).map((entry) => entry.key);
		expect(keys).toContain("row:1");
		expect(keys).toContain("header:finished");
		expect(new Set(keys).size).toBe(keys.length);
	});
});

describe("totalSpeed", () => {
	it("adds up only downloads that are running", () => {
		expect(totalSpeed([...items, makeItem({ id: 7, status: "paused", speed: 900 })])).toBe(300);
	});
});

describe("groupTitle", () => {
	it("gives sentence case titles", () => {
		expect(groupTitle("active")).toBe("Active");
		expect(groupTitle("finished")).toBe("Finished");
	});
});

describe("toastTone", () => {
	it("keeps known levels and treats others as info", () => {
		expect(toastTone("error")).toBe("error");
		expect(toastTone("warning")).toBe("warning");
		expect(toastTone("something")).toBe("info");
	});
});
