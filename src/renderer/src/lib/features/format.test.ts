import { describe, expect, it } from "vitest";
import { ago, bytes, eta, formatBytes, percent, percentLabel, playtime, sizeRange, speed } from "./format";

describe("percentLabel", () => {
	it("rounds to a whole percent and shows a dash when unknown", () => {
		expect(percentLabel(null)).toBe("--");
		expect(percentLabel(0)).toBe("0%");
		expect(percentLabel(41.6)).toBe("42%");
	});
});

describe("sizeRange", () => {
	it("joins done and total and leaves out missing sides", () => {
		expect(sizeRange(1536, 4 * 1024 ** 2)).toBe("1.5 KB / 4.0 MB");
		expect(sizeRange(undefined, 2 * 1024 ** 3)).toBe("2.0 GB");
		expect(sizeRange(1024, null)).toBe("1.0 KB");
		expect(sizeRange(undefined, undefined)).toBe("");
	});
});

describe("bytes", () => {
	it("formats sizes with the right unit", () => {
		expect(bytes(0)).toBe("0 B");
		expect(bytes(512)).toBe("512 B");
		expect(bytes(1536)).toBe("1.5 KB");
		expect(bytes(5 * 1024 ** 2)).toBe("5.0 MB");
		expect(bytes(12 * 1024 ** 3)).toBe("12 GB");
	});

	it("treats invalid numbers as zero", () => {
		expect(bytes(Number.NaN)).toBe("0 B");
		expect(bytes(-4)).toBe("0 B");
	});
});

describe("formatBytes", () => {
	it("keeps one decimal for prefix sizes", () => {
		expect(formatBytes(null)).toBe("--");
		expect(formatBytes(512)).toBe("512 B");
		expect(formatBytes(1536)).toBe("1.5 KB");
		expect(formatBytes(10 * 1024)).toBe("10.0 KB");
		expect(formatBytes(5 * 1024 ** 2)).toBe("5.0 MB");
		expect(formatBytes(2 * 1024 ** 3)).toBe("2.0 GB");
	});
});

describe("speed", () => {
	it("adds a per second suffix", () => {
		expect(speed(2 * 1024 ** 2)).toBe("2.0 MB/s");
	});
});

describe("percent", () => {
	it("clamps to 0..100 and handles empty totals", () => {
		expect(percent(50, 200)).toBe(25);
		expect(percent(300, 200)).toBe(100);
		expect(percent(10, 0)).toBe(0);
	});
});

describe("eta", () => {
	it("shows seconds, minutes and hours", () => {
		expect(eta(null)).toBe("--");
		expect(eta(42)).toBe("42s");
		expect(eta(125)).toBe("2m 05s");
		expect(eta(3900)).toBe("1h 05m");
	});
});

describe("playtime", () => {
	it("shows minutes and hours", () => {
		expect(playtime(30)).toBe("0 min");
		expect(playtime(45 * 60)).toBe("45 min");
		expect(playtime(3 * 3600 + 5 * 60)).toBe("3h 5m");
	});
});

describe("ago", () => {
	const now = Date.parse("2026-05-01T12:00:00Z");

	it("describes elapsed time", () => {
		expect(ago(null, now)).toBe("never");
		expect(ago("not a date", now)).toBe("never");
		expect(ago("2026-05-01T11:59:40Z", now)).toBe("just now");
		expect(ago("2026-05-01T11:30:00Z", now)).toBe("30m ago");
		expect(ago("2026-05-01T09:00:00Z", now)).toBe("3h ago");
		expect(ago("2026-04-28T12:00:00Z", now)).toBe("3d ago");
	});
});
