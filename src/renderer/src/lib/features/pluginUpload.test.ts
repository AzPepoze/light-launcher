import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
	isArchivePath,
	planPluginUploads,
	uploadPlugin,
	type UploadFile
} from "./pluginUpload";

function fakeFile(name: string): File {
	return new File(["data"], name);
}

function item(path: string): UploadFile {
	return { path, file: fakeFile(path.split("/").pop() ?? path) };
}

/** The paths of each plugin, so the tests read as the plugins the user will see. */
function pluginPaths(plugins: UploadFile[][]): string[][] {
	return plugins.map((files) => files.map((file) => file.path));
}

function jsonResponse(body: unknown): Response {
	return new Response(JSON.stringify(body), {
		headers: { "Content-Type": "application/json" }
	});
}

describe("isArchivePath", () => {
	it("accepts .llplugin and .zip in any case", () => {
		expect(isArchivePath("name.llplugin")).toBe(true);
		expect(isArchivePath("Name.ZIP")).toBe(true);
	});

	it("rejects other files", () => {
		expect(isArchivePath("plugin.json")).toBe(false);
		expect(isArchivePath("zipper.txt")).toBe(false);
	});
});

describe("planPluginUploads", () => {
	it("takes one archive as one plugin", () => {
		const files = [item("name.llplugin")];

		expect(planPluginUploads(files)).toEqual({ plugins: [files], skipped: [] });
	});

	it("takes one folder with plugin.json at its top level as one plugin", () => {
		const files = [item("myplugin/plugin.json"), item("myplugin/src/main.js")];

		expect(planPluginUploads(files)).toEqual({ plugins: [files], skipped: [] });
	});

	it("takes each archive as its own plugin, in drop order", () => {
		const plan = planPluginUploads([item("one.llplugin"), item("two.zip"), item("three.LLPLUGIN")]);

		expect(pluginPaths(plan.plugins)).toEqual([["one.llplugin"], ["two.zip"], ["three.LLPLUGIN"]]);
		expect(plan.skipped).toEqual([]);
	});

	it("takes each plugin folder as its own plugin with only its own files", () => {
		const plan = planPluginUploads([
			item("one/plugin.json"),
			item("one/main.js"),
			item("two/plugin.json"),
			item("two/src/index.js")
		]);

		expect(pluginPaths(plan.plugins)).toEqual([
			["one/plugin.json", "one/main.js"],
			["two/plugin.json", "two/src/index.js"]
		]);
		expect(plan.skipped).toEqual([]);
	});

	it("takes a mix of archives and folders as separate plugins", () => {
		const plan = planPluginUploads([
			item("archive.llplugin"),
			item("folder/plugin.json"),
			item("other.zip")
		]);

		expect(pluginPaths(plan.plugins)).toEqual([
			["archive.llplugin"],
			["folder/plugin.json"],
			["other.zip"]
		]);
	});

	it("skips a folder without plugin.json and still takes the valid plugins", () => {
		const plan = planPluginUploads([item("good.llplugin"), item("loose/src/main.js")]);

		expect(pluginPaths(plan.plugins)).toEqual([["good.llplugin"]]);
		expect(plan.skipped).toEqual([{ name: "loose", reason: "not-a-plugin" }]);
	});

	it("skips a loose text file by name", () => {
		const plan = planPluginUploads([item("notes.txt"), item("good.zip")]);

		expect(pluginPaths(plan.plugins)).toEqual([["good.zip"]]);
		expect(plan.skipped).toEqual([{ name: "notes.txt", reason: "not-a-plugin" }]);
	});

	it("skips plugin.json loose at the top level", () => {
		const plan = planPluginUploads([item("plugin.json")]);

		expect(plan).toEqual({
			plugins: [],
			skipped: [{ name: "plugin.json", reason: "not-a-plugin" }]
		});
	});

	it("skips a plugin.json two folders down, and hints that the folder holds plugins", () => {
		const plan = planPluginUploads([item("myplugin/sub/plugin.json")]);

		expect(plan).toEqual({
			plugins: [],
			skipped: [{ name: "myplugin", reason: "nested-plugins" }]
		});
	});

	it("skips a folder with files only three folders down", () => {
		const plan = planPluginUploads([item("myplugin/a/b/plugin.json")]);

		expect(plan).toEqual({
			plugins: [],
			skipped: [{ name: "myplugin", reason: "not-a-plugin" }]
		});
	});

	it("skips an archive inside a folder, since only top-level archives are plugins", () => {
		const plan = planPluginUploads([item("bundle/inner.zip")]);

		expect(plan).toEqual({
			plugins: [],
			skipped: [{ name: "bundle", reason: "not-a-plugin" }]
		});
	});

	it("does not expand a folder that holds several plugin folders, and says why", () => {
		const plan = planPluginUploads([
			item("bundle/one/plugin.json"),
			item("bundle/two/plugin.json"),
			item("good.llplugin")
		]);

		expect(pluginPaths(plan.plugins)).toEqual([["good.llplugin"]]);
		expect(plan.skipped).toEqual([{ name: "bundle", reason: "nested-plugins" }]);
	});

	it("reports only junk as skipped and no plugins", () => {
		const plan = planPluginUploads([item("a.txt"), item("b.png"), item("empty/readme.md")]);

		expect(plan.plugins).toEqual([]);
		expect(plan.skipped).toEqual([
			{ name: "a.txt", reason: "not-a-plugin" },
			{ name: "b.png", reason: "not-a-plugin" },
			{ name: "empty", reason: "not-a-plugin" }
		]);
	});

	it("returns nothing for an empty drop", () => {
		expect(planPluginUploads([])).toEqual({ plugins: [], skipped: [] });
	});
});

