<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import type { LibraryState } from "./libraryState.svelte";
	import SavedFilters from "./SavedFilters.svelte";
	import SidebarGroupField from "./SidebarGroupField.svelte";
	import { searchPlaceholder } from "./sidebarGroups";

	const SYNC_HINT = "Adds the games from your account to the Library";

	let {
		library,
		canSync,
		syncing,
		onSync
	}: {
		library: LibraryState;
		/** Only sources that can sync games get the built-in Sync now button. */
		canSync: boolean;
		syncing: boolean;
		onSync: () => void;
	} = $props();

	const actions = $derived(
		(library.current?.actions ?? []).filter((action) => action.scope === "library")
	);
	const placeholder = $derived(searchPlaceholder(library.current));
</script>

<aside class="sidebar" aria-label="Library tools">
	{#if canSync || actions.length > 0}
		<div class="actions">
			{#if canSync}
				<Button size="sm" icon="sync" title={SYNC_HINT} disabled={syncing} onclick={onSync}>
					{syncing ? "Syncing..." : "Sync now"}
				</Button>
			{/if}
			{#each actions as action (action.id)}
				<Button
					size="sm"
					icon={action.icon}
					loading={library.runningAction === action.id}
					disabled={library.runningAction !== ""}
					onclick={() => library.runSourceAction(action)}
				>
					{action.label}
				</Button>
			{/each}
		</div>
	{/if}

	<label class="field">
		<span class="label">Search</span>
		<input
			class="input"
			type="search"
			{placeholder}
			value={library.filters.query}
			oninput={(event) => library.setQuery(event.currentTarget.value)}
		/>
	</label>

	{#key library.section}
		{#each library.sidebarGroups as group (group.id)}
			<SidebarGroupField {library} {group} />
		{/each}
	{/key}

	{#if library.filtersActive}
		<Button size="sm" onclick={() => library.clearFilters()}>Clear filters</Button>
	{/if}

	{#if !library.usesBrowse}
		<SavedFilters {library} />
	{/if}
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

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
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
