import { fromDownload, type Transfer, type TransferState } from "../../lib/features/transfer";
import type { DownloadItem, InstallProgress } from "../../lib/features/types";

export type RowGroup = "active" | "queued" | "failed" | "finished";

export const GROUP_ORDER: RowGroup[] = ["active", "queued", "failed", "finished"];

const GROUP_TITLES: Record<RowGroup, string> = {
	active: "Active",
	queued: "Queued",
	failed: "Failed",
	finished: "Finished"
};

const GROUP_OF_STATE: Record<TransferState, RowGroup> = {
	active: "active",
	paused: "active",
	queued: "queued",
	failed: "failed",
	done: "finished"
};

/** One list row: a download and the install that follows it, if any. */
export interface TransferRow {
	item: DownloadItem;
	install: InstallProgress | null;
	transfer: Transfer;
	/** The game title from the install, or null when it is only the file name. */
	gameTitle: string | null;
	group: RowGroup;
}

export type RowGroups = Record<RowGroup, TransferRow[]>;

export type ListEntry =
	| { kind: "header"; key: string; group: RowGroup; count: number }
	| { kind: "row"; key: string; row: TransferRow };

export function groupTitle(group: RowGroup): string {
	return GROUP_TITLES[group];
}

/** Matches each install to its download by id, so both share one row. */
export function buildRows(items: DownloadItem[], installs: InstallProgress[]): TransferRow[] {
	return items.map((item) => {
		const install = installs.find((candidate) => candidate.downloadId === item.id) ?? null;
		const transfer = fromDownload(item, install ?? undefined);
		return {
			item,
			install,
			transfer,
			gameTitle: gameTitleOf(item, install),
			group: GROUP_OF_STATE[transfer.state]
		};
	});
}

export function gameTitleOf(item: DownloadItem, install: InstallProgress | null): string | null {
	if (!install || install.title === item.filename) return null;
	return install.title;
}

export function groupRows(rows: TransferRow[]): RowGroups {
	const groups: RowGroups = { active: [], queued: [], failed: [], finished: [] };
	for (const row of rows) groups[row.group].push(row);
	return groups;
}

/** The flat list the page renders. The finished rows are left out while that group is closed. */
export function listEntries(groups: RowGroups, finishedOpen: boolean): ListEntry[] {
	const entries: ListEntry[] = [];
	for (const group of GROUP_ORDER) {
		const rows = groups[group];
		if (rows.length === 0) continue;
		entries.push({ kind: "header", key: `header:${group}`, group, count: rows.length });
		if (group === "finished" && !finishedOpen) continue;
		for (const row of rows) {
			entries.push({ kind: "row", key: `row:${row.item.id}`, row });
		}
	}
	return entries;
}

export function totalSpeed(items: DownloadItem[]): number {
	return items
		.filter((item) => item.status === "downloading")
		.reduce((sum, item) => sum + item.speed, 0);
}

/** The tone of a launcher message. Unknown levels show as info. */
export function toastTone(level: string): "info" | "warning" | "error" | "success" {
	if (level === "error" || level === "warning" || level === "success") return level;
	return "info";
}
