<script lang="ts">
	import type { LibrarySource } from "../../lib/features/types";

	let {
		sections,
		selected,
		onSelect
	}: {
		sections: LibrarySource[];
		selected: string;
		onSelect: (source: string) => void;
	} = $props();
</script>

<nav class="sections" aria-label="Library sections">
	<ul>
		{#each sections as section (section.source)}
			{@const active = section.source === selected}
			<li>
				<button
					class="section"
					class:active
					aria-current={active ? "true" : undefined}
					onclick={() => onSelect(section.source)}
				>
					<span class="label">{section.label}</span>
					<span class="count">{section.count}</span>
				</button>
			</li>
		{/each}
	</ul>
</nav>

<style lang="scss">
	.sections {
		position: sticky;
		top: 0;
	}

	ul {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.section {
		display: flex;
		width: 100%;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 10px 12px;
		border: none;
		border-left: 2px solid transparent;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-main);
		font: inherit;
		font-size: 1rem;
		text-align: left;
		cursor: pointer;
		transition:
			background var(--motion-quick) var(--ease-out),
			border-color var(--motion-quick) var(--ease-out);

		&:hover {
			background: var(--bg-elevated);
		}

		&.active {
			border-left-color: var(--accent-primary);
			background: var(--bg-elevated);
			font-weight: 600;
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	.label {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.count {
		flex-shrink: 0;
		padding: 1px 10px;
		border-radius: var(--radius-pill);
		background: var(--bg-surface);
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-muted);
	}

	@media (max-width: 760px) {
		.sections {
			position: static;
		}

		ul {
			flex-direction: row;
			flex-wrap: wrap;
		}
	}
</style>
