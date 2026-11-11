import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { isArchivePath, planPluginUpload, uploadPlugin, type UploadFile } from "./pluginUpload";

function fakeFile(name: string): File {
	return new File(["data"], name);
}

function item(path: string): UploadFile {
	return { path, file: fakeFile(path.split("/").pop() ?? path) };
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

describe("planPluginUpload", () => {
	it("accepts one archive file", () => {
		const files = [item("name.llplugin")];

		expect(planPluginUpload(files)).toEqual({ ok: true, files });
	});

	it("accepts a folder with plugin.json at its top level", () => {
		const files = [item("myplugin/plugin.json"), item("myplugin/src/main.js")];

		expect(planPluginUpload(files)).toEqual({ ok: true, files });
	});

	it("rejects a folder without plugin.json", () => {
		const plan = planPluginUpload([item("myplugin/src/main.js")]);

		expect(plan).toEqual({ ok: false, problem: "not-a-plugin" });
	});

	it("rejects plugin.json nested deeper than the folder", () => {
		const plan = planPluginUpload([item("myplugin/sub/plugin.json")]);

		expect(plan).toEqual({ ok: false, problem: "not-a-plugin" });
	});

	it("rejects a loose file that is not an archive", () => {
		const plan = planPluginUpload([item("readme.txt")]);

		expect(plan).toEqual({ ok: false, problem: "not-a-plugin" });
	});

	it("rejects two plugin folders at once", () => {
		const plan = planPluginUpload([item("one/plugin.json"), item("two/plugin.json")]);

		expect(plan).toEqual({ ok: false, problem: "several-plugins" });
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
