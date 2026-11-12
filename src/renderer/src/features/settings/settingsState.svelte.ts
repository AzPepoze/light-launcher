import { get } from "svelte/store";
import { notifications } from "@stores/notificationStore";
import { settingsStore, type UserSettings } from "@stores/settingsStore";
import {
	folderInputError,
	gamesFoundText,
	gamesInFolder
} from "@lib/features/scanFolders";
import { sectionScroll } from "@lib/features/settingsSection";
import { createLogger } from "@lib/logger";
import type { AppSettings, ScanFolderConfig, ScannedFolderGroup } from "@shared";
import * as service from "@lib/settingsService";
import {
	DEFAULT_APP_SETTINGS,
	appSettingsWith,
	mergeAppSettings,
	scanFolderConfig,
	toggledTheme,
	withAddedScanFolder,
	withRemovedScanFolder,
	withScanFolderConfig
} from "./settingsData";

const log = createLogger("Settings");
const RESTART_DELAY_MS = 1500;

/** What the Settings page shows and the actions behind its controls. */
export class SettingsState {
	appSettings = $state<AppSettings>({ ...DEFAULT_APP_SETTINGS });
	userSettings = $state<UserSettings>(get(settingsStore));
	loading = $state(true);
	saving = $state(false);
	savingDiscordId = $state(false);
	discordClientId = $state("");

	groups = $state<ScannedFolderGroup[]>([]);
	rescanNotes = $state<Record<string, string>>({});
	rescanningFolder = $state<string | null>(null);
	scanError = $state("");
	pathInput = $state("");
	addError = $state("");
	adding = $state(false);

	watched = $derived(this.appSettings.ScanFolders ?? []);

	get canPickFolder(): boolean {
		return typeof window !== "undefined" && typeof window.llShell?.pickFolder === "function";
	}

	async start(): Promise<void> {
		await Promise.all([this.loadSettings(true), this.loadGroups(false)]);
	}

	/** Reloads the saved settings and the scan results after a nested editor saves. */
	async refresh(): Promise<void> {
		await Promise.all([this.loadSettings(false), this.loadGroups(false)]);
	}

	subscribeUserSettings(): () => void {
		return settingsStore.subscribe((value) => {
			this.userSettings = value;
		});
	}

	async toggleTransparentMode(): Promise<void> {
		const next = appSettingsWith(this.appSettings, {
			TransparentMode: !this.appSettings.TransparentMode
		});
		await this.saveAndRestart(next, "Restarting app to apply transparency changes...");
	}

	async toggleNativeWayland(): Promise<void> {
		const next = appSettingsWith(this.appSettings, {
			NativeWayland: !this.appSettings.NativeWayland
		});
		await this.saveAndRestart(next, "Restarting app to apply display platform changes...");
	}

	toggleTheme(): void {
		this.applyUserSettings({ ...this.userSettings, theme: toggledTheme(this.userSettings.theme) });
	}

	setTransparency(value: number): void {
		this.applyUserSettings({ ...this.userSettings, transparency: value });
	}

	async browseBackgroundImage(): Promise<void> {
		try {
			const path = await service.pickBackgroundImage();
			if (!path) return;
			this.applyUserSettings({ ...this.userSettings, backgroundImagePath: path });
			notifications.success("Background image updated");
		} catch (err) {
			log.error("Failed to pick image", err);
			notifications.error("Failed to pick image");
		}
	}

	clearBackgroundImage(): void {
		this.applyUserSettings({ ...this.userSettings, backgroundImagePath: "" });
		notifications.info("Background image cleared");
	}

	async togglePlaytimeTracking(): Promise<void> {
		const enabled = this.appSettings.TrackPlaytime === false;
		const next = appSettingsWith(this.appSettings, { TrackPlaytime: enabled });
		try {
			await this.persist(next);
			notifications.success(enabled ? "Playtime tracking enabled" : "Playtime tracking disabled");
		} catch (err) {
			log.error("Failed to save playtime setting", err);
			notifications.error(`Failed to save settings: ${err}`);
		}
	}

	async toggleDiscordPresence(): Promise<void> {
		const enabled = this.appSettings.DiscordRichPresence === false;
		const next = appSettingsWith(this.appSettings, { DiscordRichPresence: enabled });
		try {
			await this.persist(next);
			notifications.success(
				enabled ? "Discord Rich Presence enabled" : "Discord Rich Presence disabled"
			);
		} catch (err) {
			log.error("Failed to save Discord presence setting", err);
			notifications.error(`Failed to save settings: ${err}`);
		}
	}

	setDiscordClientId(value: string): void {
		this.discordClientId = value;
	}

	async saveDiscordClientId(): Promise<void> {
		if (this.savingDiscordId) return;
		this.savingDiscordId = true;
		try {
			const clientId = this.discordClientId.trim();
			const next = appSettingsWith(this.appSettings, { DiscordClientId: clientId });
			await this.persist(next);
			this.discordClientId = clientId;
			notifications.success("Discord Application ID saved");
		} catch (err) {
			log.error("Failed to save Discord client id", err);
			notifications.error(`Failed to save settings: ${err}`);
		} finally {
			this.savingDiscordId = false;
		}
	}

	async browsePrefixDirectory(): Promise<void> {
		try {
			const folder = await service.pickPrefixDirectory();
			if (!folder) return;
			const next = appSettingsWith(this.appSettings, { CustomPrefixDir: folder });
			await this.persist(next);
			notifications.success("Updated Wine prefix storage directory");
		} catch (err) {
			log.error("Failed to select prefix directory", err);
			notifications.error(`Failed to select prefix directory: ${err}`);
		}
	}

