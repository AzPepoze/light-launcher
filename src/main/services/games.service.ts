import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import { ConfigService } from "./config.service";
import { PathsService } from "./paths.service";
import type { GameInfo, ScannedFolderGroup } from "../../shared/types/games.types";
import type { LaunchOptions, ScanFolderConfig } from "../../shared/types/config.types";
import {
	DefaultExcludeNames,
	DefaultHeight,
	DefaultMemoryValue,
	DefaultMultiplier,
	DefaultRefreshRate,
	DefaultWidth
} from "../../shared/constants";

export class GamesService {
	static async getAllGames(): Promise<GameInfo[]> {
		const configs = await ConfigService.listGameConfigs();
		const settings = await ConfigService.loadAppSettings();
		const scanFolders = settings.ScanFolders || [];

		const games: GameInfo[] = [];

		for (const gameConfig of configs) {
			let name = gameConfig.Name;
			if (!name) {
				const base = path.basename(gameConfig.GamePath || gameConfig.LauncherPath);
				name = path.parse(base).name;
			}

			let cleanedPath = path.normalize(gameConfig.LauncherPath || gameConfig.GamePath);
			if (!cleanedPath) continue;

			let inScanFolder = false;
			for (const sf of scanFolders) {
				const sfCleaned = path.normalize(sf);
				const rel = path.relative(sfCleaned, cleanedPath);
				if (!rel.startsWith("..") && !path.isAbsolute(rel) && rel !== "") {
					inScanFolder = true;
					break;
				}
			}

			games.push({
				name,
				path: cleanedPath,
				icon: "",
				config: gameConfig,
				isRecent: false,
				isAutoScanned: inScanFolder
			});
		}

		return games;
	}

	static async removeGame(executablePath: string): Promise<void> {
		const cfg = await ConfigService.loadGameConfig(executablePath);
		if (!cfg) {
			throw new Error(`Could not find game config to remove for path: ${executablePath}`);
		}

		const configDir = PathsService.getExecutableConfigPath(cfg.Name, cfg.ID);
		if (fsSync.existsSync(configDir)) {
			await fs.rm(configDir, { recursive: true, force: true });
		}
	}

	static async searchExecutables(
		folderPath: string,
		maxDepth: number,
		excludeNames: string[]
	): Promise<string[]> {
		const executables: string[] = [];
		const cleanFolder = path.normalize(folderPath);

		const excludeRegexes: RegExp[] = [];
		for (const pattern of excludeNames) {
			if (!pattern) continue;
			let regexStr = pattern.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*");
			try {
				excludeRegexes.push(new RegExp(`^${regexStr}$`, "i"));
			} catch (e) {
				// skip invalid regex
			}
		}

		async function scanDir(currentDir: string, currentDepth: number): Promise<void> {
			if (maxDepth !== -1 && currentDepth > maxDepth) {
				return;
			}

			let entries: fsSync.Dirent[];
			try {
				entries = await fs.readdir(currentDir, { withFileTypes: true });
			} catch {
				return;
			}

			for (const entry of entries) {
				const name = entry.name;
				const isExcluded = excludeRegexes.some((r) => r.test(name));
				if (isExcluded) continue;

				const fullPath = path.join(currentDir, name);
				if (entry.isDirectory()) {
					await scanDir(fullPath, currentDepth + 1);
				} else if (entry.isFile() && name.toLowerCase().endsWith(".exe")) {
					executables.push(fullPath);
				}
			}
		}

		await scanDir(cleanFolder, 0);
		return executables;
	}

