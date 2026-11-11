import { describe, expect, it, vi } from "vitest";
import {
	coverKey,
	coverRoute,
	createCoverCache,
	type BlobLoader,
	type CoverSource
} from "./coverCache";

const local = (id: number, coverLocal = "covers/a.jpg"): CoverSource => ({
	id,
	cover: "https://remote.example/c.jpg",
	coverLocal
});

/** A loader that counts downloads and hands out a new URL for each blob. */
function fakeLoader(overrides: Partial<BlobLoader> = {}) {
	let counter = 0;
	const loader = {
		fetchBlob: vi.fn(async (path: string) => new Blob([path])),
		createUrl: vi.fn(() => `blob:${++counter}`),
		revokeUrl: vi.fn(),
		...overrides
	};
	return loader;
}

describe("coverKey", () => {
	it("is null when the game has no local cover", () => {
		expect(coverKey({ id: 1, cover: "https://x", coverLocal: null })).toBeNull();
	});

	it("changes when the cover file changes", () => {
		expect(coverKey(local(1, "covers/a.jpg"))).not.toBe(coverKey(local(1, "covers/b.jpg")));
	});
});

describe("createCoverCache", () => {
	it("downloads the cover route once and reuses the URL", async () => {
		const loader = fakeLoader();
		const cache = createCoverCache(loader);

		const first = await cache.load(local(7));
		const second = await cache.load(local(7));

		expect(first).toBe("blob:1");
		expect(second).toBe("blob:1");
		expect(loader.fetchBlob).toHaveBeenCalledTimes(1);
		expect(loader.fetchBlob).toHaveBeenCalledWith(coverRoute(7));
		expect(loader.revokeUrl).not.toHaveBeenCalled();
	});

	it("shares one download between callers that ask at the same time", async () => {
		const loader = fakeLoader();
		const cache = createCoverCache(loader);

		const [a, b] = await Promise.all([cache.load(local(3)), cache.load(local(3))]);

		expect(a).toBe(b);
		expect(loader.fetchBlob).toHaveBeenCalledTimes(1);
	});

	it("returns null without a download when the game has no local cover", async () => {
		const loader = fakeLoader();
		const cache = createCoverCache(loader);

		expect(await cache.load({ id: 4, cover: "https://x", coverLocal: null })).toBeNull();
		expect(loader.fetchBlob).not.toHaveBeenCalled();
	});

	it("revokes the old URL when the cover file changes", async () => {
		const loader = fakeLoader();
		const cache = createCoverCache(loader);

		await cache.load(local(5, "covers/a.jpg"));
		const next = await cache.load(local(5, "covers/b.jpg"));

		expect(next).toBe("blob:2");
		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
	});

	it("revokes the URL when the cover is removed", async () => {
		const loader = fakeLoader();
		const cache = createCoverCache(loader);

		await cache.load(local(6));
		expect(await cache.load({ id: 6, cover: null, coverLocal: null })).toBeNull();

		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
	});

	it("returns null when the download fails, so the caller can fall back", async () => {
		const loader = fakeLoader({
			fetchBlob: vi.fn(async () => {
				throw new Error("HTTP 404");
			})
		});
		const cache = createCoverCache(loader);

		expect(await cache.load(local(8))).toBeNull();
	});

	it("drops a download that the game no longer asks for", async () => {
		const finishers: Array<(blob: Blob) => void> = [];
		const loader = fakeLoader({
			fetchBlob: vi.fn(() => new Promise<Blob>((resolve) => finishers.push(resolve)))
		});
		const cache = createCoverCache(loader);

		const stale = cache.load(local(9, "covers/old.jpg"));
		const fresh = cache.load(local(9, "covers/new.jpg"));
		finishers.forEach((finish) => finish(new Blob(["x"])));

		expect(await stale).toBeNull();
		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
		expect(await fresh).toBe("blob:2");
	});
});
