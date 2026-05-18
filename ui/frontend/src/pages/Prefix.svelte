<script lang="ts">
	import {
		GetSystemToolsStatus,
		PickFolder,
		ScanProtonVersions,
	} from "@bindings/light-launcher/internal/app/app";
	import * as core from "@bindings/light-launcher/internal/types/models";
	import PrefixTools from "@components/prefix/PrefixTools.svelte";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { createLaunchOptions } from "@lib/formService";
	import { createLogger } from "@lib/logger";
	import * as service from "@lib/prefixService";
	import { onMount } from "svelte";

	const log = createLogger("Prefix");

	// State
	let availablePrefixes: string[] = [];
	let baseDir = "";
	let prefixPath = "";
	let chosenProton: core.ProtonTool | null = null; // The actual ProtonTool object
	let chosenProtonName = ""; // Display name for dropdown
	let protonVersions: core.ProtonTool[] = [];
	let protonDisplayNames: string[] = []; // Display names for dropdown
	let systemStatus: core.SystemToolsStatus | null = null;
	let newPrefixName = "";
	let isLoading = false;
	let runningToolName = "";
	let showCreateInput = false;

	// Config
	let prefixOptions: core.LaunchOptions = createLaunchOptions();

	// Reactively sync DisplayName changes to ProtonTool object
	$: if (chosenProtonName) {
		chosenProton =
			protonVersions.find((t) => t.DisplayName === chosenProtonName) ||
			null;
	}

	async function refreshPrefixes(autoSelect = true) {
		const data = await service.getPrefixData();
		availablePrefixes = data.availablePrefixes;
		baseDir = data.baseDir;

		if (autoSelect) {
			if (!prefixPath && availablePrefixes.length > 0) {
				selectPrefix(availablePrefixes[0]);
			} else if (!prefixPath) {
				prefixPath = baseDir + "/Default";
			}
		}
	}

	onMount(async () => {
		log.info("onMount starting");
		try {
			const [tools, status] = await Promise.all([
				ScanProtonVersions(),
				GetSystemToolsStatus(),
			]);
			log.debug("Scanned proton versions", {
				count: tools.length,
				tools,
			});
			log.debug("System status", status);
			systemStatus = status;
			protonVersions = tools;
			protonDisplayNames = protonVersions.map((t) => t.DisplayName);
			log.debug("Available proton displays", protonDisplayNames);
			// Don't set default selectedProton here - let it be set when prefix is loaded
			await refreshPrefixes();
		} catch (err) {
			log.error("onMount error", err);
		}
	});

	async function selectPrefix(name: string) {
		log.info("selectPrefix called", { name });
		const result = await service.getPrefixConfig(name, baseDir);
		log.debug("getPrefixConfig result", result);
		prefixPath = result.path;
		if (result.options) {
			prefixOptions = { ...prefixOptions, ...result.options };
			log.debug("Loaded proton name", result.selectedProton);
			if (result.selectedProton) {
				// Set the DisplayName so it syncs to ProtonTool object via reactive statement
				chosenProtonName = result.selectedProton;
				log.info("Set chosenProtonName", { chosenProtonName });
			} else if (protonVersions.length > 0) {
				// Fallback to first proton if none saved
				chosenProtonName = protonVersions[0].DisplayName;
				log.info("No saved proton, using first", {
					chosenProtonName,
				});
			}
		} else {
			prefixOptions = createLaunchOptions();
			// Ensure a proton is selected when creating new prefix
			if (protonVersions.length > 0 && !chosenProton) {
				chosenProtonName = protonVersions[0].DisplayName;
				log.info("New prefix, setting first proton", {
					chosenProtonName,
				});
			}
		}
	}

	async function handleSaveConfig() {
		log.info("handleSaveConfig called");
		log.debug("Current state", {
			chosenProton,
			chosenProtonName,
			protonPath: prefixOptions.ProtonPath,
		});
		await service.savePrefixDefaults(
			prefixPath,
			prefixOptions,
			chosenProton,
		);
		log.info("Config saved successfully");
	}

	async function handleBrowse() {
		try {
			const path = await PickFolder();
			if (path) prefixPath = path;
		} catch (err) {
			log.error("Browse folder error", err);
		}
	}

	async function handleCreatePrefix() {
		if (!newPrefixName) return;
		const name = newPrefixName;
		await service.createNewPrefix(name);
		newPrefixName = "";
		showCreateInput = false;
		await refreshPrefixes(false);
		await selectPrefix(name);
	}

	async function handleRemovePrefix(name: string) {
		await service.deletePrefix(name);
		await refreshPrefixes(false);
		if (availablePrefixes.length > 0) {
			await selectPrefix(availablePrefixes[0]);
		} else {
			prefixPath = baseDir + "/Default";
		}
	}

	async function runTool(tool: string) {
		if (isLoading) return;
		isLoading = true;
		runningToolName = tool;

		try {
			await service.executePrefixTool(
				prefixPath,
				tool,
				chosenProton?.Path || "",
			);
		} catch (err) {
			// Error handled in service via notification
		} finally {
			setTimeout(() => {
				isLoading = false;
				runningToolName = "";
			}, 500);
		}
	}

	$: currentPrefixName = prefixPath.startsWith(baseDir)
		? prefixPath.replace(baseDir + "/", "")
		: prefixPath;
