import { describe, expect, it } from "vitest";
import { fromDownload, fromInstall, stageOf } from "./transfer";
import type { DownloadItem, InstallProgress } from "./types";

function download(overrides: Partial<DownloadItem> = {}): DownloadItem {
	return {
		id: 1,
		gameId: null,
		url: "https://example.com/example.zip",
		filename: "example.zip",
		size: 100,
		bytesDone: 50,
		status: "downloading",
		error: null,
		speed: 2048,
		eta: 12,
		connections: 4,
		rateLimited: false,
		createdAt: "2026-01-01T00:00:00Z",
		finishedAt: null,
		...overrides
	};
}

function install(overrides: Partial<InstallProgress> = {}): InstallProgress {
	return {
		downloadId: 1,
		gameId: 7,
		title: "Sample Game",
		phase: "extracting",
		done: 25,
		total: 100,
		progress: 0.25,
		speed: 1024,
		eta: 30,
		engine: null,
		state: "active",
		error: null,
		...overrides
	};
}

describe("fromDownload", () => {
	it("maps a running download to the download stage", () => {
		const transfer = fromDownload(download());
		expect(transfer).toMatchObject({
			id: "download:1",
			kind: "download",
			title: "example.zip",
			stage: "download",
			percent: 50,
			speed: 2048,
			eta: 12,
			done: 50,
			total: 100,
			state: "active"
		});
		expect(transfer.subtitle).toBeUndefined();
	});

	it("keeps the error and the failed state", () => {
		const transfer = fromDownload(download({ status: "error", error: "Connection reset" }));
		expect(transfer.state).toBe("failed");
		expect(transfer.error).toBe("Connection reset");
	});

	it("leaves the percent empty while the size is unknown", () => {
		expect(fromDownload(download({ size: 0 })).percent).toBeNull();
	});

	it("settles a finished download with no install as done", () => {
		const transfer = fromDownload(download({ status: "done", bytesDone: 100 }));
		expect(transfer.stage).toBe("done");
		expect(transfer.state).toBe("done");
	});

	it("takes stage and progress from its install", () => {
		const transfer = fromDownload(download({ status: "done", bytesDone: 100 }), install());
		expect(transfer).toMatchObject({
			id: "download:1",
			kind: "download",
			title: "Sample Game",
			subtitle: "example.zip",
			stage: "extract",
			percent: 25,
			done: 25,
			total: 100,
			state: "active"
		});
	});

	it("leaves out the subtitle when the install title is the file name", () => {
		const transfer = fromDownload(download({ status: "done" }), install({ title: "example.zip" }));
		expect(transfer.subtitle).toBeUndefined();
	});
});

describe("fromInstall", () => {
	it("builds an install row", () => {
		expect(fromInstall(install())).toMatchObject({
			id: "install:1",
			kind: "install",
			title: "Sample Game",
			stage: "extract",
			percent: 25,
			state: "active"
		});
	});

	it("uses no percent while the progress is unknown", () => {
		expect(fromInstall(install({ progress: null })).percent).toBeNull();
	});

	it("marks a failed install as failed with its error", () => {
		const transfer = fromInstall(install({ state: "error", error: "Archive is damaged" }));
		expect(transfer.state).toBe("failed");
		expect(transfer.error).toBe("Archive is damaged");
	});
});

describe("stageOf", () => {
	it("maps install phases to stages", () => {
		expect(stageOf(install({ phase: "extracting" }))).toBe("extract");
		expect(stageOf(install({ phase: "unpacking" }))).toBe("extract");
		expect(stageOf(install({ phase: "finalizing" }))).toBe("install");
		expect(stageOf(install({ phase: "done" }))).toBe("done");
	});

	it("reports done once the install state is done", () => {
		expect(stageOf(install({ phase: "finalizing", state: "done" }))).toBe("done");
	});
});
