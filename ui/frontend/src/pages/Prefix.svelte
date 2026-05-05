<script lang="ts">
	import {
		CreatePrefix,
		GetPrefixBaseDir,
		GetSystemToolsStatus,
		ListPrefixes,
		LoadPrefixConfig,
		PickFolder,
		RunPrefixTool,
		SavePrefixConfig,
		RemovePrefix,
		ScanProtonVersions,
	} from "@bindings/light-launcher/internal/app/app";
	import * as core from "@bindings/light-launcher/internal/types/models";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import PrefixList from "@components/prefix/PrefixList.svelte";
	import PrefixTools from "@components/prefix/PrefixTools.svelte";
	import { createLaunchOptions } from "@lib/formService";
	import { notifications } from "@stores/notificationStore";
	import { onMount } from "svelte";

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
		prefixOptions = createLaunchOptions();
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

	async function handleRemovePrefix(name: string) {
		if (name === "Default") {
			notifications.add("Cannot delete Default prefix", "error");
			return;
		}
		await notifications.withNotification(
			(async () => {
				await RemovePrefix(name);
				await refreshPrefixes();
				if (availablePrefixes.length > 0) {
					await selectPrefix(availablePrefixes[0]);
				} else {
					prefixPath = baseDir + "/Default";
				}
			})(),
			{
				success: `Deleted prefix "${name}"`,
				error: "Failed to delete prefix",
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
		<PrefixList
			bind:availablePrefixes
			{currentPrefixName}
			bind:newPrefixName
			onSelectPrefix={selectPrefix}
			onCreatePrefix={handleCreatePrefix}
			onRemovePrefix={handleRemovePrefix}
		/>
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

			<PrefixTools {runningToolName} onRunTool={runTool} />

			<div class="config-card glass">
				<div class="section-header-row">
					<h3>Default Configuration</h3>
					<button class="btn primary sm" on:click={handleSaveConfig}
						>Save Defaults</button
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
