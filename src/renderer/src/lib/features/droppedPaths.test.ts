import { describe, expect, it } from "vitest";
import { pathsFromUriList } from "./droppedPaths";

describe("pathsFromUriList", () => {
	it("returns the local path of each file:// line", () => {
		const uriList = "file:///home/me/Games/Game.exe\r\nfile:///home/me/Games/Other.exe\r\n";
		expect(pathsFromUriList(uriList)).toEqual([
			"/home/me/Games/Game.exe",
			"/home/me/Games/Other.exe",
		]);
	});

	it("decodes percent-encoded characters", () => {
		expect(pathsFromUriList("file:///home/me/My%20Game%21.exe")).toEqual([
			"/home/me/My Game!.exe",
		]);
	});

	it("strips the leading slash before a Windows drive letter", () => {
		expect(pathsFromUriList("file:///C:/Games/Game.exe")).toEqual(["C:/Games/Game.exe"]);
	});

	it("skips comments and non-file lines", () => {
		const uriList = "# comment\nhttps://example.com/page\nfile:///home/me/Game.exe";
		expect(pathsFromUriList(uriList)).toEqual(["/home/me/Game.exe"]);
	});

	it("returns nothing for an empty list", () => {
		expect(pathsFromUriList("")).toEqual([]);
	});
});
