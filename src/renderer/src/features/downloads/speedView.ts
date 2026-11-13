export const SAMPLE_COUNT = 60;
export const SAMPLE_INTERVAL_MS = 1000;

/** Adds a sample and keeps only the newest `max`. Returns a new array. */
export function pushSample(samples: number[], value: number, max: number = SAMPLE_COUNT): number[] {
	const next = [...samples, value];
	return next.length > max ? next.slice(next.length - max) : next;
}

/** "Limit 4 MB/s" from a limit in KB/s, or null when there is no limit. */
export function speedLimitLabel(kbps: number): string | null {
	if (kbps <= 0) return null;
	const mbps = Math.round((kbps / 1024) * 100) / 100;
	return `Limit ${mbps} MB/s`;
}
