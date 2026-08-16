import { describe, it, expect } from "vitest";
import { getCategoryColor, hashString, PASTEL_PALETTE } from "./colors";

describe("Logger Pastel Palette & Hashing", () => {
	it("returns deterministic color for the same category name", () => {
		const color1 = getCategoryColor("IPC");
		const color2 = getCategoryColor("IPC");
		const color3 = getCategoryColor("ipc");
		expect(color1.hex).toBe(color2.hex);
		expect(color1.hex).toBe(color3.hex);
	});

	it("hashes different categories across the pastel palette", () => {
		const categories = ["IPC", "Runner", "Proton", "Games", "Migration", "LSFG", "Config"];
		const colors = new Set(categories.map((c) => getCategoryColor(c).name));
		expect(colors.size).toBeGreaterThan(1);
	});

	it("handles empty or whitespace category gracefully", () => {
		const color = getCategoryColor("");
		expect(color).toBeDefined();
		expect(color.hex).toBe(PASTEL_PALETTE[0].hex);
	});

	it("provides ANSI, Hex, and CSS background colors for every pastel color", () => {
		for (const color of PASTEL_PALETTE) {
			expect(color.ansi).toContain("\x1b[38;2;");
			expect(color.hex).toMatch(/^#[0-9a-fA-F]{6}$/);
			expect(color.cssBg).toContain("rgba(");
			expect(color.cssBorder).toContain("rgba(");
		}
	});
});
