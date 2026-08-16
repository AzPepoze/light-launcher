import { contextBridge, ipcRenderer } from "electron";

const electronAPI = {
	invoke: async <T = any>(method: string, payload: any = {}): Promise<T> => {
		const response = await ipcRenderer.invoke("api", { method, payload });
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
