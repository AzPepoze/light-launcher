<script lang="ts">
	import Badge from "@components/shared/Badge.svelte";
	import Button from "@components/shared/Button.svelte";
	import CoverImage from "@components/shared/CoverImage.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import Notice from "@components/shared/Notice.svelte";
	import ProgressBar from "@components/shared/ProgressBar.svelte";
	import Section from "@components/shared/Section.svelte";
	import Tabs from "@components/shared/Tabs.svelte";
	import { notifications } from "@stores/notificationStore";
	import { onMount } from "svelte";
	import { SvelteMap } from "svelte/reactivity";
	import { ago } from "../../lib/features/format";
	import { downloadsApi } from "../../lib/features/downloadsApi";
	import { libraryApi } from "../../lib/features/libraryApi";
	import { errorMessage, listen } from "../../lib/features/rpc";
	import type { Game, LibrarySource } from "../../lib/features/types";
	import { armEndSentinel } from "./endSentinel";
	import {
		ALL_SOURCES,
		UPDATE_CHECK_EVENT,
		backendResult,
		checkedSummary,
		emptyCounts,
		gamesOfSource,
		gamesWithUpdates,
		groupBySource,
		updateSourceTabs,
		historyCover,
		historyStep,
		issueLevel,
		issueMessage,
		lastCheckedLabel,
		latestCheckByGame,
		liveHistoryItem,
		matchesFilter,
		newestChangelog,
		nextCheckLabel,
		pagingCursor,
		pickUpdateFile,
		progressLabel,
		progressValue,
		runMessage,
		skippedNote,
		tally,
		versionPair,
		visibleHistory,
		type HistoryFilter,
		type RunCounts,
		type UpdateCheckEvent,
		type UpdateCheckResult,
		type UpdateHistoryItem,
		type UpdateIssue,
		type UpdateStatus
	} from "./updatesView";

	let {
		sources = [],
		onStatus
	}: {
		/** The library sources, for the labels and the source tabs. */
		sources?: LibrarySource[];
		/** Called with each status the page reads, so the Library tab can show the count. */
		onStatus?: (status: UpdateStatus) => void;
	} = $props();

	const HISTORY_PAGE = 25;
	const REFRESH_DELAY_MS = 300;

	const FILTERS: { value: HistoryFilter; label: string }[] = [
		{ value: "attention", label: "Updates and failures" },
		{ value: "all", label: "All checks" },
		{ value: "update", label: "Updates" },
		{ value: "error", label: "Failed" },
		{ value: "current", label: "Up to date" }
	];

	interface RunProgress {
		done: number;
		total: number;
		current: string | null;
		skipped: number;
		issues: UpdateIssue[];
	}

	interface RunSummary {
		total: number;
		counts: RunCounts;
		skipped: number;
	}

	let status = $state<UpdateStatus | null>(null);
	let loadError = $state("");
	let source = $state(ALL_SOURCES);
	/** The games of the chosen source. Null while it loads or for all sources. */
	let checking = $state(false);
	let progress = $state<RunProgress | null>(null);
	let counts = $state<RunCounts>(emptyCounts());
	let summary = $state<RunSummary | null>(null);
	let updatingId = $state<number | null>(null);
	let filter = $state<HistoryFilter>("attention");
	let history = $state<UpdateHistoryItem[]>([]);
	let historyError = $state("");
	let historyLoading = $state(false);
	let historyHasMore = $state(false);
	let sentinel = $state<HTMLElement>();
	let expanded = new SvelteMap<number, boolean>();
	let changelogs = new SvelteMap<number, string>();
	let changelogLoading = new SvelteMap<number, boolean>();
	let liveId = 0;
	let refreshTimer: ReturnType<typeof setTimeout> | undefined;

	const games = $derived(status?.games ?? []);
	const lastCheckedAt = $derived(status?.lastCheckedAt ?? null);
	const nextRunAt = $derived(status?.nextRunAt ?? null);
	const nextCheck = $derived(nextCheckLabel(nextRunAt));
	const labels = $derived(Object.fromEntries(sources.map((item) => [item.source, item.label])));
	const visibleGames = $derived(gamesOfSource(games, source));
	const updateSourceIds = $derived(sources.filter((item) => item.updateCheck).map((item) => item.source));
	const sourceTabs = $derived(updateSourceTabs(sources, updateSourceIds, games));
	const rows = $derived(visibleHistory(history, filter));
	const groups = $derived(groupBySource(visibleGames, labels, latestCheckByGame(history)));
	const doneCount = $derived(progress?.done ?? 0);
	const totalCount = $derived(progress?.total ?? 0);

	$effect(() => armEndSentinel(sentinel, historyHasMore && !historyLoading && !historyError, () => void loadHistory(false)));

	onMount(() => {
		const stopListening = listen<UpdateCheckEvent>(UPDATE_CHECK_EVENT, handleEvent);
			void loadStatus();
		void loadHistory(true);
		return () => {
			stopListening();
			clearTimeout(refreshTimer);
		};
	});

	/** Reads the saved results. Never starts a check. */
	async function loadStatus() {
		try {
			status = await libraryApi.updateStatus();
			onStatus?.(status);
			loadError = "";
			if (status.running && !checking) {
				checking = true;
				progress = status.progress
					? { done: status.progress.done, total: status.progress.total, current: null, skipped: 0, issues: [] }
					: null;
			}
		} catch (err) {
			loadError = errorMessage(err);
		}
	}


	/** Picks a source tab. The server pages that source's history, so its pages stay full. */
	function selectSource(next: string) {
		source = next;
		history = [];
		historyHasMore = false;
		void loadHistory(true);
	}

	/** Starts a check. A press while one runs does nothing. */
	async function check() {
		if (checking) return;
		checking = true;
		progress = null;
		counts = emptyCounts();
		summary = null;
		try {
			await libraryApi.checkUpdates();
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			checking = false;
			await loadStatus();
		}
	}

	function handleEvent(event: UpdateCheckEvent) {
		if (event.state === "running") {
			recordRunning(event);
			return;
		}
		finishRun(event);
	}

	function recordRunning(event: UpdateCheckEvent) {
		if (event.done === 0) {
			counts = emptyCounts();
			summary = null;
		}
		checking = true;
		progress = {
			done: event.done,
			total: event.total,
			current: event.current?.title ?? null,
			skipped: event.skipped,
			issues: event.issues
		};
		if (event.result) recordResult(event.result);
	}

	function recordResult(result: UpdateCheckResult) {
		counts = tally(counts, result);
		if (result.result === "update") scheduleRefresh();
		if (matchesFilter(filter, result.result)) {
			liveId -= 1;
			history = [liveHistoryItem(result, liveId, new Date().toISOString()), ...history];
		}
	}

	function finishRun(event: UpdateCheckEvent) {
		checking = false;
		progress = null;
		summary = { total: event.total, counts, skipped: event.skipped };
		for (const issue of event.issues) {
			const send = issueLevel(issue) === "warning" ? notifications.warning : notifications.error;
			send(issueMessage(issue));
		}
		const report = counts.failed > 0 ? notifications.info : notifications.success;
		report(runMessage(event.total, counts));
		void loadStatus();
		void loadHistory(true);
	}

	function scheduleRefresh() {
		clearTimeout(refreshTimer);
		refreshTimer = setTimeout(() => void loadStatus(), REFRESH_DELAY_MS);
	}

	function setFilter(next: HistoryFilter) {
		filter = next;
		void loadHistory(true);
	}

	async function loadHistory(reset: boolean) {
		if (!reset && (historyLoading || !historyHasMore)) return;
		historyLoading = true;
		historyError = "";
		const before = reset ? 0 : pagingCursor(history);
		try {
			const page = await libraryApi.updateHistory({
				limit: HISTORY_PAGE,
				before,
				result: backendResult(filter),
				source: source === ALL_SOURCES ? "" : source
			});
			history = reset ? page : [...history, ...page];
			historyHasMore = page.length === HISTORY_PAGE;
		} catch (err) {
			historyError = errorMessage(err);
		} finally {
			historyLoading = false;
		}
	}

	async function toggleChangelog(game: Game) {
		const open = !expanded.get(game.id);
		expanded.set(game.id, open);
		if (!open || changelogs.has(game.id)) return;
		changelogLoading.set(game.id, true);
		try {
			const detail = await libraryApi.get(game.id);
			changelogs.set(game.id, newestChangelog(detail.changes) ?? "");
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			changelogLoading.set(game.id, false);
		}
	}

	async function update(game: Game) {
		updatingId = game.id;
		try {
			const detail = await libraryApi.get(game.id);
			const choice = pickUpdateFile(detail.files);
			if (choice.kind === "single") {
				await downloadsApi.enqueueFile(game.id, choice.file.id);
				notifications.success(`Added "${choice.file.label}" to downloads.`);
			} else if (choice.kind === "none") {
				notifications.info(`${game.title} has no saved download. Add one in Library.`);
			} else {
				notifications.info(`${game.title} has several downloads. Pick one in Library.`);
			}
		} catch (err) {
			notifications.error(errorMessage(err));
		} finally {
			updatingId = null;
		}
	}
