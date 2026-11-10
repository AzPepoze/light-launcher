import { rpc } from "./rpc";
import type { DownloadAction, DownloadItem, DownloadsListResponse } from "./types";

const ACTION_METHODS: Record<DownloadAction, string> = {
	pause: "DownloadPause",
	resume: "DownloadResume",
	cancel: "DownloadCancel",
	retry: "DownloadRetry",
	remove: "DownloadRemove"
};

export const downloadsApi = {
	list: () => rpc<DownloadsListResponse>("DownloadsList"),
	enqueue: (url: string, gameId: number | null) =>
		rpc<DownloadItem>("DownloadEnqueue", { url, gameId }),
	act: (action: DownloadAction, id: number) => rpc<void>(ACTION_METHODS[action], { id }),
	setSpeedLimit: (speedLimitKBps: number) => rpc<void>("DownloadSetSpeedLimit", { speedLimitKBps })
};
