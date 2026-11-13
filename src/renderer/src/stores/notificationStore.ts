import { derived, get, writable, type Readable } from "svelte/store";
import { ago } from "../lib/features/format";

export type NotificationType = "info" | "success" | "warning" | "error";

export interface NotificationAction {
	label: string;
	run: () => void;
}

export interface NotificationOptions {
	/** Where the notice came from, such as "Updates". */
	source?: string;
	action?: NotificationAction;
	/** Keeps the notice in the history only, with no toast. */
	silent?: boolean;
}

/** One notice in the history. Its toast, if any, shares the id. */
export interface NotificationEntry {
	id: number;
	message: string;
	type: NotificationType;
	/** Time of the latest occurrence, in ms since the epoch. */
	time: number;
	/** How many identical notices were merged into this entry. */
	count: number;
	read: boolean;
	source?: string;
	/** In memory only: functions are not saved. */
	action?: NotificationAction;
}

/** A notice shown on screen for a few seconds. */
export interface Toast {
	id: number;
	message: string;
	type: NotificationType;
	count: number;
}

export type DayGroup = "today" | "yesterday" | "earlier";

export interface DayGroupEntries {
	day: DayGroup;
	entries: NotificationEntry[];
}

export interface KeyValueStorage {
	getItem(key: string): string | null;
	setItem(key: string, value: string): void;
}

export const HISTORY_LIMIT = 100;
export const DEDUPE_WINDOW_MS = 3_000;
export const TOAST_MS = 5_000;
export const ERROR_TOAST_MS = 8_000;
export const HISTORY_KEY = "notifications-v1";

const TYPES: readonly NotificationType[] = ["info", "success", "warning", "error"];
const DAY_MS = 24 * 60 * 60 * 1000;

interface NewNotice {
	message: string;
	type: NotificationType;
	source?: string;
	action?: NotificationAction;
}

interface Recorded {
	history: NotificationEntry[];
	entry: NotificationEntry;
	/** True when the notice repeated the newest entry and was merged into it. */
	merged: boolean;
}

/** Whether a notice repeats an entry: same text and type, within the dedupe window. */
export function isRepeat(
	entry: Pick<NotificationEntry, "message" | "type" | "time">,
	notice: Pick<NewNotice, "message" | "type">,
	now: number
): boolean {
	return (
		entry.message === notice.message &&
		entry.type === notice.type &&
		now - entry.time <= DEDUPE_WINDOW_MS
	);
}

/** Keeps only the newest entries. */
export function trimHistory(
	history: NotificationEntry[],
	limit: number = HISTORY_LIMIT
): NotificationEntry[] {
	return history.slice(0, limit);
}

/** Adds a notice to the front of the history, or merges it into the newest entry when it repeats. */
export function recordInHistory(
	history: NotificationEntry[],
	notice: NewNotice,
	id: number,
	now: number
): Recorded {
	const [newest, ...older] = history;
	if (newest && isRepeat(newest, notice, now)) {
		const merged = { ...newest, time: now, count: newest.count + 1, read: false };
		return { history: [merged, ...older], entry: merged, merged: true };
	}
	const entry: NotificationEntry = {
		id,
		message: notice.message,
		type: notice.type,
		time: now,
		count: 1,
		read: false,
		source: notice.source,
		action: notice.action
	};
	return { history: trimHistory([entry, ...history]), entry, merged: false };
}

export function countUnread(history: NotificationEntry[]): number {
	return history.filter((entry) => !entry.read).length;
}

/** Start of the local day that holds `time`, in ms. */
function startOfDay(time: number): number {
	const day = new Date(time);
	day.setHours(0, 0, 0, 0);
	return day.getTime();
}

export function dayGroup(time: number, now: number): DayGroup {
	const today = startOfDay(now);
	if (time >= today) return "today";
	if (time >= startOfDay(today - DAY_MS / 2)) return "yesterday";
	return "earlier";
}

/** Splits the history into the day groups it has entries for, newest group first. */
export function groupByDay(history: NotificationEntry[], now: number): DayGroupEntries[] {
	const groups: DayGroupEntries[] = [];
	for (const entry of history) {
		const day = dayGroup(entry.time, now);
		const last = groups[groups.length - 1];
		if (last?.day === day) last.entries.push(entry);
		else groups.push({ day, entries: [entry] });
	}
	return groups;
}

/** Relative time such as "3m ago", using the shared formatter. */
export function relativeTime(time: number, now: number): string {
	return ago(new Date(time).toISOString(), now);
}

export function toastDuration(type: NotificationType): number {
	return type === "error" ? ERROR_TOAST_MS : TOAST_MS;
}

function isEntry(value: unknown): value is NotificationEntry {
	if (typeof value !== "object" || value === null) return false;
	const entry = value as Record<string, unknown>;
	return (
		typeof entry.id === "number" &&
		typeof entry.message === "string" &&
		TYPES.includes(entry.type as NotificationType) &&
		typeof entry.time === "number" &&
		typeof entry.count === "number" &&
		entry.count >= 1 &&
		typeof entry.read === "boolean" &&
		(entry.source === undefined || typeof entry.source === "string")
	);
}

/** Reads the saved history. Anything unreadable gives an empty history. */
export function loadHistory(storage: KeyValueStorage | null): NotificationEntry[] {
	try {
		const raw = storage?.getItem(HISTORY_KEY);
		if (!raw) return [];
		const parsed: unknown = JSON.parse(raw);
		if (!Array.isArray(parsed)) return [];
		return trimHistory(parsed.filter(isEntry).sort((a, b) => b.time - a.time));
	} catch {
		return [];
	}
}

