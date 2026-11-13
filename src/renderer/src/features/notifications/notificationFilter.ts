import type { TabItem } from "../../components/shared/tabs";
import type { NotificationEntry } from "../../stores/notificationStore";

export type NotificationFilter = "all" | "unread" | "errors";

export interface NotificationCounts {
	all: number;
	unread: number;
	errors: number;
}

export const NOTIFICATION_FILTERS: readonly NotificationFilter[] = ["all", "unread", "errors"];

const FILTER_LABELS: Record<NotificationFilter, string> = {
	all: "All",
	unread: "Unread",
	errors: "Errors"
};

const FILTER_MATCHERS: Record<NotificationFilter, (entry: NotificationEntry) => boolean> = {
	all: () => true,
	unread: (entry) => !entry.read,
	errors: (entry) => entry.type === "error"
};

/** Whether a tab id names a notification filter. */
export function isNotificationFilter(value: string): value is NotificationFilter {
	return NOTIFICATION_FILTERS.some((filter) => filter === value);
}

export function filterNotifications(
	history: NotificationEntry[],
	filter: NotificationFilter
): NotificationEntry[] {
	return history.filter(FILTER_MATCHERS[filter]);
}

export function countNotifications(history: NotificationEntry[]): NotificationCounts {
	return {
		all: filterNotifications(history, "all").length,
		unread: filterNotifications(history, "unread").length,
		errors: filterNotifications(history, "errors").length
	};
}

/** Tab strip for the page, with each filter's count as its badge. */
export function filterTabs(counts: NotificationCounts): TabItem[] {
	return NOTIFICATION_FILTERS.map((id) => ({
		id,
		label: FILTER_LABELS[id],
		badge: counts[id]
	}));
}
