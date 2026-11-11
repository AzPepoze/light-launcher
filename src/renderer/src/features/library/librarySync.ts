import type { LibrarySyncProgress } from "../../lib/features/types";

const FAILED_MESSAGE = "The sync failed. Check the plugin, then try again.";

function countsText(progress: LibrarySyncProgress): string {
	const parts = [`${progress.imported} added`, `${progress.skipped} unchanged`];
	if (progress.failed > 0) parts.push(`${progress.failed} failed`);
	return parts.join(", ");
}

/** The sentence shown next to a source's Sync now button. */
export function syncMessage(progress: LibrarySyncProgress): string {
	if (progress.phase === "error") return progress.message || FAILED_MESSAGE;
	if (progress.phase === "done") return `Sync finished. ${countsText(progress)}.`;
	return `Syncing. ${countsText(progress)}.`;
}