	static async getAutoScannedGames(): Promise<ScannedFolderGroup[]> {
		const settings = await ConfigService.loadAppSettings();
		if (!settings || !settings.ScanFolderConfigs) {
			return [];
		}

		const manualGames = await this.getAllGames();
		const manualPaths = new Set(
			manualGames.filter((g) => !g.isAutoScanned).map((g) => path.normalize(g.path).toLowerCase())
		);

		const existingConfigsList = await ConfigService.listGameConfigs();
		const existingConfigsMap = new Map<string, LaunchOptions>();
		for (const cfg of existingConfigsList) {
			if (cfg.GamePath) existingConfigsMap.set(path.normalize(cfg.GamePath).toLowerCase(), cfg);
			if (cfg.LauncherPath)
				existingConfigsMap.set(path.normalize(cfg.LauncherPath).toLowerCase(), cfg);
		}

		const blacklist = new Set(
			(settings.Blacklist || []).map((p) => path.normalize(p).toLowerCase())
		);

		const groups: ScannedFolderGroup[] = [];

		for (const folderCfg of settings.ScanFolderConfigs) {
			const cleanedFolder = path.normalize(folderCfg.Path);
			const folderName = path.basename(cleanedFolder);

			const executables = await this.searchExecutables(
				cleanedFolder,
				folderCfg.Depth ?? 2,
				folderCfg.ExcludeNames || DefaultExcludeNames
			);

			const games: GameInfo[] = [];

			for (const exePath of executables) {
				const normExe = path.normalize(exePath).toLowerCase();
				if (blacklist.has(normExe) || manualPaths.has(normExe)) {
					continue;
				}

				const name = path.parse(path.basename(exePath)).name;
				let cfg = existingConfigsMap.get(normExe);

				if (!cfg) {
					const defaultPrefix = path.join(PathsService.getPrefixBaseDirectory(), "Default");
					cfg = {
						ID: ConfigService.generateId(),
						Name: name,
						LauncherPath: exePath,
						GamePath: exePath,
						UseGamePath: false,
						PrefixPath: defaultPrefix,
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
					await ConfigService.saveGameConfig(cfg);
					existingConfigsMap.set(normExe, cfg);
				}

				games.push({
					name,
					path: exePath,
					icon: "",
					config: cfg,
					isRecent: false,
					isAutoScanned: true
				});
			}

			groups.push({
				folderPath: cleanedFolder,
				folderName,
				games
			});
		}

		return groups;
	}

	static async addScanFolder(folderPath: string): Promise<void> {
		const settings = await ConfigService.loadAppSettings();
		const cleaned = path.normalize(folderPath);

		if (!settings.ScanFolders.includes(cleaned)) {
			settings.ScanFolders.push(cleaned);
		}

		const exists = settings.ScanFolderConfigs.some((cfg) => path.normalize(cfg.Path) === cleaned);
		if (!exists) {
			settings.ScanFolderConfigs.push({
				Path: cleaned,
				Depth: 2,
				ExcludeNames: [...DefaultExcludeNames]
			});
		}

		await ConfigService.saveAppSettings(settings);
	}

	static async removeScanFolder(folderPath: string): Promise<void> {
		const settings = await ConfigService.loadAppSettings();
		const cleaned = path.normalize(folderPath);

		settings.ScanFolders = settings.ScanFolders.filter((f) => path.normalize(f) !== cleaned);
		settings.ScanFolderConfigs = settings.ScanFolderConfigs.filter(
			(cfg) => path.normalize(cfg.Path) !== cleaned
		);

		await ConfigService.saveAppSettings(settings);
	}

	static async updateScanFolderConfig(
		folderPath: string,
		depth: number,
		excludeNames: string[]
	): Promise<void> {
		const settings = await ConfigService.loadAppSettings();
		const cleaned = path.normalize(folderPath);

		let found = false;
		for (const cfg of settings.ScanFolderConfigs) {
			if (path.normalize(cfg.Path) === cleaned) {
				cfg.Depth = depth;
				cfg.ExcludeNames = excludeNames;
				found = true;
				break;
			}
		}

		if (!found) {
			settings.ScanFolderConfigs.push({
				Path: cleaned,
				Depth: depth,
				ExcludeNames: excludeNames
			});
			if (!settings.ScanFolders.some((f) => path.normalize(f) === cleaned)) {
				settings.ScanFolders.push(cleaned);
			}
		}

		await ConfigService.saveAppSettings(settings);
	}

	static async getScanFolderConfig(folderPath: string): Promise<ScanFolderConfig> {
		const settings = await ConfigService.loadAppSettings();
		const cleaned = path.normalize(folderPath);

		const cfg = settings.ScanFolderConfigs.find((c) => path.normalize(c.Path) === cleaned);
		if (cfg) return cfg;

		return {
			Path: cleaned,
			Depth: 2,
			ExcludeNames: [...DefaultExcludeNames]
		};
	}

	static async blacklistGame(executablePath: string): Promise<void> {
		const settings = await ConfigService.loadAppSettings();
		const cleaned = path.normalize(executablePath);

		if (!settings.Blacklist.some((p) => path.normalize(p) === cleaned)) {
			settings.Blacklist.push(cleaned);
			await ConfigService.saveAppSettings(settings);
		}
	}

	static async unblacklistGame(executablePath: string): Promise<void> {
		const settings = await ConfigService.loadAppSettings();
		const cleaned = path.normalize(executablePath);

		settings.Blacklist = settings.Blacklist.filter((p) => path.normalize(p) !== cleaned);
		await ConfigService.saveAppSettings(settings);
	}
}
