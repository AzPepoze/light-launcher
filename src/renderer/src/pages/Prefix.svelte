<script lang="ts">
	import { ScanProtonVersions } from "@lib/api";
	import * as core from "@shared";
	import PrefixList from "@components/prefix/PrefixList.svelte";
	import PrefixTools from "@components/prefix/PrefixTools.svelte";
	import PrefixHeroPanel from "@components/prefix/PrefixHeroPanel.svelte";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { createLaunchOptions } from "@lib/formService";
	import { createLogger } from "@lib/logger";
	import * as service from "@lib/prefixService";
	import { onMount } from "svelte";

	const log = createLogger("Prefix");

	let availablePrefixes: string[] = [];
	let baseDir = "";
	let prefixPath = "";
	let chosenProton: core.ProtonTool | null = null;
	let chosenProtonName = "";
	let protonVersions: core.ProtonTool[] = [];
	let protonDisplayNames: string[] = [];
	let newPrefixName = "";
	let isPageLoading = true;
	let isSaving = false;
	let runningToolName = "";
	let isToolRunning = false;

	let prefixOptions: core.LaunchOptions = createLaunchOptions();

	$: if (chosenProtonName) {
		chosenProton =
			protonVersions.find((t) => t.DisplayName === chosenProtonName) ||
			null;
	}

	$: currentPrefixName = prefixPath.startsWith(baseDir)
		? prefixPath.replace(baseDir + "/", "")
		: prefixPath.split("/").filter(Boolean).pop() || "Custom";

	async function refreshPrefixes(autoSelect = true) {
		const data = await service.getPrefixData();
		availablePrefixes = data.availablePrefixes;
		baseDir = data.baseDir;

		if (autoSelect) {
			if (!prefixPath && availablePrefixes.length > 0) {
				await selectPrefix(availablePrefixes[0]);
			} else if (!prefixPath) {
				prefixPath = baseDir + "/Default";
			}
		}
	}

	onMount(async () => {
		log.info("onMount starting");
		try {
			const tools = await ScanProtonVersions();
			protonVersions = tools;
			protonDisplayNames = protonVersions.map((t) => t.DisplayName);
			await refreshPrefixes();
		} catch (err) {
			log.error("onMount error", err);
		} finally {
			isPageLoading = false;
		}
	});

	async function selectPrefix(name: string) {
		log.info("selectPrefix called", { name });
		const result = await service.getPrefixConfig(name, baseDir);
		prefixPath = result.path;
		if (result.options) {
			prefixOptions = { ...prefixOptions, ...result.options };
			if (result.selectedProton) {
				chosenProtonName = result.selectedProton;
			} else if (protonVersions.length > 0) {
				chosenProtonName = protonVersions[0].DisplayName;
			}
		} else {
			prefixOptions = createLaunchOptions();
			if (protonVersions.length > 0 && !chosenProton) {
				chosenProtonName = protonVersions[0].DisplayName;
			}
		}
	}

	async function handleSaveConfig() {
		if (isSaving) return;
		isSaving = true;
		try {
			await service.savePrefixDefaults(
				prefixPath,
				prefixOptions,
				chosenProton,
			);
		} finally {
			isSaving = false;
		}
	}

	async function handleCreatePrefix() {
		if (!newPrefixName.trim()) return;
		const name = newPrefixName.trim();
		await service.createNewPrefix(name);
		newPrefixName = "";
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
		if (isToolRunning) return;
		isToolRunning = true;
		runningToolName = tool;

		try {
			await service.executePrefixTool(
				prefixPath,
				tool,
				chosenProton?.Path || "",
			);
		} finally {
			setTimeout(() => {
				isToolRunning = false;
				runningToolName = "";
			}, 500);
		}
	}
</script>

