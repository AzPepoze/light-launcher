<script lang="ts">
	import type { LibraryState } from "./libraryState.svelte";
	import { SORT_LABELS, sourceFilterLabel, type LibrarySort } from "./libraryFilters";

	let { library }: { library: LibraryState } = $props();

	let saveName = $state("");
	let saving = $state(false);

	const sortOptions = Object.entries(SORT_LABELS) as [LibrarySort, string][];

	function toggleTag(tag: string) {
		const selected = library.filters.tags.includes(tag);
		library.filters.tags = selected
			? library.filters.tags.filter((item) => item !== tag)
			: [...library.filters.tags, tag];
	}

	function confirmSave() {
		library.saveCurrentFilter(saveName);
		saveName = "";
		saving = false;
	}
</script>

<div class="bar">
	<div class="row">
		<label class="search">
			<span class="material-icons" aria-hidden="true">search</span>
			<span class="sr-only">Search your library</span>
			<input
				class="input"
				type="search"
				placeholder="Search title, developer or tag"
				bind:value={library.filters.search}
			/>
		</label>

		<label class="sort">
			<span class="sr-only">Sort by</span>
			<select class="input" bind:value={library.filters.sort}>
				{#each sortOptions as [value, label] (value)}
					<option {value}>{label}</option>
				{/each}
			</select>
		</label>

		<select class="input" aria-label="Source" bind:value={library.filters.source}>
			<option value="">All sources</option>
			{#each library.sources as item (item.source)}
				<option value={item.source}>{sourceFilterLabel(item)}</option>
			{/each}
		</select>
	</div>

	<div class="row chips" role="group" aria-label="Quick filters">
		<button
			class="btn sm"
			class:primary={library.filters.installed}
			aria-pressed={library.filters.installed}
			onclick={() => (library.filters.installed = !library.filters.installed)}
		>
			Installed
		</button>
		<button
			class="btn sm"
			class:primary={library.filters.updates}
			aria-pressed={library.filters.updates}
			onclick={() => (library.filters.updates = !library.filters.updates)}
		>
			Updates
		</button>
		<button
			class="btn sm"
			class:primary={library.filters.favorite}
			aria-pressed={library.filters.favorite}
			onclick={() => (library.filters.favorite = !library.filters.favorite)}
		>
			Favorites
		</button>
		{#each library.tags.slice(0, 12) as item (item.tag)}
			<button
				class="btn sm"
				class:primary={library.filters.tags.includes(item.tag)}
				aria-pressed={library.filters.tags.includes(item.tag)}
				onclick={() => toggleTag(item.tag)}
			>
				{item.tag}
			</button>
		{/each}
	</div>

	<div class="row chips">
		{#each library.savedFilters as saved (saved.id)}
			<span class="saved">
				<button class="btn sm secondary" onclick={() => library.applySavedFilter(saved)}>
					{saved.name}
				</button>
				<button
					class="remove"
					aria-label="Delete saved filter {saved.name}"
					onclick={() => library.deleteSavedFilter(saved.id)}
				>
					<span class="material-icons">close</span>
				</button>
			</span>
		{/each}

		{#if library.filtersActive}
			{#if saving}
				<input
					class="input name"
					placeholder="Filter name"
					aria-label="Filter name"
					bind:value={saveName}
					onkeydown={(event) => event.key === "Enter" && confirmSave()}
				/>
				<button class="btn sm primary" onclick={confirmSave}>Save</button>
				<button class="btn sm" onclick={() => (saving = false)}>Cancel</button>
			{:else}
				<button class="btn sm" onclick={() => (saving = true)}>Save filter</button>
				<button class="btn sm" onclick={() => library.resetFilters()}>Clear</button>
			{/if}
		{/if}
	</div>
</div>

<style lang="scss">
	.bar {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-bottom: 18px;
	}

	.row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
	}

	.search {
		position: relative;
		flex: 1 1 220px;
		display: flex;
		align-items: center;

		> .material-icons {
			position: absolute;
			left: 12px;
			font-size: 18px;
			color: var(--text-dim);
			pointer-events: none;
		}

		input {
			width: 100%;
			padding-left: 40px;
		}
	}

	select.input {
		flex: 0 1 180px;
		padding: 10px 12px;
	}

	.chips {
		gap: 6px;
	}

	.saved {
		display: inline-flex;
		align-items: center;
		gap: 2px;
	}

	.remove {
		display: grid;
		place-items: center;
		width: 26px;
		height: 26px;
		border: none;
		border-radius: var(--radius-pill);
		background: transparent;
		color: var(--text-dim);
		cursor: pointer;

		.material-icons {
			font-size: 16px;
		}

		&:hover,
		&:focus-visible {
			color: var(--text-main);
			outline: 2px solid var(--accent-primary);
		}
	}

	.name {
		flex: 0 1 180px;
		padding: 8px 12px;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
</style>
