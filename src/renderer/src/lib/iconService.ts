import { GetExeIcon } from "@lib/api";

// Promise-based in-flight and result cache
const iconPromises = new Map<string, Promise<string>>();

interface QueueItem {
	filePath: string;
	resolve: (value: string) => void;
	reject: (err: any) => void;
}

const queue: QueueItem[] = [];
let activeCount = 0;
const CONCURRENCY_LIMIT = 4;

function processQueue() {
	if (activeCount >= CONCURRENCY_LIMIT || queue.length === 0) {
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
				console.error("Failed to load exe icon:", err);
				item.resolve("");
			} finally {
				activeCount--;
				processQueue();
			}
		})();
	}
}

/**
 * Loads an executable icon asynchronously through a concurrency-limited queue.
 * Reuses promises to cache results and prevent duplicate concurrent requests.
 */
export function loadExeIcon(filePath: string): Promise<string> {
	if (!filePath) return Promise.resolve("");

	let promise = iconPromises.get(filePath);
	if (!promise) {
		promise = new Promise<string>((resolve, reject) => {
			queue.push({ filePath, resolve, reject });
			processQueue();
		});
		iconPromises.set(filePath, promise);
	}
	return promise;
}

/**
 * Loads multiple exe icons in parallel (still throttled by the concurrency-limited queue)
 */
export async function loadExeIcons(...filePaths: string[]): Promise<(string | null)[]> {
	return Promise.all(filePaths.map((path) => loadExeIcon(path).catch(() => "")));
}
