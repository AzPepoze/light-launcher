import type {
	DownloadItem,
	DownloadsListResponse,
	DownloadStatus,
	InstallPhase,
	InstallProgress
} from "./types";
import type {
	WireDownloadProgress,
	WireDownloadsStatus,
	WireDownloadView,
	WireInstallProgress
} from "./wire";

const STATUSES: readonly DownloadStatus[] = [
	"queued",
	"downloading",
	"paused",
	"done",
	"error",
	"canceled"
];

const PHASES: readonly InstallPhase[] = ["extracting", "unpacking", "finalizing", "done"];

function toStatus(state: string): DownloadStatus {
	return STATUSES.find((status) => status === state) ?? "queued";
}

function toPhase(phase: string): InstallPhase {
	return PHASES.find((known) => known === phase) ?? "extracting";
}

/** Seconds left, or null while the speed or the total is unknown. */
export function estimateEta(remaining: number, speed: number): number | null {
	if (remaining <= 0 || speed <= 0) return null;
	return remaining / speed;
}

function orNull(value: string | undefined): string | null {
	return value || null;
}

export function toDownloadItem(wire: WireDownloadView): DownloadItem {
	return {
		id: wire.id,
		gameId: wire.gameId || null,
		url: wire.url,
		filename: wire.filename,
		size: wire.total,
		bytesDone: wire.received,
		status: toStatus(wire.state),
		error: orNull(wire.error),
		speed: wire.speed,
		eta: estimateEta(wire.total - wire.received, wire.speed),
		connections: wire.connections,
		rateLimited: wire.rateLimited,
		createdAt: wire.createdAt,
		finishedAt: wire.finishedAt ?? null
	};
}

/** Updates one row from a progress event. Rows that are not in the list are left out. */
export function applyProgress(
	items: DownloadItem[],
	progress: WireDownloadProgress
): DownloadItem[] {
	return items.map((item) => {
		if (item.id !== progress.id) return item;
		const size = progress.total > 0 ? progress.total : item.size;
		return {
			...item,
			status: toStatus(progress.state),
			size,
			bytesDone: progress.received,
			speed: progress.speed,
			eta: estimateEta(size - progress.received, progress.speed),
			error: orNull(progress.error)
		};
	});
}

export function hasDownload(items: DownloadItem[], id: number): boolean {
	return items.some((item) => item.id === id);
}

/** The title comes from the library. When the game has none, the caller's fallback (the file name) is used. */
export function toInstallProgress(wire: WireInstallProgress, fallbackTitle: string): InstallProgress {
	return {
		downloadId: wire.downloadId,
		gameId: wire.gameId,
		title: wire.title || fallbackTitle,
		phase: toPhase(wire.phase),
		done: wire.done,
		total: wire.total,
		progress: wire.progress,
		speed: wire.speed,
		eta: wire.eta,
		engine: orNull(wire.engine),
		state: wire.state === "done" || wire.state === "error" ? wire.state : "active",
		error: orNull(wire.error)
	};
}

export function toDownloadsList(
	views: WireDownloadView[],
	status: WireDownloadsStatus
): DownloadsListResponse {
	const items = views.map(toDownloadItem);
	const fileNames = new Map(items.map((item) => [item.id, item.filename]));
	return {
		items,
		installs: (status.installs ?? []).map((install) =>
			toInstallProgress(install, fileNames.get(install.downloadId) ?? "")
		),
		speedLimitKBps: status.speedLimitKBps,
		sevenZip: status.sevenZip
	};
}