describe("uploadPlugin", () => {
	beforeEach(() => {
		vi.stubGlobal("window", {
			location: { hash: "#t=secret", pathname: "/", search: "" },
			history: { replaceState: vi.fn() }
		});
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it("posts each file with its relative path as the part name and file name", async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValue(jsonResponse({ success: true, data: { path: "/x" } }));
		vi.stubGlobal("fetch", fetchMock);
		const files = [item("myplugin/plugin.json"), item("myplugin/src/main.js")];

		await uploadPlugin(files);

		const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
		expect(url).toBe("/api/plugin-upload");
		expect(init.method).toBe("POST");
		const form = init.body as FormData;
		const parts = [...form.entries()] as [string, File][];
		expect(parts.map(([name, file]) => [name, file.name])).toEqual([
			["myplugin/plugin.json", "myplugin/plugin.json"],
			["myplugin/src/main.js", "myplugin/src/main.js"]
		]);
	});

	it("sends the token header and does not set a content type", async () => {
		const fetchMock = vi
			.fn()
			.mockResolvedValue(jsonResponse({ success: true, data: { path: "/x" } }));
		vi.stubGlobal("fetch", fetchMock);

		await uploadPlugin([item("name.llplugin")]);

		const init = (fetchMock.mock.calls[0] as [string, RequestInit])[1];
		const headers = init.headers as Record<string, string>;
		expect(headers["X-LL-Token"]).toBe("secret");
		expect(headers["Content-Type"]).toBeUndefined();
	});

	it("returns the server path", async () => {
		vi.stubGlobal(
			"fetch",
			vi
				.fn()
				.mockResolvedValue(jsonResponse({ success: true, data: { path: "/tmp/up/name.llplugin" } }))
		);

		await expect(uploadPlugin([item("name.llplugin")])).resolves.toBe("/tmp/up/name.llplugin");
	});

	it("throws the server error text when the upload fails", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue(jsonResponse({ success: false, error: "archive is too large" }))
		);

		await expect(uploadPlugin([item("name.llplugin")])).rejects.toThrow("archive is too large");
	});
});
