<script lang="ts">
	import type { SettingsNavGroup } from "@lib/features/settingsSection";
	import { settingsNavFilter } from "@lib/features/settingsSection";

	let {
		groups,
		active,
		onSelect
	}: {
		groups: SettingsNavGroup[];
		active: string | null;
		onSelect: (id: string) => void;
	} = $props();

	const SEARCH_PLACEHOLDER = "Search settings";

	let query = $state("");
	const matches = $derived(settingsNavFilter(groups, query));
	const firstMatch = $derived(matches[0]?.items[0]?.id ?? null);
	const hasNoMatch = $derived(query.trim() !== "" && matches.length === 0);

	function onSearchKeydown(event: KeyboardEvent): void {
		if (event.key === "Enter" && firstMatch) {
			event.preventDefault();
			onSelect(firstMatch);
		} else if (event.key === "Escape") {
			query = "";
		}
	}
</script>

<nav class="rail" aria-label="Settings sections">
	<input
		class="input search"
		type="search"
		aria-label={SEARCH_PLACEHOLDER}
		placeholder={SEARCH_PLACEHOLDER}
		bind:value={query}
		onkeydown={onSearchKeydown}
	/>

	{#if hasNoMatch}
		<p class="no-match">No matching sections.</p>
	{/if}

	<div class="list">
		<ul>
			{#each matches as group (group.label)}
				<li class="group">
					<span class="group-name">{group.label}</span>
					<ul>
						{#each group.items as item (item.id)}
							<li>
								<button
									type="button"
									class="item"
									class:active={item.id === active}
									aria-current={item.id === active ? "true" : undefined}
									onclick={() => onSelect(item.id)}
								>
									{item.label}
								</button>
							</li>
						{/each}
					</ul>
				</li>
			{/each}
		</ul>
	</div>
</nav>

<style lang="scss">
	/* Sticks at the top of the settings scroll area while the sections scroll past. */
	.rail {
		position: sticky;
		top: 0;
		align-self: start;
		box-sizing: border-box;
		width: 100%;
		padding: 4px 0;
	}

	.search {
		width: 100%;
		margin-bottom: 16px;
		padding: 8px 12px;
		font-size: 0.9rem;
	}

	.no-match {
		margin: 0 0 16px;
		padding: 0 12px;
		color: var(--text-dim);
		font-size: 0.85rem;
	}

	ul {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.group ul {
		gap: 2px;
		margin-top: 6px;
	}

	.group-name {
		padding: 0 12px;
		color: var(--text-dim);
		font-size: 0.72rem;
		font-weight: 800;
		letter-spacing: 0.5px;
		text-transform: uppercase;
	}

	.item {
		position: relative;
		display: block;
		width: 100%;
		padding: 7px 12px;
		border: 0;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-muted);
		font: inherit;
		font-size: 0.9rem;
		font-weight: 600;
		text-align: left;
		cursor: pointer;
		transition: background var(--transition-fast), color var(--transition-fast);

		&:hover {
			background: var(--bg-elevated);
			color: var(--text-main);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}

		&.active {
			background: var(--bg-elevated);
			color: var(--text-main);

			/* Same accent bar as the main nav, inset so it sits inside the rounded row. */
			&::before {
				content: "";
				position: absolute;
				left: 0;
				top: 6px;
				bottom: 6px;
				width: 3px;
				border-radius: 0 var(--radius-pill) var(--radius-pill) 0;
				background: var(--accent-primary);
				box-shadow: 0 0 12px var(--accent-glow);
			}
		}
	}

	/* Under 900px the groups become one horizontal row of chips. The row scrolls on its own. */
	@media (max-width: 899px) {
		.rail {
			top: 0;
			z-index: 10;
			padding: 8px 0;
			background: var(--bg-base);
		}

		.list {
			overflow-x: auto;
		}

		ul {
			flex-direction: row;
			gap: 8px;
		}

		.group {
			display: flex;
			align-items: center;
			gap: 8px;
		}

		.group ul {
			margin-top: 0;
		}

		.group-name {
			display: none;
		}

		.item {
			width: auto;
			white-space: nowrap;
			border: 1px solid var(--glass-border);
			border-radius: var(--radius-pill);
			padding: 6px 14px;

			&.active {
				border-color: var(--accent-primary);

				&::before {
					display: none;
				}
			}
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.item {
			transition: none;
		}
	}
</style>
