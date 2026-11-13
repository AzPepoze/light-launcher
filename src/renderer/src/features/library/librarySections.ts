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

/** The human name of a source. Falls back to the id when the overview has no label for it. */
export function sourceLabel(sources: LibrarySource[], source: string): string {
	const known = sources.find((item) => item.source === source)?.label;
	if (known) return known;
	return source === LOCAL_SOURCE ? LOCAL_LABEL : source;
}
