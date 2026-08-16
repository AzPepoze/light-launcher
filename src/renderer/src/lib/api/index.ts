import type {
	AppSettings,
	LaunchOptions,
	ScanFolderConfig,
	GameInfo,
	ScannedFolderGroup,
	RunningSession,
	SystemInfo,
	SystemUsage,
	SystemToolsStatus,
	UtilsStatus,
	LsfgProfileData,
	ProtonTool,
	PrefixConfigWithProton,
	ProtonVariant,
	GitHubRelease
} from "../../../../shared/index";

declare global {
	interface Window {
		electronAPI: {
			invoke: <T = any>(method: string, payload?: any) => Promise<T>;
			send: (channel: string, ...args: any[]) => void;
			on: (channel: string, callback: (...args: any[]) => void) => () => void;
		};
	}
}

function sanitizePayload<T>(payload: T): T {
	if (payload === undefined || payload === null) return payload;
	try {
		return JSON.parse(JSON.stringify(payload));
	} catch {
		return payload;
	}
}

async function invoke<T>(method: string, payload: any = {}): Promise<T> {
	if (typeof window !== "undefined" && window.electronAPI) {
		return window.electronAPI.invoke<T>(method, sanitizePayload(payload));
	}
	throw new Error("electronAPI is not available on window");
}

export function onEvent(channel: string, callback: (...args: any[]) => void): () => void {
	if (typeof window !== "undefined" && window.electronAPI) {
		return window.electronAPI.on(`event:${channel}`, callback);
	}
	return () => {};
}

// App & Window Methods
export const GetInitialLauncherPath = () => invoke<string>("GetInitialLauncherPath");
export const GetInitialGamePath = () => invoke<string>("GetInitialGamePath");
export const GetShouldEditLsfg = () => invoke<boolean>("GetShouldEditLsfg");
export const IsDir = (path: string) => invoke<boolean>("IsDir", { path });
export const GetExeIcon = (executablePath: string) =>
	invoke<string>("GetExeIcon", { executablePath });
export const GetTotalRam = () => invoke<number>("GetTotalRam");
export const GetImageBase64 = (imagePath: string) =>
	invoke<string>("GetImageBase64", { imagePath });
export const PickFile = () => invoke<string>("PickFile");
export const PickFolder = () => invoke<string>("PickFolder");
export const PickFileCustom = (
	title: string,
	filters: { displayName: string; pattern: string }[]
) => invoke<string>("PickFileCustom", { title, filters });
export const OpenURL = (url: string) => invoke<void>("OpenURL", { url });
export const CloseWindow = () => invoke<void>("CloseWindow");
export const RestartApp = () => invoke<void>("RestartApp");

// Config & Settings
export const GetAppSettings = () => invoke<AppSettings>("GetAppSettings");
export const SaveAppSettings = (settings: AppSettings) =>
	invoke<void>("SaveAppSettings", { settings });
export const GetConfig = (executablePath: string) =>
	invoke<LaunchOptions | null>("GetConfig", { executablePath });
export const SaveGameConfig = (options: LaunchOptions) =>
	invoke<void>("SaveGameConfig", { options });

// Games Library
export const GetAllGames = () => invoke<GameInfo[]>("GetAllGames");
export const RemoveGame = (executablePath: string) =>
	invoke<void>("RemoveGame", { executablePath });
export const SearchExecutables = (folderPath: string, maxDepth: number, excludeNames: string[]) =>
	invoke<string[]>("SearchExecutables", { folderPath, maxDepth, excludeNames });
export const GetAutoScannedGames = () => invoke<ScannedFolderGroup[]>("GetAutoScannedGames");
export const AddScanFolder = (folderPath: string) => invoke<void>("AddScanFolder", { folderPath });
export const RemoveScanFolder = (folderPath: string) =>
	invoke<void>("RemoveScanFolder", { folderPath });
export const UpdateScanFolderConfig = (folderPath: string, depth: number, excludeNames: string[]) =>
	invoke<void>("UpdateScanFolderConfig", { folderPath, depth, excludeNames });
export const GetScanFolderConfig = (folderPath: string) =>
	invoke<ScanFolderConfig>("GetScanFolderConfig", { folderPath });
