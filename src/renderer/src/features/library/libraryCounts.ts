import type { LibraryStats } from "../../lib/features/types";

/** The header line and the Updates badge both read the games behind their installed version. */
export function updateCountOf(stats: LibraryStats): number {
	return stats.updates;
}

/** The line under the Library title. The updates part shows only when there are some. */
export function librarySubtitle(stats: LibraryStats): string {
	const base = `${stats.total} games · ${stats.installed} installed`;
	const updates = updateCountOf(stats);
	return updates ? `${base} · ${updates} updates` : base;
}
