import { describe, expect, it } from "vitest";
import {
	HOME_VIEW_STORAGE_KEY,
	defaultHomeView,
	readHomeView,
	writeHomeView,
	type HomeViewStorage
} from "./homeViewMemory";

function fakeStorage(initial: Record<string, string> = {}): HomeViewStorage & { data: Record<string, string> } {
	const data = { ...initial };
	return {
		data,
		getItem: (key) => (key in data ? data[key] : null),
		setItem: (key, value) => {
			data[key] = value;
		}
	};
}

function throwingStorage(): HomeViewStorage {
	return {
		getItem: () => {
			throw new Error("blocked");
		},
		setItem: () => {
			throw new Error("blocked");
		}
	};
}

describe("homeViewMemory", () => {
	it("round trips scroll and view", () => {
		const storage = fakeStorage();
		writeHomeView(storage, { scrollTop: 420, view: "list-grid" });
		expect(readHomeView(storage)).toEqual({ scrollTop: 420, view: "list-grid" });
	});

	it("returns defaults when nothing is saved", () => {
		expect(readHomeView(fakeStorage())).toEqual(defaultHomeView());
	});

	it("returns defaults for bad JSON", () => {
		const storage = fakeStorage({ [HOME_VIEW_STORAGE_KEY]: "{not json" });
		expect(readHomeView(storage)).toEqual(defaultHomeView());
	});

	it("returns defaults for non-object JSON", () => {
		const storage = fakeStorage({ [HOME_VIEW_STORAGE_KEY]: "42" });
		expect(readHomeView(storage)).toEqual(defaultHomeView());
	});

	it("falls back to the default view when the view is unknown", () => {
		const storage = fakeStorage({
			[HOME_VIEW_STORAGE_KEY]: JSON.stringify({ scrollTop: 10, view: "carousel" })
		});
		expect(readHomeView(storage)).toEqual({ scrollTop: 10, view: "grid" });
	});

	it("resets an invalid scroll value to zero", () => {
		const storage = fakeStorage({
			[HOME_VIEW_STORAGE_KEY]: JSON.stringify({ scrollTop: -5, view: "sidebar-grid" })
		});
		expect(readHomeView(storage)).toEqual({ scrollTop: 0, view: "sidebar-grid" });
	});

	it("returns defaults when storage throws on read", () => {
		expect(readHomeView(throwingStorage())).toEqual(defaultHomeView());
	});

	it("does not throw when storage throws on write", () => {
		expect(() => writeHomeView(throwingStorage(), { scrollTop: 1, view: "grid" })).not.toThrow();
	});
});
