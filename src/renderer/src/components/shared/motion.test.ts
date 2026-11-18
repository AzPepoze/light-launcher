import { describe, expect, it } from "vitest";
import { ENTER_MS, EXIT_MS, fadeScale, menuPanelConfig, motionMs, slideIn } from "./motion";

describe("motionMs", () => {
	it("returns 0 for both kinds when reduced motion is on", () => {
		expect(motionMs("enter", true)).toBe(0);
		expect(motionMs("exit", true)).toBe(0);
	});

	it("returns the enter and exit durations otherwise", () => {
		expect(motionMs("enter", false)).toBe(ENTER_MS);
		expect(motionMs("exit", false)).toBe(EXIT_MS);
	});

	it("makes exit shorter than enter", () => {
		expect(EXIT_MS).toBeLessThan(ENTER_MS);
	});
});

describe("motion helpers", () => {
	it("scales the dialog from 0.96 when motion is on", () => {
		expect(fadeScale("enter", false).start).toBe(0.96);
	});

	it("removes the dialog scale and slide when reduced motion is on", () => {
		expect(fadeScale("exit", true)).toMatchObject({ duration: 0, start: 1 });
		expect(slideIn("enter", true)).toMatchObject({ duration: 0, x: 0, y: 0 });
	});

	it("slides toasts in from 24 px to the right", () => {
		expect(slideIn("enter", false).x).toBe(24);
	});

	it("moves a bottom menu down from above and a top menu up from below", () => {
		const bottom = menuPanelConfig("enter", "bottom", false).css as (t: number, u: number) => string;
		const top = menuPanelConfig("enter", "top", false).css as (t: number, u: number) => string;

		expect(bottom(0, 1)).toContain("translateY(-6px)");
		expect(top(0, 1)).toContain("translateY(6px)");
	});

	it("keeps menu panels still when reduced motion is on", () => {
		const css = menuPanelConfig("enter", "bottom", true).css as (t: number, u: number) => string;

		expect(menuPanelConfig("exit", "bottom", true).duration).toBe(0);
		expect(css(0, 1)).toContain("translateY(0px)");
		expect(css(0, 1)).toContain("scale(1)");
	});
});
