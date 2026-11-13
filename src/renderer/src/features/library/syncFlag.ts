import type { LibrarySource } from "../../lib/features/types";
import { LOCAL_SOURCE } from "./librarySections";

/**
 * Whether the built-in Sync now button shows. A boolean sync flag from the overview decides.
 * Until the flag is sent, plugin sources keep the current behavior and show it.
 */
export function canSyncGames(source: LibrarySource | undefined): boolean {
	if (!source || source.source === LOCAL_SOURCE) return false;
	return source.sync;
}
