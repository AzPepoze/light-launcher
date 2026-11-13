import type { Transfer, TransferState } from "../../lib/features/transfer";
import type { DownloadAction, DownloadStatus } from "../../lib/features/types";
import type { TransferRow } from "./downloadGroups";

export interface RowAction {
	action: DownloadAction;
	icon: string;
	label: string;
}

export interface RowBadge {
	tone: "neutral" | "warning" | "danger";
	label: string;
}

export interface RowMeta {
	speed?: number;
	eta?: number | null;
	done: number;
	total?: number;
}

const PAUSE: RowAction = { action: "pause", icon: "pause", label: "Pause" };
const RESUME: RowAction = { action: "resume", icon: "play_arrow", label: "Resume" };
const CANCEL: RowAction = { action: "cancel", icon: "close", label: "Cancel" };
const RETRY: RowAction = { action: "retry", icon: "refresh", label: "Retry" };
const REMOVE: RowAction = { action: "remove", icon: "delete", label: "Remove from list" };

const LIVE_STATUSES: DownloadStatus[] = ["downloading", "paused", "queued"];
const ARCHIVE_EXTENSIONS = [".zip", ".7z", ".rar", ".tar", ".gz", ".xz"];
const INSTALLER_EXTENSIONS = [".exe", ".msi", ".apk"];

/** The buttons a row offers. A row that is installing offers none. */
export function rowActions(row: TransferRow): RowAction[] {
	const status = row.item.status;
	const actions: RowAction[] = [];
	if (status === "downloading") actions.push(PAUSE);
	if (status === "paused" || status === "queued") actions.push(RESUME);
	if (status === "error" || status === "canceled") actions.push(RETRY);
	if (LIVE_STATUSES.includes(status)) actions.push(CANCEL);
	if (row.group === "finished" || row.group === "failed") actions.push(REMOVE);
	return actions;
}

const BADGES: Partial<Record<TransferState, RowBadge>> = {
	paused: { tone: "warning", label: "Paused" },
	queued: { tone: "neutral", label: "Queued" },
	failed: { tone: "danger", label: "Failed" }
};

export function rowBadge(transfer: Transfer): RowBadge | null {
	return BADGES[transfer.state] ?? null;
}

/** Speed and time left only show while a file is moving. */
export function rowMeta(transfer: Transfer): RowMeta {
	const moving = transfer.state === "active" && transfer.stage === "download";
	return {
		speed: moving ? transfer.speed : undefined,
		eta: moving ? transfer.eta : undefined,
		done: transfer.done,
		total: transfer.total > 0 ? transfer.total : undefined
	};
}

/** 0 to 1 for the bar. An unknown total is a sliding bar only while the row is active. */
export function progressValue(transfer: Transfer): number | null {
	if (transfer.state === "done") return 1;
	if (transfer.percent === null) return transfer.state === "active" ? null : 0;
	return transfer.percent / 100;
}

export function progressTone(state: TransferState): "accent" | "success" | "danger" | "muted" {
	if (state === "done") return "success";
	if (state === "failed") return "danger";
	if (state === "paused" || state === "queued") return "muted";
	return "accent";
}

export function fileIcon(filename: string): string {
	const name = filename.toLowerCase();
	if (ARCHIVE_EXTENSIONS.some((extension) => name.endsWith(extension))) return "folder_zip";
	if (INSTALLER_EXTENSIONS.some((extension) => name.endsWith(extension))) return "install_desktop";
	return "draft";
}
