<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import ChoiceRow from "@components/shared/ChoiceRow.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import type { SidebarGroup } from "../../lib/features/types";
	import { filterChoices, showsChoiceSearch } from "./choiceFilter";
	import { ENGINE_GROUP, TAG_GROUP } from "./libraryFilters";
	import type { LibraryState } from "./libraryState.svelte";

	let { library, group }: { library: LibraryState; group: SidebarGroup } = $props();

	const ANY_OPTION = { value: "", label: "Any" };
	const EMPTY_TEXT: Record<string, string> = {
		[ENGINE_GROUP]: "No engines yet",
		[TAG_GROUP]: "No tags yet"
	};

	/** Text typed into a built-in group's search. Dynamic groups search remotely instead. */
	let localQuery = $state("");
	const remoteQuery = $derived(library.browse.choiceSearch[group.id] ?? "");
	const query = $derived(group.dynamic ? remoteQuery : localQuery);
	const showSearch = $derived(showsChoiceSearch(group.options.length, group.searchable));
	/** Local groups filter the loaded options. Dynamic groups already sent the query to the source. */
	const view = $derived(
		group.dynamic
			? {
					choices: filterChoices(group.options, "", (value) => isChosen(group.id, value)).choices,
					noMatches: remoteQuery.trim() !== "" && group.options.length === 0
				}
			: filterChoices(group.options, localQuery, (value) => isChosen(group.id, value))
	);

	function typed(text: string) {
		if (group.dynamic) library.searchChoices(group, text);
		else localQuery = text;
	}

	function selectValue(id: string): string {
		const current = library.filters.values[id];
		return typeof current === "string" ? current : "";
	}

	function isChosen(id: string, value: string): boolean {
		const current = library.filters.values[id];
		if (Array.isArray(current)) return current.includes(value);
		return current === value && value !== "";
	}

	function pick(value: string) {
		if (group.type === "single") library.toggleSingle(group.id, value);
		else library.toggleMulti(group.id, value);
	}

	function isOn(id: string): boolean {
		return library.filters.values[id] === true;
	}

	function emptyText(): string {
		return EMPTY_TEXT[group.id] ?? (group.dynamic ? "No choices." : "None declared.");
	}

	function clearSearch() {
		typed("");
	}
</script>

{#snippet heading()}
	<div class="heading">
		<span class="label">{group.label}</span>
		{#if group.dynamic && library.choicesLoaded(group.id)}
			<IconButton
				icon="refresh"
				size={32}
				label="Refresh choices"
				disabled={library.choicesLoading(group.id)}
				onclick={() => library.refreshChoices(group)}
			/>
		{/if}
	</div>
{/snippet}

{#if group.dynamic && !library.choicesLoaded(group.id)}
	<div class="group">
		{@render heading()}
		{#if library.choicesFailed(group.id)}
			<Button size="sm" icon="sync" onclick={() => library.refreshChoices(group)}>Retry</Button>
		{:else}
			<LoadingState label="Loading choices..." />
		{/if}
	</div>
{:else if group.type === "toggles"}
	<div class="group" role="group" aria-label={group.label}>
		{@render heading()}
		<div class="choices">
			{#each group.options as option (option.value)}
				<ChoiceRow
					selected={isOn(option.value)}
					onToggle={() => library.setToggle(option.value, !isOn(option.value))}
				>
					{option.label}
				</ChoiceRow>
			{/each}
		</div>
	</div>
{:else if group.type === "select"}
	<div class="group">
		{@render heading()}
		<Dropdown
			aria-label={group.label}
			options={[ANY_OPTION, ...group.options]}
			value={selectValue(group.id)}
			searchable={group.searchable}
			onChange={(value) => library.setSelect(group.id, value)}
		/>
	</div>
{:else if group.type === "multi" || group.type === "single"}
	<div class="group" role="group" aria-label={group.label}>
		{@render heading()}
		{#if showSearch}
			<div class="search">
				<Icon name="search" size={16} class="search-icon" />
				<input
					class="input"
					type="search"
					placeholder="Search {group.label.toLowerCase()}"
					aria-label="Search {group.label.toLowerCase()}"
					value={query}
					oninput={(event) => typed(event.currentTarget.value)}
				/>
				{#if query}
					<IconButton icon="close" size={32} label="Clear search" onclick={clearSearch} />
				{/if}
			</div>
		{/if}
		{#if group.options.length === 0 && query === ""}
			<p class="muted">{emptyText()}</p>
		{:else}
			<div class="choices">
				{#each view.choices as option (option.value)}
					<ChoiceRow
						selected={isChosen(group.id, option.value)}
						onToggle={() => pick(option.value)}
						icon={option.icon}
						count={option.count}
					>
						{option.label}
					</ChoiceRow>
				{/each}
			</div>
			{#if view.noMatches}
				<p class="muted">No matches</p>
			{/if}
		{/if}
	</div>
{:else if group.type === "sort"}
	<div class="group">
		<span class="label">{group.label || "Sort"}</span>
		<Dropdown
			aria-label={group.label || "Sort"}
			options={group.options}
			value={library.filters.sort}
			onChange={(value) => library.setSort(value)}
		/>
		{#if group.orderable}
			<Button size="sm" onclick={() => library.toggleOrder()}>
				{library.filters.order === "asc" ? "Ascending" : "Descending"}
			</Button>
		{/if}
	</div>
{/if}

<style lang="scss">
	/* 8px between each group label and its controls. */
	.group {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0;
		padding: 0;
		border: none;
		min-width: 0;
	}

	.heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.label {
		padding: 0;
		color: var(--text-muted);
		font-size: 0.8125rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.choices {
		display: flex;
		flex-direction: column;
		gap: 4px;
		max-height: 240px;
		overflow-y: auto;
		/* Keeps the scrollbar beside the counts instead of on top of them. */
		scrollbar-gutter: stable;
		padding-right: 10px;
	}

	/* The magnifier and clear button sit inside the input box. */
	.search {
		position: relative;
	}

	.search .input {
		box-sizing: border-box;
		width: 100%;
		padding-left: 34px;
		padding-right: 36px;
	}

	.search :global(.search-icon) {
		position: absolute;
		top: 50%;
		left: 10px;
		transform: translateY(-50%);
		color: var(--text-muted);
		pointer-events: none;
	}

	.search :global(.icon-button) {
		position: absolute;
		top: 50%;
		right: 2px;
		transform: translateY(-50%);
	}

	.muted {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.875rem;
	}
</style>
