import { writable } from "svelte/store";

export interface Notification {
	id: number;
	message: string;
	type: "info" | "error" | "success";
}

const { subscribe, update } = writable<Notification[]>([]);

const notificationStore = {
	subscribe,
	add: (message: any, type: "info" | "error" | "success" = "info") => {
		const msg = typeof message === "string" ? message : (message?.message || String(message));
		const id = Date.now();
		update((n) => [...n, { id, message: msg, type }]);
		setTimeout(() => {
			update((n) => n.filter((i) => i.id !== id));
		}, 5000);
	},
	remove: (id: number) => {
		update((n) => n.filter((i) => i.id !== id));
	},
	error: (message: any) => notificationStore.add(message, "error"),
	success: (message: any) => notificationStore.add(message, "success"),
	info: (message: any) => notificationStore.add(message, "info"),
	warning: (message: string) => {
		notificationStore.error(message);
	},
	/**
	 * Wrapper for async operations with automatic notifications
	 */
	withNotification: async <T>(
		promise: Promise<T>,
		options: {
			pending?: string;
			success?: string;
			error?: string;
		},
	): Promise<T> => {
		if (options.pending) notificationStore.info(options.pending);

		try {
			const result = await promise;
			if (options.success) notificationStore.success(options.success);
			return result;
		} catch (err) {
			const message = (err as any)?.message || String(err);
			const errorMsg = options.error || message;
			notificationStore.error(errorMsg);
			throw err;
		}
	},
};

export const notifications = notificationStore;
