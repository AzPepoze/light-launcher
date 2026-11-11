import { describe, expect, it } from "vitest";
import { PageCache } from "./pageCache";

describe("PageCache", () => {
	it("returns what was stored and nothing for other pages", () => {
		const cache = new PageCache<string>(3);
		cache.set(4, "four");
		expect(cache.get(4)).toBe("four");
		expect(cache.has(5)).toBe(false);
		expect(cache.get(5)).toBeUndefined();
	});

	it("drops the least recently used page when a new one does not fit", () => {
		const cache = new PageCache<string>(2);
		cache.set(0, "a");
		cache.set(1, "b");
		cache.set(2, "c");
		expect(cache.has(0)).toBe(false);
		expect(cache.has(1)).toBe(true);
		expect(cache.has(2)).toBe(true);
	});

	it("keeps a page that was touched recently", () => {
		const cache = new PageCache<string>(2);
		cache.set(0, "a");
		cache.set(1, "b");
		cache.touch(0);
		cache.set(2, "c");
		expect(cache.has(0)).toBe(true);
		expect(cache.has(1)).toBe(false);
	});

	it("holds at most the capacity when many pages are loaded", () => {
		const cache = new PageCache<number>(20);
		for (let page = 0; page < 25; page++) cache.set(page, page);
		expect(cache.size).toBe(20);
		expect(cache.has(4)).toBe(false);
		expect(cache.has(5)).toBe(true);
		expect(cache.has(24)).toBe(true);
	});

	it("replaces a page in place without evicting another", () => {
		const cache = new PageCache<string>(2);
		cache.set(0, "a");
		cache.set(1, "b");
		cache.set(1, "B");
		expect(cache.get(0)).toBe("a");
		expect(cache.get(1)).toBe("B");
	});

	it("touching a missing page does nothing", () => {
		const cache = new PageCache<string>(2);
		cache.touch(9);
		expect(cache.size).toBe(0);
	});

	it("clears every page", () => {
		const cache = new PageCache<string>(2);
		cache.set(0, "a");
		cache.clear();
		expect(cache.entries()).toEqual([]);
	});
});
