import {
	GetPrefixBaseDir,
	GetListGpus,
	DetectLosslessDll,
	ScanProtonVersions,
	GetInitialLauncherPath,
	GetExeIcon,
	ListPrefixes,
	GetSystemToolsStatus,
	RunGame,
} from "@bindings/light-launcher/internal/app/app";
import * as core from "@bindings/light-launcher/internal/types/models";
import { notifications } from "@stores/notificationStore";
import { runState } from "@stores/runState";
import { get } from "svelte/store";
import { mergeOptions } from "./formService";
import { loadConfigForGame } from "./runConfig";
import { Window } from "@wailsio/runtime";

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
	if (currentOptions.RunnerPath) {
		await loadConfigForGame(
			currentOptions.RunnerPath, 
			currentOptions, 
			prefixPath, 
			baseDirectory, 
			selectedPrefixName, 
			protonVersions, 
			onConfigUpdate
		);
		if (!launcherIcon) {
			launcherIcon = (await GetExeIcon(currentOptions.RunnerPath)) || "";
		}
	}

	// Handle initial path passed from OS (e.g., tray or file open)
	const initialOsPath = await GetInitialLauncherPath();
	if (initialOsPath) {
		if (!currentOptions.RunnerPath && !currentOptions.GamePath) {
			currentOptions.RunnerPath = initialOsPath;
			launcherIcon = (await GetExeIcon(initialOsPath)) || "";
			if (!currentOptions.Name || currentOptions.Name === "Launcher") {
				currentOptions.Name = initialOsPath.split(/[/\\]/).pop()?.replace(/\.exe$/i, "") || "Launcher";
			}
		} else if (!currentOptions.GamePath || currentOptions.GamePath === currentOptions.RunnerPath) {
			mainExecutablePath = initialOsPath;
			currentOptions.GamePath = initialOsPath;
			gameIcon = (await GetExeIcon(initialOsPath)) || "";
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
	if (currentOptions.RunnerPath && !launcherIcon) {
		launcherIcon = (await GetExeIcon(currentOptions.RunnerPath)) || "";
	}
	if (currentOptions.GamePath && !gameIcon) {
		gameIcon = (await GetExeIcon(currentOptions.GamePath)) || "";
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
	showLogsWindow: boolean
): Promise<boolean> {
	if (!launchOptions.RunnerPath) {
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

	await executeLaunch(launchOptions, selectedProtonName, protonVersions, showLogsWindow);
	return false;
}

/**
 * Directly executes the game launch without further validation.
 */
export async function executeLaunch(
	launchOptions: core.LaunchOptions,
	selectedProtonName: string,
	protonVersions: core.ProtonTool[],
	showLogsWindow: boolean
): Promise<void> {
	const matchedProton = protonVersions.find((tool) => tool.DisplayName === selectedProtonName);
	launchOptions.ProtonPath = matchedProton ? matchedProton.Path : (selectedProtonName.includes("/") ? selectedProtonName : "");

	try {
		await RunGame(launchOptions, showLogsWindow);
		Window.Close();
	} catch (error) {
		console.error("[EXECUTE] Launch failed:", error);
		notifications.add(`Launch failed: ${error}`, "error");
	}
}
