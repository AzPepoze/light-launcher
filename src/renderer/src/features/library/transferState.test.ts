import { describe, expect, it } from "vitest";
import type { DownloadItem, InstallProgress } from "../../lib/features/types";
import { transferKinds } from "./transferState";

function download(id: number, gameId: number | null, status: DownloadItem["status"]): DownloadItem {
	return {
		id,
		gameId,
		url: "https://example.test/file",
		filename: "file.zip",
		size: 0,
		bytesDone: 0,
		status,
		error: null,
		speed: 0,
		eta: null,
		connections: 1,
		rateLimited: false,
		createdAt: "2026-01-01T00:00:00Z",
		finishedAt: null
	};
}

function install(downloadId: number, gameId: number, state: InstallProgress["state"]): InstallProgress {
	return {
		downloadId,
		gameId,
		title: "Example Game",
		phase: "extracting",
		done: 0,
		total: 0,
		progress: null,
		speed: 0,
		eta: null,
		engine: null,
		state,
		error: null
	};
}

describe("transferKinds", () => {
	it("marks a game whose download is running as downloading", () => {
		const kinds = transferKinds([download(1, 10, "downloading")], []);
		expect(kinds.get(10)).toBe("downloading");
	});

	it("ignores queued, finished and downloads with no game", () => {
		const kinds = transferKinds(
			[download(1, 10, "queued"), download(2, 11, "done"), download(3, null, "downloading")],
			[]
		);
		expect(kinds.size).toBe(0);
	});

	it("marks a game whose install is running as installing, over its download", () => {
		const kinds = transferKinds([download(1, 10, "downloading")], [install(1, 10, "active")]);
		expect(kinds.get(10)).toBe("installing");
	});

	it("ignores installs that are done or failed", () => {
		const kinds = transferKinds([], [install(1, 10, "done"), install(2, 11, "error")]);
		expect(kinds.size).toBe(0);
	});
});
