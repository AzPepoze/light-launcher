import { applyProgress, hasDownload, toInstallProgress } from "../../lib/features/downloadsMap";
import { downloadsApi } from "../../lib/features/downloadsApi";
import {
	DOWNLOAD_PROGRESS_EVENT,
	INSTALL_PROGRESS_EVENT,
	TOAST_EVENT
} from "../../lib/features/events";
import { errorMessage, listen } from "../../lib/features/rpc";
import type {
	DownloadAction,
	DownloadItem,
	DownloadsListResponse,
	InstallProgress,
	Toast
} from "../../lib/features/types";
import type { WireDownloadProgress, WireInstallProgress } from "../../lib/features/wire";

/** Adds or replaces an item by id, keeping the list order. */
function upsert<T>(list: T[], item: T, idOf: (value: T) => number): T[] {
	const id = idOf(item);
	return list.some((existing) => idOf(existing) === id)
		? list.map((existing) => (idOf(existing) === id ? item : existing))
		: [...list, item];
}

export class DownloadsState {
	items = $state<DownloadItem[]>([]);
	installs = $state<InstallProgress[]>([]);
	/** The latest message from the Go side, such as the 7-Zip warning. */
	toast = $state<Toast | null>(null);
	speedLimitKBps = $state(0);
	loading = $state(true);
	error = $state("");

	async load() {
		this.loading = true;
		try {
			this.applyList(await downloadsApi.list());
			this.error = "";
		} catch (err) {
			this.error = errorMessage(err);
		} finally {
			this.loading = false;
		}
	}

	private applyList(response: DownloadsListResponse) {
		this.items = response.items;
		this.installs = response.installs;
		this.speedLimitKBps = response.speedLimitKBps;
	}

	/** Subscribes to live events. Returns a function that stops them. */
	subscribe(): () => void {
		const stops = [
			listen<WireDownloadProgress>(DOWNLOAD_PROGRESS_EVENT, (progress) =>
				this.onProgress(progress)
			),
			listen<WireInstallProgress>(INSTALL_PROGRESS_EVENT, (install) => this.onInstall(install)),
			listen<Toast>(TOAST_EVENT, (toast) => {
				this.toast = toast;
			})
		];
		return () => stops.forEach((stop) => stop());
	}

	/** A row that is not in the list yet is loaded again, so nothing is lost. */
	private onProgress(progress: WireDownloadProgress) {
		if (!hasDownload(this.items, progress.id)) {
			void this.refreshItems();
			return;
		}
		this.items = applyProgress(this.items, progress);
	}

	private onInstall(wire: WireInstallProgress) {
		const title = this.items.find((item) => item.id === wire.downloadId)?.filename ?? "";
		const install = toInstallProgress(wire, title);
		this.installs = upsert(this.installs, install, (value) => value.downloadId);
	}

	/** Reloads the rows only, so live install progress is kept. */
	private async refreshItems() {
		try {
			this.items = (await downloadsApi.list()).items;
		} catch (err) {
			this.error = errorMessage(err);
		}
	}

	async act(action: DownloadAction, id: number) {
		await this.run(() => downloadsApi.act(action, id));
	}

	async enqueue(url: string, gameId: number | null = null) {
		await this.run(async () => {
			this.items = upsert(this.items, await downloadsApi.enqueue(url, gameId), (v) => v.id);
		});
	}

	async setSpeedLimit(kbps: number) {
		await this.run(async () => {
			await downloadsApi.setSpeedLimit(kbps);
			this.speedLimitKBps = kbps;
		});
	}

	private async run(action: () => Promise<void>) {
		try {
			await action();
			this.error = "";
		} catch (err) {
			this.error = errorMessage(err);
		}
	}
}
