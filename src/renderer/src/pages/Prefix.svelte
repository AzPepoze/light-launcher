<script lang="ts">
	import { ScanProtonVersions, GetRunningSessions, GetAllGames, GetPrefixStats } from "@lib/api";
	import * as core from "@shared";
	import PrefixList from "@components/prefix/PrefixList.svelte";
	import PrefixTools from "@components/prefix/PrefixTools.svelte";
	import PrefixHeroPanel from "@components/prefix/PrefixHeroPanel.svelte";
	import ConfigForm from "@components/shared/ConfigForm.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { createLaunchOptions } from "@lib/formService";
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
	let activeTab: "tools" | "config" | "advanced" = "tools";
	let runningMap: Record<string, string[]> = {};
	let runningPoll: ReturnType<typeof setInterval> | null = null;
	let prefixStatsMap: Record<string, core.PrefixStats> = {};
	let prefixGameCounts: Record<string, number> = {};

	$: if (chosenProtonName) {
		chosenProton =
			protonVersions.find((t) => t.DisplayName === chosenProtonName) ||
			null;
	}

	$: currentPrefixName = prefixPath.startsWith(baseDir)
		? prefixPath.replace(baseDir + "/", "")
		: prefixPath.split("/").filter(Boolean).pop() || "Custom";

	function formatBytes(bytes: number | null | undefined): string {
		if (bytes == null) return "--";
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
		return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
	}

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

	onMount(async () => {
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

				<section class="panel tab-panel-wrapper">
					<div class="tabs" role="tablist">
						<button
							type="button"
							class="tab"
							class:active={activeTab === 'tools'}
							on:click={() => (activeTab = 'tools')}
							role="tab"
							aria-selected={activeTab === 'tools'}
						>
							<span class="material-icons">build</span>
							Wine Tools
						</button>
						<button
							type="button"
							class="tab"
							class:active={activeTab === 'config'}
							on:click={() => (activeTab = 'config')}
							role="tab"
							aria-selected={activeTab === 'config'}
						>
							<span class="material-icons">tune</span>
							Defaults
						</button>
						<button
							type="button"
							class="tab"
							class:active={activeTab === 'advanced'}
							on:click={() => (activeTab = 'advanced')}
							role="tab"
							aria-selected={activeTab === 'advanced'}
						>
							<span class="material-icons">settings</span>
							Advanced
						</button>
						<span class="tabs-meta">6 tools · 5 toggles</span>
					</div>

					{#if activeTab === 'tools'}
						<div class="tab-content" role="tabpanel">
							<p class="tab-hint">Launch utilities inside the selected prefix — uses your selected Proton + <span class="kbd">umu-run</span></p>
							<PrefixTools {runningToolName} onRunTool={runTool} />
						</div>
					{:else if activeTab === 'config'}
						<div class="tab-content" role="tabpanel">
							<header class="panel-header" style="margin-bottom: 16px">
								<span class="material-icons panel-icon">tune</span>
								<div>
									<h3>Default configuration</h3>
									<p>Applied when launching games with this prefix</p>
								</div>
							</header>
							<ConfigForm bind:options={prefixOptions} />
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
						</div>
					{:else}
						<div class="tab-content" role="tabpanel">
							<div class="advanced-grid">
								<div class="advanced-card">
									<h4>Symlinks</h4>
									<p>Repair broken prefix symlinks after moving storage.</p>
									<button class="btn sm" type="button">Repair</button>
								</div>
								<div class="advanced-card">
									<h4>Duplicate Prefix</h4>
									<p>Clone config + Wine bottle into a new prefix.</p>
									<button class="btn sm" type="button" on:click={() => {
										const base = currentPrefixName.replace(/[^a-zA-Z0-9_-]/g, '') || 'Default';
										newPrefixName = `${base}-copy`;
										handleCreatePrefix();
									}}>Duplicate "{currentPrefixName}"</button>
								</div>
							</div>
						</div>
					{/if}
				</section>
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

		.material-icons {
			font-size: 28px;
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

	.tab-panel-wrapper {
		padding: 14px 16px 18px;
	}

	.tabs {
		display: flex;
		align-items: flex-end;
		gap: 8px;
		border-bottom: 2px solid var(--glass-border);
		padding-bottom: 0;
		margin-bottom: 18px;
		flex-shrink: 0;
	}

	.tab {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 10px 16px;
		border-radius: 12px 12px 0 0;
		border: 2px solid transparent;
		border-bottom: none;
		background: transparent;
		color: var(--text-muted);
		font-weight: 800;
		font-size: 0.8rem;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		cursor: pointer;
		transition: all var(--transition-fast);

		.material-icons {
			font-size: 18px;
		}

		&:hover {
			color: var(--text-main);
		}

		&.active {
			background: var(--bg-surface);
			border-color: var(--glass-border);
			color: var(--text-main);
			transform: translateY(2px);
		}
	}

	.tabs-meta {
		margin-left: auto;
		align-self: center;
		font-size: 0.7rem;
		font-weight: 700;
		color: var(--text-muted);
		background: var(--bg-elevated);
		padding: 4px 8px;
		border-radius: var(--radius-pill);
		border: 1px solid var(--glass-border);
		white-space: nowrap;
	}

	.tab-content {
		animation: fadeIn 150ms ease;
	}

	.tab-hint {
		margin: 0 0 12px;
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--text-muted);

		.kbd {
			font-family: monospace;
			font-size: 0.7rem;
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.1);
			padding: 2px 6px;
			border-radius: 6px;
			color: var(--text-dim);
		}
	}

	.advanced-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}

	.advanced-card {
		padding: 16px;
		background: var(--bg-elevated);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);

		h4 {
			margin: 0 0 6px;
			font-size: 0.85rem;
			font-weight: 800;
			letter-spacing: 0.5px;
			text-transform: uppercase;
		}

		p {
			margin: 0 0 10px;
			font-size: 0.78rem;
			font-weight: 600;
			color: var(--text-muted);
			line-height: 1.5;
		}
	}

	.save-bar {
		display: flex;
		justify-content: flex-end;
		padding: 16px 0 4px;
		margin-top: 8px;
		z-index: 1;
	}

	.save-btn {
		min-width: 180px;
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
			max-height: 320px;
			min-height: 0;
			overflow: hidden;
		}

		.prefix-main {
			height: auto;
			overflow: visible;
			padding-right: 0;
			padding-bottom: 0;
		}

		.advanced-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
