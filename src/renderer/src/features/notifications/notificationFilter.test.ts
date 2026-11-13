import { describe, expect, it } from "vitest";
import type { NotificationEntry } from "../../stores/notificationStore";
import {
	NOTIFICATION_FILTERS,
	countNotifications,
	filterNotifications,
	filterTabs,
	isNotificationFilter
} from "./notificationFilter";

function entry(overrides: Partial<NotificationEntry>): NotificationEntry {
	return {
		id: 1,
		message: "Saved",
		type: "info",
		time: 1_000,
		count: 1,
		read: false,
		...overrides
	};
}

const history: NotificationEntry[] = [
	entry({ id: 1, type: "error", read: false }),
	entry({ id: 2, type: "info", read: true }),
	entry({ id: 3, type: "error", read: true }),
	entry({ id: 4, type: "success", read: false })
];

describe("notification filters", () => {
	it("keeps every entry for the all filter", () => {
		expect(filterNotifications(history, "all").map((e) => e.id)).toEqual([1, 2, 3, 4]);
	});

	it("keeps only unread entries for the unread filter", () => {
		expect(filterNotifications(history, "unread").map((e) => e.id)).toEqual([1, 4]);
	});

	it("keeps only error entries for the errors filter", () => {
		expect(filterNotifications(history, "errors").map((e) => e.id)).toEqual([1, 3]);
	});

	it("counts each filter", () => {
		expect(countNotifications(history)).toEqual({ all: 4, unread: 2, errors: 2 });
	});

	it("counts an empty history as zero", () => {
		expect(countNotifications([])).toEqual({ all: 0, unread: 0, errors: 0 });
	});

	it("builds the tabs in order with the counts as badges", () => {
		expect(filterTabs({ all: 4, unread: 2, errors: 0 })).toEqual([
			{ id: "all", label: "All", badge: 4 },
			{ id: "unread", label: "Unread", badge: 2 },
			{ id: "errors", label: "Errors", badge: 0 }
		]);
	});
});

describe("isNotificationFilter", () => {
	it("accepts every filter id", () => {
		expect(NOTIFICATION_FILTERS.every(isNotificationFilter)).toBe(true);
	});

	it("rejects values that are not filter ids", () => {
		expect(isNotificationFilter("")).toBe(false);
		expect(isNotificationFilter("All")).toBe(false);
		expect(isNotificationFilter("today")).toBe(false);
		expect(isNotificationFilter("error")).toBe(false);
	});
});
