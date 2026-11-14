<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import type { LibraryState } from "./libraryState.svelte";
	import SavedFilters from "./SavedFilters.svelte";
	import SidebarGroupField from "./SidebarGroupField.svelte";

	const SEARCH_PLACEHOLDER = "Title, developer or tag";

	let { library }: { library: LibraryState } = $props();
</script>

<aside class="sidebar" aria-label="Library filters">
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
</aside>

<style lang="scss">
	.sidebar {
		display: flex;
		flex-direction: column;
		gap: 18px;
		box-sizing: border-box;
		min-width: 0;
		padding: 16px 14px 16px 16px;
	}

	/* From 900px the filters stick at the top of the content and scroll on their own. */
	@media (min-width: 900px) {
		.sidebar {
			position: sticky;
			top: 76px;
			max-height: calc(100vh - 116px);
			overflow-y: auto;
			/* Keeps the scrollbar from shifting the content when the list grows. */
			scrollbar-gutter: stable;
		}
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
