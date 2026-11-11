const STORAGE_KEY = "fpsMeter";

/** Reads the saved choice. Returns false when storage is missing or blocked. */
export function loadFpsMeterVisible(): boolean {
	try {
		return localStorage.getItem(STORAGE_KEY) === "on";
	} catch {
		return false;
	}
}

export function saveFpsMeterVisible(visible: boolean): void {
	try {
		localStorage.setItem(STORAGE_KEY, visible ? "on" : "off");
	} catch {
		// Storage is blocked; the meter still works for this session.
	}
}

export function isFpsMeterShortcut(e: KeyboardEvent): boolean {
	return e.ctrlKey && e.shiftKey && !e.repeat && e.key.toLowerCase() === "f";
}

export function isTypingTarget(target: EventTarget | null): boolean {
	if (!(target instanceof HTMLElement)) return false;
	return target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable;
}
