<script lang="ts">
	import { onMount } from "svelte";
	import {
		PickFile,
		PickFolder,
		ScanProtonVersions,
		RunGame,
		GetConfig,
		ListPrefixes,
		GetPrefixBaseDir,
		GetSystemToolsStatus,
		LoadPrefixConfig,
		GetInitialLauncherPath,
		DetectLosslessDll,
		GetExeIcon,
		GetListGpus,
		SaveGameConfig,
	} from "@bindings/light-launcher/internal/app/app";
	import * as core from "@bindings/light-launcher/internal/types/models";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import SlideButton from "@components/shared/SlideButton.svelte";
	import ExecutableSelector from "@components/run/ExecutableSelector.svelte";
	import PrefixSelector from "@components/run/PrefixSelector.svelte";
	import ProtonSelector from "@components/run/ProtonSelector.svelte";
	import LaunchButton from "@components/run/LaunchButton.svelte";
	import MissingDependenciesModal from "@components/run/MissingDependenciesModal.svelte";
	import { notifications } from "@stores/notificationStore";
	import { runState } from "@stores/runState";
	import { get } from "svelte/store";
	import { Window } from "@wailsio/runtime";
	import { createLaunchOptions } from "@lib/formService";

	// Component State
	let mounted = false;

	// Game Selection
	let mainExePath = "";
	let gameIcon = "";
	let launcherIcon = "";
	let prefixPath = "";
	let baseDir = "";
	let selectedPrefixName = "Default";

	// Prefix & Utilities
	let availablePrefixes: string[] = [];

	// Proton
	let protonVersions: core.ProtonTool[] = [];
	let protonOptions: string[] = [];
	let selectedProton = "";
	let isLoadingProton = true;

	// Game exe toggle state - keep in sync with options
	let useGamePath = false;

	$: if (options) useGamePath = options.UseGamePath;
	$: if (options) options.UseGamePath = useGamePath;

	// UI State
	let showLogsWindow = false;
	let showValidationModal = false;
	let missingToolsList: string[] = [];
	let systemStatus: core.SystemToolsStatus = {
		hasGamescope: false,
		hasMangoHud: false,
		hasGameMode: false,
		hasVulkanInfo: false,
	};

	let options: core.LaunchOptions = createLaunchOptions();
	let gpuList: string[] = [];
	let isSaving = false;

	import { loadConfigForGame, loadConfigForPrefix } from "@lib/runConfig";

	async function handleSave() {
		isSaving = true;
		try {
			await SaveGameConfig(options);
			notifications.add("Configuration saved!", "success");
		} catch (err) {
			notifications.add(`Failed to save: ${err}`, "error");
		} finally {
			isSaving = false;
		}
	}

	function handleConfigUpdate(newOpts: core.LaunchOptions, pPath: string, pName: string, proton: string) {
		options = newOpts;
		if (pPath) prefixPath = pPath;
		if (pName) selectedPrefixName = pName;
		if (proton && selectedProton === "") selectedProton = proton;
	}

	async function doLoadConfigForGame(path: string) {
		await loadConfigForGame(path, options, prefixPath, baseDir, selectedPrefixName, protonVersions, handleConfigUpdate);
	}

	async function doLoadConfigForPrefix(name: string) {
		await loadConfigForPrefix(name, options, prefixPath, baseDir, protonVersions, handleConfigUpdate);
	}

	onMount(async () => {
		try {
			baseDir = await GetPrefixBaseDir();
			
			// Load LSFG Resources
			const [gpus, dll] = await Promise.all([
				GetListGpus(),
				DetectLosslessDll()
			]);
			gpuList = gpus || [];
			if (dll && !options.Extras.Lsfg.DllPath) {
				options.Extras.Lsfg.DllPath = dll;
			}

			const s = get(runState);
			if (s) {
				if (s.mainExePath) {
					mainExePath = s.mainExePath;
					options.GamePath = s.mainExePath;
				}
				if (s.gameIcon) gameIcon = s.gameIcon;
				if (s.launcherIcon) launcherIcon = s.launcherIcon;
				if (s.prefixPath) prefixPath = s.prefixPath;
				if (s.selectedPrefixName)
					selectedPrefixName = s.selectedPrefixName;
				if (s.selectedProton) selectedProton = s.selectedProton;
				if (s.options) {
					options = { ...options, ...s.options };
				}
			}

			if (options.RunnerPath) {
				await doLoadConfigForGame(options.RunnerPath);
				if (!launcherIcon) {
					const icon = await GetExeIcon(options.RunnerPath);
					if (icon) launcherIcon = icon;
				}
			}

			const initialPath = await GetInitialLauncherPath();
			if (initialPath) {
				// Only set as game/launcher if not already set, or if explicitly passed from tray
				if (!options.RunnerPath && !options.GamePath) {
					// No prior state - set initial path as launcher
					options.RunnerPath = initialPath;
					const icon = await GetExeIcon(initialPath);
					if (icon) launcherIcon = icon;
					// Set default name from runner filename
					if (!options.Name || options.Name === "Launcher") {
						options.Name = initialPath.split(/[/\\]/).pop()?.replace(/\.exe$/i, "") || "Launcher";
					}
				} else if (
					!options.GamePath ||
					options.GamePath === options.RunnerPath
				) {
					// Prior state has launcher but no game - set initial path as game
					mainExePath = initialPath;
					options.GamePath = initialPath;
					const icon = await GetExeIcon(initialPath);
					if (icon) gameIcon = icon;
					await doLoadConfigForGame(initialPath);
				}
			}

			// Load icons for any paths that don't have icons yet
			if (options.RunnerPath && !launcherIcon) {
				const icon = await GetExeIcon(options.RunnerPath);
				if (icon) launcherIcon = icon;
			}
			if (options.GamePath && !gameIcon) {
				const icon = await GetExeIcon(options.GamePath);
				if (icon) gameIcon = icon;
			}

			const [tools, prefixes, base, sysStatus] = await Promise.all([
				ScanProtonVersions(),
				ListPrefixes(),
				GetPrefixBaseDir(),
				GetSystemToolsStatus(),
			]);
			if (tools) {
				protonVersions = tools;
				protonOptions = tools.map((t) => t.DisplayName);
				if (protonOptions.length > 0 && !selectedProton) {
					selectedProton = protonOptions[0];
				}
			}
			availablePrefixes = Array.isArray(prefixes)
				? prefixes
				: ["Default"];
			baseDir = base;
			systemStatus = sysStatus;

			if (!prefixPath) {
				prefixPath = baseDir + "/Default";
				selectedPrefixName = "Default";
				await doLoadConfigForPrefix("Default");
			}
		} catch (err) {
			console.error("Failed to initialize:", err);
		} finally {
			isLoadingProton = false;
			mounted = true;
		}
	});

	$: if (mounted) {
		runState.set({
			mainExePath,
			gameIcon,
			launcherIcon,
			prefixPath,
			selectedPrefixName,
			selectedProton,
			options,
		});
	}

	async function handlePrefixChange(name: string) {
		if (name !== "Custom...") {
			prefixPath = baseDir + "/" + name;
			selectedPrefixName = name;
			await doLoadConfigForPrefix(name);
		}
	}

	async function handleBrowseGame() {
		try {
			const path = await PickFile();
			if (path) {
				mainExePath = path;
				options = { ...options, GamePath: path };
			}
		} catch (err) {
			console.error("[GAME] Error loading game:", err);
		}
	}

	async function handleBrowseLauncher() {
		try {
			const path = await PickFile();
			if (path) {
				options = { ...options, RunnerPath: path };
				
				// Set default name if not set
				if (!options.Name || options.Name === "Launcher") {
					options.Name = path.split(/[/\\]/).pop()?.replace(/\.exe$/i, "") || "Launcher";
				}

				if (!mainExePath) {
					options = { ...options, GamePath: path };
				}

				await doLoadConfigForGame(path);
			}
		} catch (err) {
			console.error("[LAUNCHER] Error selecting launcher:", err);
		}
	}

	async function handleBrowsePrefix() {
		try {
			const path = await PickFolder();
			if (path) {
				prefixPath = path;
				selectedPrefixName = path.split('/').filter(Boolean).pop() || "Custom";
			}
		} catch (err) {
			console.error(err);
		}
	}

	function handleProtonChange(value: string) {
		selectedProton = value;
	}

	async function handleLaunch() {
		if (!options.RunnerPath) {
			notifications.add(
				"Please select a launcher executable.",
				"error",
			);
			return;
		}

		if (options.Extras.Lsfg.Enabled && !options.Extras.Lsfg.DllPath) {
			notifications.add("LSFG-VK requires Lossless.dll.", "error");
			return;
		}

		missingToolsList = [];
		if (options.Extras.Gamescope.Enabled && !systemStatus.hasGamescope)
			missingToolsList.push("Gamescope");
		if (options.Extras.EnableMangoHud && !systemStatus.hasMangoHud)
			missingToolsList.push("MangoHud");
		if (options.Extras.EnableGamemode && !systemStatus.hasGameMode)
			missingToolsList.push("GameMode");
		if (options.Extras.Lsfg.Enabled && !systemStatus.hasVulkanInfo)
			missingToolsList.push("Vulkan-Tools");
		if (missingToolsList.length > 0) {
			showValidationModal = true;
			return;
		}
		await proceedToLaunch();
	}

	async function proceedToLaunch() {
		showValidationModal = false;

		const tool = protonVersions.find(
			(p) => p.DisplayName === selectedProton,
		);
		let cleanName = selectedProton;
		if (cleanName.startsWith("(Steam) ")) {
			cleanName = cleanName.substring(8);
		}

		options.PrefixPath = prefixPath;
		options.ProtonPattern = cleanName;
		options.ProtonPath = tool ? tool.Path : "";

		try {
			await RunGame(options, showLogsWindow);
			Window.Close();
		} catch (err) {
			console.error("[EXECUTE] Launch failed:", err);
			notifications.add(`Launch failed: ${err}`, "error");
		}
	}
