import { describe, expect, it } from "vitest";
import { pushSample, speedLimitLabel } from "./speedView";

describe("pushSample", () => {
	it("appends a sample", () => {
		expect(pushSample([1, 2], 3)).toEqual([1, 2, 3]);
	});

	it("keeps only the newest samples", () => {
		expect(pushSample([1, 2, 3], 4, 3)).toEqual([2, 3, 4]);
	});

	it("does not change the array it was given", () => {
		const samples = [1, 2];
		pushSample(samples, 3);
		expect(samples).toEqual([1, 2]);
	});
});

describe("speedLimitLabel", () => {
	it("is null when there is no limit", () => {
		expect(speedLimitLabel(0)).toBeNull();
	});

	it("shows the limit in MB/s", () => {
		expect(speedLimitLabel(4096)).toBe("Limit 4 MB/s");
		expect(speedLimitLabel(512)).toBe("Limit 0.5 MB/s");
	});
});
