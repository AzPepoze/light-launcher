import { describe, expect, it } from "vitest";
import {
	addLoadingGroups,
	folderBaseName,
	loadedGroups,
	pathMatchKey,
	upsertGroup,
	withLibraryCovers,
	type CoveredGame,
	type HomeFolderGroup
} from "./scanGroups";

const game = { name: "Alpha", path: "C:/Games/Alpha/alpha.exe" } as never;

describe("folderBaseName", () => {
	it("returns the last segment for Windows and POSIX paths", () => {
		expect(folderBaseName("C:\\Games\\Library")).toBe("Library");
		expect(folderBaseName("/home/user/Games/")).toBe("Games");
	});

	it("returns the path itself when it has no separator", () => {
		expect(folderBaseName("Games")).toBe("Games");
	});
});

describe("addLoadingGroups", () => {
	it("adds an empty loading group per folder", () => {
		expect(addLoadingGroups([], ["D:/Lib"])).toEqual([
			{ folderPath: "D:/Lib", folderName: "Lib", games: [], loading: true }
		]);
	});

	it("keeps groups that are already listed", () => {
		const listed: HomeFolderGroup[] = [
			{ folderPath: "D:/Lib", folderName: "Lib", games: [game], loading: false }
		];
		const next = addLoadingGroups(listed, ["D:/Lib", "E:/More"]);
		expect(next).toHaveLength(2);
		expect(next[0]).toBe(listed[0]);
		expect(next[1]).toMatchObject({ folderPath: "E:/More", loading: true });
	});
});

describe("upsertGroup", () => {
	const loading: HomeFolderGroup[] = [
		{ folderPath: "D:/Lib", folderName: "Lib", games: [], loading: true },
		{ folderPath: "E:/More", folderName: "More", games: [], loading: true }
	];

	it("replaces the group with the same folderPath and marks it loaded", () => {
		const next = upsertGroup(loading, { folderPath: "D:/Lib", folderName: "Lib", games: [game] });
		expect(next).toEqual([
			{ folderPath: "D:/Lib", folderName: "Lib", games: [game], loading: false },
			loading[1]
		]);
	});

	it("appends a group whose folder is not listed", () => {
		const next = upsertGroup(loading, { folderPath: "F:/New", folderName: "New", games: [] });
		expect(next).toHaveLength(3);
		expect(next[2]).toEqual({ folderPath: "F:/New", folderName: "New", games: [], loading: false });
	});

	it("does not change the input list", () => {
		upsertGroup(loading, { folderPath: "D:/Lib", folderName: "Lib", games: [] });
		expect(loading[0].loading).toBe(true);
	});
});

describe("loadedGroups", () => {
	it("marks every group of a full result as loaded", () => {
		expect(loadedGroups([{ folderPath: "D:/Lib", folderName: "Lib", games: [] }])).toEqual([
			{ folderPath: "D:/Lib", folderName: "Lib", games: [], loading: false }
		]);
	});
});

describe("pathMatchKey", () => {
	it("matches paths across case, separators and a trailing slash", () => {
		expect(pathMatchKey("C:\\Games\\Alpha\\")).toBe(pathMatchKey("c:/games/alpha"));
	});
});

describe("withLibraryCovers", () => {
	const scanned = { name: "Alpha", path: "C:\\Games\\Alpha\\alpha.exe" } as never;
	const library = {
		name: "Alpha",
		path: "c:/games/alpha/alpha.exe",
		id: "7",
		cover: "https://example.test/alpha.jpg",
		coverLocal: "covers/7.jpg"
	} as CoveredGame;
	const groups: HomeFolderGroup[] = [
		{ folderPath: "C:/Games", folderName: "Games", games: [scanned], loading: false }
	];

	it("copies id, cover and coverLocal from the library game with the same path", () => {
		const [group] = withLibraryCovers(groups, [library]);
		expect(group.games[0]).toMatchObject({ id: "7", cover: library.cover, coverLocal: "covers/7.jpg" });
	});

	it("leaves a scanned game without a library match unchanged", () => {
		const [group] = withLibraryCovers(groups, []);
		expect(group.games[0]).toBe(scanned);
	});

	it("does not change the input groups", () => {
		withLibraryCovers(groups, [library]);
		expect(groups[0].games[0]).toBe(scanned);
	});
});
