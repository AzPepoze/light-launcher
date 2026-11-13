import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
	DEDUPE_WINDOW_MS,
	ERROR_TOAST_MS,
	HISTORY_KEY,
	HISTORY_LIMIT,
	TOAST_MS,
	countUnread,
	createNotificationStore,
	dayGroup,
	groupByDay,
	isRepeat,
	loadHistory,
	recordInHistory,
	relativeTime,
	trimHistory,
	type KeyValueStorage,
	type NotificationEntry
} from "./notificationStore";

function memoryStorage(
	initial: Record<string, string> = {}
): KeyValueStorage & { data: Map<string, string> } {
	const data = new Map(Object.entries(initial));
	return {
		data,
		getItem: (key) => data.get(key) ?? null,
		setItem: (key, value) => {
			data.set(key, value);
		}
	};
}

function entry(overrides: Partial<NotificationEntry> = {}): NotificationEntry {
	return {
		id: 1,
		message: "Saved",
		type: "success",
		time: 1_000,
		count: 1,
		read: false,
		...overrides
	};
}

describe("pure notification logic", () => {
	it("treats the same text and type within 3 seconds as a repeat", () => {
		const saved = entry({ time: 10_000 });
		expect(isRepeat(saved, { message: "Saved", type: "success" }, 10_000 + DEDUPE_WINDOW_MS)).toBe(
			true
		);
		expect(
			isRepeat(saved, { message: "Saved", type: "success" }, 10_000 + DEDUPE_WINDOW_MS + 1)
		).toBe(false);
		expect(isRepeat(saved, { message: "Saved", type: "error" }, 10_000)).toBe(false);
		expect(isRepeat(saved, { message: "Other", type: "success" }, 10_000)).toBe(false);
	});

	it("merges a repeat into the newest entry and counts it", () => {
		const first = recordInHistory([], { message: "Saved", type: "success" }, 1, 1_000);
		const second = recordInHistory(first.history, { message: "Saved", type: "success" }, 2, 2_000);
		expect(second.merged).toBe(true);
		expect(second.history).toHaveLength(1);
		expect(second.entry).toMatchObject({ id: 1, count: 2, time: 2_000, read: false });
	});

	it("adds a new entry at the front when the message differs", () => {
		const first = recordInHistory([], { message: "Saved", type: "success" }, 1, 1_000);
		const second = recordInHistory(first.history, { message: "Copied", type: "success" }, 2, 1_500);
		expect(second.merged).toBe(false);
		expect(second.history.map((item) => item.message)).toEqual(["Copied", "Saved"]);
	});

	it("keeps only the newest 100 entries", () => {
		const many = Array.from({ length: HISTORY_LIMIT + 5 }, (_, index) => entry({ id: index }));
		const trimmed = trimHistory(many);
		expect(trimmed).toHaveLength(HISTORY_LIMIT);
		expect(trimmed[0].id).toBe(0);
	});

	it("counts unread entries, not counts of repeats", () => {
		expect(countUnread([entry({ count: 4 }), entry({ id: 2, read: true }), entry({ id: 3 })])).toBe(
			2
		);
	});

	it("groups entries into today, yesterday and earlier", () => {
		const now = new Date(2026, 10, 13, 15).getTime();
		expect(dayGroup(new Date(2026, 10, 13, 1).getTime(), now)).toBe("today");
		expect(dayGroup(new Date(2026, 10, 12, 23).getTime(), now)).toBe("yesterday");
		expect(dayGroup(new Date(2026, 10, 11, 23).getTime(), now)).toBe("earlier");

		const groups = groupByDay(
			[
				entry({ id: 1, time: new Date(2026, 10, 13, 14).getTime() }),
				entry({ id: 2, time: new Date(2026, 10, 13, 9).getTime() }),
				entry({ id: 3, time: new Date(2026, 10, 12, 9).getTime() })
			],
			now
		);
		expect(groups.map((group) => [group.day, group.entries.map((item) => item.id)])).toEqual([
			["today", [1, 2]],
			["yesterday", [3]]
		]);
	});

	it("shows relative time with the shared formatter", () => {
		const now = 10 * 60_000;
		expect(relativeTime(now - 3 * 60_000, now)).toBe("3m ago");
		expect(relativeTime(now, now)).toBe("just now");
	});

	it("reads a saved history and drops invalid entries", () => {
		const valid = entry({ id: 5, time: 50 });
		const storage = memoryStorage({
			[HISTORY_KEY]: JSON.stringify([valid, { id: "bad" }, { ...valid, id: 6, type: "loud" }])
		});
		expect(loadHistory(storage).map((item) => item.id)).toEqual([5]);
	});

	it("returns an empty history for broken saved data", () => {
		expect(loadHistory(memoryStorage({ [HISTORY_KEY]: "{not json" }))).toEqual([]);
		expect(loadHistory(memoryStorage({ [HISTORY_KEY]: '{"a":1}' }))).toEqual([]);
		expect(loadHistory(null)).toEqual([]);
	});
});

