import { describe, expect, it } from "vitest";
import type { GameInfo, ScannedFolderGroup } from "@shared";
import {
	excludeNamesText,
	folderInputError,
	folderKey,
	gamesFoundText,
	gamesInFolder,
	normalizeExcludeNames,
	validateDepth
} from "./scanFolders";

const game = { name: "game", path: "game.exe" } as GameInfo;

describe("validateDepth", () => {
	it.each([
		["-1", -1],
		["1", 1],
		[" 2 ", 2],
		["999", 999]
	])("accepts %s", (text, depth) => {
		expect(validateDepth(text)).toEqual({ ok: true, depth });
	});

	it.each([
		["", "Enter a whole number."],
		["   ", "Enter a whole number."],
		["1.5", "Enter a whole number."],
		["abc", "Enter a whole number."],
		["0", "Use -1 for no limit, or a number from 1 to 999."],
		["-2", "Use -1 for no limit, or a number from 1 to 999."],
		["1000", "Use -1 for no limit, or a number from 1 to 999."]
	])("rejects %j", (text, message) => {
		expect(validateDepth(text)).toEqual({ ok: false, message });
	});
});

describe("normalizeExcludeNames", () => {
	it("trims, drops blanks and keeps the first spelling of a repeat", () => {
		expect(normalizeExcludeNames(" Unity.* , ,*64,unity.*,, redist ")).toEqual([
			"Unity.*",
			"*64",
			"redist"
		]);
	});

	it("returns an empty list for blank input", () => {
		expect(normalizeExcludeNames("  ,  ")).toEqual([]);
		expect(normalizeExcludeNames("")).toEqual([]);
	});

	it("joins names back into the editor text", () => {
		expect(excludeNamesText(["Unity.*", "redist"])).toBe("Unity.*, redist");
		expect(excludeNamesText([])).toBe("");
	});
});

describe("folderKey and folderInputError", () => {
	it("treats case and a trailing separator as the same folder", () => {
		expect(folderKey(" /games/Alpha/ ")).toBe("/games/alpha");
		expect(folderKey("C:\\Games\\Alpha\\")).toBe("c:\\games\\alpha");
	});

	it("asks for a path when the input is blank", () => {
		expect(folderInputError("   ", [])).toBe("Enter a folder path.");
	});

	it("refuses a folder that is already watched", () => {
		expect(folderInputError("/games/Alpha/", ["/games/alpha"])).toBe(
			"This folder is already watched."
		);
	});

	it("accepts a new folder", () => {
		expect(folderInputError("/games/beta", ["/games/alpha"])).toBeNull();
	});
});

describe("gamesInFolder and gamesFoundText", () => {
	const groups: ScannedFolderGroup[] = [
		{ folderPath: "/games/alpha", folderName: "alpha", games: [game, game] },
		{ folderPath: "/games/beta", folderName: "beta", games: [] }
	];

	it("counts the games of the matching folder", () => {
		expect(gamesInFolder(groups, "/games/Alpha/")).toBe(2);
		expect(gamesInFolder(groups, "/games/beta")).toBe(0);
	});

	it("returns null for a folder that was not scanned", () => {
		expect(gamesInFolder(groups, "/games/gamma")).toBeNull();
	});

	it("writes the count in plain words", () => {
		expect(gamesFoundText(0)).toBe("Found 0 games");
		expect(gamesFoundText(1)).toBe("Found 1 game");
		expect(gamesFoundText(12)).toBe("Found 12 games");
	});
});
