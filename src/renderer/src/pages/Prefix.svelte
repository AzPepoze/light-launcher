<script lang="ts">
	import { ScanProtonVersions, GetRunningSessions, GetAllGames, GetPrefixStats } from "@lib/api";
	import * as core from "@shared";
	import PrefixList from "@components/prefix/PrefixList.svelte";
	import PrefixTools from "@components/prefix/PrefixTools.svelte";
	import PrefixHeroPanel from "@components/prefix/PrefixHeroPanel.svelte";
	import PrefixDrives from "@components/prefix/PrefixDrives.svelte";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import SettingRow from "@components/shared/SettingRow.svelte";
	import Tabs from "@components/shared/Tabs.svelte";
	import type { TabItem } from "@components/shared/tabs";
	import { createLaunchOptions } from "@lib/formService";
	import { formatBytes } from "@lib/features/format";
	import { confirmLeave, leaveGuard } from "@stores/leaveGuard";
	import { createLogger } from "@lib/logger";
	import * as service from "@lib/prefixService";
	import { onMount, onDestroy } from "svelte";

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
	/** The options and Proton version as last loaded or saved; the dirty check compares against these. */
	let savedOptions: core.LaunchOptions = createLaunchOptions();
	let savedProtonName = "";
	$: hasUnsavedChanges = isDirty(prefixOptions, chosenProtonName, savedOptions, savedProtonName);
	type PrefixTab = "tools" | "config" | "advanced" | "drives";
	const prefixTabs: TabItem[] = [
		{ id: "tools", label: "Wine Tools", icon: "build" },
		{ id: "config", label: "Defaults", icon: "tune" },
		{ id: "advanced", label: "Advanced", icon: "settings" },
		{ id: "drives", label: "Drives", icon: "hard_drive" }
	];
	let activeTab: PrefixTab = "tools";

	function selectTab(id: string) {
		activeTab = id as PrefixTab;
	}
	let runningMap: Record<string, string[]> = {};
	let runningPoll: ReturnType<typeof setInterval> | null = null;
	let prefixStatsMap: Record<string, core.PrefixStats> = {};
	let prefixGameCounts: Record<string, number> = {};

	$: if (chosenProtonName) {
		chosenProton =
			protonVersions.find((t) => t.DisplayName === chosenProtonName) ||
			null;
	}

	function getToolBlockReason(proton: core.ProtonTool | null, path: string): string {
		if (!proton) return "Choose a Proton version first";
		if (!path.trim()) return "Choose a prefix folder first";
		return "";
	}

	$: toolBlockReason = getToolBlockReason(chosenProton, prefixPath);

	$: currentPrefixName = prefixPath.startsWith(baseDir)
		? prefixPath.replace(baseDir + "/", "")
		: prefixPath.split("/").filter(Boolean).pop() || "Custom";

	function formatCreated(ts: number | null | undefined): string {
		if (!ts) return "Created --";
		try {
			const d = new Date(ts);
			return `Created ${d.toLocaleDateString("en-GB", { day: "2-digit", month: "short" })}`;
		} catch {
			return "Created --";
		}
	}

	function formatGamesCount(n: number): string {
		if (n === 0) return "0 games";
		if (n === 1) return "1 game linked";
		return `${n} games linked`;
	}

	$: currentStats = prefixStatsMap[currentPrefixName];
	$: currentGameCount = prefixGameCounts[currentPrefixName] ?? 0;
	$: heroInfo = {
		size: formatBytes(currentStats?.sizeBytes),
		games: formatGamesCount(currentGameCount),
		created: formatCreated(currentStats?.createdAt)
	};

	function extractPrefixName(prefixPath: string): string {
		if (!prefixPath) return "";
		const parts = prefixPath.split("/").filter(Boolean);
		return parts[parts.length - 1] || "";
	}

	async function refreshRunningMap() {
		try {
			const [sessions, games] = await Promise.all([GetRunningSessions(), GetAllGames()]);
			const gamePrefixMap = new Map<string, string>();
			for (const g of games || []) {
				const pName = extractPrefixName((g as core.GameInfo).config?.PrefixPath || "");
				if (g.path) gamePrefixMap.set(g.path, pName);
				// also map by name fallback
				if ((g as core.GameInfo).name) gamePrefixMap.set((g as core.GameInfo).name, pName);
			}
			const next: Record<string, string[]> = {};
			for (const s of sessions || []) {
				const pName = gamePrefixMap.get(s.gamePath) || gamePrefixMap.get(s.gameName) || "";
				if (!pName) continue;
				if (!next[pName]) next[pName] = [];
				next[pName].push(s.gameName || s.gamePath.split("/").pop() || "Game");
			}
			runningMap = next;
		} catch (e) {
			// silent - running check is best effort
		}
	}

	async function refreshPrefixMeta() {
		try {
			const games = await GetAllGames();
			const counts: Record<string, number> = {};
			for (const g of games || []) {
				const pName = extractPrefixName((g as core.GameInfo).config?.PrefixPath || "");
				if (!pName) continue;
				counts[pName] = (counts[pName] || 0) + 1;
			}
			prefixGameCounts = counts;

			const statsEntries = await Promise.all(
				availablePrefixes.map(async (name) => {
					try {
						const s = await GetPrefixStats(name);
						return [name, s] as const;
					} catch {
						return [name, { name, createdAt: null, sizeBytes: null }] as const;
					}
				})
			);
			const nextMap: Record<string, core.PrefixStats> = {};
			for (const [name, stats] of statsEntries) nextMap[name] = stats;
			prefixStatsMap = nextMap;
		} catch {
			// best effort
		}
	}

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
		await refreshPrefixMeta();
	}

	function isDirty(
		options: core.LaunchOptions,
		protonName: string,
		saved: core.LaunchOptions,
		savedName: string
	): boolean {
		return protonName !== savedName || JSON.stringify(options) !== JSON.stringify(saved);
	}

	function copyOptions(options: core.LaunchOptions): core.LaunchOptions {
		return JSON.parse(JSON.stringify(options));
	}

	function markSaved() {
		savedOptions = copyOptions(prefixOptions);
		savedProtonName = chosenProtonName;
	}

	function discardChanges() {
		prefixOptions = copyOptions(savedOptions);
		chosenProtonName = savedProtonName;
	}

	onMount(async () => {
		leaveGuard.set({ isDirty: () => hasUnsavedChanges, save: handleSaveConfig });
		log.info("onMount starting");
		try {
			const tools = await ScanProtonVersions();
			protonVersions = tools;
			protonDisplayNames = protonVersions.map((t) => t.DisplayName);
			await refreshPrefixes();
			await refreshRunningMap();
			runningPoll = setInterval(refreshRunningMap, 2000);
		} catch (err) {
			log.error("onMount error", err);
		} finally {
			isPageLoading = false;
		}
	});

	onDestroy(() => {
		if (runningPoll) clearInterval(runningPoll);
		leaveGuard.set(null);
	});

	async function selectPrefix(name: string) {
		if (!(await confirmLeave())) return;
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
		markSaved();
	}

	/** Saves the options and Proton version. Resolves false when the save failed. */
	async function handleSaveConfig(): Promise<boolean> {
		if (isSaving) return false;
		isSaving = true;
		try {
			await service.savePrefixDefaults(prefixPath, prefixOptions, chosenProton);
			markSaved();
			return true;
		} catch {
			return false;
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

	async function duplicateCurrentPrefix() {
		const base = currentPrefixName.replace(/[^a-zA-Z0-9_-]/g, "") || "Default";
		newPrefixName = `${base}-copy`;
		await handleCreatePrefix();
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
		subtitle="Wine and Proton prefixes, runtimes and launch defaults"
	/>

		{#if isPageLoading}
		<div class="loading-state">
			<Icon name="progress_activity" size={28} class="spin" />
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
					{runningMap}
					prefixStats={prefixStatsMap}
				/>
			</aside>

			<main class="prefix-main">
				<PrefixHeroPanel
					{currentPrefixName}
					{protonDisplayNames}
					bind:chosenProtonName
					bind:prefixPath
					heroSize={heroInfo.size}
					heroGamesLinked={heroInfo.games}
					heroCreated={heroInfo.created}
				/>

				<section class="tab-card">
					<div class="tabs-row">
						<Tabs
							tabs={prefixTabs}
							value={activeTab}
							onChange={selectTab}
							aria-label="Prefix sections"
						/>
					</div>

					{#if activeTab === 'tools'}
						<div class="tab-content" role="tabpanel">
							<PrefixTools {runningToolName} blockReason={toolBlockReason} onRunTool={runTool} />
						</div>
					{:else if activeTab === 'config'}
						<div class="tab-content" role="tabpanel">
							<ConfigForm bind:options={prefixOptions} />
						</div>
					{:else if activeTab === 'drives'}
						<div class="tab-content" role="tabpanel">
							<PrefixDrives {prefixPath} />
						</div>
					{:else}
						<div class="tab-content" role="tabpanel">
							<SettingRow title="Repair symlinks" help="Fix broken prefix symlinks after moving storage.">
								<Button size="sm">Repair</Button>
							</SettingRow>
							<SettingRow title="Duplicate prefix" help="Copy the config and Wine files into a new prefix.">
								<Button size="sm" onclick={duplicateCurrentPrefix}>Duplicate</Button>
							</SettingRow>
						</div>
					{/if}
				</section>

				{#if hasUnsavedChanges}
					<div class="save-bar" role="status">
						<p class="save-text">Unsaved changes</p>
						<div class="save-actions">
							<Button onclick={discardChanges} disabled={isSaving}>Discard</Button>
							<Button
								variant="primary"
								onclick={handleSaveConfig}
								loading={isSaving}
								disabled={isSaving}
							>
								Save changes
							</Button>
						</div>
					</div>
				{/if}
			</main>
		</div>
	{/if}
</div>

<style lang="scss">
	.prefix-page {
		display: flex;
		flex-direction: column;
		gap: 0;
		flex: 1;
		min-height: 0;
		height: 100%;
		max-height: 100%;
	}

	.loading-state :global(.spin) {
		animation: ll-spin 1.2s linear infinite;
	}

	@keyframes fadeIn {
		from { opacity: 0; transform: translateY(4px); }
		to { opacity: 1; transform: translateY(0); }
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

		:global(.spin) {
			color: var(--accent-primary);
		}
	}

	.prefix-layout {
		display: grid;
		grid-template-columns: minmax(260px, 280px) minmax(0, 1fr);
		gap: 20px;
		flex: 1;
		min-height: 0;
		overflow: hidden;
		align-items: stretch;
	}

	.prefix-sidebar {
		height: 100%;
		min-height: 0;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		align-self: stretch;
		/* keep visible while right scrolls */
		position: sticky;
		top: 0;

		:global(.prefix-list-panel) {
			flex: 1;
			min-height: 0;
			height: 100%;
		}
	}

	.prefix-main {
		display: flex;
		flex-direction: column;
		gap: 16px;
		min-width: 0;
		height: 100%;
		min-height: 0;
		overflow-y: auto;
		overflow-x: hidden;
		padding-right: 6px;
		padding-bottom: 20px;
		scrollbar-width: thin;
		scrollbar-color: var(--glass-border-bright) transparent;

		&::-webkit-scrollbar {
			width: 8px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: var(--glass-border-bright);
			border-radius: 10px;
		}
	}

	.tab-card {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 16px 20px 20px;
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-lg);
	}

	.tabs-row {
		flex-shrink: 0;
	}

	.tab-content {
		animation: fadeIn 150ms ease;
	}

	/* Sits at the bottom of the column, and stays in view when the tab content scrolls. */
	.save-bar {
		position: sticky;
		bottom: 12px;
		margin-top: auto;
		z-index: 2;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
		padding: 12px 14px 12px 20px;
		background: var(--bg-elevated);
		border: 1px solid var(--glass-border-bright);
		border-radius: var(--radius-lg);
		box-shadow: var(--glass-shadow);
	}

	.save-text {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text-main);
	}

	.save-actions {
		display: flex;
		gap: 8px;
	}

	@media (max-width: 900px) {
		.prefix-page {
			height: auto;
			min-height: 0;
		}

		.prefix-layout {
			grid-template-columns: 1fr;
			overflow: visible;
			flex: none;
		}

		.prefix-sidebar {
			position: static;
			height: auto;
			min-width: 0;
		}

		.prefix-main {
			height: auto;
			overflow: visible;
			padding-right: 0;
			padding-bottom: 0;
		}
	}
</style>