describe("createNotificationStore", () => {
	let clock: number;
	const now = () => clock;

	beforeEach(() => {
		vi.useFakeTimers();
		clock = 1_000_000;
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it("adds a toast and a history entry with the warning type", () => {
		const store = createNotificationStore({ now });
		store.warning("Check the API key");
		expect(get(store.toasts)).toEqual([
			{ id: expect.any(Number), message: "Check the API key", type: "warning", count: 1 }
		]);
		expect(get(store.history)[0]).toMatchObject({ type: "warning", read: false, count: 1 });
	});

	it("gives every notice a unique id", () => {
		const store = createNotificationStore({ now });
		store.info("First");
		store.info("Second");
		store.info("Third");
		const ids = get(store.history).map((item) => item.id);
		expect(new Set(ids).size).toBe(3);
	});

	it("does not add a repeat within 3 seconds, but counts it", () => {
		const store = createNotificationStore({ now });
		store.error("Launch failed");
		clock += 1_000;
		store.error("Launch failed");
		expect(get(store.toasts)).toHaveLength(1);
		expect(get(store.toasts)[0].count).toBe(2);
		expect(get(store.history)).toHaveLength(1);
		expect(get(store.history)[0].count).toBe(2);
	});

	it("adds a new entry for the same text after 3 seconds", () => {
		const store = createNotificationStore({ now });
		store.error("Launch failed");
		clock += DEDUPE_WINDOW_MS + 1;
		store.error("Launch failed");
		expect(get(store.history)).toHaveLength(2);
	});

	it("keeps errors for 8 seconds and other toasts for 5", () => {
		const store = createNotificationStore({ now });
		store.error("Bad");
		store.success("Good");
		vi.advanceTimersByTime(TOAST_MS);
		expect(get(store.toasts).map((toast) => toast.type)).toEqual(["error"]);
		vi.advanceTimersByTime(ERROR_TOAST_MS - TOAST_MS);
		expect(get(store.toasts)).toEqual([]);
	});

	it("keeps a silent notice in the history without a toast", () => {
		const store = createNotificationStore({ now });
		store.info("Background sync", { silent: true, source: "Updates" });
		expect(get(store.toasts)).toEqual([]);
		expect(get(store.history)[0]).toMatchObject({ message: "Background sync", source: "Updates" });
	});

	it("pauses a toast while hovered and resumes with the time left", () => {
		const store = createNotificationStore({ now });
		store.info("Hover me");
		const id = get(store.toasts)[0].id;

		vi.advanceTimersByTime(2_000);
		clock += 2_000;
		store.pause(id);
		vi.advanceTimersByTime(60_000);
		expect(get(store.toasts)).toHaveLength(1);

		store.resume(id);
		vi.advanceTimersByTime(TOAST_MS - 2_000 - 1);
		expect(get(store.toasts)).toHaveLength(1);
		vi.advanceTimersByTime(1);
		expect(get(store.toasts)).toEqual([]);
	});

	it("tracks unread entries and marks them read", () => {
		const store = createNotificationStore({ now });
		store.info("One");
		store.info("Two");
		expect(get(store.unreadCount)).toBe(2);

		const first = get(store.history).find((item) => item.message === "One");
		store.markRead(first!.id);
		expect(get(store.unreadCount)).toBe(1);

		store.markAllRead();
		expect(get(store.unreadCount)).toBe(0);
	});

	it("removes an entry from the history and its toast", () => {
		const store = createNotificationStore({ now });
		store.info("Remove me");
		const id = get(store.history)[0].id;
		store.remove(id);
		expect(get(store.history)).toEqual([]);
		expect(get(store.toasts)).toEqual([]);
	});

	it("clears the history but keeps the toasts on screen", () => {
		const store = createNotificationStore({ now });
		store.info("Gone from history");
		store.clearHistory();
		expect(get(store.history)).toEqual([]);
		expect(get(store.toasts)).toHaveLength(1);
	});

	it("saves the history without actions and restores it", () => {
		const storage = memoryStorage();
		const run = vi.fn();
		const store = createNotificationStore({ now, storage });
		store.error("Install failed", { source: "Downloads", action: { label: "Retry", run } });

		const saved = JSON.parse(storage.data.get(HISTORY_KEY) ?? "[]");
		expect(saved[0]).toMatchObject({ message: "Install failed", source: "Downloads" });
		expect(saved[0]).not.toHaveProperty("action");

		const restored = createNotificationStore({ now, storage });
		expect(get(restored.history)[0]).toMatchObject({
			message: "Install failed",
			source: "Downloads"
		});
		expect(get(restored.history)[0].action).toBeUndefined();
	});

	it("starts new ids after the highest saved id", () => {
		const storage = memoryStorage({
			[HISTORY_KEY]: JSON.stringify([
				{ id: 41, message: "Old", type: "info", time: 1, count: 1, read: true }
			])
		});
		const store = createNotificationStore({ now, storage });
		store.info("New");
		expect(get(store.history)[0].id).toBe(42);
	});

	it("still works when storage throws", () => {
		const broken: KeyValueStorage = {
			getItem: () => {
				throw new Error("blocked");
			},
			setItem: () => {
				throw new Error("full");
			}
		};
		const store = createNotificationStore({ now, storage: broken });
		expect(() => store.info("Still shown")).not.toThrow();
		expect(get(store.history)).toHaveLength(1);
	});

	it("turns an error into a plain message in withNotification", async () => {
		const store = createNotificationStore({ now });
		await expect(
			store.withNotification(Promise.reject(new Error("Disk full (code -1)")), {
				error: "Save failed"
			})
		).rejects.toThrow("Disk full");
		expect(get(store.history)[0]).toMatchObject({ message: "Save failed", type: "error" });
	});
});
