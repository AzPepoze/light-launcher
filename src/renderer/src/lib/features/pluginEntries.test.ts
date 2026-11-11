import { describe, expect, it } from "vitest";
import {
	collectDroppedFiles,
	filesFromInput,
	isSkippedPath,
	toRelativePath,
	walkEntry
} from "./pluginEntries";

type FakeEntry = Record<string, unknown>;

function fakeFileEntry(fullPath: string): FakeEntry {
	return {
		fullPath,
		isFile: true,
		isDirectory: false,
		file: (success: (file: File) => void) =>
			success(new File(["x"], fullPath.split("/").pop() ?? ""))
	};
}

/** A directory whose reader returns the children in batches of the given size, then an empty batch. */
function fakeDirectoryEntry(fullPath: string, children: FakeEntry[], batchSize = 100): FakeEntry {
	return {
		fullPath,
		isFile: false,
		isDirectory: true,
		createReader: () => {
			let offset = 0;
			return {
				readEntries: (success: (batch: FakeEntry[]) => void) => {
					const batch = children.slice(offset, offset + batchSize);
					offset += batch.length;
					success(batch);
				}
			};
		}
	};
}

function pathsOf(files: { path: string }[]): string[] {
	return files.map((item) => item.path).sort();
}

describe("isSkippedPath", () => {
	it("skips node_modules and .git folders at any depth", () => {
		expect(isSkippedPath("myplugin/node_modules/x/index.js")).toBe(true);
		expect(isSkippedPath("myplugin/.git/config")).toBe(true);
		expect(isSkippedPath(".git")).toBe(true);
	});

	it("keeps similar names", () => {
		expect(isSkippedPath("myplugin/node_modules_old/a.js")).toBe(false);
		expect(isSkippedPath("myplugin/.gitignore")).toBe(false);
	});
});

describe("toRelativePath", () => {
	it("removes the leading slash of a dropped entry path", () => {
		expect(toRelativePath("/myplugin/plugin.json")).toBe("myplugin/plugin.json");
		expect(toRelativePath("/name.llplugin")).toBe("name.llplugin");
	});
});

describe("walkEntry", () => {
	it("lists a single dropped file by its name", async () => {
		const files = await walkEntry(fakeFileEntry("/name.llplugin") as unknown as FileSystemEntry);

		expect(pathsOf(files)).toEqual(["name.llplugin"]);
	});

	it("walks a folder recursively with paths from the dropped root", async () => {
		const folder = fakeDirectoryEntry("/myplugin", [
			fakeFileEntry("/myplugin/plugin.json"),
			fakeDirectoryEntry("/myplugin/src", [fakeFileEntry("/myplugin/src/main.js")])
		]);

		const files = await walkEntry(folder as unknown as FileSystemEntry);

		expect(pathsOf(files)).toEqual(["myplugin/plugin.json", "myplugin/src/main.js"]);
	});

	it("skips node_modules and .git folders", async () => {
		const folder = fakeDirectoryEntry("/myplugin", [
			fakeFileEntry("/myplugin/plugin.json"),
			fakeDirectoryEntry("/myplugin/node_modules", [fakeFileEntry("/myplugin/node_modules/a.js")]),
			fakeDirectoryEntry("/myplugin/.git", [fakeFileEntry("/myplugin/.git/HEAD")])
		]);

		const files = await walkEntry(folder as unknown as FileSystemEntry);

		expect(pathsOf(files)).toEqual(["myplugin/plugin.json"]);
	});

	it("reads every batch until the reader returns an empty one", async () => {
		const children = Array.from({ length: 5 }, (_, i) => fakeFileEntry(`/big/f${i}.txt`));
		const folder = fakeDirectoryEntry("/big", children, 2);

		const files = await walkEntry(folder as unknown as FileSystemEntry);

		expect(files).toHaveLength(5);
	});
});

describe("collectDroppedFiles", () => {
	it("walks each dropped file entry and ignores non-file items", async () => {
		const dataTransfer = {
			items: [
				{ kind: "file", webkitGetAsEntry: () => fakeFileEntry("/name.llplugin") },
				{ kind: "string", webkitGetAsEntry: () => null }
			]
		};

		const files = await collectDroppedFiles(dataTransfer as unknown as DataTransfer);

		expect(pathsOf(files)).toEqual(["name.llplugin"]);
	});

	it("walks dropped folders", async () => {
		const folder = fakeDirectoryEntry("/myplugin", [fakeFileEntry("/myplugin/plugin.json")]);
		const dataTransfer = {
			items: [{ kind: "file", webkitGetAsEntry: () => folder }]
		};

		const files = await collectDroppedFiles(dataTransfer as unknown as DataTransfer);

		expect(pathsOf(files)).toEqual(["myplugin/plugin.json"]);
	});
});

describe("filesFromInput", () => {
	it("uses the file name for a picked file", () => {
		const file = new File(["x"], "name.llplugin");

		expect(filesFromInput([file])).toEqual([{ path: "name.llplugin", file }]);
	});

	it("uses the relative path for a picked folder and skips node_modules", () => {
		const plugin = new File(["x"], "plugin.json");
		const dependency = new File(["x"], "index.js");
		Object.defineProperty(plugin, "webkitRelativePath", { value: "myplugin/plugin.json" });
		Object.defineProperty(dependency, "webkitRelativePath", {
			value: "myplugin/node_modules/x/index.js"
		});

		const files = filesFromInput([plugin, dependency]);

		expect(files.map((item) => item.path)).toEqual(["myplugin/plugin.json"]);
	});
});
