<script lang="ts">
	import CreatePrefixForm from "./CreatePrefixForm.svelte";
	import PrefixListItem from "./PrefixListItem.svelte";

	export let availablePrefixes: string[] = [];
	export let currentPrefixName: string = "";
	export let newPrefixName: string = "";
	export let onSelectPrefix: (name: string) => void;
	export let onCreatePrefix: () => void;
	export let onRemovePrefix: (name: string) => void;

	let showCreateForm = false;

	function openCreateForm() {
		showCreateForm = true;
	}

	function closeCreateForm() {
		showCreateForm = false;
	}

	function handleCreate(name: string) {
		newPrefixName = name;
		showCreateForm = false;
		onCreatePrefix();
	}
</script>

<div class="prefix-list-panel">
	<header class="panel-header">
		<h2>Prefixes</h2>
		<span class="count">{availablePrefixes.length}</span>
	</header>

	<div class="prefix-list" role="listbox" aria-label="Available prefixes">
		{#if availablePrefixes.length === 0 && !showCreateForm}
			<p class="list-empty-hint">No prefixes yet</p>
		{/if}

		{#each availablePrefixes as name (name)}
			<PrefixListItem
				{name}
				{currentPrefixName}
				{onSelectPrefix}
				{onRemovePrefix}
			/>
		{/each}

		{#if showCreateForm}
			<CreatePrefixForm
				{availablePrefixes}
				onCreate={handleCreate}
				onCancel={closeCreateForm}
			/>
		{:else}
			<button
				type="button"
				class="list-row add-row"
				on:click={openCreateForm}
			>
				<span class="material-icons row-icon">add</span>
				<span class="add-label">New prefix</span>
			</button>
		{/if}
	</div>
</div>

<style lang="scss">
	.prefix-list-panel {
		display: flex;
		flex-direction: column;
		border-radius: var(--radius-lg);
		overflow: hidden;
		border: 2px solid var(--glass-border);
		background: var(--bg-surface);
		height: 100%;
		max-height: 100%;
		box-sizing: border-box;
	}

	.panel-header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 16px 18px;
		border-bottom: 2px solid var(--glass-border);
		flex-shrink: 0;

		h2 {
			font-size: 0.9rem;
			font-weight: 800;
			margin: 0;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 0.8px;
			flex: 1;
		}
	}

	.count {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 22px;
		height: 22px;
		padding: 0 6px;
		border-radius: var(--radius-pill);
		background: var(--bg-elevated);
		border: 2px solid var(--glass-border);
		font-size: 0.7rem;
		font-weight: 800;
		color: var(--text-muted);
	}

	.prefix-list {
		flex: 1;
		overflow-y: auto;
		padding: 10px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-height: 0;

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: var(--glass-border-bright);
			border-radius: 10px;
		}
	}

	.list-empty-hint {
		margin: 8px 12px 4px;
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--text-dim);
		text-align: center;
	}

	.list-row {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		border-radius: var(--radius-md);
		box-sizing: border-box;
	}

	.add-row {
		padding: 11px 12px;
		border: 2px dashed var(--glass-border);
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		text-align: left;
		transition:
			border-color var(--transition-fast),
			background var(--transition-fast),
			color var(--transition-fast);

		.row-icon {
			font-size: 20px;
			color: var(--accent-primary);
		}

		.add-label {
			font-size: 0.85rem;
			font-weight: 700;
		}

		&:hover {
			border-color: var(--accent-secondary);
			border-style: solid;
			background: var(--bg-elevated);
			color: var(--text-main);
		}
	}
</style>