function saveHistory(storage: KeyValueStorage | null, history: NotificationEntry[]): void {
	try {
		// Actions are functions, so they are left out.
		const saved = history.map(({ action: _action, ...entry }) => entry);
		storage?.setItem(HISTORY_KEY, JSON.stringify(saved));
	} catch {
		// Storage can be full or blocked; the in-memory history still works.
	}
}

function messageText(message: unknown): string {
	if (typeof message === "string") return message;
	const text = (message as { message?: unknown } | null)?.message;
	return typeof text === "string" && text ? text : String(message);
}

interface ToastTimer {
	handle?: ReturnType<typeof setTimeout>;
	remaining: number;
	startedAt: number;
}

export interface NotificationStoreDeps {
	now?: () => number;
	storage?: KeyValueStorage | null;
}

export interface NotificationStore {
	/** On-screen toasts. */
	toasts: Readable<Toast[]>;
	/** Every notice, newest first. */
	history: Readable<NotificationEntry[]>;
	unreadCount: Readable<number>;
	add(message: unknown, type?: NotificationType, options?: NotificationOptions): void;
	info(message: unknown, options?: NotificationOptions): void;
	success(message: unknown, options?: NotificationOptions): void;
	warning(message: unknown, options?: NotificationOptions): void;
	error(message: unknown, options?: NotificationOptions): void;
	/** Hides a toast now. The history keeps the notice. */
	dismiss(id: number): void;
	/** Removes a notice from the history and hides its toast. */
	remove(id: number): void;
	pause(id: number): void;
	resume(id: number): void;
	markRead(id: number): void;
	markAllRead(): void;
	clearHistory(): void;
	withNotification<T>(
		promise: Promise<T>,
		options: { pending?: string; success?: string; error?: string }
	): Promise<T>;
}

export function createNotificationStore({
	now = Date.now,
	storage = null
}: NotificationStoreDeps = {}): NotificationStore {
	const initial = loadHistory(storage);
	const history = writable<NotificationEntry[]>(initial);
	const toasts = writable<Toast[]>([]);
	const unreadCount = derived(history, countUnread);
	const timers = new Map<number, ToastTimer>();
	let nextId = initial.reduce((max, entry) => Math.max(max, entry.id), 0) + 1;

	function setHistory(next: NotificationEntry[]): void {
		history.set(next);
		saveHistory(storage, next);
	}

	function updateHistory(change: (entries: NotificationEntry[]) => NotificationEntry[]): void {
		setHistory(change(get(history)));
	}

	function dismiss(id: number): void {
		clearTimeout(timers.get(id)?.handle);
		timers.delete(id);
		toasts.update((list) => list.filter((toast) => toast.id !== id));
	}

	function startToast(entry: NotificationEntry): void {
		const duration = toastDuration(entry.type);
		toasts.update((list) => [
			...list,
			{ id: entry.id, message: entry.message, type: entry.type, count: entry.count }
		]);
		timers.set(entry.id, {
			handle: setTimeout(() => dismiss(entry.id), duration),
			remaining: duration,
			startedAt: now()
		});
	}

	function updateToastCount(entry: NotificationEntry): void {
		toasts.update((list) =>
			list.map((toast) => (toast.id === entry.id ? { ...toast, count: entry.count } : toast))
		);
	}

	function add(
		message: unknown,
		type: NotificationType = "info",
		options: NotificationOptions = {}
	): void {
		const recorded = recordInHistory(
			get(history),
			{ message: messageText(message), type, source: options.source, action: options.action },
			nextId++,
			now()
		);
		setHistory(recorded.history);
		if (recorded.merged) updateToastCount(recorded.entry);
		else if (!options.silent) startToast(recorded.entry);
	}

	function pause(id: number): void {
		const timer = timers.get(id);
		if (!timer?.handle) return;
		clearTimeout(timer.handle);
		timer.handle = undefined;
		timer.remaining -= now() - timer.startedAt;
	}

	function resume(id: number): void {
		const timer = timers.get(id);
		if (!timer || timer.handle) return;
		timer.startedAt = now();
		timer.handle = setTimeout(() => dismiss(id), timer.remaining);
	}

	function withNotification<T>(
		promise: Promise<T>,
		options: { pending?: string; success?: string; error?: string }
	): Promise<T> {
		if (options.pending) store.info(options.pending);
		return promise.then(
			(result) => {
				if (options.success) store.success(options.success);
				return result;
			},
			(err: unknown) => {
				store.error(options.error || messageText(err));
				throw err;
			}
		);
	}

	const store: NotificationStore = {
		toasts: { subscribe: toasts.subscribe },
		history: { subscribe: history.subscribe },
		unreadCount,
		add,
		info: (message, options) => add(message, "info", options),
		success: (message, options) => add(message, "success", options),
		warning: (message, options) => add(message, "warning", options),
		error: (message, options) => add(message, "error", options),
		dismiss,
		remove(id) {
			dismiss(id);
			updateHistory((entries) => entries.filter((entry) => entry.id !== id));
		},
		pause,
		resume,
		markRead(id) {
			updateHistory((entries) =>
				entries.map((entry) => (entry.id === id ? { ...entry, read: true } : entry))
			);
		},
		markAllRead() {
			updateHistory((entries) => entries.map((entry) => ({ ...entry, read: true })));
		},
		clearHistory() {
			setHistory([]);
		},
		withNotification
	};

	return store;
}

/** localStorage when the page has it, otherwise null. */
function browserStorage(): KeyValueStorage | null {
	if (typeof window === "undefined") return null;
	try {
		return window.localStorage ?? null;
	} catch {
		return null;
	}
}

export const notifications = createNotificationStore({
	now: () => Date.now(),
	storage: browserStorage()
});
