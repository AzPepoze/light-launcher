import { describe, expect, it } from "vitest";
import type { UploadFile } from "./pluginUpload";
import { PluginInstallQueue, type QueueView } from "./pluginInstallQueue";

function item(path: string): UploadFile {
	return { path, file: new File(["x"], path) };
}

/** An upload that resolves with a server path named after the first file. */
function serverPath(files: UploadFile[]): Promise<string> {
	return Promise.resolve(`/uploads/${files[0].path}`);
}

function batch(...names: string[]): UploadFile[][] {
	return names.map((name) => [item(name)]);
}

describe("PluginInstallQueue", () => {
	it("starts idle", () => {
		const queue = new PluginInstallQueue(serverPath);

		expect(queue.view).toMatchObject({ phase: "idle", position: 0, total: 0 });
		expect(queue.busy).toBe(false);
	});

	it("uploads the first plugin and waits for review", async () => {
		const queue = new PluginInstallQueue(serverPath);

		await queue.start(batch("one.llplugin", "two.llplugin"));

		expect(queue.view).toMatchObject({
			phase: "review",
			position: 1,
			total: 2,
			installed: 0,
			source: "/uploads/one.llplugin"
		});
		expect(queue.busy).toBe(true);
	});

	it("uploads each plugin only when its turn comes", async () => {
		const uploaded: string[] = [];
		const queue = new PluginInstallQueue(async (files) => {
			uploaded.push(files[0].path);
			return `/uploads/${files[0].path}`;
		});

		await queue.start(batch("one.llplugin", "two.llplugin", "three.llplugin"));
		expect(uploaded).toEqual(["one.llplugin"]);

		await queue.markInstalled();
		expect(uploaded).toEqual(["one.llplugin", "two.llplugin"]);
	});

	it("counts installed plugins and ends with done", async () => {
		const queue = new PluginInstallQueue(serverPath);

		await queue.start(batch("one.llplugin", "two.llplugin"));
		await queue.markInstalled();
		await queue.markInstalled();

		expect(queue.view).toMatchObject({ phase: "done", installed: 2, total: 2, source: "" });
		expect(queue.busy).toBe(false);
	});

	it("skips only the current plugin and goes on with the next", async () => {
		const queue = new PluginInstallQueue(serverPath);

		await queue.start(batch("one.llplugin", "two.llplugin", "three.llplugin"));
		await queue.skip();

		expect(queue.view).toMatchObject({
			phase: "review",
			position: 2,
			installed: 0,
			source: "/uploads/two.llplugin"
		});
	});

	it("ends with done when the last plugin is skipped", async () => {
		const queue = new PluginInstallQueue(serverPath);

		await queue.start(batch("one.llplugin"));
		await queue.skip();

		expect(queue.view).toMatchObject({ phase: "done", installed: 0, total: 1 });
	});

	it("shows an upload failure and continues only after it is dismissed", async () => {
		const queue = new PluginInstallQueue(async (files) => {
			if (files[0].path === "bad.zip") throw new Error("archive is too large");
			return `/uploads/${files[0].path}`;
		});

		await queue.start(batch("bad.zip", "good.zip"));

		expect(queue.view).toMatchObject({ phase: "failed", failure: "archive is too large", position: 1 });
		expect(queue.busy).toBe(true);

		await queue.dismissFailure();

		expect(queue.view).toMatchObject({ phase: "review", position: 2, source: "/uploads/good.zip" });
	});

	it("uses a generic message for an upload error without text", async () => {
		const queue = new PluginInstallQueue(() => Promise.reject("boom"));

		await queue.start(batch("one.llplugin"));

		expect(queue.view.failure).toBe("Something went wrong.");
	});

	it("ignores skip, markInstalled and dismissFailure when they do not apply", async () => {
		const queue = new PluginInstallQueue(serverPath);

		await queue.skip();
		await queue.markInstalled();
		await queue.dismissFailure();
		expect(queue.view.phase).toBe("idle");

		await queue.start(batch("one.llplugin"));
		await queue.dismissFailure();
		expect(queue.view).toMatchObject({ phase: "review", installed: 0 });
	});

	it("does not start a new batch while one is busy", async () => {
		const queue = new PluginInstallQueue(serverPath);

		await queue.start(batch("one.llplugin"));
		await queue.start(batch("two.llplugin"));

		expect(queue.view).toMatchObject({ phase: "review", total: 1, source: "/uploads/one.llplugin" });
	});

	it("starts a new batch after the last one is done and resets the counts", async () => {
		const queue = new PluginInstallQueue(serverPath);
		await queue.start(batch("one.llplugin"));
		await queue.markInstalled();

		await queue.start(batch("two.llplugin", "three.llplugin"));

		expect(queue.view).toMatchObject({ phase: "review", position: 1, total: 2, installed: 0 });
	});

	it("ignores an empty start", async () => {
		const queue = new PluginInstallQueue(serverPath);

		await queue.start([]);

		expect(queue.view.phase).toBe("idle");
	});

	it("tells subscribers about each change and stops after unsubscribe", async () => {
		const queue = new PluginInstallQueue(serverPath);
		const phases: string[] = [];
		const stop = queue.subscribe((view: QueueView) => phases.push(view.phase));

		await queue.start(batch("one.llplugin"));
		stop();
		await queue.markInstalled();

		expect(phases).toEqual(["idle", "uploading", "review"]);
	});
});
