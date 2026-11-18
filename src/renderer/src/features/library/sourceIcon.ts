import { LOCAL_SOURCE } from "./librarySections";

const FALLBACK_ICON = "store";
const ICONS = new Map<string, string>([[LOCAL_SOURCE, "folder_open"]]);

/** The Material icon for a source in the source rail. Unknown sources get the store icon. */
export function sourceIcon(source: string): string {
	return ICONS.get(source) ?? FALLBACK_ICON;
}