<div class="prefix-page">
	<PageHeader
		title="Prefix Manager"
		icon="folder_shared"
		subtitle="Wine/Proton prefixes, runtime, and default launch options"
	/>

	{#if isPageLoading}
		<div class="loading-state">
			<span class="material-icons spin">progress_activity</span>
			<span>Loading prefixes…</span>
		</div>
	{:else}
		<div class="prefix-layout">
			<aside class="prefix-sidebar">
				<PrefixList
					{availablePrefixes}
					{currentPrefixName}
					bind:newPrefixName
					onSelectPrefix={selectPrefix}
					onCreatePrefix={handleCreatePrefix}
					onRemovePrefix={handleRemovePrefix}
				/>
			</aside>

			<main class="prefix-main">
				<PrefixHeroPanel
					{currentPrefixName}
					{protonDisplayNames}
					bind:chosenProtonName
					bind:prefixPath
				/>

				<section class="panel tools-panel">
					<header class="panel-header">
						<span class="material-icons panel-icon">build</span>
						<div>
							<h3>Wine tools</h3>
							<p>Launch utilities inside the selected prefix</p>
						</div>
					</header>
					<PrefixTools {runningToolName} onRunTool={runTool} />
				</section>

				<section class="panel config-panel">
					<header class="panel-header">
						<span class="material-icons panel-icon">tune</span>
						<div>
							<h3>Default configuration</h3>
							<p>Applied when launching games with this prefix</p>
						</div>
					</header>
					<ConfigForm bind:options={prefixOptions} />
				</section>

				<div class="save-bar">
					<button
						class="btn primary save-btn"
						type="button"
						on:click={handleSaveConfig}
						disabled={isSaving}
					>
						<span class="material-icons btn-icon" class:spin={isSaving}
							>{isSaving ? "sync" : "save"}</span
						>
						{isSaving ? "Saving…" : "Save Defaults"}
					</button>
				</div>
			</main>
		</div>
	{/if}
</div>

<style lang="scss">
	.prefix-page {
		display: flex;
		flex-direction: column;
		gap: 0;
		min-height: 0;
	}

	.btn-icon {
		font-size: 18px;
		margin-right: 6px;
		vertical-align: middle;
	}

	.btn.primary .btn-icon {
		margin-right: 8px;
	}

	.btn.primary:disabled .material-icons.spin,
	.loading-state .spin {
		animation: spin 1.2s linear infinite;
	}

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}

	.loading-state {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		padding: 80px 24px;
		color: var(--text-muted);
		font-size: 0.95rem;
		font-weight: 700;

		.material-icons {
			font-size: 28px;
			color: var(--accent-primary);
		}
	}

	.prefix-layout {
		display: grid;
		grid-template-columns: minmax(240px, 280px) minmax(0, 1fr);
		gap: 24px;
		align-items: stretch;
	}

	.prefix-sidebar {
		height: 100%;
	}

	.prefix-main {
		display: flex;
		flex-direction: column;
		gap: 20px;
		min-width: 0;
	}

	.panel {
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-lg);
		padding: 22px 24px;
	}

	.panel-header {
		display: flex;
		align-items: flex-start;
		gap: 14px;
		margin-bottom: 20px;

		.panel-icon {
			font-size: 22px;
			color: var(--accent-primary);
			margin-top: 2px;
		}

		h3 {
			margin: 0 0 4px;
			font-size: 1rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		p {
			margin: 0;
			font-size: 0.8rem;
			font-weight: 600;
			color: var(--text-muted);
		}
	}



	.tools-panel :global(.tools-grid) {
		margin-bottom: 0;
	}

	.config-panel :global(.config-form) {
		margin-top: 0;
	}

	.save-bar {
		position: sticky;
		bottom: 0;
		display: flex;
		justify-content: flex-end;
		padding: 24px 0 8px;
		margin-top: 8px;
		z-index: 10;
		background: transparent;
		pointer-events: none;
	}

	.save-btn {
		pointer-events: auto;
		min-width: 200px;
	}

	@media (max-width: 900px) {
		.prefix-layout {
			grid-template-columns: 1fr;
		}

		.prefix-sidebar {
			position: static;
			height: auto;
			max-height: 320px;
			min-height: 0;
		}


	}
</style>
