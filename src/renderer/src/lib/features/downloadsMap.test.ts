import { describe, expect, it } from "vitest";
import {
	applyProgress,
	estimateEta,
	hasDownload,
	toDownloadItem,
	toDownloadsList,
	toInstallProgress
} from "./downloadsMap";
import type { DownloadItem } from "./types";
import type { WireDownloadState, WireDownloadView } from "./wire";

function makeView(overrides: Partial<WireDownloadView> = {}): WireDownloadView {
	return {
		id: 1,
		gameId: 0,
		url: "https://example.test/file.zip",
		filename: "file.zip",
		dest: "/downloads/file.zip",
		state: "downloading",
		received: 250,
		total: 1000,
		speed: 250,
		connections: 4,
		install: false,
		rateLimited: false,
		createdAt: "2026-01-01T00:00:00Z",
		...overrides
	};
}

function makeItem(overrides: Partial<DownloadItem> = {}): DownloadItem {
	return {
		id: 1,
		gameId: null,
		url: "https://example.test/file.zip",
		filename: "file.zip",
		size: 1000,
		bytesDone: 250,
		status: "downloading",
		error: null,
		speed: 250,
		eta: 3,
		connections: 4,
		rateLimited: false,
		createdAt: "2026-01-01T00:00:00Z",
		finishedAt: null,
		...overrides
	};
}

describe("estimateEta", () => {
	it("divides the bytes left by the speed", () => {
		expect(estimateEta(750, 250)).toBe(3);
	});

	it("is null without a speed or with nothing left", () => {
		expect(estimateEta(750, 0)).toBeNull();
		expect(estimateEta(0, 250)).toBeNull();
		expect(estimateEta(-5, 250)).toBeNull();
	});
});

describe("toDownloadItem", () => {
	it("renames the Go fields and treats gameId 0 as no game", () => {
		expect(toDownloadItem(makeView())).toEqual({
			id: 1,
			gameId: null,
			url: "https://example.test/file.zip",
			filename: "file.zip",
			size: 1000,
			bytesDone: 250,
			status: "downloading",
			error: null,
			speed: 250,
			eta: 3,
			connections: 4,
			rateLimited: false,
			createdAt: "2026-01-01T00:00:00Z",
			finishedAt: null
		});
	});

	it("keeps the game id and the finish time", () => {
		const item = toDownloadItem(
			makeView({ gameId: 9, state: "done", finishedAt: "2026-01-02T00:00:00Z", speed: 0 })
		);
		expect(item.gameId).toBe(9);
		expect(item.status).toBe("done");
		expect(item.finishedAt).toBe("2026-01-02T00:00:00Z");
		expect(item.eta).toBeNull();
	});

	it("falls back to queued for an unknown state", () => {
		expect(toDownloadItem(makeView({ state: "weird" as WireDownloadState })).status).toBe("queued");
	});
});

describe("applyProgress", () => {
	it("merges a progress event into the existing row and keeps the rest", () => {
		const items = [makeItem(), makeItem({ id: 2, filename: "other.zip" })];
		const updated = applyProgress(items, {
			id: 1,
			state: "downloading",
			received: 500,
			total: 1000,
			speed: 500
		});
		expect(updated[0]).toMatchObject({
			id: 1,
			filename: "file.zip",
			bytesDone: 500,
			speed: 500,
			eta: 1
		});
		expect(updated[1]).toBe(items[1]);
	});

	it("records the error and the new state", () => {
		const [item] = applyProgress([makeItem()], {
			id: 1,
			state: "error",
			received: 250,
			total: 1000,
			speed: 0,
			error: "connection reset"
		});
		expect(item.status).toBe("error");
		expect(item.error).toBe("connection reset");
		expect(item.eta).toBeNull();
	});

	it("keeps the old size when the event has no total", () => {
		const [item] = applyProgress([makeItem()], {
			id: 1,
			state: "downloading",
			received: 300,
			total: 0,
			speed: 100
		});
		expect(item.size).toBe(1000);
	});
});

describe("hasDownload", () => {
	it("finds a row by id", () => {
		expect(hasDownload([makeItem({ id: 3 })], 3)).toBe(true);
		expect(hasDownload([makeItem({ id: 3 })], 4)).toBe(false);
	});
});

describe("toInstallProgress", () => {
	it("maps an active install with the title and eta the API sent", () => {
		const install = toInstallProgress(
			{
				downloadId: 1,
				gameId: 7,
				title: "Dust Runner",
				state: "active",
				phase: "unpacking",
				done: 250,
				total: 1000,
				progress: 0.25,
				speed: 250,
				eta: 3,
				engine: "7-Zip"
			},
			"file.zip"
		);
		expect(install).toEqual({
			downloadId: 1,
			gameId: 7,
			title: "Dust Runner",
			phase: "unpacking",
			done: 250,
			total: 1000,
			progress: 0.25,
			speed: 250,
			eta: 3,
			engine: "7-Zip",
			state: "active",
			error: null
		});
	});

	it("falls back to the file name when the API has no title", () => {
		const install = toInstallProgress(
			{
				downloadId: 1,
				gameId: 0,
				state: "active",
				phase: "extracting",
				done: 0,
				total: 0,
				progress: null,
				speed: 0,
				eta: null
			},
			"file.zip"
		);
		expect(install.title).toBe("file.zip");
		expect(install.eta).toBeNull();
	});

	it("keeps the done phase and the error text", () => {
		const done = toInstallProgress(
			{
				downloadId: 1,
				gameId: 7,
				state: "done",
				phase: "done",
				done: 1,
				total: 1,
				progress: 1,
				speed: 0,
				eta: null
			},
			"x"
		);
		expect(done.phase).toBe("done");
		expect(done.state).toBe("done");

		const failed = toInstallProgress(
			{
				downloadId: 1,
				gameId: 7,
				state: "error",
				phase: "extracting",
				done: 0,
				total: 0,
				progress: null,
				speed: 0,
				eta: null,
				error: "bad archive"
			},
			"x"
		);
		expect(failed.state).toBe("error");
		expect(failed.error).toBe("bad archive");
		expect(failed.progress).toBeNull();
	});
});

describe("toDownloadsList", () => {
	it("returns the rows, running installs, speed limit and 7-Zip state from the status", () => {
		const list = toDownloadsList([makeView({ id: 5, filename: "game.7z" })], {
			installs: [
				{
					downloadId: 5,
					gameId: 2,
					state: "active",
					phase: "extracting",
					done: 10,
					total: 100,
					progress: 0.1,
					speed: 10,
					eta: 9
				}
			],
			sevenZip: false,
			speedLimitKBps: 512
		});
		expect(list.items).toHaveLength(1);
		expect(list.installs).toHaveLength(1);
		expect(list.installs[0].title).toBe("game.7z");
		expect(list.installs[0].eta).toBe(9);
		expect(list.speedLimitKBps).toBe(512);
		expect(list.sevenZip).toBe(false);
	});

	it("treats a missing install list as empty", () => {
		const list = toDownloadsList([], { installs: null, sevenZip: true, speedLimitKBps: 0 });
		expect(list.installs).toEqual([]);
		expect(list.sevenZip).toBe(true);
	});
});