</script>

<div class="toolbar">
	<p class="last-checked">
		{lastCheckedLabel(lastCheckedAt)}
		{#if nextCheck}
			<span class="next-check"> · {nextCheck}</span>
		{/if}
	</p>
	<Button variant="primary" size="sm" icon="refresh" loading={checking} onclick={() => check()}>
		{checking ? "Checking..." : "Check now"}
	</Button>
</div>

{#if sourceTabs.length > 1}
	<div class="source-tabs">
		<Tabs tabs={sourceTabs} value={source} onChange={selectSource} aria-label="Update sources" />
	</div>
{/if}

{#if checking}
	<div class="progress" role="status">
		<ProgressBar
			value={progressValue(doneCount, totalCount)}
			label={progressLabel(doneCount, totalCount)}
			height={4}
			active
		/>
		<p class="progress-text">
			<span>{progressLabel(doneCount, totalCount)}</span>
			{#if progress?.current}
				<span class="current">{progress.current}</span>
			{/if}
		</p>
		{#if progress && progress.skipped > 0}
			<p class="quiet">{skippedNote(progress.skipped, progress.issues)}</p>
		{/if}
	</div>
{:else if summary}
	<div class="summary">
		<p class="summary-text">{checkedSummary(summary.total)}</p>
		{#if summary.counts.updates > 0}
			<Badge tone="accent">{summary.counts.updates} with updates</Badge>
		{/if}
		{#if summary.counts.failed > 0}
			<Badge tone="danger">{summary.counts.failed} failed</Badge>
		{/if}
		{#if summary.skipped > 0}
			<p class="quiet">{skippedNote(summary.skipped, [])}</p>
		{/if}
	</div>
{/if}

{#if loadError && !status}
	<Notice tone="error">
		<span>Could not load updates.</span>
		<Button variant="text" size="sm" icon="refresh" onclick={() => loadStatus()}>Retry</Button>
	</Notice>
{/if}

{#if !status}
	{#if !loadError}
		<LoadingState label="Loading updates" />
	{/if}
{:else if games.length === 0}
	<EmptyState
		icon={lastCheckedAt ? "task_alt" : "update"}
		title={lastCheckedAt ? "Everything is up to date" : "Never checked"}
		hint={lastCheckedAt ? lastCheckedLabel(lastCheckedAt) : "Check now to look for new releases."}
	>
		<Button variant="primary" size="sm" icon="refresh" loading={checking} onclick={() => check()}>
			Check now
		</Button>
	</EmptyState>
{:else if visibleGames.length === 0}
	<EmptyState icon="task_alt" title="No updates from this source." />
{:else}
	<div class="groups">
		{#each groups as group (group.source)}
			<Section title={group.label} help={gamesWithUpdates(group.games.length)}>
				{#snippet actions()}
					<Badge tone="accent">{group.games.length}</Badge>
				{/snippet}
				<ul class="rows" aria-label="Games with updates from {group.label}">
					{#each group.games as game (game.id)}
						{@const pair = versionPair(game)}
						{@const open = expanded.get(game.id) ?? false}
						<li class="row">
							<div class="thumb">
								{#if game.cover || game.coverLocal}
									<CoverImage {game} />
								{/if}
							</div>
							<div class="main">
								<p class="title">{game.title}</p>
								<div class="versions">
									<Badge>{pair.from}</Badge>
									<Icon name="arrow_forward" size={16} />
									<Badge tone="accent">{pair.to}</Badge>
								</div>
								<div class="changelog" class:open>
									<div class="changelog-inner">
										{#if changelogLoading.get(game.id)}
											<p class="muted">Loading notes...</p>
										{:else if changelogs.get(game.id)}
											<p class="notes">{changelogs.get(game.id)}</p>
										{:else}
											<p class="muted">No notes for this update.</p>
										{/if}
									</div>
								</div>
							</div>
							<div class="actions">
								<Button variant="text" size="sm" onclick={() => toggleChangelog(game)}>
									<Icon name="expand_more" size={18} class="chevron {open ? 'open' : ''}" />
									Notes
								</Button>
								<Button
									variant="primary"
									size="sm"
									loading={updatingId === game.id}
									disabled={updatingId !== null}
									onclick={() => update(game)}
								>
									Update
								</Button>
							</div>
						</li>
					{/each}
				</ul>
			</Section>
		{/each}
	</div>
{/if}

<Section title="History">
	<div class="filters" role="group" aria-label="Filter checks">
		{#each FILTERS as option (option.value)}
			<Button
				variant={filter === option.value ? "primary" : "secondary"}
				size="sm"
				onclick={() => setFilter(option.value)}
			>
				{option.label}
			</Button>
		{/each}
	</div>
	<ol class="timeline">
		{#each rows as entry (entry.id)}
			{@const step = historyStep(entry)}
			<li class="event" class:fresh={entry.id < 0}>
				<div class="thumb small">
					<CoverImage game={historyCover(entry)} />
				</div>
				<span class="marker {step.tone}">
					<Icon name={step.icon} size={18} filled />
				</span>
				<div class="event-body">
					<p class="event-title">{entry.title}</p>
					<p class="event-detail">
						{#if entry.result === "update"}
							<span>{entry.from ?? "?"}</span>
							<Icon name="arrow_forward" size={14} />
							<span>{entry.to ?? "?"}</span>
						{:else}
							{step.title}
						{/if}
					</p>
				</div>
				<span class="when">{ago(entry.checkedAt)}</span>
			</li>
		{:else}
			{#if !historyLoading && !historyHasMore}
				<p class="muted">Nothing here yet.</p>
			{/if}
		{/each}
	</ol>
	<div class="end" bind:this={sentinel}>
		{#if historyLoading}
			<LoadingState label="Loading more checks" />
		{:else if historyError}
			<Notice tone="error">
				<span>Could not load checks.</span>
				<Button variant="text" size="sm" icon="refresh" onclick={() => loadHistory(history.length === 0)}>
					Retry
				</Button>
			</Notice>
		{:else if !historyHasMore && history.length > 0}
			<p class="muted">No more checks.</p>
		{/if}
	</div>
</Section>

<style lang="scss">
	.toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 12px;
	}

	.last-checked {
		margin: 0;
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.next-check {
		color: var(--text-muted);
	}

	.source-tabs {
		margin-bottom: 16px;
	}

	.progress {
		display: flex;
		flex-direction: column;
		gap: 6px;
		margin: 0 0 20px;
	}

	.progress-text {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 12px;
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.current {
		color: var(--text-main);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 100%;
	}

	.quiet {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.summary {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px 12px;
		margin: 0 0 20px;
	}

	.summary-text {
		margin: 0;
		font-size: 0.9rem;
		color: var(--text-main);
	}

	.groups {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-bottom: 16px;
	}

	.rows,
	.timeline {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.row {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 10px 12px;
		border-radius: var(--radius-md);
		background: var(--bg-elevated);
	}

	.thumb {
		width: 56px;
		height: 56px;
		flex-shrink: 0;
		border-radius: var(--radius-sm);
		overflow: hidden;
		background: var(--bg-surface);

		&.small {
			width: 36px;
			height: 36px;
		}
	}

	.main {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.title {
		margin: 0;
		font-weight: 700;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.versions {
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--text-muted);
	}

	.changelog {
		display: grid;
		grid-template-rows: 0fr;
		transition: grid-template-rows var(--motion-base) var(--ease-out);

		&.open {
			grid-template-rows: 1fr;
		}
	}

	.changelog-inner {
		min-height: 0;
		overflow: hidden;
	}

	.notes {
		margin: 4px 0 0;
		padding-top: 6px;
		color: var(--text-muted);
		font-size: 0.85rem;
		white-space: pre-line;
	}

	.muted {
		margin: 4px 0 0;
		padding-top: 6px;
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-shrink: 0;
	}

	.actions :global(.chevron) {
		transition: transform var(--motion-base) var(--ease-out);
	}

	.actions :global(.chevron.open) {
		transform: rotate(180deg);
	}

	.filters {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 12px;
	}

	.timeline {
		gap: 0;
	}

	.event {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 8px 4px;
		font-size: 0.85rem;

		&.fresh {
			animation: updates-fade-in var(--motion-base) var(--ease-out) both;
		}
	}

	.marker {
		display: inline-flex;
		color: var(--text-muted);

		&.success {
			color: var(--success);
		}

		&.accent {
			color: var(--accent-primary);
		}

		&.danger {
			color: var(--danger);
		}
	}

	.event-body {
		flex: 1;
		min-width: 0;
	}

	.event-title {
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.event-detail {
		display: flex;
		align-items: center;
		gap: 6px;
		margin: 2px 0 0;
		color: var(--text-muted);
	}

	.when {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.end {
		display: flex;
		justify-content: center;
		min-height: 8px;
		margin-top: 12px;
	}

	@keyframes updates-fade-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.changelog,
		.actions :global(.chevron) {
			transition: none;
		}

		.event.fresh {
			animation: none;
		}
	}
</style>
