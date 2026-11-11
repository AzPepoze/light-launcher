export interface FrameStats {
	fps: number;
	p50: number;
	p95: number;
	worst: number;
	longFrames: number;
	displayHz: number;
}

const EMPTY_STATS: FrameStats = { fps: 0, p50: 0, p95: 0, worst: 0, longFrames: 0, displayHz: 0 };

const PERCENTILE_WINDOW = 120;
const SECOND_MS = 1000;
const SECOND_TOLERANCE_MS = 1;
const LONG_FRAME_FACTOR = 1.5;
const RATE_TOLERANCE = 0.04;

export const COMMON_REFRESH_RATES = [30, 48, 50, 60, 72, 75, 90, 100, 120, 144, 165, 180, 240];

/** Nearest-rank percentile. `p` is between 0 and 1. `sorted` must be ascending. */
export function percentile(sorted: ArrayLike<number>, p: number): number {
	if (sorted.length === 0) return 0;
	const rank = Math.ceil(p * sorted.length) - 1;
	const index = Math.min(sorted.length - 1, Math.max(0, rank));
	return sorted[index];
}

/** Turns a frame time in ms into a refresh rate in Hz, or 0 when unknown. */
export function snapToRefreshRate(frameMs: number): number {
	if (!Number.isFinite(frameMs) || frameMs <= 0) return 0;
	const hz = 1000 / frameMs;
	return nearestCommonRate(hz) ?? Math.round(hz);
}

/**
 * Summarizes frame intervals (ms), oldest first.
 * Percentiles and display rate use the last 120 intervals. Worst and long frames use all of them.
 * fps counts the intervals that fit in the last second.
 */
export function summarize(intervalsMs: ArrayLike<number>): FrameStats {
	if (intervalsMs.length === 0) return { ...EMPTY_STATS };
	const recent = sortedTail(intervalsMs, PERCENTILE_WINDOW);
	const p50 = percentile(recent, 0.5);
	return {
		fps: framesInLastSecond(intervalsMs),
		p50,
		p95: percentile(recent, 0.95),
		worst: maxOf(intervalsMs),
		longFrames: countLongFrames(intervalsMs, p50),
		displayHz: snapToRefreshRate(percentile(recent, 0.05))
	};
}

/** Fixed-size ring of frame intervals. Pushing never allocates. */
export class FrameHistory {
	private readonly values: Float64Array;
	private next = 0;
	private size = 0;

	constructor(capacity: number) {
		this.values = new Float64Array(capacity);
	}

	push(intervalMs: number): void {
		this.values[this.next] = intervalMs;
		this.next = (this.next + 1) % this.values.length;
		this.size = Math.min(this.size + 1, this.values.length);
	}

	/** Newest intervals whose total fits in `windowMs`, oldest first. */
	recentWithin(windowMs: number): Float64Array {
		let total = 0;
		let count = 0;
		while (count < this.size) {
			const interval = this.stepsBack(count);
			if (total + interval > windowMs) break;
			total += interval;
			count++;
		}
		const out = new Float64Array(count);
		for (let i = 0; i < count; i++) {
			out[i] = this.stepsBack(count - 1 - i);
		}
		return out;
	}

	/** Interval `steps` back from the newest one (0 is the newest). */
	private stepsBack(steps: number): number {
		const capacity = this.values.length;
		return this.values[(this.next - 1 - steps + capacity) % capacity];
	}
}

function nearestCommonRate(hz: number): number | undefined {
	let best: number | undefined;
	let bestError = Infinity;
	for (const rate of COMMON_REFRESH_RATES) {
		const error = Math.abs(hz - rate) / rate;
		if (error < bestError) {
			bestError = error;
			best = rate;
		}
	}
	return bestError <= RATE_TOLERANCE ? best : undefined;
}

function sortedTail(values: ArrayLike<number>, count: number): Float64Array {
	const start = Math.max(0, values.length - count);
	const tail = new Float64Array(values.length - start);
	for (let i = start; i < values.length; i++) {
		tail[i - start] = values[i];
	}
	return tail.sort();
}

function framesInLastSecond(intervalsMs: ArrayLike<number>): number {
	let total = 0;
	let frames = 0;
	for (let i = intervalsMs.length - 1; i >= 0; i--) {
		total += intervalsMs[i];
		if (total > SECOND_MS + SECOND_TOLERANCE_MS) break;
		frames++;
	}
	return frames;
}

function countLongFrames(intervalsMs: ArrayLike<number>, medianMs: number): number {
	const threshold = medianMs * LONG_FRAME_FACTOR;
	let count = 0;
	for (let i = 0; i < intervalsMs.length; i++) {
		if (intervalsMs[i] > threshold) count++;
	}
	return count;
}

function maxOf(values: ArrayLike<number>): number {
	let max = 0;
	for (let i = 0; i < values.length; i++) {
		if (values[i] > max) max = values[i];
	}
	return max;
}
