import { describe, expect, it } from "vitest";
import { cleanPath, compileWildcardPatterns, isSubPath } from "./pathUtils";

describe("pathUtils", () => {
	it("cleanPath normalizes and trims path strings", () => {
		expect(cleanPath("   /home/user/games/../games/test.exe   ")).toBe("/home/user/games/test.exe");
		expect(cleanPath("")).toBe("");
	});

	it("isSubPath correctly detects sub-paths", () => {
		expect(isSubPath("/home/user/games", "/home/user/games/game1/game.exe")).toBe(true);
		expect(isSubPath("/home/user/games", "/home/user/other/game.exe")).toBe(false);
		expect(isSubPath("/home/user/games", "/home/user/games")).toBe(false);
	});

	it("compileWildcardPatterns creates valid regexes", () => {
		const regexes = compileWildcardPatterns(["*.exe", "unins*"]);
		expect(regexes.length).toBe(2);
		expect(regexes[0].test("game.EXE")).toBe(true);
		expect(regexes[0].test("game.txt")).toBe(false);
		expect(regexes[1].test("unins000.exe")).toBe(true);
	});
});
