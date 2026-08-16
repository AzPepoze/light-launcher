import type { WebContents } from "electron";
import { AppService } from "../services/app.service";
import { ConfigService } from "../services/config.service";
import { GamesService } from "../services/games.service";
import { LsfgService } from "../services/lsfg.service";
import { PrefixService } from "../services/prefix.service";
import { ProtonService } from "../services/proton.service";
import { RunnerService } from "../services/runner.service";
import { SessionService } from "../services/session.service";
import { SystemService } from "../services/system.service";

export class IpcRouter {
	static async dispatch(method: string, payload: any, sender: WebContents): Promise<any> {
		switch (method) {
			// App Service
			case "GetInitialLauncherPath":
				return AppService.getInitialLauncherPath();
			case "GetInitialGamePath":
				return AppService.getInitialGamePath();
			case "GetShouldEditLsfg":
				return AppService.getShouldEditLsfg();
			case "IsDir":
				return AppService.isDir(payload.path);
			case "GetExeIcon":
				return AppService.getExeIcon(payload.executablePath);
			case "GetTotalRam":
				return AppService.getTotalRam();
			case "GetImageBase64":
				return AppService.getImageBase64(payload.imagePath);
			case "PickFile":
				return AppService.pickFile();
			case "PickFolder":
				return AppService.pickFolder();
			case "PickFileCustom":
				return AppService.pickFileCustom(payload.title, payload.filters);
			case "OpenURL":
				return AppService.openExternal(payload.url);
			case "CloseWindow":
				return AppService.closeWindow();
			case "RestartApp":
				return AppService.restartApp();

			// Config & Settings
			case "GetAppSettings":
				return ConfigService.loadAppSettings();
			case "SaveAppSettings":
				return ConfigService.saveAppSettings(payload.settings);
			case "GetConfig":
				return ConfigService.loadGameConfig(payload.executablePath);
			case "SaveGameConfig":
				return ConfigService.saveGameConfig(payload.options);

			// Games Service
			case "GetAllGames":
				return GamesService.getAllGames();
			case "RemoveGame":
				return GamesService.removeGame(payload.executablePath);
			case "SearchExecutables":
				return GamesService.searchExecutables(
					payload.folderPath,
					payload.maxDepth,
					payload.excludeNames
				);
			case "GetAutoScannedGames":
				return GamesService.getAutoScannedGames();
			case "AddScanFolder":
				return GamesService.addScanFolder(payload.folderPath);
			case "RemoveScanFolder":
				return GamesService.removeScanFolder(payload.folderPath);
			case "UpdateScanFolderConfig":
				return GamesService.updateScanFolderConfig(
					payload.folderPath,
					payload.depth,
					payload.excludeNames
				);
			case "GetScanFolderConfig":
				return GamesService.getScanFolderConfig(payload.folderPath);
			case "BlacklistGame":
				return GamesService.blacklistGame(payload.executablePath);
			case "UnblacklistGame":
				return GamesService.unblacklistGame(payload.executablePath);

			// Runner Service
			case "RunGame":
				return RunnerService.runGame(payload.options, payload.showLogs);

			// System Service
			case "GetSystemToolsStatus":
				return SystemService.getSystemToolsStatus();
			case "GetSystemInfo":
				return SystemService.getSystemInfo();
			case "GetSystemUsage":
				return SystemService.getSystemUsage();
			case "GetShaderCacheSize":
				return SystemService.getShaderCacheSize();
			case "ClearShaderCache":
				return SystemService.clearShaderCache();
			case "DropCaches":
				return SystemService.dropCaches();
			case "ClearSwap":
				return SystemService.clearSwap();
			case "CleanupProcesses":
				return SystemService.cleanupProcesses();
			case "GetListGpus":
				return SystemService.getListGpus();

			// Prefix Service
			case "ListPrefixes":
				return PrefixService.listPrefixes();
			case "CreatePrefix":
				return PrefixService.createPrefix(payload.name);
			case "GetPrefixBaseDir":
				return PrefixService.getPrefixBaseDir();
			case "RemovePrefix":
				return PrefixService.removePrefix(payload.name);
			case "SavePrefixConfig":
				return PrefixService.savePrefixConfig(payload.prefixName, payload.options);
			case "LoadPrefixConfig":
				return PrefixService.loadPrefixConfig(payload.prefixName);
			case "LoadPrefixConfigWithProton":
				return PrefixService.loadPrefixConfigWithProton(payload.prefixName);
			case "RunPrefixTool":
				return PrefixService.runPrefixTool(
					payload.prefixPath,
					payload.toolName,
					payload.protonPath
				);

			// Proton Service
			case "ScanProtonVersions":
				return ProtonService.scanProtonVersions();
			case "GetProtonVariants":
				return ProtonService.getProtonVariants();
			case "GetProtonReleases":
				return ProtonService.getProtonReleases(payload.variantID);
			case "InstallProtonVersion":
				return ProtonService.installProtonVersion(
					payload.url,
					payload.version,
					(percent, message) => {
						sender.send("event:install-proton-progress", { percent, message });
					}
				);

			// LSFG Service
			case "GetUtilsStatus":
				return LsfgService.getUtilsStatus();
			case "GetLsfgProfileForGame":
				return LsfgService.getLsfgProfileForGame(payload.name, payload.gamePath);
			case "DetectLosslessDll":
				return LsfgService.detectLosslessDll();
			case "SaveLsfgProfile":
				return LsfgService.saveLsfgProfile(
					payload.profileName,
					payload.gamePath,
					payload.multiplier,
					payload.performanceMode,
					payload.dllPath,
					payload.gpu,
					payload.flowScale,
					payload.pacing,
					payload.allowFp16
				);
			case "DisableLsfgProfile":
				return LsfgService.disableLsfgProfile(payload.profileName, payload.gamePath);
			case "RemoveProfile":
				return LsfgService.removeProfile(payload.mainExecutablePath);
			case "InstallLsfg":
				return LsfgService.installLsfg((percent, message) => {
					sender.send("event:lsfg-install-progress", { percent, message });
				});
			case "UninstallLsfg":
				return LsfgService.uninstallLsfg();

			// Session Service
			case "GetRunningSessions":
				return SessionService.getRunningSessions();
			case "KillSession":
				return SessionService.killSession(payload.pid);

			default:
				throw new Error(`Unknown IPC method: ${method}`);
		}
	}
}
