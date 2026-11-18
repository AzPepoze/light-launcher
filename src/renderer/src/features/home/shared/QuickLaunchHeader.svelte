<script lang="ts">
	import Dropdown, { asOptions } from "@components/shared/Dropdown.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import PageHeader from "@components/shared/PageHeader.svelte";
	import { tooltip } from "@components/shared/tooltip";

	$: prefixItems = asOptions(prefixes);
	import SelectionActions from "./SelectionActions.svelte";

	export let isSelectionMode: boolean;
	export let selectedCount: number;
	export let prefixes: string[];
	export let selectedPrefixFilter: string;
	export let searchQuery: string;
	export let currentView: "grid" | "list-grid" | "sidebar-grid";

	export let onBulkRemove: () => void;
	export let onToggleSelectionMode: () => void;
	export let onSelectAll: () => void;
	export let onShowAddModal: () => void;
	export let onShowHelpModal: () => void;
</script>

<div class="quick-launch-header">
<PageHeader title="Quick Launch">
	<div class="header-controls">
	{#if isSelectionMode}
		<SelectionActions {selectedCount} {onBulkRemove} {onToggleSelectionMode} {onSelectAll} />
	{:else}
		<button class="add-btn" on:click={onShowAddModal} use:tooltip={"Add Game"} aria-label="Add Game">
			<Icon name="add" size={20} />
		</button>

		<button
			class="select-mode-btn"
			on:click={onToggleSelectionMode}
			use:tooltip={"Bulk Remove"} aria-label="Bulk Remove"
			class:active={isSelectionMode}
		>
			<Icon name="checklist" size={20} />
		</button>

		<button
			class="help-btn"
			on:click={onShowHelpModal}
			use:tooltip={"How it works"} aria-label="How it works"
		>
			<Icon name="help_outline" size={24} />
		</button>
	{/if}

	<div class="prefix-filter-container">
		<Dropdown
			options={prefixItems}
			bind:value={selectedPrefixFilter}
			placeholder="All Prefixes"
		/>
	</div>

	<div class="search-container">
		<Icon name="search" size={20} class="search-icon" />
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
				<Icon name="close" size={14} />
			</button>
		{/if}
	</div>

	<div class="view-switcher">
		<button
			class="view-btn"
			class:active={currentView === "grid"}
			on:click={() => (currentView = "grid")}
			use:tooltip={"Grid View"} aria-label="Grid View"
		>
			<Icon name="grid_view" size={18} />
		</button>
		<button
			class="view-btn"
			class:active={currentView === "list-grid"}
			on:click={() => (currentView = "list-grid")}
			use:tooltip={"List View"} aria-label="List View"
		>
			<Icon name="view_list" size={18} />
		</button>
		<button
			class="view-btn"
			class:active={currentView === "sidebar-grid"}
			on:click={() => (currentView = "sidebar-grid")}
			use:tooltip={"Sidebar View"} aria-label="Sidebar View"
		>
			<Icon name="view_sidebar" size={18} />
		</button>
	</div>
	</div>
</PageHeader>
</div>

<style lang="scss">
	/* PageHeader keeps its actions at full width; let them shrink on narrow screens. */
	.quick-launch-header :global(.header-actions) {
		flex-shrink: 1;
		min-width: 0;
		max-width: 100%;
	}

	.header-controls {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		flex-wrap: wrap;
		gap: 12px;
	}

	@media (max-width: 600px) {
		.prefix-filter-container,
		.search-container {
			flex: 1 1 100%;
			max-width: none;
		}
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
		flex: 0 1 280px;
		min-width: 200px;
		transition: border-color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);

		&:focus-within {
			border-color: var(--accent-primary);
			box-shadow: 0 0 0 4px var(--accent-glow);
			transform: scale(1.01);
		}

		:global(.search-icon) {
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
