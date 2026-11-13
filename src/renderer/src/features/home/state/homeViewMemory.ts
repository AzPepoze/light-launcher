export const HOME_VIEW_STORAGE_KEY = "home-view";

export type HomeViewMode = "grid" | "list-grid" | "sidebar-grid";

export interface HomeView {
	scrollTop: number;
	view: HomeViewMode;
}

export interface HomeViewStorage {
	getItem(key: string): string | null;
	setItem(key: string, value: string): void;
}

const HOME_VIEW_MODES: readonly HomeViewMode[] = ["grid", "list-grid", "sidebar-grid"];
const DEFAULT_VIEW: HomeViewMode = "grid";

export function defaultHomeView(): HomeView {
	return { scrollTop: 0, view: DEFAULT_VIEW };
}

export function readHomeView(storage: HomeViewStorage): HomeView {
	try {
		const raw = storage.getItem(HOME_VIEW_STORAGE_KEY);
		return raw === null ? defaultHomeView() : parseHomeView(JSON.parse(raw));
	} catch {
		return defaultHomeView();
	}
}

export function writeHomeView(storage: HomeViewStorage, value: HomeView): void {
	try {
		storage.setItem(HOME_VIEW_STORAGE_KEY, JSON.stringify(value));
	} catch {
		// Storage can be full or blocked; the view just is not remembered.
	}
}

function parseHomeView(value: unknown): HomeView {
	if (typeof value !== "object" || value === null) return defaultHomeView();
	const record = value as Record<string, unknown>;
	return {
		scrollTop: parseScrollTop(record.scrollTop),
		view: parseViewMode(record.view)
	};
}

function parseScrollTop(value: unknown): number {
	if (typeof value !== "number" || !Number.isFinite(value) || value < 0) return 0;
	return Math.round(value);
}

function parseViewMode(value: unknown): HomeViewMode {
	return HOME_VIEW_MODES.find((mode) => mode === value) ?? DEFAULT_VIEW;
}
