import type { SettingsNavGroup } from "./settingsSection";

/** Pixels below the top of the scroll area where a section counts as reached. */
export const SPY_LINE_PX = 24;

/** The section ids in nav order, which is also page order. */
export function navSectionIds(groups: SettingsNavGroup[]): string[] {
	return groups.flatMap((group) => group.items.map((item) => item.id));
}

/**
 * The section the reader is in: the last one whose top has reached the line.
 * Above the first section, the first one is active. At the bottom, the last one is.
 * `tops` maps a section id to its top edge, measured from the scroll area's top.
 */
export function activeSectionId(
	order: string[],
	tops: Record<string, number>,
	line: number,
	atBottom: boolean
): string | null {
	const measured = order.filter((id) => id in tops);
	if (measured.length === 0) return null;
	if (atBottom) return measured[measured.length - 1];

	const reached = measured.filter((id) => tops[id] <= line);
	return reached.length > 0 ? reached[reached.length - 1] : measured[0];
}
