import {
	PickFile,
	PickFolder,
	SaveGameConfig,
} from "@lib/api";
import * as core from "@shared";
import { notifications } from "@stores/notificationStore";
import { runState } from "@stores/runState";
import { createLaunchOptions } from "@lib/formService";
import * as service from "@lib/runService";
import { loadConfigForGame, loadConfigForPrefix } from "@lib/runConfig";
import { ProtonState } from "./ProtonState.svelte";

export class RunPageState {
	mounted = $state(false);

	// Game Selection
	mainExePath = $state("");
	gameIcon = $state("");
	launcherIcon = $state("");
	prefixPath = $state("");
	baseDir = $state("");
	selectedPrefixName = $state("Default");

	// Prefix & Utilities
	availablePrefixes = $state<string[]>([]);

	// Sub-states
	proton = new ProtonState();

	// Game exe toggle state
	useGamePath = $state(false);

	// UI State
	showLogsWindow = $state(false);
	showValidationModal = $state(false);
	missingToolsList = $state<string[]>([]);
	systemStatus = $state<core.SystemToolsStatus>({
		hasGamescope: false,
		hasMangoHud: false,
		hasGameMode: false,
		hasVulkanInfo: false,
	});

	options = $state<core.LaunchOptions>(createLaunchOptions());
	gpuList = $state<string[]>([]);
	isSaving = $state(false);

	constructor() {
		// Sync useGamePath and options.UseGamePath
		$effect(() => {
			if (this.options) this.useGamePath = this.options.UseGamePath;
		});
		$effect(() => {
			if (this.options) this.options.UseGamePath = this.useGamePath;
		});

		// Sync options.PrefixPath and options.ProtonPath, and runState
		$effect(() => {
			if (this.mounted) {
				this.options.PrefixPath = this.prefixPath;
				this.options.ProtonPath = this.proton.selectedProton;
				runState.set({
					mainExePath: this.mainExePath,
					gameIcon: this.gameIcon,
					launcherIcon: this.launcherIcon,
					prefixPath: this.prefixPath,
					selectedPrefixName: this.selectedPrefixName,
					selectedProton: this.proton.selectedProton,
					options: this.options,
				});
			}
		});

		// Update prefix default proton
		$effect(() => {
			if (this.mounted && this.selectedPrefixName && this.proton.protonVersions.length > 0) {
				this.proton.updatePrefixDefaultProton(this.selectedPrefixName);
			}
		});

		// Fallback to prefix default proton
		$effect(() => {
			if (this.mounted && !this.options.UseCustomProton && this.proton.prefixDefaultProton) {
				this.proton.selectedProton = this.proton.prefixDefaultProton;
			}
		});
	}

	async initialize() {
		try {
			const data = await service.initializeRunPage(this.options, this.handleConfigUpdate.bind(this));
			
			this.baseDir = data.baseDir;
			this.gpuList = data.gpuList;
			this.proton.protonVersions = data.protonVersions;
			this.proton.protonOptions = data.protonOptions;
			this.availablePrefixes = data.availablePrefixes;
			this.systemStatus = data.systemStatus;
			this.launcherIcon = data.launcherIcon;
			this.gameIcon = data.gameIcon;
			this.mainExePath = data.mainExePath;

			if (this.proton.protonOptions.length > 0 && !this.proton.selectedProton) {
				this.proton.selectedProton = this.proton.protonOptions[0];
			}

			if (!this.prefixPath) {
				this.prefixPath = this.baseDir + "/Default";
				this.selectedPrefixName = "Default";
				await this.doLoadConfigForPrefix("Default");
			}
		} catch (err) {
			console.error("Failed to initialize:", err);
		} finally {
			this.proton.isLoadingProton = false;
			this.mounted = true;
		}
	}

	handleConfigUpdate(newOpts: core.LaunchOptions, pPath: string, pName: string, proton: string) {
		this.options = { ...newOpts };
		if (pPath) this.prefixPath = pPath;
		if (pName) this.selectedPrefixName = pName;
		if (proton) {
			this.proton.selectedProton = proton;
			if (proton && !this.proton.protonOptions.includes(proton)) {
				this.proton.protonOptions = [...this.proton.protonOptions, proton];
			}
		}
	}

	async handleSave() {
		this.isSaving = true;
		try {
			await SaveGameConfig(this.options);
			notifications.add("Configuration saved!", "success");
		} catch (err) {
			notifications.add(`Failed to save: ${err}`, "error");
		} finally {
			this.isSaving = false;
		}
	}

	async doLoadConfigForGame(path: string) {
		await loadConfigForGame(path, this.options, this.prefixPath, this.baseDir, this.selectedPrefixName, this.proton.protonVersions, this.handleConfigUpdate.bind(this));
	}

	async doLoadConfigForPrefix(name: string) {
		await loadConfigForPrefix(name, this.options, this.prefixPath, this.baseDir, this.proton.protonVersions, this.handleConfigUpdate.bind(this));
	}

	async handlePrefixChange(name: string) {
		if (name !== "Custom...") {
			this.prefixPath = this.baseDir + "/" + name;
			this.selectedPrefixName = name;
			await this.doLoadConfigForPrefix(name);
		}
	}

	async handleBrowseGame() {
		try {
			const path = await PickFile();
			if (path) {
				this.mainExePath = path;
				this.options = { ...this.options, GamePath: path };
			}
		} catch (err) {
			console.error("[GAME] Error loading game:", err);
		}
	}

	async handleBrowseLauncher() {
		try {
			const path = await PickFile();
			if (path) {
				this.options = { ...this.options, LauncherPath: path };
				if (!this.options.Name || this.options.Name === "Launcher") {
					this.options.Name = path.split(/[/\\]/).pop()?.replace(/\.exe$/i, "") || "Launcher";
				}
				if (!this.mainExePath) {
					this.options = { ...this.options, GamePath: path };
				}
				await this.doLoadConfigForGame(path);
			}
		} catch (err) {
			console.error("[LAUNCHER] Error selecting launcher:", err);
		}
	}

	async handleBrowsePrefix() {
		try {
			const path = await PickFolder();
			if (path) {
				this.prefixPath = path;
				this.selectedPrefixName = path.split('/').filter(Boolean).pop() || "Custom";
			}
		} catch (err) {
			console.error(err);
		}
	}

	closeLauncherOnConfirm = true;

	async handleLaunch(closeLauncher = true) {
		const shouldShowModal = await service.validateAndLaunch(
			this.options, 
			this.systemStatus, 
			this.proton.selectedProton, 
			this.proton.protonVersions, 
			this.showLogsWindow,
			closeLauncher
		);
		
		if (shouldShowModal === true) {
			this.closeLauncherOnConfirm = closeLauncher;
			this.missingToolsList = [];
			if (this.options.Extras.Gamescope.Enabled && !this.systemStatus.hasGamescope) this.missingToolsList.push("Gamescope");
			if (this.options.Extras.EnableMangoHud && !this.systemStatus.hasMangoHud) this.missingToolsList.push("MangoHud");
			if (this.options.Extras.EnableGamemode && !this.systemStatus.hasGameMode) this.missingToolsList.push("GameMode");
			if (this.options.Extras.Lsfg.Enabled && !this.systemStatus.hasVulkanInfo) this.missingToolsList.push("Vulkan-Tools");
			this.showValidationModal = true;
		}
	}

	async proceedToLaunch() {
		this.showValidationModal = false;
		await service.executeLaunch(this.options, this.proton.selectedProton, this.proton.protonVersions, this.showLogsWindow, this.closeLauncherOnConfirm);
	}
}
