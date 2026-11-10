import { downloadsApi } from "../../lib/features/downloadsApi";
import { errorMessage, listen } from "../../lib/features/rpc";
import {
	DOWNLOAD_PROGRESS_EVENT,
	INSTALL_PROGRESS_EVENT,
	SEVEN_ZIP_MISSING_EVENT,
	type DownloadAction,
	type DownloadItem,
	type InstallProgress
} from "../../lib/features/types";

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
	sevenZipMissing = $state(false);
	speedLimitKBps = $state(0);
	loading = $state(true);
	error = $state("");

	async load() {
		this.loading = true;
		try {
			const response = await downloadsApi.list();
			this.items = response.items;
			this.installs = response.installs;
			this.sevenZipMissing = !response.sevenZip;
			this.speedLimitKBps = response.speedLimitKBps;
			this.error = "";
		} catch (err) {
			this.error = errorMessage(err);
		} finally {
			this.loading = false;
		}
	}

	/** Subscribes to live events. Returns a function that stops them. */
	subscribe(): () => void {
		const stops = [
			listen<DownloadItem>(DOWNLOAD_PROGRESS_EVENT, (item) => {
				this.items = upsert(this.items, item, (value) => value.id);
			}),
			listen<InstallProgress>(INSTALL_PROGRESS_EVENT, (install) => {
				this.installs = upsert(this.installs, install, (value) => value.downloadId);
			}),
			listen<unknown>(SEVEN_ZIP_MISSING_EVENT, () => {
				this.sevenZipMissing = true;
			})
		];
		return () => stops.forEach((stop) => stop());
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
