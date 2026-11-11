import { describe, expect, it } from "vitest";
import { pluginStatusKind } from "./pluginStatus";

const READY = { status: "stopped", enabled: true, idle: false, signed: true } as const;

describe("pluginStatusKind", () => {
	it("shows enabled for a healthy signed plugin", () => {
		expect(pluginStatusKind(READY)).toBe("enabled");
	});

	it("shows disabled before sleeping or unsigned", () => {
		expect(pluginStatusKind({ ...READY, enabled: false, idle: true, signed: false })).toBe(
			"disabled"
		);
		expect(pluginStatusKind({ ...READY, status: "disabled" })).toBe("disabled");
	});

	it("shows error for a broken plugin before everything else", () => {
		expect(pluginStatusKind({ ...READY, status: "broken", enabled: false })).toBe("error");
	});

	it("shows sleeping for an idle plugin that is enabled", () => {
		expect(pluginStatusKind({ ...READY, idle: true })).toBe("sleeping");
	});

	it("shows unsigned for an enabled plugin without a signature", () => {
		expect(pluginStatusKind({ ...READY, signed: false })).toBe("unsigned");
	});

	it("leaves a running plugin enabled", () => {
		expect(pluginStatusKind({ ...READY, status: "running" })).toBe("enabled");
	});
});
