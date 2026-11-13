import {
	readHomeView,
	writeHomeView,
	type HomeView,
	type HomeViewMode,
	type HomeViewStorage
} from "../features/home/state/homeViewMemory";

const NO_STORAGE: HomeViewStorage = {
	getItem: () => null,
	setItem: () => {}
};

function browserStorage(): HomeViewStorage {
	try {
		return window.localStorage;
	} catch {
		return NO_STORAGE;
	}
}

const WRITE_DELAY_MS = 250;

const storage = browserStorage();
let current: HomeView = readHomeView(storage);
let writeTimer: ReturnType<typeof setTimeout> | null = null;

export function getHomeView(): HomeView {
	return current;
}

export function saveHomeScroll(scrollTop: number): void {
	update({ ...current, scrollTop: Math.max(0, Math.round(scrollTop)) });
}

export function saveHomeViewMode(view: HomeViewMode): void {
	update({ ...current, view });
}

// Scrolling updates memory every frame but writes storage once it settles.
function update(next: HomeView): void {
	current = next;
	if (writeTimer !== null) clearTimeout(writeTimer);
	writeTimer = setTimeout(flush, WRITE_DELAY_MS);
}

function flush(): void {
	if (writeTimer !== null) clearTimeout(writeTimer);
	writeTimer = null;
	writeHomeView(storage, current);
}

window.addEventListener("pagehide", flush);
