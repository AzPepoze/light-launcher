import {
	PickFile,
	PickFolder,
	SaveGameConfig,
	LoadPrefixConfig,
} from "@bindings/light-launcher/internal/app/app";
import * as core from "@bindings/light-launcher/internal/types/models";
import { notifications } from "@stores/notificationStore";
import { runState } from "@stores/runState";
import { createLaunchOptions } from "./formService";
import * as service from "./runService";
import { loadConfigForGame, loadConfigForPrefix } from "./runConfig";

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

	// Proton
	protonVersions = $state<core.ProtonTool[]>([]);
	protonOptions = $state<string[]>([]);
	selectedProton = $state("");
	isLoadingProton = $state(true);

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
	prefixDefaultProton = $state("");

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
				this.options.ProtonPath = this.selectedProton;
				runState.set({
					mainExePath: this.mainExePath,
					gameIcon: this.gameIcon,
					launcherIcon: this.launcherIcon,
					prefixPath: this.prefixPath,
					selectedPrefixName: this.selectedPrefixName,
					selectedProton: this.selectedProton,
					options: this.options,
				});
			}
		});

		// Update prefix default proton
		$effect(() => {
			if (this.mounted && this.selectedPrefixName && this.protonVersions.length > 0) {
				this.updatePrefixDefaultProton(this.selectedPrefixName);
			}
		});

		// Fallback to prefix default proton
		$effect(() => {
			if (this.mounted && !this.options.UseCustomProton && this.prefixDefaultProton) {
				this.selectedProton = this.prefixDefaultProton;
			}
		});
	}

	async initialize() {
		try {
			const data = await service.initializeRunPage(this.options, this.handleConfigUpdate.bind(this));
			
			this.baseDir = data.baseDir;
			this.gpuList = data.gpuList;
			this.protonVersions = data.protonVersions;
			this.protonOptions = data.protonOptions;
			this.availablePrefixes = data.availablePrefixes;
			this.systemStatus = data.systemStatus;
			this.launcherIcon = data.launcherIcon;
			this.gameIcon = data.gameIcon;
			this.mainExePath = data.mainExePath;

			if (this.protonOptions.length > 0 && !this.selectedProton) {
				this.selectedProton = this.protonOptions[0];
			}

			if (!this.prefixPath) {
				this.prefixPath = this.baseDir + "/Default";
				this.selectedPrefixName = "Default";
				await this.doLoadConfigForPrefix("Default");
			}
		} catch (err) {
			console.error("Failed to initialize:", err);
		} finally {
			this.isLoadingProton = false;
			this.mounted = true;
		}
	}

	handleConfigUpdate(newOpts: core.LaunchOptions, pPath: string, pName: string, proton: string) {
		this.options = { ...newOpts };
		if (pPath) this.prefixPath = pPath;
		if (pName) this.selectedPrefixName = pName;
		if (proton) {
			this.selectedProton = proton;
			if (proton && !this.protonOptions.includes(proton)) {
				this.protonOptions = [...this.protonOptions, proton];
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
		await loadConfigForGame(path, this.options, this.prefixPath, this.baseDir, this.selectedPrefixName, this.protonVersions, this.handleConfigUpdate.bind(this));
	}

	async doLoadConfigForPrefix(name: string) {
		await loadConfigForPrefix(name, this.options, this.prefixPath, this.baseDir, this.protonVersions, this.handleConfigUpdate.bind(this));
	}

	async updatePrefixDefaultProton(prefixName: string) {
		if (!prefixName || prefixName === "Custom...") return;
		try {
			const cfg = await LoadPrefixConfig(prefixName);
			if (cfg && cfg.ProtonPath) {
				const match = this.protonVersions.find((p) => p.Path === cfg.ProtonPath);
				this.prefixDefaultProton = match ? match.DisplayName : cfg.ProtonPath;
			} else {
				if (this.protonVersions.length > 0) {
					this.prefixDefaultProton = this.protonVersions[0].DisplayName;
				}
			}
		} catch (e) {
			console.error("Failed to load prefix config for default proton:", e);
			if (this.protonVersions.length > 0) {
				this.prefixDefaultProton = this.protonVersions[0].DisplayName;
			}
		}
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

	handleProtonChange(value: string) {
		this.selectedProton = value;
	}

	closeLauncherOnConfirm = true;

	async handleLaunch(closeLauncher = true) {
		const shouldShowModal = await service.validateAndLaunch(
			this.options, 
			this.systemStatus, 
			this.selectedProton, 
			this.protonVersions, 
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
		await service.executeLaunch(this.options, this.selectedProton, this.protonVersions, this.showLogsWindow, this.closeLauncherOnConfirm);
	}
}