export const BlacklistGame = (executablePath: string) =>
	invoke<void>("BlacklistGame", { executablePath });
export const UnblacklistGame = (executablePath: string) =>
	invoke<void>("UnblacklistGame", { executablePath });

// Runner
export const RunGame = (options: LaunchOptions, showLogs: boolean = true) =>
	invoke<void>("RunGame", { options, showLogs });

// System Telemetry & Utilities
export const GetSystemToolsStatus = () => invoke<SystemToolsStatus>("GetSystemToolsStatus");
export const GetSystemInfo = () => invoke<SystemInfo>("GetSystemInfo");
export const GetSystemUsage = () => invoke<SystemUsage>("GetSystemUsage");
export const GetShaderCacheSize = () => invoke<string>("GetShaderCacheSize");
export const ClearShaderCache = () => invoke<void>("ClearShaderCache");
export const DropCaches = () => invoke<void>("DropCaches");
export const ClearSwap = () => invoke<void>("ClearSwap");
export const CleanupProcesses = () => invoke<void>("CleanupProcesses");
export const GetListGpus = () => invoke<string[]>("GetListGpus");

// Prefix Management
export const ListPrefixes = () => invoke<string[]>("ListPrefixes");
export const CreatePrefix = (name: string) => invoke<void>("CreatePrefix", { name });
export const GetPrefixBaseDir = () => invoke<string>("GetPrefixBaseDir");
export const RemovePrefix = (name: string) => invoke<void>("RemovePrefix", { name });
export const SavePrefixConfig = (prefixName: string, options: LaunchOptions) =>
	invoke<void>("SavePrefixConfig", { prefixName, options });
export const LoadPrefixConfig = (prefixName: string) =>
	invoke<LaunchOptions>("LoadPrefixConfig", { prefixName });
export const LoadPrefixConfigWithProton = (prefixName: string) =>
	invoke<PrefixConfigWithProton>("LoadPrefixConfigWithProton", { prefixName });
export const RunPrefixTool = (prefixPath: string, toolName: string, protonPath: string) =>
	invoke<void>("RunPrefixTool", { prefixPath, toolName, protonPath });

// Proton Runtimes
export const ScanProtonVersions = () => invoke<ProtonTool[]>("ScanProtonVersions");
export const GetProtonVariants = () => invoke<ProtonVariant[]>("GetProtonVariants");
export const GetProtonReleases = (variantID: string) =>
	invoke<GitHubRelease[]>("GetProtonReleases", { variantID });
export const InstallProtonVersion = (url: string, version: string) =>
	invoke<void>("InstallProtonVersion", { url, version });

// LSFG Tooling
export const GetUtilsStatus = () => invoke<UtilsStatus>("GetUtilsStatus");
export const GetLsfgProfileForGame = (name: string, gamePath: string) =>
	invoke<LsfgProfileData | null>("GetLsfgProfileForGame", { name, gamePath });
export const DetectLosslessDll = () => invoke<string>("DetectLosslessDll");
export const SaveLsfgProfile = (
	profileName: string,
	gamePath: string,
	multiplier: number,
	performanceMode: boolean,
	dllPath: string,
	gpu: string,
	flowScale: string,
	pacing: string,
	allowFp16: boolean
) =>
	invoke<void>("SaveLsfgProfile", {
		profileName,
		gamePath,
		multiplier,
		performanceMode,
		dllPath,
		gpu,
		flowScale,
		pacing,
		allowFp16
	});
export const DisableLsfgProfile = (profileName: string, gamePath: string) =>
	invoke<void>("DisableLsfgProfile", { profileName, gamePath });
export const RemoveProfile = (mainExecutablePath: string) =>
	invoke<void>("RemoveProfile", { mainExecutablePath });
export const InstallLsfg = () => invoke<void>("InstallLsfg");
export const UninstallLsfg = () => invoke<void>("UninstallLsfg");

// Sessions
export const GetRunningSessions = () => invoke<RunningSession[]>("GetRunningSessions");
export const KillSession = (pid: number) => invoke<void>("KillSession", { pid });
