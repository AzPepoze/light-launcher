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
	} from "@bindings/light-launcher-wails/backend/app";
	import * as core from "@bindings/light-launcher/pkg/core/models";
	import Dropdown from "../components/Dropdown.svelte";
	import ConfigForm from "../components/ConfigForm.svelte";
	import SlideButton from "../components/SlideButton.svelte";
	import Modal from "../components/Modal.svelte";
	import ExecutableSelector from "../components/ExecutableSelector.svelte";
	import { notifications } from "../notificationStore";
	import { runState } from "../stores/runState";
	import { get } from "svelte/store";
	import { Window } from "@wailsio/runtime";
	import { createLaunchOptions } from "../lib/formService";



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
	let haveGameExe = false;

	$: if (options) haveGameExe = options.HaveGameExe;
	$: if (options) options.HaveGameExe = haveGameExe;

	// UI State
	let showLogsWindow = false;
	let showValidationModal = false;
	let missingToolsList: string[] = [];
	let systemStatus: core.SystemToolsStatus = { hasGamescope: false, hasMangoHud: false, hasGameMode: false };

	// Config
	let options: core.LaunchOptions = createLaunchOptions();

	async function loadConfigForGame(path: string) {
		try {
			const config = await GetConfig(path);
			if (config) {
				prefixPath = config.PrefixPath;
				if (prefixPath.startsWith(baseDir)) {
					selectedPrefixName = prefixPath.replace(baseDir + "/", "");
				} else {
					selectedPrefixName = "Custom...";
				}
				applyConfigToOptions(config);
			} else {
				await loadConfigForPrefix(selectedPrefixName);
			}

			// Auto-detect Lossless.dll if not already set
			if (!options.LsfgDllPath) {
				try {
					const dll = await DetectLosslessDll();
					if (dll) {
						options.LsfgDllPath = dll;
					}
				} catch (err) {
					console.error("Failed to detect Lossless.dll:", err);
				}
			}
		} catch (err) {}
	}

	async function loadConfigForPrefix(name: string) {
		if (name === "Custom...") return;
		try {
			const config = await LoadPrefixConfig(name);
			if (config) {
				const savedMainExePath = options.MainExecutablePath;
				const savedLauncherPath = options.LauncherPath;
				const savedHaveGameExe = options.HaveGameExe;
				const savedPrefixPath = options.PrefixPath;

				applyConfigToOptions(config);

				// Restore Target
				options.MainExecutablePath = savedMainExePath;
				options.LauncherPath = savedLauncherPath;
				options.HaveGameExe = savedHaveGameExe;

				if (savedPrefixPath) options.PrefixPath = savedPrefixPath;
			}
		} catch (err) {}
	}

	function applyConfigToOptions(config: core.LaunchOptions) {
		const match = protonVersions.find((p) => p.Path === config.ProtonPath);
		if (match) {
			selectedProton = match.DisplayName;
		} else if (config.ProtonPattern) {
			selectedProton = config.ProtonPattern;
		}

		options.CustomArgs = config.CustomArgs || "";
		options.EnableMangoHud = config.EnableMangoHud;
		options.EnableGamemode = config.EnableGamemode;
		options.EnableLsfgVk = config.EnableLsfgVk;
		options.LsfgMultiplier = config.LsfgMultiplier || "2";
		options.LsfgPerfMode = config.LsfgPerfMode;
		options.LsfgDllPath = config.LsfgDllPath || "";
		options.LsfgGpu = config.LsfgGpu || "";
		options.LsfgFlowScale = config.LsfgFlowScale || "0.8";
		options.LsfgPacing = config.LsfgPacing || "none";
		options.LsfgAllowFp16 = config.LsfgAllowFp16 || false;
		// IMPORTANT: Only apply LauncherPath from config if we don't already have one selected
		if (!options.LauncherPath && config.LauncherPath) {
			options.LauncherPath = config.LauncherPath;
		}
		options.HaveGameExe = config.HaveGameExe === true; // Default to false (launcher-only)

		// CRITICAL: If HaveGameExe is false, MainExecutablePath must equal LauncherPath (launcher-only mode)
		// If HaveGameExe is true, MainExecutablePath should be a separate game exe
		if (!options.HaveGameExe && options.LauncherPath) {
			options.MainExecutablePath = options.LauncherPath;
			console.log(
				"[CONFIG] HaveGameExe=false: enforcing MainExecutablePath=LauncherPath (launcher-only mode)",
			);
			console.log("[CONFIG]   MainExecutablePath set to:", options.LauncherPath);
		} else if (options.HaveGameExe && config.MainExecutablePath) {
			// Only load MainExecutablePath from config if HaveGameExe is true (separate game exe)
			options.MainExecutablePath = config.MainExecutablePath;
			console.log("[CONFIG] HaveGameExe=true: loaded separate game exe from config");
			console.log("[CONFIG]   MainExecutablePath set to:", config.MainExecutablePath);
		}

		options.EnableGamescope = config.EnableGamescope;
		options.GamescopeW = config.GamescopeW || "1920";
		options.GamescopeH = config.GamescopeH || "1080";
		options.GamescopeR = config.GamescopeR || "60";
		options.EnableMemoryMin = config.EnableMemoryMin;
		options.MemoryMinValue = config.MemoryMinValue || "4G";
	}

	onMount(async () => {
		try {
			const s = get(runState);
			if (s) {
				if (s.mainExePath) {
					mainExePath = s.mainExePath;
					options.MainExecutablePath = s.mainExePath;
				}
				if (s.gameIcon) gameIcon = s.gameIcon;
				if (s.launcherIcon) launcherIcon = s.launcherIcon;
				if (s.prefixPath) prefixPath = s.prefixPath;
				if (s.selectedPrefixName) selectedPrefixName = s.selectedPrefixName;
				if (s.selectedProton) selectedProton = s.selectedProton;
				if (s.options) {
					options = { ...options, ...s.options };
				}
			}

			if (options.LauncherPath) {
				await loadConfigForGame(options.LauncherPath);
				if (!launcherIcon) {
					const icon = await GetExeIcon(options.LauncherPath);
					if (icon) launcherIcon = icon;
				}
			}

			const initialPath = await GetInitialLauncherPath();
			if (initialPath) {
				// Only set as game/launcher if not already set, or if explicitly passed from tray
				if (!options.LauncherPath && !options.MainExecutablePath) {
					// No prior state - set initial path as launcher
					options.LauncherPath = initialPath;
					const icon = await GetExeIcon(initialPath);
					if (icon) launcherIcon = icon;
				} else if (!options.MainExecutablePath || options.MainExecutablePath === options.LauncherPath) {
					// Prior state has launcher but no game - set initial path as game
					mainExePath = initialPath;
					options.MainExecutablePath = initialPath;
					const icon = await GetExeIcon(initialPath);
					if (icon) gameIcon = icon;
					await loadConfigForGame(initialPath);
				}
			}

			// Load icons for any paths that don't have icons yet
			if (options.LauncherPath && !launcherIcon) {
				const icon = await GetExeIcon(options.LauncherPath);
				if (icon) launcherIcon = icon;
			}
			if (options.MainExecutablePath && !gameIcon) {
				const icon = await GetExeIcon(options.MainExecutablePath);
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
			availablePrefixes = Array.isArray(prefixes) ? prefixes : ["Default"];
			baseDir = base;
			systemStatus = sysStatus;

			if (!prefixPath) {
				prefixPath = baseDir + "/Default";
				selectedPrefixName = "Default";
				await loadConfigForPrefix("Default");
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
			await loadConfigForPrefix(name);
		}
	}

	async function handleBrowseGame() {
		try {
			const path = await PickFile();
			if (path) {
				console.log("[GAME] Selected game exe:", path);
				console.log("[GAME] Current LauncherPath before game selection:", options.LauncherPath);
				mainExePath = path;
				// Use object spread to trigger Svelte reactivity
				options = { ...options, MainExecutablePath: path };
				console.log("[GAME] Set options.MainExecutablePath to:", options.MainExecutablePath);
				console.log("[GAME] LauncherPath after game selection:", options.LauncherPath);
				console.log("[GAME] Full options object:", JSON.stringify(options));
				// NOTE: Do NOT load config for game exe
				// Game exe is only for LSFG profile matching
				// Configuration is ALWAYS saved under launcher exe path only
			}
		} catch (err) {
			console.error("[GAME] Error loading game:", err);
		}
	}

	async function handleBrowseLauncher() {
		try {
			const path = await PickFile();
			if (path) {
				console.log("[LAUNCHER] Selected launcher exe:", path);
				options = { ...options, LauncherPath: path };
				console.log("[LAUNCHER] Set options.LauncherPath to:", options.LauncherPath);
				console.log("[LAUNCHER] Full options object after assignment:", JSON.stringify(options));

				// Only set MainExecutablePath if user has not explicitly selected a separate game exe
				if (!mainExePath) {
					console.log(
						"[LAUNCHER] No separate game exe selected by user, initializing MainExecutablePath to launcher",
					);
					options = { ...options, MainExecutablePath: path };
					console.log("[LAUNCHER] Set MainExecutablePath to launcher path:", options.MainExecutablePath);
				} else {
					console.log(
						"[LAUNCHER] User already selected separate game exe, keeping MainExecutablePath:",
						mainExePath,
					);
				}

				// Load config for the launcher
				// applyConfigToOptions will enforce UseGameExe if true
				console.log("[LAUNCHER] Loading config for launcher path...");
				await loadConfigForGame(path);
				console.log("[LAUNCHER] Config loaded, final MainExecutablePath:", options.MainExecutablePath);
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
				selectedPrefixName = "Custom...";
			}
		} catch (err) {
			console.error(err);
		}
	}

	function handleProtonChange(value: string) {
		selectedProton = value;
	}

	async function handleLaunch() {
		if (!options.LauncherPath) {
			notifications.add("Please select a launcher executable.", "error");
			return;
		}

		if (options.EnableLsfgVk && !options.LsfgDllPath) {
			notifications.add("LSFG-VK requires Lossless.dll.", "error");
			return;
		}

		missingToolsList = [];
		if (options.EnableGamescope && !systemStatus.hasGamescope) missingToolsList.push("Gamescope");
		if (options.EnableMangoHud && !systemStatus.hasMangoHud) missingToolsList.push("MangoHud");
		if (options.EnableGamemode && !systemStatus.hasGameMode) missingToolsList.push("GameMode");
		if (missingToolsList.length > 0) {
			showValidationModal = true;
			return;
		}
		await proceedToLaunch();
	}

	async function proceedToLaunch() {
		showValidationModal = false;
		console.log("\n============ PROCEED TO LAUNCH ============");

		// DEBUG: Log state at execution time
		console.log("[EXECUTE] Step 1 - Initial state");
		console.log("[EXECUTE]   options.LauncherPath:", options.LauncherPath);
		console.log("[EXECUTE]   options.MainExecutablePath:", options.MainExecutablePath);
		console.log("[EXECUTE]   mainExePath variable:", mainExePath);
		console.log("[EXECUTE]   Full options:", JSON.stringify(options));

		const tool = protonVersions.find((p) => p.DisplayName === selectedProton);
		let cleanName = selectedProton;
		if (cleanName.startsWith("(Steam) ")) {
			cleanName = cleanName.substring(8);
		}

		// Config is ALWAYS saved to launcher path via SaveGameConfig backend logic
		// MainExecutablePath remains the actual executable to run
		// LauncherPath is provided to SaveGameConfig for config storage
		options.PrefixPath = prefixPath;
		options.ProtonPattern = cleanName;
		options.ProtonPath = tool ? tool.Path : "";

		console.log("[EXECUTE] Step 2 - Final options object before RunGame:");
		console.log(JSON.stringify(options, null, 2));
		console.log("============ ABOUT TO CALL RunGame ============\n");

		try {
			console.log("[EXECUTE] Calling RunGame with LauncherPath:", options.LauncherPath);
			console.log("[EXECUTE] Calling RunGame with MainExecutablePath:", options.MainExecutablePath);
			console.log("[EXECUTE] Calling RunGame with full options:", JSON.stringify(options, null, 2));
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

	<!-- Executable Selector Component -->
	<ExecutableSelector
		launcherPath={options.LauncherPath}
		mainExePath={options.MainExecutablePath}
		bind:haveGameExe
		bind:launcherIcon
		bind:gameIcon
		onBrowseLauncher={handleBrowseLauncher}
		onBrowseGame={handleBrowseGame}
	/>

	<!-- Main Form Container -->
	<div class="form-container">
		<div class="form-group">
			<label for="winePrefix">WINEPREFIX</label>
			<div class="input-group">
				<div class="dropdown-wrapper">
					<Dropdown
						options={[...availablePrefixes, "Custom..."]}
						bind:value={selectedPrefixName}
						onChange={handlePrefixChange}
					/>
				</div>
				<button on:click={handleBrowsePrefix} class="btn">Browse</button>
			</div>
			{#if selectedPrefixName === "Custom..." || !prefixPath.startsWith(baseDir)}
				<div class="path-display">{prefixPath}</div>
			{/if}
		</div>

		<div class="form-group">
			<label for="protonVersion">Proton Version</label>
			<div id="protonVersion">
				<Dropdown
					options={protonOptions}
					bind:value={selectedProton}
					placeholder={isLoadingProton ? "Scanning..." : "Select Version"}
					disabled={isLoadingProton}
					onChange={handleProtonChange}
				/>
			</div>
		</div>

		<ConfigForm bind:options />

		<div class="form-group">
			<SlideButton bind:checked={showLogsWindow} label="Show Logs" subtitle="Open logs in terminal" />
		</div>

		<Modal show={showValidationModal} title="Missing Dependencies" onClose={() => (showValidationModal = false)}>
			<div class="warning-modal-content">
				<div class="warning-icon">
					<span class="material-icons" style="font-size: 48px; color: #ef4444;">warning</span>
				</div>
				<p>The following requested features are not installed on your system:</p>
				<div class="missing-list">
					{#each missingToolsList as tool}
						<span class="tool-tag">{tool}</span>
					{/each}
				</div>
				<p class="question">Do you want to launch the game without these features?</p>
				<div class="modal-actions">
					<button class="btn secondary" on:click={() => (showValidationModal = false)}>Cancel</button>
					<button class="btn primary" on:click={proceedToLaunch}>Launch Anyway</button>
				</div>
			</div>
		</Modal>

		<div class="action-area">
			<button class="btn primary launch-btn" on:click={handleLaunch}>
				<span class="material-icons run-icon">rocket_launch</span>
				<span>LAUNCH GAME</span>
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

	.form-group label {
		display: block;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-muted);
		margin-bottom: 8px;
	}
	.input-group {
		display: flex;
		gap: 12px;
		width: 100%;
		.dropdown-wrapper {
			flex: 1;
		}
	}
	.path-display {
		margin-top: 8px;
		font-size: 0.75rem;
		color: var(--text-dim);
		word-break: break-all;
		padding: 8px;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 6px;
	}
	.warning-modal-content {
		text-align: center;
		.warning-icon {
			font-size: 3rem;
			margin-bottom: 16px;
			display: flex;
			justify-content: center;

		}
		p {
			color: var(--text-main);
			line-height: 1.5;
		}
		.missing-list {
			margin: 16px 0;
			display: flex;
			flex-wrap: wrap;
			justify-content: center;
			gap: 12px;
			.tool-tag {
				background: rgba(239, 68, 68, 0.1);
				color: #ef4444;
				padding: 6px 16px;
				border-radius: 20px;
				font-size: 0.9rem;
				border: 1px solid rgba(239, 68, 68, 0.2);
				font-weight: bold;
			}
		}
		.question {
			margin-top: 24px;
			font-weight: 600;
			color: var(--accent-secondary);
		}
	}
	.modal-actions {
		display: flex;
		gap: 12px;
		margin-top: 32px;
		button {
			flex: 1;
			padding: 12px;
			font-weight: 600;
		}
	}
	.action-area {
		position: sticky;
		bottom: -32px; /* Locks to the bottom edge of the view */
		margin: 48px -32px -32px -32px; /* Pull out of parent padding */
		padding: 32px;
		z-index: 10;
		background: linear-gradient(to top, var(--glass-bg) 70%, transparent);
		display: flex;
		justify-content: center;
		pointer-events: none; /* Let clicks pass through the gradient area */
	}

	.launch-btn {
		pointer-events: auto; /* Re-enable clicks for the button */
		width: 100%;
		max-width: 600px;
		height: 60px;
		padding: 0 32px;
		border-radius: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		background: var(--accent-primary) !important;
		color: var(--glass-bg) !important;
		border: none;
		box-shadow: 0 15px 45px rgba(0, 0, 0, 0.5);
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
		font-weight: 800;
		font-size: 1.1rem;
		text-transform: uppercase;
		letter-spacing: 1.5px;

		span {
			display: block;
		}

		&:hover {
			background: var(--accent-secondary) !important;
			transform: translateY(-4px) scale(1.02);
			box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);

			.run-icon {
				transform: rotate(-10deg) scale(1.1);
			}
		}

		&:active {
			transform: translateY(-2px) scale(0.98);
		}

		.run-icon {
			font-size: 28px;
			transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
		}
	}

	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 10px 20px;
		border-radius: 10px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.2s ease;
		border: 1px solid var(--glass-border);
		background: rgba(255, 255, 255, 0.05);
		color: var(--text-main);
		&:hover {
			background: rgba(255, 255, 255, 0.1);
			border-color: var(--glass-border-bright);
		}
		&.primary {
			background: var(--accent-primary);
			border: none;
			color: #000;
			&:hover {
				background: var(--accent-secondary);
			}
		}
	}
	.page-title {
		font-size: 2rem;
		font-weight: bold;
		color: var(--text-main);
		margin: 0 0 24px 0;
	}
</style>
