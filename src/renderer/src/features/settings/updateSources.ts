import type { LibrarySource } from "@lib/features/types";
import { LOCAL_SOURCE } from "../library/librarySections";
import { sourceIcon } from "../library/sourceIcon";

/** One row in the Updates settings source list. */
export interface UpdateSource {
	id: string;
	label: string;
	/** Material icon name. */
	icon: string;
}

/**
 * The library sources that can check for updates. Local and hand-added sources never run
 * update checks, so they stay out. One row per source id, sorted by label.
 */
export function updateCheckSources(sources: LibrarySource[]): UpdateSource[] {
	const rows = sources
		.filter(
			(source) => source.kind !== "local" && source.source !== LOCAL_SOURCE && source.updateCheck
		)
		.map((source) => ({
			id: source.source,
			label: source.label.trim() || source.source,
			icon: sourceIcon(source.source)
		}));
	const byId = new Map<string, UpdateSource>();
	for (const row of rows) byId.set(row.id, row);
	return [...byId.values()].sort((a, b) => a.label.localeCompare(b.label));
}
