import { toDownloadItem, toDownloadsList } from "./downloadsMap";
import { rpc } from "./rpc";
import type { DownloadAction, DownloadItem, DownloadsListResponse } from "./types";
import type { WireDownloadsStatus, WireDownloadView, WireQueuedDownload } from "./wire";

const ACTION_METHODS: Record<Exclude<DownloadAction, "remove">, string> = {
	pause: "DownloadPause",
	resume: "DownloadResume",
	cancel: "DownloadCancel",
	retry: "DownloadRetry"
};

export const downloadsApi = {
	list: async (): Promise<DownloadsListResponse> => {
		const [views, status] = await Promise.all([
			rpc<WireDownloadView[] | null>("DownloadsList"),
			rpc<WireDownloadsStatus>("DownloadsStatus")
		]);
		return toDownloadsList(views ?? [], status);
	},

	/** A game file is installed after it downloads. A direct link is only saved. */
	enqueue: async (url: string, gameId: number | null): Promise<DownloadItem> =>
		toDownloadItem(
			await rpc<WireDownloadView>("DownloadEnqueue", {
				url,
				gameId: gameId ?? 0,
				install: gameId !== null
			})
		),

	/**
	 * Queues a saved file of a game. The host names the file by id, so the link comes from the library.
	 * The game's plugin prepares it, which can wait for the person to finish a browser step.
	 */
	enqueueFile: async (gameId: number, fileId: number): Promise<number> => {
		const queued = await rpc<WireQueuedDownload>("LibraryDownloadFile", { gameId, fileId });
		return queued.id;
	},

	/** Remove keeps the file on disk. */
	act: (action: DownloadAction, id: number) =>
		action === "remove"
			? rpc<void>("DownloadRemove", { id, deleteFile: false })
			: rpc<void>(ACTION_METHODS[action], { id }),

	setSpeedLimit: (speedLimitKBps: number) => rpc<void>("DownloadSetSpeedLimit", { speedLimitKBps })
};
