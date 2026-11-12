<script lang="ts">
	import { onMount, untrack } from "svelte";
	import { libraryApi } from "../../lib/features/libraryApi";
	import { errorMessage } from "../../lib/features/rpc";
	import type { FilterOption, Game, LibrarySource, SidebarGroup, SidebarOption } from "../../lib/features/types";
	import {
		browseQuery,
		hasBrowseFilters,
		initialBrowseFilters,
		setBrowseQuery,
		setMulti,
		setOrder,
		setSelect,
		setSort,
		setToggle
	} from "./browseFilters";
	import GameTile from "./GameTile.svelte";
	import { tileShapeFor } from "./sourceTileShape";

	/** Typing in a group search box waits this long before asking the source. */
	const SEARCH_DELAY_MS = 150;

	/** One rendered choice: fixed options carry an icon, dynamic ones a count. */
	type BrowseOption = Pick<SidebarOption, "value" | "label" | "icon" | "count">;

	let {
		source,
		onOpenGame
	}: {
		source: LibrarySource;
		onOpenGame: (game: Game) => void;
	} = $props();

	let filters = $state(initialBrowseFilters(untrack(() => source.sidebar)));
	let games = $state<Game[]>([]);
	let total = $state<number | null>(null);
	let nextCursor = $state<string | null>(null);
	let loading = $state(false);
	let error = $state("");
	let optionError = $state<Record<string, string>>({});
	let dynamic = $state<Record<string, FilterOption[]>>({});
	let optionSearch = $state<Record<string, string>>({});

	let generation = 0;
	let optionGenerations: Record<string, number> = {};
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	let optionTimers: Record<string, ReturnType<typeof setTimeout>> = {};

	const shape = $derived(tileShapeFor(source.tile));
	const sortGroup = $derived(source.sidebar.find((group) => group.type === "sort"));

	onMount(() => {
		void loadPage();
		for (const group of source.sidebar) {
			if (group.dynamic) void loadOptions(group, "");
		}
		return () => {
			clearTimeout(searchTimer);
			for (const timer of Object.values(optionTimers)) clearTimeout(timer);
		};
	});

	async function loadPage(cursor = ""): Promise<void> {
		const run = ++generation;
		loading = true;
		error = "";
		try {
			const page = await libraryApi.browse(browseQuery(source.source, filters, cursor));
			if (run !== generation) return;
			games = cursor ? [...games, ...page.games] : page.games;
			if (page.total !== null) total = page.total;
			nextCursor = page.nextCursor;
		} catch (err) {
			if (run !== generation) return;
			error = errorMessage(err);
		} finally {
			if (run === generation) loading = false;
		}
	}

	function reload(): void {
		games = [];
		total = null;
		nextCursor = null;
		void loadPage();
	}

	function loadMore(): void {
		if (nextCursor === null || loading) return;
		void loadPage(nextCursor);
	}

	function searchChanged(text: string): void {
		filters = setBrowseQuery(filters, text);
		clearTimeout(searchTimer);
		searchTimer = setTimeout(reload, SEARCH_DELAY_MS);
	}

	function toggleChanged(id: string, on: boolean): void {
		filters = setToggle(filters, id, on);
		reload();
	}

	function selectChanged(group: SidebarGroup, value: string): void {
		filters = setSelect(filters, group.id, value);
		reload();
	}

	function multiChanged(group: SidebarGroup, value: string): void {
		const current = filters.values[group.id];
		const selected = Array.isArray(current) ? current : [];
		const next = selected.includes(value)
			? selected.filter((item) => item !== value)
			: [...selected, value];
		filters = setMulti(filters, group.id, next);
		reload();
	}

	function sortChanged(value: string): void {
		filters = setSort(filters, value);
		reload();
	}

	function orderToggled(): void {
		filters = setOrder(filters, filters.order === "asc" ? "desc" : "asc");
		reload();
	}

	function clearFilters(): void {
		clearTimeout(searchTimer);
		filters = initialBrowseFilters(source.sidebar);
		reload();
	}

	function optionsOf(group: SidebarGroup): BrowseOption[] {
		if (!group.dynamic) return group.options;
		return (dynamic[group.id] ?? []).map((option) => ({ ...option, icon: null }));
	}

	function isSelected(group: SidebarGroup, value: string): boolean {
		const current = filters.values[group.id];
		return Array.isArray(current) && current.includes(value);
	}

	async function loadOptions(group: SidebarGroup, query: string): Promise<void> {
		const run = (optionGenerations[group.id] ?? 0) + 1;
		optionGenerations[group.id] = run;
		try {
			const options = await libraryApi.filters(source.source, group.id, query);
			if (optionGenerations[group.id] !== run) return;
			dynamic = { ...dynamic, [group.id]: options };
			optionError = { ...optionError, [group.id]: "" };
		} catch (err) {
			if (optionGenerations[group.id] !== run) return;
			optionError = { ...optionError, [group.id]: errorMessage(err) };
		}
	}

	function optionSearchChanged(group: SidebarGroup, text: string): void {
		optionSearch = { ...optionSearch, [group.id]: text };
		clearTimeout(optionTimers[group.id]);
		optionTimers = {
			...optionTimers,
			[group.id]: setTimeout(() => void loadOptions(group, text), SEARCH_DELAY_MS)
		};
	}
</script>

