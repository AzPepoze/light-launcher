import { writable } from "svelte/store";
import { navigateTo } from "./navigationStore";

export type LibraryTab = "games" | "updates" | "downloads" | "incoming";

/** The Library tab on screen. It stays set between visits to the page. */
export const libraryTab = writable<LibraryTab>("games");

/** Opens the Library page on one of its tabs. */
export function openLibraryTab(tab: LibraryTab): void {
	libraryTab.set(tab);
	navigateTo("library");
}
