import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SpeedSampler } from "./speedSampler";

describe("SpeedSampler", () => {
	beforeEach(() => vi.useFakeTimers());

	afterEach(() => vi.useRealTimers());

	it("does not sample while inactive", () => {
		const sample = vi.fn();
		const sampler = new SpeedSampler(sample);

		vi.advanceTimersByTime(3000);

		expect(sample).not.toHaveBeenCalled();
		expect(sampler.running).toBe(false);
	});

	it("samples once a second while active", () => {
		const sample = vi.fn();
		const sampler = new SpeedSampler(sample);

		sampler.setActive(true);
		vi.advanceTimersByTime(3000);

		expect(sample).toHaveBeenCalledTimes(3);
		expect(sampler.running).toBe(true);
	});

	it("stops sampling when it becomes inactive", () => {
		const sample = vi.fn();
		const sampler = new SpeedSampler(sample);

		sampler.setActive(true);
		vi.advanceTimersByTime(1000);
		expect(sample).toHaveBeenCalledTimes(1);

		sampler.setActive(false);
		vi.advanceTimersByTime(3000);

		expect(sample).toHaveBeenCalledTimes(1);
		expect(sampler.running).toBe(false);
	});

	it("keeps one timer when set active again", () => {
		const sample = vi.fn();
		const sampler = new SpeedSampler(sample);

		sampler.setActive(true);
		sampler.setActive(true);
		vi.advanceTimersByTime(2000);

		expect(sample).toHaveBeenCalledTimes(2);
	});

	it("stops the timer when stopped", () => {
		const sample = vi.fn();
		const sampler = new SpeedSampler(sample);

		sampler.setActive(true);
		sampler.stop();
		vi.advanceTimersByTime(2000);

		expect(sample).not.toHaveBeenCalled();
		expect(sampler.running).toBe(false);
	});
});
