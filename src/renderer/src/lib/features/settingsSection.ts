/** The element id of the "Game folders" section on the Settings page. */
export const GAME_FOLDERS_SECTION = "game-folders";

export interface SectionScroll {
	id: string;
	behavior: ScrollBehavior;
}

/** Turns a pending section request into a scroll. Reduced motion jumps instead of animating. */
export function sectionScroll(section: string | null, reducedMotion: boolean): SectionScroll | null {
	if (!section) return null;
	return { id: section, behavior: reducedMotion ? "auto" : "smooth" };
}
