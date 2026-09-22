import { GetExeIcon } from "@lib/api";
import { createLogger } from "./logger";

const log = createLogger("iconService");

const iconPromises = new Map<string, Promise<string>>();

interface QueueItem {
	filePath: string;
	resolve: (value: string) => void;
}

const queue: QueueItem[] = [];
let activeCount = 0;
let paused = false;
const CONCURRENCY_LIMIT = 8;

/** Pause extraction while the user scrolls so spawned extractors don't steal CPU from rendering. */
export function setIconQueuePaused(value: boolean): void {
	if (paused === value) return;
	paused = value;
	if (!paused) processQueue();
}

function processQueue() {
	if (paused || activeCount >= CONCURRENCY_LIMIT || queue.length === 0) {
		return;
	}

	while (activeCount < CONCURRENCY_LIMIT && queue.length > 0) {
		const item = queue.shift();
		if (!item) continue;

		activeCount++;
		(async () => {
			try {
				const icon = await GetExeIcon(item.filePath);
				item.resolve(icon || "");
			} catch (err) {
				log.error("Failed to load exe icon", err);
				item.resolve("");
			} finally {
				activeCount--;
				processQueue();
			}
		})();
	}
}

/** Loads an exe icon through the concurrency-limited queue, reusing in-flight promises. */
export function loadExeIcon(filePath: string): Promise<string> {
	if (!filePath) return Promise.resolve("");

	let promise = iconPromises.get(filePath);
	if (!promise) {
		promise = new Promise<string>((resolve) => {
			queue.push({ filePath, resolve });
			processQueue();
		});
		iconPromises.set(filePath, promise);
	}
	return promise;
}
