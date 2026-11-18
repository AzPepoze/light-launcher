<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import type { LibraryState } from "./libraryState.svelte";
	import SavedFilters from "./SavedFilters.svelte";
	import SidebarGroupField from "./SidebarGroupField.svelte";

	const SEARCH_PLACEHOLDER = "Title, developer or tag";

	let { library }: { library: LibraryState } = $props();
</script>

<div class="sidebar">
	<label class="field">
		<span class="label">Search</span>
		<input
			class="input"
			type="search"
			placeholder={SEARCH_PLACEHOLDER}
			value={library.filters.query}
			oninput={(event) => library.setQuery(event.currentTarget.value)}
		/>
	</label>

	{#each library.sidebarGroups as group (group.id)}
		<SidebarGroupField {library} {group} />
	{/each}

	{#if library.filtersActive}
		<Button size="sm" onclick={() => library.clearFilters()}>Clear filters</Button>
	{/if}

	<SavedFilters {library} />
</div>

<style lang="scss">
	.sidebar {
		display: flex;
		flex-direction: column;
		gap: 18px;
		box-sizing: border-box;
		min-width: 0;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.label {
		color: var(--text-muted);
		font-size: 0.8125rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
</style>
