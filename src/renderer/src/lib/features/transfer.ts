import { percent } from "./format";
import type { DownloadItem, DownloadStatus, InstallPhase, InstallProgress } from "./types";

export type TransferKind = "download" | "install";
export type TransferStage = "download" | "extract" | "install" | "done";
export type TransferState = "active" | "paused" | "queued" | "done" | "failed";

/** One row of a transfer list, built from a download and its install. */
export interface Transfer {
	id: string;
	kind: TransferKind;
	title: string;
	subtitle?: string;
	stage: TransferStage;
	/** 0 to 100, or null while the total is unknown. */
	percent: number | null;
	/** Bytes per second. */
	speed: number;
	/** Seconds left, or null while unknown. */
	eta: number | null;
	done: number;
	total: number;
	state: TransferState;
	error?: string;
}

const INSTALL_STAGES: Record<InstallPhase, TransferStage> = {
	extracting: "extract",
	unpacking: "extract",
	finalizing: "install",
	done: "done"
};

const DOWNLOAD_STATES: Record<DownloadStatus, TransferState> = {
	queued: "queued",
	downloading: "active",
	paused: "paused",
	done: "done",
	error: "failed",
	canceled: "failed"
};

const INSTALL_STATES: Record<InstallProgress["state"], TransferState> = {
	active: "active",
	done: "done",
	error: "failed"
};

/** The stage an install is in. A finished install is always done. */
export function stageOf(install: InstallProgress): TransferStage {
	if (install.state === "done") return "done";
	return INSTALL_STAGES[install.phase];
}

export function fromDownload(item: DownloadItem, install?: InstallProgress): Transfer {
	if (install) return fromDownloadWithInstall(item, install);
	return {
		id: downloadId(item.id),
		kind: "download",
		title: item.filename,
		stage: item.status === "done" ? "done" : "download",
		percent: sizePercent(item.bytesDone, item.size),
		speed: item.speed,
		eta: item.eta,
		done: item.bytesDone,
		total: item.size,
		state: DOWNLOAD_STATES[item.status],
		error: item.error ?? undefined
	};
}

export function fromInstall(install: InstallProgress): Transfer {
	return {
		id: installId(install.downloadId),
		kind: "install",
		title: install.title,
		stage: stageOf(install),
		percent: install.progress === null ? null : Math.round(install.progress * 100),
		speed: install.speed,
		eta: install.eta,
		done: install.done,
		total: install.total,
		state: INSTALL_STATES[install.state],
		error: install.error ?? undefined
	};
}

/** The install takes over the row of its download: stage, state and progress come from the install. */
function fromDownloadWithInstall(item: DownloadItem, install: InstallProgress): Transfer {
	const shownFilename = install.title === item.filename ? undefined : item.filename;
	return {
		...fromInstall(install),
		id: downloadId(item.id),
		kind: "download",
		subtitle: shownFilename,
		error: install.error ?? item.error ?? undefined
	};
}

function sizePercent(done: number, total: number): number | null {
	return total > 0 ? percent(done, total) : null;
}

function downloadId(id: number): string {
	return `download:${id}`;
}

function installId(linkedDownloadId: number): string {
	return `install:${linkedDownloadId}`;
}