</script>

<div class="run-container">
	<div class="header-row">
		<h1 class="page-title">Launch Configuration</h1>
	</div>

	<div class="form-group profile-name-group">
		<label for="profileName">Profile Name</label>
		<input
			id="profileName"
			type="text"
			class="input profile-input"
			bind:value={options.Name}
			placeholder="Enter a name for this profile..."
		/>
	</div>

	<!-- Executable Selector Component -->
	<ExecutableSelector
		runnerPath={options.RunnerPath}
		gamePath={options.GamePath}
		bind:useGamePath
		bind:launcherIcon
		bind:gameIcon
		onBrowseLauncher={handleBrowseLauncher}
		onBrowseGame={handleBrowseGame}
	/>

	<!-- Main Form Container -->
	<div class="form-container">
		<PrefixSelector
			bind:availablePrefixes
			bind:selectedPrefixName
			bind:prefixPath
			{baseDir}
			onPrefixChange={handlePrefixChange}
			onBrowsePrefix={handleBrowsePrefix}
		/>

		<ProtonSelector
			bind:protonOptions
			bind:selectedProton
			bind:isLoadingProton
			onProtonChange={handleProtonChange}
		/>

		<ConfigForm bind:options />

		<div class="form-group">
			<SlideButton
				bind:checked={showLogsWindow}
				label="Show Logs"
				subtitle="Open logs in terminal"
			/>
		</div>

		<MissingDependenciesModal
			show={showValidationModal}
			missingTools={missingToolsList}
			onClose={() => (showValidationModal = false)}
			onConfirm={proceedToLaunch}
		/>

		<div class="actions-row">
			<button class="btn secondary save-btn" on:click={handleSave} disabled={isSaving}>
				<span class="material-icons">{isSaving ? "sync" : "save"}</span>
				{isSaving ? "Saving..." : "Save Configuration"}
			</button>
			<LaunchButton onLaunch={handleLaunch} />
		</div>
	</div>
