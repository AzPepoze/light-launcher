import { describe, expect, it } from "vitest";
import { fromDownload } from "../../lib/features/transfer";
import type { DownloadItem, DownloadStatus } from "../../lib/features/types";
import { buildRows } from "./downloadGroups";
import {
	fileIcon,
	progressTone,
	progressValue,
	rowActions,
	rowBadge,
	rowMeta
} from "./rowView";

function makeItem(overrides: Partial<DownloadItem> = {}): DownloadItem {
	return {
		id: 1,
		gameId: null,
		url: "https://example.test/example.zip",
		filename: "example.zip",
		size: 200,
		bytesDone: 50,
		status: "downloading",
		error: null,
		speed: 1024,
		eta: 10,
		connections: 4,
		rateLimited: false,
		createdAt: "2026-01-01T00:00:00Z",
		finishedAt: null,
		...overrides
	};
}

function rowFor(status: DownloadStatus) {
	return buildRows([makeItem({ status })], [])[0];
}

function actionNames(status: DownloadStatus): string[] {
	return rowActions(rowFor(status)).map((action) => action.action);
}

describe("rowActions", () => {
	it("pauses a running download and cancels it", () => {
		expect(actionNames("downloading")).toEqual(["pause", "cancel"]);
	});

	it("resumes a paused or queued download", () => {
		expect(actionNames("paused")).toEqual(["resume", "cancel"]);
		expect(actionNames("queued")).toEqual(["resume", "cancel"]);
	});

	it("retries and removes a failed or canceled download", () => {
		expect(actionNames("error")).toEqual(["retry", "remove"]);
		expect(actionNames("canceled")).toEqual(["retry", "remove"]);
	});

	it("only removes a finished download", () => {
		expect(actionNames("done")).toEqual(["remove"]);
	});

	it("offers no action while the install runs", () => {
		const [row] = buildRows(
			[makeItem({ status: "done" })],
			[
				{
					downloadId: 1,
					gameId: 7,
					title: "Sample Game",
					phase: "extracting",
					done: 10,
					total: 100,
					progress: 0.1,
					speed: 0,
					eta: null,
					engine: null,
					state: "active",
					error: null
				}
			]
		);
		expect(rowActions(row)).toEqual([]);
	});

	it("gives every action a label and an icon", () => {
		for (const status of ["downloading", "paused", "error", "done"] as const) {
			for (const action of rowActions(rowFor(status))) {
				expect(action.label).not.toBe("");
				expect(action.icon).not.toBe("");
			}
		}
	});
});

describe("rowBadge", () => {
	it("names paused, queued and failed rows only", () => {
		expect(rowBadge(fromDownload(makeItem({ status: "paused" })))).toEqual({
			tone: "warning",
			label: "Paused"
		});
		expect(rowBadge(fromDownload(makeItem({ status: "queued" })))?.label).toBe("Queued");
		expect(rowBadge(fromDownload(makeItem({ status: "canceled" })))?.label).toBe("Failed");
		expect(rowBadge(fromDownload(makeItem({ status: "downloading" })))).toBeNull();
		expect(rowBadge(fromDownload(makeItem({ status: "done" })))).toBeNull();
	});
});

describe("rowMeta", () => {
	it("shows speed and time left only while downloading", () => {
		expect(rowMeta(fromDownload(makeItem()))).toEqual({
			speed: 1024,
			eta: 10,
			done: 50,
			total: 200
		});
		expect(rowMeta(fromDownload(makeItem({ status: "paused" })))).toMatchObject({
			speed: undefined,
			eta: undefined
		});
	});

	it("leaves the total out when it is unknown", () => {
		expect(rowMeta(fromDownload(makeItem({ size: 0 }))).total).toBeUndefined();
	});
});

describe("progressValue", () => {
	it("is full for a done row and sliding while an unknown total is active", () => {
		expect(progressValue(fromDownload(makeItem({ status: "done" })))).toBe(1);
		expect(progressValue(fromDownload(makeItem({ size: 0 })))).toBeNull();
	});

	it("is zero for an unknown total that is not active", () => {
		expect(progressValue(fromDownload(makeItem({ size: 0, status: "paused" })))).toBe(0);
	});

	it("turns the percent into a 0 to 1 value", () => {
		expect(progressValue(fromDownload(makeItem({ size: 200, bytesDone: 50 })))).toBe(0.25);
	});
});

describe("progressTone", () => {
	it("follows the state", () => {
		expect(progressTone("active")).toBe("accent");
		expect(progressTone("done")).toBe("success");
		expect(progressTone("failed")).toBe("danger");
		expect(progressTone("paused")).toBe("muted");
	});
});

describe("fileIcon", () => {
	it("picks an icon by file type", () => {
		expect(fileIcon("example.ZIP")).toBe("folder_zip");
		expect(fileIcon("setup.exe")).toBe("install_desktop");
		expect(fileIcon("notes.txt")).toBe("draft");
	});
});
