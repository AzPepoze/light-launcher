import { contextBridge, ipcRenderer } from "electron";

function sanitizePayload<T>(payload: T): T {
	if (payload === undefined || payload === null) return payload;
	try {
		return JSON.parse(JSON.stringify(payload));
	} catch {
		return payload;
	}
}

const electronAPI = {
	invoke: async <T = any>(method: string, payload: any = {}): Promise<T> => {
		const clean = sanitizePayload(payload);
		const response = await ipcRenderer.invoke("api", { method, payload: clean });
		if (!response.success) {
			throw new Error(response.error || `IPC error calling ${method}`);
		}
		return response.data as T;
	},
	send: (channel: string, ...args: any[]): void => {
		ipcRenderer.send(channel, ...args);
	},
	on: (channel: string, callback: (...args: any[]) => void): (() => void) => {
		const listener = (_event: any, ...args: any[]) => callback(...args);
		ipcRenderer.on(channel, listener);
		return () => {
			ipcRenderer.removeListener(channel, listener);
		};
	}
};

contextBridge.exposeInMainWorld("electronAPI", electronAPI);

export type ElectronAPI = typeof electronAPI;