	async resetPrefixDirectory(): Promise<void> {
		try {
			const next = appSettingsWith(this.appSettings, { CustomPrefixDir: "" });
			await this.persist(next);
			notifications.success(
				"Reset Wine prefix storage to default (~/.config/light-launcher/prefixes)"
			);
		} catch (err) {
			log.error("Failed to reset prefix directory", err);
			notifications.error(`Failed to reset prefix directory: ${err}`);
		}
	}

	async restoreGame(path: string): Promise<void> {
		try {
			await service.unblacklistGame(path);
			this.appSettings = appSettingsWith(this.appSettings, {
				Blacklist: (this.appSettings.Blacklist ?? []).filter((item) => item !== path)
			});
			notifications.success("Game restored to library");
			await this.loadSettings(false);
		} catch (err) {
			log.error("Failed to restore game", err);
			notifications.error(`Failed to restore game: ${err}`);
		}
	}

	configFor(folder: string): ScanFolderConfig {
		return scanFolderConfig(this.appSettings, folder);
	}

	async loadGroups(force: boolean): Promise<boolean> {
		try {
			this.groups = await service.loadScannedFolders(force);
			this.scanError = "";
			return true;
		} catch (err) {
			log.error("Failed to scan game folders", err);
			this.scanError = "Could not scan the game folders. Try again.";
			return false;
		}
	}

	async rescanFolder(folder: string): Promise<void> {
		this.rescanningFolder = folder;
		try {
			if (!(await this.loadGroups(true))) return;
			const count = gamesInFolder(this.groups, folder) ?? 0;
			this.rescanNotes = { ...this.rescanNotes, [folder]: gamesFoundText(count) };
		} finally {
			this.rescanningFolder = null;
		}
	}

	async addFolder(path: string): Promise<void> {
		const folder = path.trim();
		this.addError = folderInputError(folder, this.watched) ?? "";
		if (this.addError) return;

		this.adding = true;
		try {
			if (!(await service.isDirectory(folder))) {
				this.addError = "That folder does not exist.";
				return;
			}
			await service.addScanFolder(folder);
			this.appSettings = withAddedScanFolder(this.appSettings, folder);
			this.pathInput = "";
			notifications.success("Monitored folder added");
			await Promise.all([this.loadSettings(false), this.loadGroups(false)]);
		} catch (err) {
			log.error("Failed to add game folder", err);
			this.addError = "Could not add this folder. Try again.";
		} finally {
			this.adding = false;
		}
	}

	async pickAndAddFolder(): Promise<void> {
		try {
			const picked = await window.llShell?.pickFolder?.();
			if (picked) await this.addFolder(picked);
		} catch (err) {
			log.error("Folder picker failed", err);
			this.addError = "Could not open the folder picker. Type the path instead.";
		}
	}

	async removeFolder(folder: string): Promise<void> {
		try {
			await service.removeScanFolder(folder);
			this.appSettings = withRemovedScanFolder(this.appSettings, folder);
			notifications.info("Monitored folder removed");
			await Promise.all([this.loadSettings(false), this.loadGroups(false)]);
		} catch (err) {
			log.error("Failed to remove game folder", err);
			this.scanError = "Could not remove this folder. Try again.";
		}
	}

	async saveFolderConfig(folder: string, depth: number, excludeNames: string[]): Promise<void> {
		await service.updateScanFolderConfig(folder, depth, excludeNames);
		this.appSettings = withScanFolderConfig(this.appSettings, {
			Path: folder,
			Depth: depth,
			ExcludeNames: excludeNames
		});
		await Promise.all([this.loadSettings(false), this.loadGroups(false)]);
	}

	/** Scrolls to a pending settings section. Returns false when none is pending. */
	scrollToSection(section: string | null): boolean {
		const reducedMotion =
			typeof window.matchMedia === "function" &&
			window.matchMedia("(prefers-reduced-motion: reduce)").matches;
		const scroll = sectionScroll(section, reducedMotion);
		if (!scroll) return false;
		document
			.getElementById(scroll.id)
			?.scrollIntoView({ behavior: scroll.behavior, block: "start" });
		return true;
	}

	private async loadSettings(showLoading: boolean): Promise<void> {
		if (showLoading) this.loading = true;
		try {
			const loaded = await service.loadAppSettings();
			this.appSettings = mergeAppSettings(loaded, this.appSettings);
			this.discordClientId = this.appSettings.DiscordClientId ?? "";
		} finally {
			if (showLoading) this.loading = false;
		}
	}

	private async persist(next: AppSettings): Promise<void> {
		this.saving = true;
		try {
			await service.saveAppSettings(next);
			this.appSettings = next;
		} finally {
			this.saving = false;
		}
	}

	private async saveAndRestart(next: AppSettings, message: string): Promise<void> {
		this.saving = true;
		try {
			await service.saveAppSettings(next);
			notifications.info(message);
			setTimeout(() => {
				void service.restartApp().catch((err) => log.error("Failed to restart app", err));
			}, RESTART_DELAY_MS);
		} catch (err) {
			log.error("Failed to save setting", err);
			notifications.error("Failed to save setting");
		} finally {
			this.saving = false;
		}
	}

	private applyUserSettings(next: UserSettings): void {
		this.userSettings = next;
		settingsStore.set(next);
	}
}
