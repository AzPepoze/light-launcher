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

/** The top status line while the library loads. Null once it is loaded. */
export function homeStatus(isLoading: boolean): HomeStatus | null {
	if (!isLoading) return null;
	return { text: "Loading games...", percent: null };
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