</script>

<div class="prefix-container">
	<PageHeader
		title="Prefix Manager"
		icon="folder_shared"
		subtitle="Manage Wine/Proton prefixes"
	/>

	<!-- Prefix Selector Bar -->
	<div class="prefix-selector">
		<div class="prefix-tabs">
			{#each availablePrefixes as name}
				<div
					class="prefix-tab"
					class:active={currentPrefixName === name}
					on:click={() => selectPrefix(name)}
					on:keydown={(e) =>
						e.key === "Enter" && selectPrefix(name)}
					role="tab"
					tabindex="0"
					title={name}
				>
					<span class="material-icons tab-icon">folder</span>
					<span class="tab-name">{name}</span>
					{#if name !== "Default"}
						<span
							class="tab-delete"
							title="Delete"
							role="button"
							tabindex="0"
							on:click|stopPropagation={() =>
								handleRemovePrefix(name)}
							on:keydown|stopPropagation={(e) =>
								e.key === "Enter" &&
								handleRemovePrefix(name)}
						>
							<span class="material-icons">close</span>
						</span>
					{/if}
				</div>
			{/each}

			{#if showCreateInput}
				<div class="create-inline">
					<input
						type="text"
						placeholder="Name..."
						bind:value={newPrefixName}
						on:keydown={(e) => {
							if (e.key === "Enter") handleCreatePrefix();
							if (e.key === "Escape") {
								showCreateInput = false;
								newPrefixName = "";
							}
						}}
						class="create-input"
					/>
					<button
						class="create-confirm"
						on:click={handleCreatePrefix}
					>
						<span class="material-icons">check</span>
					</button>
				</div>
			{/if}

			<button
				class="prefix-tab add-tab"
				on:click={() => (showCreateInput = !showCreateInput)}
				title="New Prefix"
			>
				<span class="material-icons"
					>{showCreateInput ? "close" : "add"}</span
				>
			</button>
		</div>
	</div>

	<!-- Active Prefix Info -->
	<div class="prefix-path-bar">
		<span class="material-icons path-icon">terminal</span>
		<span class="path-text" title={prefixPath}>{prefixPath}</span>
		<button class="browse-btn" on:click={handleBrowse}>
			<span class="material-icons">folder_open</span>
			Browse
		</button>
	</div>

	<!-- Runtime Selector -->
	<div class="runtime-row">
		<div class="runtime-label">
			<span class="material-icons">science</span>
			<span>Runtime</span>
		</div>
		<div class="runtime-dropdown">
			<Dropdown
				options={protonDisplayNames}
				bind:value={chosenProtonName}
			/>
		</div>
	</div>

	<!-- Tools Row -->
	<PrefixTools {runningToolName} onRunTool={runTool} />

	<!-- Default Config -->
	<div class="config-section">
		<div class="config-header">
			<h3>Default Configuration</h3>
			<button class="btn primary sm" on:click={handleSaveConfig}>
				<span
					class="material-icons"
					style="font-size: 16px; margin-right: 6px;">save</span
				>
				Save Defaults
			</button>
		</div>
		<ConfigForm bind:options={prefixOptions} />
	</div>
</div>

<style lang="scss">
	.prefix-container {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	/* ---- Prefix Tab Bar ---- */
	.prefix-selector {
		width: 100%;
	}

	.prefix-tabs {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		align-items: center;
	}

	.prefix-tab {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 18px;
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-pill);
		background: var(--bg-surface);
		color: var(--text-muted);
		font-size: 0.85rem;
		font-weight: 700;
		cursor: pointer;
		transition:
			transform var(--transition-spring),
			background var(--transition-fast),
			border-color var(--transition-fast),
			color var(--transition-fast);
		outline: none;
		position: relative;

		.tab-icon {
			font-size: 16px;
		}

		.tab-name {
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		&:hover {
			background: var(--bg-elevated);
			border-color: var(--accent-secondary);
			color: var(--text-main);
			transform: scale(1.03);

			.tab-delete {
				opacity: 1;
			}
		}

		&:active {
			transform: scale(0.96);
		}

		&.active {
			background: var(--accent-primary);
			border-color: var(--accent-primary);
			color: #ffffff;
			box-shadow: 0 4px 14px var(--accent-glow);

			.tab-icon {
				color: #ffffff;
			}

			.tab-delete {
				color: rgba(255, 255, 255, 0.7);
				&:hover {
					color: #ffffff;
					background: rgba(0, 0, 0, 0.15);
				}
			}
		}

		&.add-tab {
			width: 40px;
			height: 40px;
			padding: 0;
			justify-content: center;
			border-style: dashed;

			.material-icons {
				font-size: 20px;
			}

			&:hover {
				border-color: var(--accent-primary);
				color: var(--accent-primary);
				border-style: solid;
			}
		}
	}

	.tab-delete {
		background: transparent;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		border-radius: 50%;
		opacity: 0;
		transition:
			opacity var(--transition-fast),
			color var(--transition-fast),
			background var(--transition-fast);
		padding: 0;
		margin-left: -2px;

		.material-icons {
			font-size: 14px;
		}

		&:hover {
			background: rgba(255, 74, 122, 0.2);
			color: var(--danger);
		}
	}

	.create-inline {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.create-input {
		padding: 8px 14px;
		background: var(--bg-input);
		border: 2px solid var(--accent-primary);
		border-radius: var(--radius-pill);
		color: var(--text-main);
		font-size: 0.85rem;
		font-weight: 700;
		outline: none;
		width: 140px;
		transition: transform var(--transition-fast);

		&:focus {
			transform: scale(1.01);
		}
	}

	.create-confirm {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border: none;
		border-radius: 50%;
		background: var(--accent-primary);
		color: #ffffff;
		cursor: pointer;
		transition: transform var(--transition-spring);

		.material-icons {
			font-size: 18px;
		}

		&:hover {
			transform: scale(1.1);
		}
		&:active {
			transform: scale(0.9);
		}
	}

	/* ---- Path Bar ---- */
	.prefix-path-bar {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 20px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		overflow: hidden;

		.path-icon {
			font-size: 18px;
			color: var(--accent-secondary);
			flex-shrink: 0;
		}

		.path-text {
			flex: 1;
			font-size: 0.85rem;
			font-weight: 600;
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			font-family: monospace;
		}

		.browse-btn {
			display: flex;
			align-items: center;
			gap: 6px;
			padding: 6px 14px;
			background: var(--bg-elevated);
			border: 2px solid rgba(255, 255, 255, 0.05);
			border-radius: var(--radius-pill);
			color: var(--text-muted);
			font-size: 0.8rem;
			font-weight: 700;
			cursor: pointer;
			flex-shrink: 0;
			transition:
				transform var(--transition-spring),
				border-color var(--transition-fast),
				color var(--transition-fast);

			.material-icons {
				font-size: 16px;
			}

			&:hover {
				border-color: var(--accent-secondary);
				color: var(--text-main);
				transform: scale(1.05);
			}
			&:active {
				transform: scale(0.95);
			}
		}
	}

	/* ---- Runtime Row ---- */
	.runtime-row {
		display: flex;
		align-items: center;
		gap: 16px;

		.runtime-label {
			display: flex;
			align-items: center;
			gap: 8px;
			font-size: 0.9rem;
			font-weight: 800;
			color: var(--accent-primary);
			text-transform: uppercase;
			letter-spacing: 0.5px;
			white-space: nowrap;
			flex-shrink: 0;

			.material-icons {
				font-size: 20px;
			}
		}

		.runtime-dropdown {
			flex: 1;
			max-width: 400px;
		}
	}

	/* ---- Config Section ---- */
	.config-section {
		padding: 24px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.config-header {
		display: flex;
		justify-content: space-between;
		align-items: center;

		h3 {
			margin: 0;
			font-size: 1.1rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}
	}
</style>
