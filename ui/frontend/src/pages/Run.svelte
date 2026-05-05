<script lang="ts">
	import { onMount } from "svelte";
	import {
		PickFile,
		PickFolder,
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
	import { createLaunchOptions, mergeOptions } from "@lib/formService";
	import * as service from "@lib/runService";
	import { loadConfigForGame, loadConfigForPrefix } from "@lib/runConfig";

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
		options = { ...newOpts };
		if (pPath) prefixPath = pPath;
		if (pName) selectedPrefixName = pName;
		if (proton) {
			selectedProton = proton;
			if (proton && !protonOptions.includes(proton)) {
				protonOptions = [...protonOptions, proton];
			}
		}
	}

	async function doLoadConfigForGame(path: string) {
		await loadConfigForGame(path, options, prefixPath, baseDir, selectedPrefixName, protonVersions, handleConfigUpdate);
	}

	async function doLoadConfigForPrefix(name: string) {
		await loadConfigForPrefix(name, options, prefixPath, baseDir, protonVersions, handleConfigUpdate);
	}

	onMount(async () => {
		try {
			const data = await service.initializeRunPage(options, handleConfigUpdate);
			
			baseDir = data.baseDir;
			gpuList = data.gpuList;
			protonVersions = data.protonVersions;
			protonOptions = data.protonOptions;
			availablePrefixes = data.availablePrefixes;
			systemStatus = data.systemStatus;
			launcherIcon = data.launcherIcon;
			gameIcon = data.gameIcon;
			mainExePath = data.mainExePath;

			if (protonOptions.length > 0 && !selectedProton) {
				selectedProton = protonOptions[0];
			}

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
		const shouldShowModal = await service.validateAndLaunch(
			options, 
			systemStatus, 
			selectedProton, 
			protonVersions, 
			showLogsWindow
		);
		
		if (shouldShowModal === true) {
			missingToolsList = [];
			if (options.Extras.Gamescope.Enabled && !systemStatus.hasGamescope) missingToolsList.push("Gamescope");
			if (options.Extras.EnableMangoHud && !systemStatus.hasMangoHud) missingToolsList.push("MangoHud");
			if (options.Extras.EnableGamemode && !systemStatus.hasGameMode) missingToolsList.push("GameMode");
			if (options.Extras.Lsfg.Enabled && !systemStatus.hasVulkanInfo) missingToolsList.push("Vulkan-Tools");
			showValidationModal = true;
		}
	}

	async function proceedToLaunch() {
		showValidationModal = false;
		await service.executeLaunch(options, selectedProton, protonVersions, showLogsWindow);
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
			<div class="launch-wrapper">
				<LaunchButton onLaunch={handleLaunch} />
			</div>
			<button class="icon-btn save-btn" on:click={handleSave} disabled={isSaving} title="Save Configuration">
				<span class="material-icons">{isSaving ? "sync" : "save"}</span>
			</button>
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
		position: sticky;
		bottom: -32px;
		margin: 48px -32px -32px -32px;
		padding: 32px;
		z-index: 10;
		background: linear-gradient(to top, var(--glass-bg) 80%, transparent);
		display: flex;
		align-items: center;
		gap: 16px;
		pointer-events: none;
	}

	.launch-wrapper {
		flex: 1;
		pointer-events: auto;
	}

	.save-btn {
		pointer-events: auto;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 60px;
		height: 60px;
		border-radius: 14px;
		background: var(--glass-surface);
		color: var(--text-muted);
		border: 1px solid var(--glass-border);
		cursor: pointer;
		transition: all 0.2s;

		&:hover:not(:disabled) {
			background: var(--glass-border);
			color: var(--text-main);
			border-color: var(--glass-border-bright);
			transform: scale(1.05);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
			.material-icons {
				animation: spin 2s linear infinite;
			}
		}

		.material-icons {
			font-size: 24px;
		}
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}
</style>
