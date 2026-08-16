import { ipcMain } from "electron";
import { IpcRouter } from "./router";

export function registerIpcHandlers(): void {
	ipcMain.handle("api", async (event, { method, payload }: { method: string; payload?: any }) => {
		try {
			const result = await IpcRouter.dispatch(method, payload || {}, event.sender);
			return { success: true, data: result };
		} catch (error: any) {
			console.error(`[IPC Error] method=${method}:`, error);
			return { success: false, error: error?.message || String(error) };
		}
	});

	ipcMain.on("window-action", (_, action: string) => {
		switch (action) {
			case "close":
				break;
		}
	});
}
