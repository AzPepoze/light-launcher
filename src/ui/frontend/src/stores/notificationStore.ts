import { writable } from "svelte/store";

export interface Notification {
	id: number;
	message: string;
	type: "info" | "error" | "success";
}

const { subscribe, update } = writable<Notification[]>([]);

const notificationStore = {
	subscribe,
	add: (message: string, type: "info" | "error" | "success" = "info") => {
		const id = Date.now();
		update((n) => [...n, { id, message, type }]);
		setTimeout(() => {
			update((n) => n.filter((i) => i.id !== id));
		}, 5000);
	},
	remove: (id: number) => {
		update((n) => n.filter((i) => i.id !== id));
	},
	error: (message: string) => {
		const id = Date.now();
		update((n) => [...n, { id, message, type: "error" }]);
		setTimeout(() => {
			update((n) => n.filter((i) => i.id !== id));
		}, 5000);
	},
	success: (message: string) => {
		const id = Date.now();
		update((n) => [...n, { id, message, type: "success" }]);
		setTimeout(() => {
			update((n) => n.filter((i) => i.id !== id));
		}, 5000);
	},
	info: (message: string) => {
		const id = Date.now();
		update((n) => [...n, { id, message, type: "info" }]);
		setTimeout(() => {
			update((n) => n.filter((i) => i.id !== id));
		}, 5000);
	},
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
			const errorMsg = options.error || String(err);
			notificationStore.error(errorMsg);
			throw err;
		}
	},
};

export const notifications = notificationStore;
