import { describe, expect, it } from "vitest";
import type { DownloadItem, InstallProgress } from "../../lib/features/types";
import {
	groupDownloads,
	installLine,
	progressPercent,
	statusLabel,
	totalSpeed
} from "./downloadGroups";

function makeItem(overrides: Partial<DownloadItem>): DownloadItem {
	return {
		id: 1,
		gameId: null,
		url: "https://example.test/file.zip",
		filename: "file.zip",
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

const items = [
	makeItem({ id: 1, status: "downloading", speed: 300 }),
	makeItem({ id: 2, status: "paused" }),
	makeItem({ id: 3, status: "queued" }),
	makeItem({ id: 4, status: "done" }),
	makeItem({ id: 5, status: "error" }),
	makeItem({ id: 6, status: "canceled" })
];

describe("groupDownloads", () => {
	it("splits downloads into active, finished and failed", () => {
		const groups = groupDownloads(items);
		expect(groups.active.map((item) => item.id)).toEqual([1, 2, 3]);
		expect(groups.finished.map((item) => item.id)).toEqual([4]);
		expect(groups.failed.map((item) => item.id)).toEqual([5, 6]);
	});
});

describe("totalSpeed", () => {
	it("adds up only downloads that are running", () => {
		expect(totalSpeed([...items, makeItem({ id: 7, status: "paused", speed: 900 })])).toBe(300);
	});
});

describe("statusLabel", () => {
	it("names every status", () => {
		expect(statusLabel("error")).toBe("Failed");
		expect(statusLabel("canceled")).toBe("Canceled");
	});
});

describe("progressPercent", () => {
	it("returns the share of bytes done", () => {
		expect(progressPercent(makeItem({ size: 200, bytesDone: 50 }))).toBe(25);
		expect(progressPercent(makeItem({ size: 0, bytesDone: 50 }))).toBe(0);
	});
});

describe("installLine", () => {
	const base: InstallProgress = {
		downloadId: 1,
		gameId: 1,
		title: "Game",
		phase: "extracting",
		done: 3 * 1024 ** 3,
		total: 6 * 1024 ** 3,
		progress: 0.5,
		speed: 0,
		eta: null,
		engine: "7-Zip",
		state: "active",
		error: null
	};

	it("joins phase, tool, percent and byte counts", () => {
		expect(installLine(base)).toBe("Extracting · 7-Zip · 50% · 3.0 GB / 6.0 GB");
	});

	it("skips parts that are unknown", () => {
		expect(
			installLine({ ...base, engine: null, progress: null, total: 0, phase: "finalizing" })
		).toBe("Finalizing");
	});
});
