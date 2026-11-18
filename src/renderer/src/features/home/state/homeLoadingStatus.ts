import type { HomeScan } from "@lib/homeScan";

export interface HomeStatus {
	text: string;
	/** Share done, from 0 to 1. Null when the total is unknown. */
	percent: number | null;
}

/** The progress shown in the header of the folder being scanned. */
export interface FolderScanStatus {
	/** Basename of the folder, matched against the folder group name. */
	folderName: string;
	text: string;
	/** Share done, from 0 to 1. Null when the total is unknown. */
	percent: number | null;
}

/** Share of a scan that is done, or null when there is no total to measure against. */
export function scanPercent(done: number, total: number): number | null {
	if (total <= 0) return null;
	return Math.min(1, Math.max(0, done / total));
}

/** The rescan lifecycle shown in the top status line. */
export type RescanPhase = "idle" | "running" | "done";

export interface RescanStatus {
	phase: RescanPhase;
	/** The last count reported by the scan, kept for the completion message. */
	found: number;
}

/**
 * The top status line: the first load, a manual rescan with its live found count, or nothing.
 * A rescan never opens a toast; it reuses the same strip as the initial loading state.
 */
export function homeStatus(
	isLoading: boolean,
	scan: HomeScan | null = null,
	rescan: RescanStatus = { phase: "idle", found: 0 }
): HomeStatus | null {
	if (isLoading) return { text: "Loading games...", percent: null };
	if (rescan.phase === "running") {
		if (scan?.state !== "running") {
			return { text: "Preparing rescan...", percent: null };
		}
		const where = scan.current ? `Rescanning ${scan.current}` : "Rescanning watched folders...";
		return {
			text: `${where} · ${scan.done} of ${scan.total} · ${scan.found} found`,
			percent: scanPercent(scan.done, scan.total)
		};
	}
	if (rescan.phase === "done") {
		return {
			text: `Rescan complete · ${rescan.found} game${rescan.found === 1 ? "" : "s"} found`,
			percent: 1
		};
	}
	return null;
}

/** The status of the folder the scan is on. Null when no folder is being scanned. */
export function runningFolderScan(scan: HomeScan | null): FolderScanStatus | null {
	if (scan?.state !== "running" || !scan.current) return null;
	return {
		folderName: scan.current,
		text: `Scanning... ${scan.done} of ${scan.total} · ${scan.found} found`,
		percent: scanPercent(scan.done, scan.total)
	};
}

/** The scan status when it belongs to the folder with this name. Null for every other folder. */
export function scanStatusForFolder(
	status: FolderScanStatus | null,
	folderName: string
): FolderScanStatus | null {
	return status?.folderName === folderName ? status : null;
}
