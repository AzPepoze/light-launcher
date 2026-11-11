import { describe, expect, it, vi } from "vitest";
import {
	coverKey,
	coverRoute,
	createMediaCache,
	screenshotKey,
	screenshotRoute,
	type BlobLoader,
	type CoverSource
} from "./mediaCache";

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

describe("routes", () => {
	it("addresses covers and indexed screenshots under the media route", () => {
		expect(coverRoute(7)).toBe("/api/media/cover/7");
		expect(screenshotRoute(7, 2)).toBe("/api/media/screenshot/7/2");
	});

	it("identifies screenshots by game, index and cached file", () => {
		expect(screenshotKey(7, 2, null)).toBeNull();
		expect(screenshotKey(7, 2, "screenshots/a.png")).not.toBe(
			screenshotKey(7, 3, "screenshots/a.png")
		);
	});
});

describe("createMediaCache", () => {
	it("downloads a cover once and reuses the URL", async () => {
		const loader = fakeLoader();
		const cache = createMediaCache(loader);

		const first = await cache.load("a", coverKey(local(7)), coverRoute(7));
		const second = await cache.load("b", coverKey(local(7)), coverRoute(7));

		expect(first).toBe("blob:1");
		expect(second).toBe("blob:1");
		expect(loader.fetchBlob).toHaveBeenCalledTimes(1);
		expect(loader.fetchBlob).toHaveBeenCalledWith(coverRoute(7));
		expect(loader.revokeUrl).not.toHaveBeenCalled();
	});

	it("shares one download between callers that ask at the same time", async () => {
		const loader = fakeLoader();
		const cache = createMediaCache(loader);

		const [a, b] = await Promise.all([
			cache.load("a", coverKey(local(3)), coverRoute(3)),
			cache.load("b", coverKey(local(3)), coverRoute(3))
		]);

		expect(a).toBe(b);
		expect(loader.fetchBlob).toHaveBeenCalledTimes(1);
	});

	it("returns null without a download when there is no local file", async () => {
		const loader = fakeLoader();
		const cache = createMediaCache(loader);

		expect(await cache.load("a", null, coverRoute(4))).toBeNull();
		expect(loader.fetchBlob).not.toHaveBeenCalled();
	});

	it("revokes the old URL when the cover file changes", async () => {
		const loader = fakeLoader();
		const cache = createMediaCache(loader);

		await cache.load("a", coverKey(local(5, "covers/a.jpg")), coverRoute(5));
		const next = await cache.load("a", coverKey(local(5, "covers/b.jpg")), coverRoute(5));

		expect(next).toBe("blob:2");
		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
	});

	it("revokes the URL when the cover is removed", async () => {
		const loader = fakeLoader();
		const cache = createMediaCache(loader);

		await cache.load("a", coverKey(local(6)), coverRoute(6));
		expect(await cache.load("a", null, coverRoute(6))).toBeNull();

		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
	});

	it("keeps a shared URL until the last scope releases it", async () => {
		const loader = fakeLoader();
		const cache = createMediaCache(loader);
		const key = coverKey(local(6));

		await cache.load("a", key, coverRoute(6));
		await cache.load("b", key, coverRoute(6));
		cache.release("a");
		expect(loader.revokeUrl).not.toHaveBeenCalled();

		cache.release("b");
		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
	});

	it("returns null when the download fails, so the caller can fall back", async () => {
		const loader = fakeLoader({
			fetchBlob: vi.fn(async () => {
				throw new Error("HTTP 404");
			})
		});
		const cache = createMediaCache(loader);

		expect(await cache.load("a", coverKey(local(8)), coverRoute(8))).toBeNull();
	});

	it("drops a download that the scope no longer asks for", async () => {
		const finishers: Array<(blob: Blob) => void> = [];
		const loader = fakeLoader({
			fetchBlob: vi.fn(() => new Promise<Blob>((resolve) => finishers.push(resolve)))
		});
		const cache = createMediaCache(loader);

		const stale = cache.load("a", coverKey(local(9, "covers/old.jpg")), coverRoute(9));
		const fresh = cache.load("a", coverKey(local(9, "covers/new.jpg")), coverRoute(9));
		finishers.forEach((finish) => finish(new Blob(["x"])));

		expect(await stale).toBeNull();
		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
		expect(await fresh).toBe("blob:2");
	});

	it("revokes every URL on dispose", async () => {
		const loader = fakeLoader();
		const cache = createMediaCache(loader);

		await cache.load("a", coverKey(local(1)), coverRoute(1));
		await cache.load("b", screenshotKey(1, 0, "screenshots/a.png"), screenshotRoute(1, 0));
		cache.dispose();

		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:1");
		expect(loader.revokeUrl).toHaveBeenCalledWith("blob:2");
	});
});
