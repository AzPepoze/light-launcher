<script lang="ts">
	import { onMount } from "svelte";
	import {
		PickFolder,
		RunPrefixTool,
		ScanProtonVersions,
		ListPrefixes,
		GetPrefixBaseDir,
		CreatePrefix,
		SavePrefixConfig,
		LoadPrefixConfig,
		GetSystemToolsStatus,
	} from "@bindings/light-launcher-wails/backend/app";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import ToolButton from "@components/shared/ToolButton.svelte";
	import * as core from "@bindings/light-launcher/pkg/core/models";
	import { notifications } from "@stores/notificationStore";
	import { createLaunchOptions } from "@lib/formService";

	// State
	let availablePrefixes: string[] = [];
	let baseDir = "";
	let prefixPath = "";
	let selectedProton = "";
	let protonVersions: any[] = [];
	let protonOptions: string[] = [];
	let systemStatus: any = null;
	let newPrefixName = "";
	let isLoading = false;
	let runningToolName = "";

	// Config
	let prefixOptions: core.LaunchOptions = createLaunchOptions();

	async function refreshPrefixes() {
		try {
			const list = await ListPrefixes();
			availablePrefixes = Array.isArray(list) ? list : [];
			baseDir = await GetPrefixBaseDir();
			if (!prefixPath && availablePrefixes.length > 0) {
				selectPrefix(availablePrefixes[0]);
			} else if (!prefixPath) {
				prefixPath = baseDir + "/Default";
			}
		} catch (err) {
			console.error(err);
			availablePrefixes = [];
		}
	}

	onMount(async () => {
		try {
			const [tools, status] = await Promise.all([
				ScanProtonVersions(),
				GetSystemToolsStatus(),
			]);
			systemStatus = status;
			protonVersions = tools;
			protonOptions = protonVersions.map((t) => t.DisplayName);
			if (protonVersions.length > 0) {
				selectedProton = protonOptions[0];
			}
			await refreshPrefixes();
		} catch (err) {
			console.error(err);
		}
	});

	async function selectPrefix(name: string) {
		prefixPath = baseDir + "/" + name;
		try {
			const config = await LoadPrefixConfig(name);
			if (config) {
				prefixOptions = { ...prefixOptions, ...config };

				if (config.ProtonPattern) {
					const match = protonVersions.find(
						(p) =>
							p.Name === config.ProtonPattern ||
							p.DisplayName === config.ProtonPattern,
					);
					if (match) selectedProton = match.DisplayName;
					else if (config.ProtonPattern)
						selectedProton = config.ProtonPattern;
				}
			} else {
				resetOptions();
			}
		} catch (e) {
			resetOptions();
		}
	}

	function resetOptions() {
		prefixOptions.CustomArgs = "";
		prefixOptions.EnableGamescope = false;
		prefixOptions.EnableMangoHud = false;
		prefixOptions.EnableGamemode = false;
		prefixOptions.EnableLsfgVk = false;
	}

	async function handleSaveConfig() {
		if (!prefixPath) return;
		const name = prefixPath.split("/").pop() || "Default";

		// Update proton info in options
		const tool = protonVersions.find(
			(p) => p.DisplayName === selectedProton,
		);
		let cleanName = selectedProton;
		if (cleanName.startsWith("(Steam) "))
			cleanName = cleanName.substring(8);

		prefixOptions.ProtonPattern = cleanName;
		prefixOptions.ProtonPath = tool ? tool.Path : "";

		await notifications.withNotification(
			SavePrefixConfig(name, prefixOptions),
			{
				success: "Prefix defaults saved!",
				error: "Failed to save configuration",
			},
		);
	}

	async function handleBrowse() {
		try {
			const path = await PickFolder();
			if (path) prefixPath = path;
		} catch (err) {
			console.error(err);
		}
	}

	async function handleCreatePrefix() {
		if (!newPrefixName) return;
		const name = newPrefixName;
		await notifications.withNotification(
			(async () => {
				await CreatePrefix(name);
				newPrefixName = "";
				await refreshPrefixes();
				selectPrefix(name);
			})(),
			{
				success: `Created prefix "${name}"`,
				error: "Failed to create prefix",
			},
		);
	}

	function handleProtonChange(value: string) {
		selectedProton = value;
	}

	async function runTool(tool: string) {
		if (isLoading) return;
		if (!prefixPath) {
			notifications.error("Please select or create a prefix first.");
			return;
		}

		isLoading = true;
		runningToolName = tool;

		let cleanName = selectedProton;
		if (cleanName.startsWith("(Steam) ")) {
			cleanName = cleanName.substring(8);
		}
		try {
			await RunPrefixTool(prefixPath, tool, cleanName);
		} catch (err) {
			notifications.error(`Failed to run ${tool}: ${err}`);
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
	<h1 class="page-title">Prefix Manager</h1>
	<div class="main-layout">
		<div class="sidebar-section glass">
			<div class="section-header">
				<h2>Available Prefixes</h2>
			</div>
			<div class="prefix-list">
				{#each availablePrefixes as name}
					<button
						class="prefix-item"
						class:active={currentPrefixName === name}
						on:click={() => selectPrefix(name)}
					>
						<span class="folder-icon">
							<span class="material-icons">folder</span>
						</span>
						<span class="name">{name}</span>
					</button>
				{/each}
				{#if availablePrefixes.length === 0}
					<div class="empty-state">
						No prefixes found in default directory.
					</div>
				{/if}
			</div>
			<div class="add-prefix-area">
				<input
					type="text"
					placeholder="New prefix..."
					bind:value={newPrefixName}
					class="input sm"
					on:keydown={(e) =>
						e.key === "Enter" && handleCreatePrefix()}
				/>
				<button class="btn primary sm" on:click={handleCreatePrefix}
					>Create</button
				>
			</div>
		</div>
		<div class="content-section">
			<div class="config-card glass">
				<div class="form-group">
					<label for="prefixPath">Selected Prefix Path</label>
					<div class="input-group">
						<input
							id="prefixPath"
							type="text"
							class="input"
							bind:value={prefixPath}
							readonly
						/>
						<button class="btn" on:click={handleBrowse}
							>Browse Other</button
						>
					</div>
				</div>
				<div class="form-group">
					<label for="protonRuntime"
						>Runtime Environment (Proton)</label
					>
					<div id="protonRuntime">
						<Dropdown
							options={protonOptions}
							bind:value={selectedProton}
							onChange={handleProtonChange}
						/>
					</div>
				</div>
			</div>

			<div class="tools-grid">
				<ToolButton
					icon="settings"
					title="Winecfg"
					subtitle="Settings"
					loading={runningToolName === "winecfg"}
					onClick={() => runTool("winecfg")}
				/>
				<ToolButton
					icon="edit"
					title="Regedit"
					subtitle="Registry"
					loading={runningToolName === "regedit"}
					onClick={() => runTool("regedit")}
				/>
				<ToolButton
					icon="terminal"
					title="CMD"
					subtitle="Terminal"
					loading={runningToolName === "cmd"}
					onClick={() => runTool("cmd")}
				/>
				<ToolButton
					icon="auto_fix_high"
					title="Winetricks"
					subtitle="Extras"
					loading={runningToolName === "winetricks"}
					onClick={() => runTool("winetricks")}
				/>
				<ToolButton
					icon="bar_chart"
					title="TaskMgr"
					subtitle="Processes"
					loading={runningToolName === "taskmgr"}
					onClick={() => runTool("taskmgr")}
				/>
				<ToolButton
					icon="folder"
					title="Explorer"
					subtitle="Files"
					loading={runningToolName === "explorer"}
					onClick={() => runTool("explorer")}
				/>
			</div>

			<div class="config-card glass">
				<div class="section-header-row">
					<h3>Default Configuration</h3>
					<button
						class="btn primary sm"
						on:click={handleSaveConfig}>Save Defaults</button
					>
				</div>
				<div class="config-form-wrapper">
					<ConfigForm bind:options={prefixOptions} />
				</div>
			</div>
		</div>
	</div>
</div>

<style lang="scss">
	.prefix-container {
		padding: 24px;
		height: 100%;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-sizing: border-box;
	}
	.page-title {
		font-size: 1.75rem;
		font-weight: 800;
		color: var(--text-main);
		margin: 0 0 20px 0;
		flex-shrink: 0;
	}
	.main-layout {
		display: grid;
		grid-template-columns: 280px 1fr;
		gap: 20px;
		flex: 1;
		min-height: 0;
		overflow: hidden;
		padding-bottom: 4px;
	}
	.sidebar-section {
		display: flex;
		flex-direction: column;
		border-radius: 16px;
		overflow: hidden;
		border: 1px solid var(--glass-border);
		background: var(--glass-surface);
		height: 100%;
		box-sizing: border-box;

		.section-header {
			padding: 16px;
			border-bottom: 1px solid var(--glass-border);
			h2 {
				font-size: 0.9rem;
				margin: 0;
				color: var(--text-dim);
				text-transform: uppercase;
				letter-spacing: 1px;
			}
		}
	}
	.prefix-list {
		flex: 1;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.prefix-item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		background: transparent;
		border: none;
		border-radius: 8px;
		color: var(--text-main);
		cursor: pointer;
		text-align: left;
		transition: all 0.2s;
		&:hover {
			background: var(--glass-hover);
		}
		&.active {
			background: var(--accent-primary);
			color: var(--glass-bg);
			font-weight: 600;
			.folder-icon {
				color: var(--glass-bg);
				opacity: 0.9;
			}
		}
		.folder-icon {
			font-size: 1.1rem;
			display: flex;
			align-items: center;
			justify-content: center;
			width: 18px;
			height: 18px;

			.material-icons {
				font-size: 1.2rem;
				color: var(--accent-primary);
			}
		}
		.name {
			flex: 1;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}
	}
	.add-prefix-area {
		padding: 16px;
		border-top: 1px solid var(--glass-border);
		display: flex;
		flex-direction: column;
		gap: 8px;
		background: var(--glass-surface);
	}
	.content-section {
		display: flex;
		flex-direction: column;
		gap: 24px;
		overflow-y: auto;
		padding: 2px 12px 24px 2px;
		height: 100%;
		box-sizing: border-box;
	}
	.config-card {
		padding: 24px;
		border-radius: 16px;
		border: 1px solid var(--glass-border);
		display: flex;
		flex-direction: column;
		gap: 20px;
		flex-shrink: 0;
	}
	.form-group {
		label {
			display: block;
			font-size: 0.8rem;
			font-weight: 600;
			color: var(--text-dim);
			margin-bottom: 8px;
		}
	}
	.input-group {
		display: flex;
		gap: 8px;
		input {
			flex: 1;
		}
	}
	.tools-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 16px;
		margin-bottom: 8px;
		flex-shrink: 0;
	}
	.empty-state {
		padding: 32px;
		text-align: center;
		color: var(--text-dim);
		font-size: 0.8rem;
	}
	.input.sm {
		padding: 8px 12px;
		font-size: 0.85rem;
	}
	.btn.sm {
		padding: 8px;
		font-size: 0.85rem;
	}

	.section-header-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		h3 {
			margin: 0;
			color: var(--text-main);
			font-size: 1.1rem;
		}
	}
</style>
