import { ipcMain } from "electron";
import { IpcRouter } from "./router";
import { LoggerService } from "../services/logger.service";

export function registerIpcHandlers(): void {
	ipcMain.handle("api", async (event, { method, payload }: { method: string; payload?: any }) => {
		try {
			LoggerService.debug("IPC", `dispatch -> ${method}`);
			const result = await IpcRouter.dispatch(method, payload || {}, event.sender);
			return { success: true, data: result };
		} catch (error: any) {
			LoggerService.error("IPC", `method=${method} failed:`, error?.message || error);
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
