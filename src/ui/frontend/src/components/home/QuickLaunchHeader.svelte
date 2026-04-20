<script lang="ts">
	import { fade } from "svelte/transition";
	import Dropdown from "../shared/Dropdown.svelte";

	export let isSelectionMode: boolean;
	export let selectedCount: number;
	export let prefixes: string[];
	export let selectedPrefixFilter: string;
	export let searchQuery: string;
	export let currentView: "grid" | "list-grid";

	export let onBulkRemove: () => void;
	export let onToggleSelectionMode: () => void;
	export let onShowAddModal: () => void;
	export let onShowHelpModal: () => void;
</script>

<div class="section-header">
	<h2 class="section-title">Quick Launch</h2>

	{#if isSelectionMode}
		<div class="selection-actions" in:fade>
			<span class="selection-count">{selectedCount} selected</span>
			<button
				class="bulk-remove-btn"
				on:click={onBulkRemove}
				disabled={selectedCount === 0}
			>
				<span class="material-icons" style="font-size: 18px;"
					>delete</span
				>
				Remove Selected
			</button>
			<button
				class="cancel-selection-btn"
				on:click={onToggleSelectionMode}
			>
				Cancel
			</button>
		</div>
	{:else}
		<button class="add-btn" on:click={onShowAddModal} title="Add Game">
			<span class="material-icons" style="font-size: 20px;">add</span>
		</button>

		<button
			class="select-mode-btn"
			on:click={onToggleSelectionMode}
			title="Bulk Remove"
			class:active={isSelectionMode}
		>
			<span class="material-icons" style="font-size: 20px;"
				>checklist</span
			>
		</button>

		<button
			class="help-btn"
			on:click={onShowHelpModal}
			title="How it works"
		>
			<span class="material-icons" style="font-size: 24px;"
				>help_outline</span
			>
		</button>
	{/if}

	<div class="prefix-filter-container">
		<Dropdown
			options={prefixes}
			bind:value={selectedPrefixFilter}
			placeholder="All Prefixes"
		/>
	</div>

	<div class="search-container">
		<span class="material-icons search-icon">search</span>
		<input
			type="text"
			placeholder="Search games..."
			bind:value={searchQuery}
			class="search-input"
		/>
		{#if searchQuery}
			<button
				class="clear-search"
				on:click={() => (searchQuery = "")}
				aria-label="Clear search"
			>
				<span class="material-icons" style="font-size: 14px;"
					>close</span
				>
			</button>
		{/if}
	</div>

	<div class="view-switcher">
		<button
			class="view-btn"
			class:active={currentView === "grid"}
			on:click={() => (currentView = "grid")}
			title="Grid View"
		>
			<span class="material-icons" style="font-size: 18px;"
				>grid_view</span
			>
		</button>
		<button
			class="view-btn"
			class:active={currentView === "list-grid"}
			on:click={() => (currentView = "list-grid")}
			title="List View"
		>
			<span class="material-icons" style="font-size: 18px;"
				>view_list</span
			>
		</button>
	</div>
</div>

<style lang="scss">
	.section-title {
		font-size: 1.2rem;
		font-weight: 800;
		color: var(--text-dim);
		opacity: 0.4;
		text-transform: uppercase;
		letter-spacing: 2px;
		margin: 0;
		line-height: 1;
		white-space: nowrap;
	}

	.section-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 20px;
		flex-wrap: wrap;
	}

	.view-switcher {
		display: flex;
		background: var(--glass-surface);
		padding: 4px;
		border-radius: 12px;
		gap: 4px;
		border: 1px solid var(--glass-border);

		.view-btn {
			background: none;
			border: none;
			color: var(--text-dim);
			opacity: 0.5;
			padding: 6px;
			cursor: pointer;
			border-radius: 8px;
			display: flex;
			align-items: center;
			justify-content: center;
			aspect-ratio: 1 / 1;
			transition: all 0.2s;

			&:hover {
				color: var(--text-main);
				background: var(--glass-border-bright);
			}

			&.active {
				color: var(--glass-bg);
				background: var(--accent-primary);
				box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
			}
		}
	}

	.help-btn {
		background: none;
		border: none;
		color: var(--text-dim);
		opacity: 0.4;
		cursor: pointer;
		padding: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s;
		border-radius: 50%;

		&:hover {
			color: var(--accent-primary);
			opacity: 1;
			background: var(--glass-surface);
			transform: scale(1.1);
		}
	}

	.select-mode-btn {
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		color: var(--text-dim);
		cursor: pointer;
		padding: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s;
		border-radius: 50%;

		&:hover {
			color: var(--text-main);
			background: var(--glass-border-bright);
			transform: scale(1.1);
		}

		&.active {
			background: var(--accent-primary);
			color: var(--glass-bg);
			border-color: transparent;
		}
	}

	.selection-actions {
		display: flex;
		align-items: center;
		gap: 12px;
		background: var(--glass-surface);
		padding: 4px 12px;
		border-radius: 16px;
		border: 1px solid var(--glass-border);

		.selection-count {
			font-size: 0.85rem;
			font-weight: 700;
			color: var(--accent-primary);
		}

		.bulk-remove-btn {
			background: #ef4444;
			color: #fff;
			border: none;
			padding: 6px 12px;
			border-radius: 8px;
			font-size: 0.8rem;
			font-weight: 800;
			cursor: pointer;
			display: flex;
			align-items: center;
			gap: 6px;
			transition: all 0.2s;

			&:hover:not(:disabled) {
				filter: brightness(1.2);
				transform: translateY(-1px);
			}

			&:disabled {
				opacity: 0.5;
				cursor: not-allowed;
			}
		}

		.cancel-selection-btn {
			background: none;
			border: none;
			color: var(--text-dim);
			font-size: 0.8rem;
			font-weight: 700;
			cursor: pointer;

			&:hover {
				color: var(--text-main);
			}
		}
	}

	.add-btn {
		background: var(--accent-primary);
		border: none;
		color: var(--glass-bg);
		cursor: pointer;
		padding: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s;
		border-radius: 50%;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);

		&:hover {
			filter: brightness(1.2);
			transform: scale(1.1) rotate(90deg);
			box-shadow: 0 6px 16px rgba(0, 0, 0, 0.3);
		}
	}

	.prefix-filter-container {
		min-width: 160px;
		max-width: 200px;

		:global(.dropdown-trigger) {
			padding: 8px 12px;
			font-size: 0.8rem;
			background: rgba(255, 255, 255, 0.05);
			border-color: rgba(255, 255, 255, 0.05);

			&:hover {
				background: rgba(255, 255, 255, 0.1);
				border-color: rgba(255, 255, 255, 0.2);
			}
		}
	}

	.search-container {
		display: flex;
		align-items: center;
		background: var(--glass-low-alpha);
		border: 1px solid var(--glass-border);
		border-radius: 12px;
		padding: 4px 10px;
		gap: 8px;
		flex: 1;
		transition: all 0.3s;

		&:focus-within {
			background: var(--glass-surface);
			border-color: var(--accent-primary);
			box-shadow: var(--glass-shadow);
		}

		.search-icon {
			color: var(--text-dim);
		}

		.search-input {
			background: none;
			border: none;
			color: var(--text-main);
			font-size: 0.9rem;
			width: 100%;
			outline: none;

			&::placeholder {
				color: var(--text-dim);
			}
		}

		.clear-search {
			background: none;
			border: none;
			color: var(--text-dim);
			cursor: pointer;
			padding: 2px;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 4px;

			&:hover {
				color: var(--text-main);
				background: var(--glass-hover);
			}
		}
	}
</style>
