import { rpc } from "@lib/features/rpc";
import type { WireDownloadSettings } from "@lib/features/wire";
import { createLogger } from "@lib/logger";
import { notifications } from "@stores/notificationStore";
import { plainMessage } from "@lib/plainMessage";
import {
	DEFAULT_DOWNLOAD_SETTINGS,
	DOWNLOAD_SAVE_DELAY_MS,
	clampDownloadSettings,
	withDownloadNumber,
	type DownloadNumberKey
} from "./downloadsSettings";

const log = createLogger("DownloadSettings");

/** The download settings on the Settings page. Changes save after a short pause. */
export class DownloadSettingsState {
	settings = $state<WireDownloadSettings>({ ...DEFAULT_DOWNLOAD_SETTINGS });
	loaded = $state(false);
	loadFailed = $state(false);
	private saveTimer: ReturnType<typeof setTimeout> | undefined;

	async load(): Promise<void> {
		try {
			const saved = await rpc<WireDownloadSettings>("DownloadsGetSettings");
			this.settings = clampDownloadSettings(saved);
			this.loaded = true;
			this.loadFailed = false;
		} catch (err) {
			log.error("Failed to load download settings", err);
			this.loadFailed = true;
		}
	}

	setNumber(key: DownloadNumberKey, text: string): void {
		const next = withDownloadNumber(this.settings, key, text);
		if (next) this.change(next);
	}

	setKeepArchives(enabled: boolean): void {
		this.change({ ...this.settings, keepArchives: enabled });
	}

	/** Saves a pending change at once, for example when the page closes. */
	flush(): void {
		if (this.saveTimer === undefined) return;
		clearTimeout(this.saveTimer);
		this.saveTimer = undefined;
		void this.save();
	}

	private change(next: WireDownloadSettings): void {
		if (!this.loaded) return;
		this.settings = next;
		clearTimeout(this.saveTimer);
		this.saveTimer = setTimeout(() => {
			this.saveTimer = undefined;
			void this.save();
		}, DOWNLOAD_SAVE_DELAY_MS);
	}

	private async save(): Promise<void> {
		try {
			const saved = await rpc<WireDownloadSettings>("DownloadsSetSettings", { ...this.settings });
			// A newer change is still waiting to save, so keep what is on screen.
			if (this.saveTimer === undefined) this.settings = clampDownloadSettings(saved);
		} catch (err) {
			log.error("Failed to save download settings", err);
			notifications.error(`Failed to save download settings: ${plainMessage(err)}`);
		}
	}
}