</div>

<style lang="scss">
	.run-container {
		display: flex;
		flex-direction: column;
		padding: 32px;
	}
	.form-container {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}




	.page-title {
		font-size: 2rem;
		font-weight: bold;
		color: var(--text-main);
		margin: 0 0 24px 0;
	}

	.profile-name-group {
		margin-bottom: 24px;
		padding: 16px;
		background: var(--glass-surface);
		border-radius: 12px;
		border: 1px solid var(--glass-border);

		label {
			display: block;
			font-size: 0.875rem;
			font-weight: 700;
			color: var(--text-muted);
			margin-bottom: 8px;
		}

		.profile-input {
			width: 100%;
			font-size: 1.1rem;
			font-weight: 600;
			background: var(--glass-hover);
			border-color: var(--glass-border-bright);

			&:focus {
				border-color: var(--accent-primary);
				background: var(--glass-surface);
			}
		}
	}

	.actions-row {
		display: flex;
		justify-content: flex-end;
		gap: 16px;
		margin-top: 8px;
		align-items: center;

		.save-btn {
			display: flex;
			align-items: center;
			gap: 8px;
			padding: 12px 24px;
			border-radius: 12px;
			font-weight: 800;
			text-transform: uppercase;
			letter-spacing: 0.5px;
			background: var(--glass-surface);
			color: var(--text-main);
			border: 1px solid var(--glass-border);
			cursor: pointer;
			transition: all 0.2s;

			&:hover:not(:disabled) {
				background: var(--glass-border);
				border-color: var(--glass-border-bright);
				transform: translateY(-1px);
			}

			&:active {
				transform: translateY(0);
			}

			&:disabled {
				opacity: 0.5;
				cursor: not-allowed;

				.material-icons {
					animation: spin 2s linear infinite;
				}
			}

			.material-icons {
				font-size: 20px;
			}
		}
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}
</style>
