import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import crypto from "crypto";
import { PathsService } from "./paths.service";
import type { AppSettings, LaunchOptions } from "../../shared/types/config.types";
import {
	DefaultExcludeNames,
	DefaultHeight,
	DefaultMemoryValue,
	DefaultMultiplier,
	DefaultRefreshRate,
	DefaultWidth
} from "../../shared/constants";

let cachedSettings: AppSettings | null = null;

export class ConfigService {
	static generateId(): string {
		return crypto.randomBytes(8).toString("hex");
	}

	static async loadJson<T>(filePath: string): Promise<T> {
		const data = await fs.readFile(filePath, "utf-8");
		return JSON.parse(data) as T;
	}

	static async saveJson(filePath: string, value: unknown): Promise<void> {
		const dir = path.dirname(filePath);
		await fs.mkdir(dir, { recursive: true });
		const data = JSON.stringify(value, null, 2);
		await fs.writeFile(filePath, data, "utf-8");
	}

	static async loadAppSettings(): Promise<AppSettings> {
		if (cachedSettings) {
			return cachedSettings;
		}

		const settingsPath = PathsService.getAppSettingsPath();
		try {
			if (fsSync.existsSync(settingsPath)) {
				const settings = await this.loadJson<AppSettings>(settingsPath);
				if (!settings.ScanFolderConfigs) {
					settings.ScanFolderConfigs = [];
				}
				if (!settings.ScanFolders) {
					settings.ScanFolders = [];
				}
				if (!settings.Blacklist) {
					settings.Blacklist = [];
				}

				// Synchronize legacy string folders to configs
				let modified = false;
				for (const sf of settings.ScanFolders) {
					const cleaned = path.normalize(sf);
					const found = settings.ScanFolderConfigs.some(
						(cfg) => path.normalize(cfg.Path) === cleaned
					);
					if (!found) {
						settings.ScanFolderConfigs.push({
							Path: cleaned,
							Depth: 2,
							ExcludeNames: [...DefaultExcludeNames]
						});
						modified = true;
					}
				}

				if (modified) {
					await this.saveJson(settingsPath, settings);
				}

				cachedSettings = settings;
				return cachedSettings;
			}
		} catch (err) {
			console.error("Error loading app settings:", err);
		}

		cachedSettings = {
			TransparentMode: true,
			ScanFolders: [],
			ScanFolderConfigs: [],
			Blacklist: []
		};
		return cachedSettings;
	}

	static async saveAppSettings(settings: AppSettings): Promise<void> {
		const settingsPath = PathsService.getAppSettingsPath();
		await this.saveJson(settingsPath, settings);
		cachedSettings = settings;
	}

	static async listGameConfigs(): Promise<LaunchOptions[]> {
		const configDir = PathsService.getConfigDirectory();
		if (!fsSync.existsSync(configDir)) {
			return [];
		}

		try {
			const entries = await fs.readdir(configDir, { withFileTypes: true });
			const configs: LaunchOptions[] = [];

			for (const entry of entries) {
				if (!entry.isDirectory()) continue;
				const configPath = path.join(configDir, entry.name, "config.json");
				if (fsSync.existsSync(configPath)) {
					try {
						const options = await this.loadJson<LaunchOptions>(configPath);
						configs.push(options);
					} catch (e) {
						console.error(`Failed to parse ${configPath}:`, e);
					}
				}
			}
			return configs;
		} catch (err) {
			console.error("Error listing game configs:", err);
			return [];
		}
	}

	static async saveGameConfig(options: LaunchOptions): Promise<void> {
		if (!options.ID) {
			options.ID = this.generateId();
		}
		const configPath = PathsService.getGameConfigFilePath(options.Name, options.ID);
		await this.saveJson(configPath, options);
	}

	static async loadGameConfig(executablePath: string): Promise<LaunchOptions | null> {
		const configs = await this.listGameConfigs();
		const cleanPath = path.normalize(executablePath).toLowerCase();

		for (const cfg of configs) {
			const cfgGame = cfg.GamePath ? path.normalize(cfg.GamePath).toLowerCase() : "";
			const cfgLaunch = cfg.LauncherPath ? path.normalize(cfg.LauncherPath).toLowerCase() : "";
			if (cfgGame === cleanPath || cfgLaunch === cleanPath) {
				return cfg;
			}
		}
		return null;
	}

	static async loadPrefixConfig(prefixName: string): Promise<LaunchOptions> {
		const prefixConfigPath = PathsService.getPrefixConfigPath(prefixName);
		if (fsSync.existsSync(prefixConfigPath)) {
			try {
				return await this.loadJson<LaunchOptions>(prefixConfigPath);
			} catch (e) {
				console.error(`Failed to parse ${prefixConfigPath}:`, e);
			}
		}

		return {
			ID: "",
			Name: prefixName,
			LauncherPath: "",
			GamePath: "",
			UseGamePath: false,
			PrefixPath: path.join(PathsService.getPrefixBaseDirectory(), prefixName),
			ProtonPath: "",
			UseCustomProton: false,
			CustomArgs: "",
			Extras: {
				EnableMangoHud: false,
				EnableGamemode: false,
				Lsfg: {
					Enabled: false,
					Multiplier: DefaultMultiplier,
					PerfMode: false,
					DllPath: "",
					Gpu: "",
					FlowScale: "1.0",
					Pacing: "smooth",
					AllowFp16: false
				},
				Gamescope: {
					Enabled: false,
					Width: DefaultWidth,
					Height: DefaultHeight,
					OutputWidth: "",
					OutputHeight: "",
					RefreshRate: DefaultRefreshRate,
					FramerateLimit: "",
					WindowMode: "borderless",
					Scaler: "auto",
					Filter: "linear",
					Sharpness: "0",
					HDR: false,
					AdaptiveSync: false,
					Mangoapp: false,
					CustomArgs: ""
				},
				Memory: {
					Enabled: false,
					Value: DefaultMemoryValue
				}
			}
		};
	}

	static async savePrefixConfig(prefixName: string, options: LaunchOptions): Promise<void> {
		const prefixConfigPath = PathsService.getPrefixConfigPath(prefixName);
		await this.saveJson(prefixConfigPath, options);
	}
}
