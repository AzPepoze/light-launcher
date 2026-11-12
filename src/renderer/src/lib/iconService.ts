import { GetExeIcon, GetImageBase64 } from "@lib/api";
import { createLogger } from "./logger";

const log = createLogger("iconService");

export type IconPriority = "high" | "low";

const iconPromises = new Map<string, Promise<string>>();

export interface GameIconResult {
	icon: string;
	source: "custom" | "exe" | null;
}

interface QueueItem {
	filePath: string;
	resolve: (value: string) => void;
	priority: IconPriority;
}

const queue: QueueItem[] = [];
let active = 0;
let activeLow = 0;
let paused = false;
const CONCURRENCY_LIMIT = 8;
const LOW_CONCURRENCY_LIMIT = 2;

/** Pause extraction while the user scrolls so spawned extractors don't steal CPU from rendering. */
export function setIconQueuePaused(value: boolean): void {
	if (paused === value) return;
	paused = value;
	if (!paused) processQueue();
}

function start(item: QueueItem) {
	active++;
	if (item.priority === "low") activeLow++;
	(async () => {
		try {
			item.resolve((await GetExeIcon(item.filePath)) || "");
		} catch (err) {
			log.error("Failed to load exe icon", err);
			item.resolve("");
		} finally {
			active--;
			if (item.priority === "low") activeLow--;
			processQueue();
		}
	})();
}

function processQueue() {
	if (paused) return;

	while (active < CONCURRENCY_LIMIT) {
		const index = queue.findIndex((item) => item.priority === "high");
		if (index === -1) break;
		start(queue.splice(index, 1)[0]);
	}

	while (active < CONCURRENCY_LIMIT && activeLow < LOW_CONCURRENCY_LIMIT) {
		const index = queue.findIndex((item) => item.priority === "low");
		if (index === -1) break;
		start(queue.splice(index, 1)[0]);
	}
}

/** Loads an exe icon through the concurrency-limited queue, reusing in-flight promises. */
export function loadExeIcon(filePath: string, priority: IconPriority = "high"): Promise<string> {
	if (!filePath) return Promise.resolve("");

	let promise = iconPromises.get(filePath);
	if (!promise) {
		promise = new Promise<string>((resolve) => {
			queue.push({ filePath, resolve, priority });
			processQueue();
		});
		iconPromises.set(filePath, promise);
	}
	return promise;
}

/** Loads a custom icon when configured, falling back to the executable icon. */
export async function loadGameIcon(
	gamePath: string,
	customIconPath: string | null = null,
	priority: IconPriority = "high"
): Promise<GameIconResult> {
	if (customIconPath) {
		try {
			const customIcon = (await GetImageBase64(customIconPath)) || "";
			if (customIcon) return { icon: customIcon, source: "custom" };
		} catch (error) {
			log.warn("Custom game icon unavailable", error);
		}
	}

	const icon = (await loadExeIcon(gamePath, priority)) || "";
	return { icon, source: icon ? "exe" : null };
}
