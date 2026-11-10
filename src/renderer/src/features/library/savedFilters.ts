import { DEFAULT_FILTERS, type LibraryFilters } from "./libraryFilters";

export interface SavedFilter {
	id: string;
	name: string;
	filters: LibraryFilters;
}

const STORAGE_KEY = "library.savedFilters";

export function parseSavedFilters(raw: string | null): SavedFilter[] {
	if (!raw) return [];
	try {
		const parsed: unknown = JSON.parse(raw);
		if (!Array.isArray(parsed)) return [];
		return parsed.filter(isSavedFilter).map((entry) => ({
			id: entry.id,
			name: entry.name,
			filters: { ...DEFAULT_FILTERS, ...entry.filters }
		}));
	} catch {
		return [];
	}
}

function isSavedFilter(value: unknown): value is SavedFilter {
	if (typeof value !== "object" || value === null) return false;
	const entry = value as Partial<SavedFilter>;
	return (
		typeof entry.id === "string" &&
		typeof entry.name === "string" &&
		typeof entry.filters === "object" &&
		entry.filters !== null
	);
}

export function loadSavedFilters(): SavedFilter[] {
	try {
		return parseSavedFilters(localStorage.getItem(STORAGE_KEY));
	} catch {
		return [];
	}
}

export function storeSavedFilters(filters: SavedFilter[]): void {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(filters));
	} catch {
		// Storage can be blocked; the filters still work for this session.
	}
}
