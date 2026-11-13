import { describe, expect, it } from "vitest";
import type { LibrarySource } from "@lib/features/types";
import {
	folderRelation,
	isInsideFolder,
	pluginFolderMatches,
	pluginFolderWarning
} from "./pluginFolders";

function makeSource(id: string, overrides: Partial<LibrarySource> = {}): LibrarySource {
	return {
		source: id,
		label: id,
		kind: "plugin",
		plugin: id,
		capabilities: [],
		sync: false,
		updateCheck: false,
		count: 0,
		tile: "poster",
		cover: "3:4",
		actions: [],
		card: null,
		sidebar: [],
		gamesFolder: null,
		...overrides
	};
}

describe("isInsideFolder", () => {
	it("is true for the parent folder itself", () => {
		expect(isInsideFolder("/games", "/games")).toBe(true);
	});

	it("is true for a folder inside the parent", () => {
		expect(isInsideFolder("/games/Sample", "/games")).toBe(true);
		expect(isInsideFolder("C:\\games\\Sample", "C:\\games")).toBe(true);
	});

	it("is false for a sibling that only starts with the same text", () => {
		expect(isInsideFolder("/games-old/Sample", "/games")).toBe(false);
	});

	it("is false for a folder above the parent", () => {
		expect(isInsideFolder("/", "/games")).toBe(false);
	});

	it("ignores case and trailing slashes", () => {
		expect(isInsideFolder("/Games/Sample/", "/games/")).toBe(true);
	});

	it("is false for an empty path", () => {
		expect(isInsideFolder("", "/games")).toBe(false);
		expect(isInsideFolder("/games/Sample", "")).toBe(false);
	});
});

describe("folderRelation", () => {
	it("is same for the same folder, whatever its case or trailing slash", () => {
		expect(folderRelation("/games", "/games")).toBe("same");
		expect(folderRelation("/Games/", "/games")).toBe("same");
		expect(folderRelation("C:\\Games\\", "c:\\games")).toBe("same");
	});

	it("is inside for a child folder", () => {
		expect(folderRelation("/games/f95", "/games")).toBe("inside");
		expect(folderRelation("/games/f95/saves", "/games")).toBe("inside");
		expect(folderRelation("C:\\games\\f95", "C:\\games")).toBe("inside");
	});

	it("is contains for a parent folder", () => {
		expect(folderRelation("/games", "/games/f95")).toBe("contains");
		expect(folderRelation("C:\\games", "C:\\games\\f95")).toBe("contains");
	});

	it("is null for unrelated folders", () => {
		expect(folderRelation("/media/games", "/games")).toBeNull();
		expect(folderRelation("/games-old", "/games")).toBeNull();
	});

	it("treats the root as the parent of every absolute path", () => {
		expect(folderRelation("/games", "/")).toBe("inside");
		expect(folderRelation("/", "/games")).toBe("contains");
		expect(folderRelation("C:\\Games", "C:\\")).toBe("inside");
		expect(folderRelation("C:\\", "C:\\Games")).toBe("contains");
	});

	it("does not match a relative path to an absolute path", () => {
		expect(folderRelation("Games", "/Games")).toBeNull();
	});

	it("is null for an empty path", () => {
		expect(folderRelation("", "/games")).toBeNull();
		expect(folderRelation("/games", "")).toBeNull();
	});
});

describe("pluginFolderMatches", () => {
	it("finds a plugin games folder that is the scan folder", () => {
		const matches = pluginFolderMatches("/games", [
			makeSource("f95", { label: "F95 Zone", gamesFolder: "/games" })
		]);
		expect(matches).toEqual([
			{ source: "f95", label: "F95 Zone", folder: "/games", relation: "same" }
		]);
	});

	it("finds a scan folder inside a plugin games folder", () => {
		const matches = pluginFolderMatches("/games/f95/mod", [
			makeSource("f95", { label: "F95 Zone", gamesFolder: "/games" })
		]);
		expect(matches.map((match) => match.relation)).toEqual(["inside"]);
	});

	it("finds a plugin games folder inside the scan folder", () => {
		const matches = pluginFolderMatches("/games", [
			makeSource("f95", { label: "F95 Zone", gamesFolder: "/games/f95" })
		]);
		expect(matches.map((match) => match.relation)).toEqual(["contains"]);
	});

	it("leaves out sources without a games folder and unrelated folders", () => {
		const matches = pluginFolderMatches("/games", [
			makeSource("manual", { label: "Local games", kind: "local" }),
			makeSource("itch", { label: "itch.io", gamesFolder: null }),
			makeSource("f95", { label: "F95 Zone", gamesFolder: "/elsewhere" })
		]);
		expect(matches).toEqual([]);
	});

	it("falls back to the source id when the label is blank", () => {
		const matches = pluginFolderMatches("/games", [
			makeSource("f95", { label: " ", gamesFolder: "/games" })
		]);
		expect(matches[0]?.label).toBe("f95");
	});

	it("lists every plugin that overlaps the folder", () => {
		const matches = pluginFolderMatches("/games", [
			makeSource("f95", { label: "F95 Zone", gamesFolder: "/games" }),
			makeSource("itch", { label: "itch.io", gamesFolder: "/games/itch" }),
			makeSource("other", { label: "Other", gamesFolder: "/media" })
		]);
		expect(matches.map((match) => match.source)).toEqual(["f95", "itch"]);
	});
});

describe("pluginFolderWarning", () => {
	it("names the plugin for the same folder", () => {
		const text = pluginFolderWarning({
			source: "f95",
			label: "F95 Zone",
			folder: "/games",
			relation: "same"
		});
		expect(text).toContain("F95 Zone");
		expect(text).toContain("can add duplicates");
	});

	it("names the plugin and its folder when one contains the other", () => {
		const inside = pluginFolderWarning({
			source: "f95",
			label: "F95 Zone",
			folder: "/games",
			relation: "inside"
		});
		expect(inside).toContain("inside F95 Zone");
		expect(inside).toContain("/games");

		const contains = pluginFolderWarning({
			source: "f95",
			label: "F95 Zone",
			folder: "/games/f95",
			relation: "contains"
		});
		expect(contains).toContain("contains F95 Zone");
		expect(contains).toContain("/games/f95");
	});
});
