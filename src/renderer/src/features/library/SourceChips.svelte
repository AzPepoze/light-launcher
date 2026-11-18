<script lang="ts">
	import type { LibrarySource } from "../../lib/features/types";

	let {
		sections,
		selected,
		onSelect,
		onAll
	}: {
		sections: LibrarySource[];
		/** The source whose filter groups the sidebar shows. */
		selected: string;
		onSelect: (source: string) => void;
		onAll: () => void;
	} = $props();

	const total = $derived(sections.reduce((sum, section) => sum + (section.count ?? 0), 0));
</script>

<nav class="chips" aria-label="Jump to a source">
	<button type="button" class="chip" onclick={onAll}>
		<span>All</span>
		<span class="count">{total}</span>
	</button>
	{#each sections as section (section.source)}
		<button
			type="button"
			class="chip"
			class:active={section.source === selected}
			aria-pressed={section.source === selected}
			onclick={() => onSelect(section.source)}
		>
			<span>{section.label}</span>
			<span class="count">{section.count}</span>
		</button>
	{/each}
</nav>

<style lang="scss">
	.chips {
		display: flex;
		gap: 8px;
		overflow-x: auto;
		scrollbar-width: none;
	}

	.chip {
		display: inline-flex;
		flex-shrink: 0;
		align-items: center;
		gap: 8px;
		height: 34px;
		padding: 0 14px;
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-pill);
		background: transparent;
		color: var(--text-muted);
		font: inherit;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition:
			background var(--transition-fast),
			color var(--transition-fast);

		&:hover {
			color: var(--text-main);
			background: var(--bg-elevated);
		}

		&.active {
			border-color: transparent;
			background: var(--text-main);
			color: var(--bg-base);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	.count {
		font-size: 0.75rem;
		font-weight: 600;
		opacity: 0.7;
	}
</style>
