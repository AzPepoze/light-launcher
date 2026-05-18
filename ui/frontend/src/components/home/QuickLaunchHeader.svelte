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
		font-size: 1.6rem;
		font-weight: 800;
		color: var(--text-main);
		text-transform: uppercase;
		letter-spacing: 2px;
		margin: 0;
		line-height: 1;
		white-space: nowrap;
	}

	.section-header {
		display: flex;
		align-items: center;
		gap: 18px;
		margin-bottom: 28px;
		flex-wrap: wrap;
		width: 100%;
	}

	.view-switcher {
		display: flex;
		background: var(--bg-surface);
		padding: 4px;
		border-radius: var(--radius-pill);
		gap: 4px;
		border: 2px solid rgba(255, 255, 255, 0.05);

		.view-btn {
			background: none;
			border: none;
			color: var(--text-muted);
			padding: 8px;
			cursor: pointer;
			border-radius: var(--radius-pill);
			display: flex;
			align-items: center;
			justify-content: center;
			aspect-ratio: 1 / 1;
			transition: transform var(--transition-spring), background var(--transition-fast), color var(--transition-fast);

			&:hover {
				color: var(--text-main);
				background: var(--bg-elevated);
				transform: scale(1.1);
			}

			&.active {
				color: var(--bg-base);
				background: var(--accent-primary);
				box-shadow: 0 4px 10px var(--accent-glow);
			}
		}
	}

	.help-btn {
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		color: var(--text-muted);
		cursor: pointer;
		padding: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: transform var(--transition-spring), color var(--transition-fast), border-color var(--transition-fast);
		border-radius: var(--radius-pill);

		&:hover {
			color: var(--accent-secondary);
			border-color: var(--accent-secondary);
			transform: scale(1.15);
		}

		&:active {
			transform: scale(0.95);
		}
	}

	.select-mode-btn {
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		color: var(--text-muted);
		cursor: pointer;
		padding: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: transform var(--transition-spring), color var(--transition-fast), border-color var(--transition-fast);
		border-radius: var(--radius-pill);

		&:hover {
			color: var(--accent-primary);
			border-color: var(--accent-primary);
			transform: scale(1.15);
		}

		&:active {
			transform: scale(0.95);
		}

		&.active {
			background: var(--accent-primary);
			color: var(--bg-base);
			border-color: var(--accent-primary);
			box-shadow: 0 4px 12px var(--accent-glow);
		}
	}

	.selection-actions {
		display: flex;
		align-items: center;
		gap: 14px;
		background: var(--bg-surface);
		padding: 6px 18px;
		border-radius: var(--radius-pill);
		border: 2px solid var(--accent-primary);

		.selection-count {
			font-size: 0.9rem;
			font-weight: 800;
			color: var(--accent-primary);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		.bulk-remove-btn {
			background: var(--danger);
			color: #fff;
			border: none;
			padding: 8px 18px;
			border-radius: var(--radius-pill);
			font-size: 0.85rem;
			font-weight: 800;
			cursor: pointer;
			display: flex;
			align-items: center;
			gap: 6px;
			transition: transform var(--transition-spring), filter var(--transition-fast);

			&:hover:not(:disabled) {
				filter: brightness(1.15);
				transform: scale(1.05);
			}

			&:active {
				transform: scale(0.95);
			}

			&:disabled {
				opacity: 0.5;
				cursor: not-allowed;
			}
		}

		.cancel-selection-btn {
			background: none;
			border: none;
			color: var(--text-muted);
			font-size: 0.85rem;
			font-weight: 800;
			cursor: pointer;
			text-transform: uppercase;
			letter-spacing: 0.5px;
			transition: color var(--transition-fast);

			&:hover {
				color: var(--text-main);
			}
		}
	}

	.add-btn {
		background: var(--accent-primary);
		border: none;
		color: var(--bg-base);
		cursor: pointer;
		padding: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: transform var(--transition-spring), background var(--transition-fast);
		border-radius: var(--radius-pill);
		box-shadow: 0 4px 12px var(--accent-glow);

		&:hover {
			background: var(--accent-hover);
			transform: scale(1.18) rotate(90deg);
		}

		&:active {
			transform: scale(0.9);
		}
	}

	.prefix-filter-container {
		min-width: 180px;
		max-width: 220px;
	}

	.search-container {
		display: flex;
		align-items: center;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-pill);
		padding: 10px 20px;
		gap: 10px;
		flex: 1;
		transition: border-color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);

		&:focus-within {
			border-color: var(--accent-primary);
			box-shadow: 0 0 0 4px var(--accent-glow);
			transform: scale(1.01);
		}

		.search-icon {
			color: var(--text-muted);
			font-size: 20px;
		}

		.search-input {
			background: none;
			border: none;
			color: var(--text-main);
			font-size: 0.95rem;
			font-weight: 600;
			width: 100%;
			outline: none;

			&::placeholder {
				color: var(--text-dim);
			}
		}

		.clear-search {
			background: none;
			border: none;
			color: var(--text-muted);
			cursor: pointer;
			padding: 2px;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 50%;
			transition: color var(--transition-fast), background var(--transition-fast);

			&:hover {
				color: var(--text-main);
				background: var(--bg-elevated);
			}
		}
	}
</style>
