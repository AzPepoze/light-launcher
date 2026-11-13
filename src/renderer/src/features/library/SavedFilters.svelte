<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import IconButton from "@components/shared/IconButton.svelte";
	import type { LibraryState } from "./libraryState.svelte";

	let { library }: { library: LibraryState } = $props();

	let saveName = $state("");
	let saving = $state(false);

	function confirmSave() {
		library.saveCurrentFilter(saveName);
		saveName = "";
		saving = false;
	}
</script>

<div class="saved">
	<span class="label">Saved filters</span>
	{#each library.savedFilters as saved (saved.id)}
		<div class="row">
			<Button size="sm" onclick={() => library.applySavedFilter(saved)}>{saved.name}</Button>
			<IconButton
				icon="close"
				size={32}
				label="Delete saved filter {saved.name}"
				onclick={() => library.deleteSavedFilter(saved.id)}
			/>
		</div>
	{/each}

	{#if library.filtersActive}
		{#if saving}
			<input
				class="input"
				placeholder="Filter name"
				aria-label="Filter name"
				bind:value={saveName}
				onkeydown={(event) => event.key === "Enter" && confirmSave()}
			/>
			<div class="row">
				<Button size="sm" variant="primary" onclick={confirmSave}>Save</Button>
				<Button size="sm" onclick={() => (saving = false)}>Cancel</Button>
			</div>
		{:else}
			<Button size="sm" onclick={() => (saving = true)}>Save filter</Button>
		{/if}
	{/if}
</div>

<style lang="scss">
	.saved {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.label {
		color: var(--text-muted);
		font-size: 0.8125rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.row {
		display: flex;
		align-items: center;
		gap: 4px;
	}
</style>
