import { SAMPLE_INTERVAL_MS } from "./speedView";

/**
 * Takes speed samples on an interval while transfers run. `setActive` is idempotent, so a
 * reactive caller can update it on every change without restarting the timer.
 */
export class SpeedSampler {
	private timer: ReturnType<typeof setInterval> | null = null;

	constructor(
		private readonly takeSample: () => void,
		private readonly intervalMs: number = SAMPLE_INTERVAL_MS
	) {}

	get running(): boolean {
		return this.timer !== null;
	}

	setActive(active: boolean): void {
		if (!active) {
			this.stop();
			return;
		}
		if (this.timer !== null) return;
		this.timer = setInterval(() => this.takeSample(), this.intervalMs);
	}

	stop(): void {
		if (this.timer === null) return;
		clearInterval(this.timer);
		this.timer = null;
	}
}
