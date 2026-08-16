import {
	GetPrefixBaseDir,
	GetListGpus,
	DetectLosslessDll,
	ScanProtonVersions,
	GetInitialLauncherPath,
	ListPrefixes,
	GetSystemToolsStatus,
	RunGame,
	CloseWindow,
} from "@lib/api";
import { loadExeIcon } from "@lib/iconService";
import * as core from "@shared";
import { notifications } from "@stores/notificationStore";
import { runState } from "@stores/runState";
import { get } from "svelte/store";
import { mergeOptions } from "./formService";
import { loadConfigForGame } from "./runConfig";

export interface RunPageInitData {
	baseDir: string;
	gpuList: string[];
	protonVersions: core.ProtonTool[];
	protonOptions: string[];
	availablePrefixes: string[];
	systemStatus: core.SystemToolsStatus;
	launcherIcon: string;
	gameIcon: string;
	mainExePath: string;
}

export type ConfigUpdateCallback = (newOpts: core.LaunchOptions, pPath: string, pName: string, proton: string) => void;

/**
 * Orchestrates the full initialization of the Launch Configuration page.
 * Fetches hardware info, system tools, and merges any existing run state.
 */
export async function initializeRunPage(
	currentOptions: core.LaunchOptions,
	onConfigUpdate: ConfigUpdateCallback
): Promise<RunPageInitData> {
	const baseDirectory = await GetPrefixBaseDir();
	
	const [availableGpus, detectedLosslessDll, detectedProtons] = await Promise.all([
		GetListGpus(),
		DetectLosslessDll(),
		ScanProtonVersions(),
	]);

	const gpuList = availableGpus || [];
	if (detectedLosslessDll && !currentOptions.Extras.Lsfg.DllPath) {
		currentOptions.Extras.Lsfg.DllPath = detectedLosslessDll;
	}

	const protonVersions = detectedProtons || [];
	const protonOptions = protonVersions.map((tool) => tool.DisplayName);

	let launcherIcon = "";
	let gameIcon = "";
	let mainExecutablePath = "";
	let prefixPath = "";
	let selectedPrefixName = "Default";

	// Restore state from global runState store
	const activeRunState = get(runState);
	if (activeRunState) {
		if (activeRunState.mainExePath) mainExecutablePath = activeRunState.mainExePath;
		if (activeRunState.gameIcon) gameIcon = activeRunState.gameIcon;
		if (activeRunState.launcherIcon) launcherIcon = activeRunState.launcherIcon;
		if (activeRunState.prefixPath) prefixPath = activeRunState.prefixPath;
		if (activeRunState.selectedPrefixName) selectedPrefixName = activeRunState.selectedPrefixName;
		if (activeRunState.options) {
			Object.assign(currentOptions, mergeOptions(currentOptions, activeRunState.options));
		}
	}

	// Auto-load config if a runner is already set
	if (currentOptions.LauncherPath) {
		await loadConfigForGame(
			currentOptions.LauncherPath, 
			currentOptions, 
			prefixPath, 
			baseDirectory, 
			selectedPrefixName, 
			protonVersions, 
			onConfigUpdate
		);
		if (!launcherIcon) {
			launcherIcon = (await loadExeIcon(currentOptions.LauncherPath)) || "";
		}
	}

	// Handle initial path passed from OS (e.g., tray or file open)
	const initialOsPath = await GetInitialLauncherPath();
	if (initialOsPath) {
		if (!currentOptions.LauncherPath && !currentOptions.GamePath) {
			currentOptions.LauncherPath = initialOsPath;
			currentOptions.GamePath = initialOsPath; // Sync GamePath
			launcherIcon = (await loadExeIcon(initialOsPath)) || "";
			if (!currentOptions.Name || currentOptions.Name === "Launcher") {
				currentOptions.Name = initialOsPath.split(/[/\\]/).pop()?.replace(/\.exe$/i, "") || "Launcher";
			}
		} else if (!currentOptions.GamePath || currentOptions.GamePath === currentOptions.LauncherPath) {
			mainExecutablePath = initialOsPath;
			currentOptions.GamePath = initialOsPath;
			gameIcon = (await loadExeIcon(initialOsPath)) || "";
			await loadConfigForGame(
				initialOsPath, 
				currentOptions, 
				prefixPath, 
				baseDirectory, 
				selectedPrefixName, 
				protonVersions, 
				onConfigUpdate
			);
		}
	}

	// Ensure icons are loaded for existing paths
	if (currentOptions.LauncherPath && !launcherIcon) {
		launcherIcon = (await loadExeIcon(currentOptions.LauncherPath)) || "";
	}
	if (currentOptions.GamePath && !gameIcon) {
		gameIcon = (await loadExeIcon(currentOptions.GamePath)) || "";
	}

	const [availablePrefixesList, baseDirAgain, currentSystemStatus] = await Promise.all([
		ListPrefixes(),
		GetPrefixBaseDir(),
		GetSystemToolsStatus(),
	]);

	return {
		baseDir: baseDirAgain,
		gpuList,
		protonVersions,
		protonOptions,
		availablePrefixes: Array.isArray(availablePrefixesList) ? availablePrefixesList : ["Default"],
		systemStatus: currentSystemStatus,
		launcherIcon,
		gameIcon,
		mainExePath: mainExecutablePath,
	};
}

/**
 * Validates dependencies and environment before launching the game.
 * Returns true if the validation modal should be shown to the user.
 */
export async function validateAndLaunch(
	launchOptions: core.LaunchOptions,
	systemStatus: core.SystemToolsStatus,
	selectedProtonName: string,
	protonVersions: core.ProtonTool[],
	showLogsWindow: boolean,
	closeLauncher = true
): Promise<boolean> {
	if (!launchOptions.LauncherPath) {
		notifications.add("Please select a launcher executable.", "error");
		return false;
	}

	if (launchOptions.Extras.Lsfg.Enabled && !launchOptions.Extras.Lsfg.DllPath) {
		notifications.add("LSFG-VK requires Lossless.dll.", "error");
		return false;
	}

	// Check for missing system tools
	const missingTools: string[] = [];
	if (launchOptions.Extras.Gamescope.Enabled && !systemStatus.hasGamescope) missingTools.push("Gamescope");
	if (launchOptions.Extras.EnableMangoHud && !systemStatus.hasMangoHud) missingTools.push("MangoHud");
	if (launchOptions.Extras.EnableGamemode && !systemStatus.hasGameMode) missingTools.push("GameMode");
	if (launchOptions.Extras.Lsfg.Enabled && !systemStatus.hasVulkanInfo) missingTools.push("Vulkan-Tools");

	if (missingTools.length > 0) {
		return true; // Show modal
	}

	await executeLaunch(launchOptions, selectedProtonName, protonVersions, showLogsWindow, closeLauncher);
	return false;
}

/**
 * Directly executes the game launch without further validation.
 */
export async function executeLaunch(
	launchOptions: core.LaunchOptions,
	selectedProtonName: string,
	protonVersions: core.ProtonTool[],
	showLogsWindow: boolean,
	closeLauncher = true
): Promise<void> {
	launchOptions.ProtonPath = selectedProtonName;

	try {
		await RunGame(launchOptions, showLogsWindow);
		if (closeLauncher) {
			CloseWindow();
		}
	} catch (error) {
		console.error("[EXECUTE] Launch failed:", error);
		const message = (error as any)?.message || String(error);
		notifications.add(`Launch failed: ${message}`, "error");
	}
}
