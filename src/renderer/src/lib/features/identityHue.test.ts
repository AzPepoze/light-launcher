import { describe, expect, it } from "vitest";
import { fnv1a, identityHue, identityInitial } from "./identityHue";

describe("fnv1a", () => {
	it("matches the published FNV-1a 32-bit test vectors", () => {
		expect(fnv1a("")).toBe(0x811c9dc5);
		expect(fnv1a("a")).toBe(0xe40c292c);
		expect(fnv1a("foobar")).toBe(0xbf9cf968);
	});
});

describe("identityHue", () => {
	it("gives the same hue for the same id on every call", () => {
		expect(identityHue("com.example.store")).toBe(identityHue("com.example.store"));
	});

	it("stays within 0 to 359 degrees", () => {
		for (const id of ["", "a", "com.example.store", "ünïcödé", "x".repeat(200)]) {
			const hue = identityHue(id);
			expect(hue).toBeGreaterThanOrEqual(0);
			expect(hue).toBeLessThan(360);
			expect(Number.isInteger(hue)).toBe(true);
		}
	});

	it("gives different ids different hues in a simple case", () => {
		expect(identityHue("a")).not.toBe(identityHue("b"));
	});
});

describe("identityInitial", () => {
	it("uses the first letter in upper case", () => {
		expect(identityInitial("example store")).toBe("E");
	});

	it("ignores leading spaces", () => {
		expect(identityInitial("  zeta")).toBe("Z");
	});

	it("keeps a whole first character outside the basic alphabet", () => {
		expect(identityInitial("ðeta")).toBe("Ð");
	});

	it("falls back to a question mark for an empty name", () => {
		expect(identityInitial("   ")).toBe("?");
	});
});
