import type { LibrarySource } from "../../lib/features/types";

/** The source id the server gives games added by hand or found in folders. */
export const LOCAL_SOURCE = "manual";
const LOCAL_LABEL = "Local games";

/** The local section when the overview does not list it, for example on a new library. */
const EMPTY_LOCAL: LibrarySource = {
	source: LOCAL_SOURCE,
	label: LOCAL_LABEL,
	kind: "local",
	plugin: null,
	capabilities: ["launch", "scan"],
	sync: false,
	updateCheck: false,
	count: 0,
	tile: "square",
	cover: "1:1",
	actions: [],
	card: null,
	sidebar: []
};

/** Local games first, then each plugin source in the order the server lists them. */
export function librarySections(sources: LibrarySource[]): LibrarySource[] {
	const local = sources.find((source) => source.source === LOCAL_SOURCE) ?? EMPTY_LOCAL;
	const plugins = sources.filter((source) => source.source !== LOCAL_SOURCE);
	return [local, ...plugins];
}

/** A section shows when it is the local one, or a plugin source with games or an error to report. */
export function showsSection(source: LibrarySource, total: number | null, hasError: boolean): boolean {
	if (source.source === LOCAL_SOURCE) return true;
	return (total ?? 0) > 0 || hasError;
}

/** The element id of a source's group on the Library page, used by the source chips to scroll to it. */
export function sectionElementId(source: string): string {
	return `library-section-${source}`;
}

/** "1 game" or "12 games". */
export function gameCount(count: number): string {
	return count === 1 ? "1 game" : `${count} games`;
}

/** The toolbar's stats line: "12 games, 4 installed" and "2 updates" when any are waiting. */
export function librarySubtitle(stats: { total: number; installed: number; updates: number }): string {
	const parts = [gameCount(stats.total), `${stats.installed} installed`];
	if (stats.updates > 0) parts.push(`${stats.updates} updates`);
	return parts.join(", ");
}

/** The human name of a source. Falls back to the id when the overview has no label for it. */
export function sourceLabel(sources: LibrarySource[], source: string): string {
	const known = sources.find((item) => item.source === source)?.label;
	if (known) return known;
	return source === LOCAL_SOURCE ? LOCAL_LABEL : source;
}
