export interface TabItem {
	id: string;
	label: string;
	icon?: string;
	/** Shown beside the label when above zero. */
	badge?: number;
}

const LAST = (ids: string[]) => ids.length - 1;

/** The tab id that a key moves to, or null when the key does nothing for tabs. */
export function nextTabId(ids: string[], current: string, key: string): string | null {
	if (ids.length === 0) return null;
	const index = Math.max(0, ids.indexOf(current));
	switch (key) {
		case "ArrowRight":
			return ids[index === LAST(ids) ? 0 : index + 1];
		case "ArrowLeft":
			return ids[index === 0 ? LAST(ids) : index - 1];
		case "Home":
			return ids[0];
		case "End":
			return ids[LAST(ids)];
		default:
			return null;
	}
}