<div class="browse">
	<aside class="filters" aria-label="{source.label} filters">
		{#each source.sidebar as group (group.id)}
			{#if group.type === "search"}
				<label class="group">
					<span class="group-label">{group.label || "Search"}</span>
					<input
						class="input"
						type="search"
						placeholder={group.placeholder || "Search"}
						value={filters.query}
						oninput={(event) => searchChanged(event.currentTarget.value)}
					/>
				</label>
			{:else if group.type === "toggles"}
				<div class="group" role="group" aria-label={group.label || group.id}>
					{#if group.label}<span class="group-label">{group.label}</span>{/if}
					{#each group.options as option (option.value)}
						<button
							class="btn sm toggle"
							class:primary={filters.values[option.value] === true}
							aria-pressed={filters.values[option.value] === true}
							onclick={() => toggleChanged(option.value, filters.values[option.value] !== true)}
						>
							{#if option.icon}<span class="material-icons" aria-hidden="true">{option.icon}</span>{/if}
							{option.label}
						</button>
					{/each}
				</div>
			{:else if group.type === "select"}
				<label class="group">
					<span class="group-label">{group.label}</span>
					<select
						class="input"
						value={typeof filters.values[group.id] === "string" ? filters.values[group.id] : ""}
						onchange={(event) => selectChanged(group, event.currentTarget.value)}
					>
						<option value="">Any</option>
						{#each optionsOf(group) as option (option.value)}
							<option value={option.value}>{option.label}</option>
						{/each}
					</select>
				</label>
			{:else if group.type === "multi"}
				<fieldset class="group choices">
					<legend class="group-label">{group.label}</legend>
					{#if group.searchable}
						<input
							class="input"
							type="search"
							placeholder="Filter {group.label.toLowerCase()}"
							value={optionSearch[group.id] ?? ""}
							oninput={(event) => optionSearchChanged(group, event.currentTarget.value)}
						/>
					{/if}
					{#if optionError[group.id]}
						<p class="error" role="alert">{optionError[group.id]}</p>
					{:else if optionsOf(group).length === 0}
						<p class="muted">{group.dynamic ? "No choices." : "None declared."}</p>
					{:else}
						{#each optionsOf(group) as option (option.value)}
							<label class="choice">
								<input
									type="checkbox"
									checked={isSelected(group, option.value)}
									onchange={() => multiChanged(group, option.value)}
								/>
								{#if option.icon}<span class="material-icons" aria-hidden="true">{option.icon}</span>{/if}
								<span>{option.label}</span>
								{#if option.count !== null}<span class="count">{option.count}</span>{/if}
							</label>
						{/each}
					{/if}
				</fieldset>
			{:else if group.type === "sort"}
				<div class="group">
					<label>
						<span class="group-label">{group.label || "Sort"}</span>
						<select
							class="input"
							value={filters.sort}
							onchange={(event) => sortChanged(event.currentTarget.value)}
						>
							{#each group.options as option (option.value)}
								<option value={option.value}>{option.label}</option>
							{/each}
						</select>
					</label>
					{#if group.orderable}
						<button class="btn sm" onclick={orderToggled}>
							{filters.order === "asc" ? "Ascending" : "Descending"}
						</button>
					{/if}
				</div>
			{/if}
		{/each}

		{#if hasBrowseFilters(filters)}
			<button class="btn sm" onclick={clearFilters}>Clear filters</button>
		{/if}
	</aside>

	<div class="results">
		{#if error && games.length === 0}
			<div class="status">
				<p>{error}</p>
				<button class="btn sm" onclick={reload}>Retry</button>
			</div>
		{:else if loading && games.length === 0}
			<p class="status">Loading games...</p>
		{:else if games.length === 0}
			<p class="status">No games match these filters.</p>
		{:else}
			<p class="summary">
				{total === null ? `${games.length} games` : `${total} games`}
				{#if loading}<span class="muted"> · loading...</span>{/if}
			</p>
			<ul class="grid">
				{#each games as game, index (game.url + "#" + index)}
					<li><GameTile {game} {shape} onOpen={onOpenGame} /></li>
				{/each}
			</ul>
			{#if error}
				<p class="error" role="alert">{error}</p>
			{/if}
			{#if nextCursor !== null}
				<button class="btn sm more" disabled={loading} onclick={loadMore}>Load more</button>
			{/if}
		{/if}
	</div>
</div>

<style lang="scss">
	.browse {
		display: grid;
		grid-template-columns: 200px minmax(0, 1fr);
		align-items: start;
		gap: 24px;
	}

	.filters {
		position: sticky;
		top: 0;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.group {
		display: flex;
		flex-direction: column;
		gap: 6px;
		margin: 0;
		padding: 0;
		border: none;
	}

	.group-label {
		padding: 0;
		color: var(--text-muted);
		font-size: 0.8125rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.toggle {
		justify-content: flex-start;
	}

	.choices {
		gap: 4px;
	}

	.choice {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.9375rem;
		cursor: pointer;
	}

	.count {
		margin-left: auto;
		color: var(--text-muted);
		font-size: 0.8125rem;
	}

	.results {
		min-width: 0;
	}

	.summary {
		margin: 0 0 12px;
		color: var(--text-muted);
		font-size: 0.875rem;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		gap: 18px 14px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	li {
		min-width: 0;
	}

	.more {
		margin-top: 18px;
	}

	.status {
		padding: 32px 16px;
		color: var(--text-muted);
		text-align: center;
	}

	.muted {
		color: var(--text-muted);
	}

	.error {
		color: var(--danger);
		font-size: 0.875rem;
	}

	@media (max-width: 760px) {
		.browse {
			grid-template-columns: minmax(0, 1fr);
		}

		.filters {
			position: static;
		}
	}
</style>
