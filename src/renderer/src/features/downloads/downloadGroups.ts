import { bytes, percent } from "../../lib/features/format";
import type { DownloadItem, DownloadStatus, InstallProgress } from "../../lib/features/types";

export interface DownloadGroups {
	active: DownloadItem[];
	finished: DownloadItem[];
	failed: DownloadItem[];
}

const ACTIVE_STATUSES: DownloadStatus[] = ["downloading", "queued", "paused"];
const FAILED_STATUSES: DownloadStatus[] = ["error", "canceled"];

export function groupDownloads(items: DownloadItem[]): DownloadGroups {
	return {
		active: items.filter((item) => ACTIVE_STATUSES.includes(item.status)),
		finished: items.filter((item) => item.status === "done"),
		failed: items.filter((item) => FAILED_STATUSES.includes(item.status))
	};
}

export function totalSpeed(items: DownloadItem[]): number {
	return items
		.filter((item) => item.status === "downloading")
		.reduce((sum, item) => sum + item.speed, 0);
}

export function statusLabel(status: DownloadStatus): string {
	const labels: Record<DownloadStatus, string> = {
		downloading: "Downloading",
		queued: "Queued",
		paused: "Paused",
		done: "Done",
		error: "Failed",
		canceled: "Canceled"
	};
	return labels[status];
}

const PHASE_LABELS: Record<InstallProgress["phase"], string> = {
	extracting: "Extracting",
	unpacking: "Unpacking",
	finalizing: "Finalizing",
	done: "Done"
};

/** One line such as "Extracting · 7-Zip · 59% · 3.7 GB / 6.3 GB". */
export function installLine(install: InstallProgress): string {
	const parts = [PHASE_LABELS[install.phase]];
	if (install.engine) parts.push(install.engine);
	if (install.progress !== null) parts.push(`${Math.round(install.progress * 100)}%`);
	if (install.total > 0) parts.push(`${bytes(install.done)} / ${bytes(install.total)}`);
	return parts.join(" · ");
}

export function progressPercent(item: DownloadItem): number {
	return percent(item.bytesDone, item.size);
}
