import type { HomeScan } from "@lib/homeScan";

export interface HomeStatus {
	text: string;
	/** Share done, from 0 to 1. Null when the total is unknown. */
	percent: number | null;
}

/** Share of a scan that is done, or null when there is no total to measure against. */
export function scanPercent(done: number, total: number): number | null {
	if (total <= 0) return null;
	return Math.min(1, Math.max(0, done / total));
}

export function scanStatusText(done: number, total: number, current = ""): string {
	if (current) return `Scanning ${current}... ${done} of ${total}`;
	if (total <= 0) return "Scanning folders...";
	return `Scanning folders... ${done} of ${total}`;
}

/** The header status line while the library loads or folders scan. Null when nothing runs. */
export function homeStatus(isLoading: boolean, scan: HomeScan | null): HomeStatus | null {
	if (isLoading) return { text: "Loading games...", percent: null };
	if (scan?.state !== "running") return null;
	return {
		text: scanStatusText(scan.done, scan.total, scan.current),
		percent: scanPercent(scan.done, scan.total)
	};
}
