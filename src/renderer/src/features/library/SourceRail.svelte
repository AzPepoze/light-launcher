<script lang="ts">
	import IconButton from "@components/shared/IconButton.svelte";
	import type { LibrarySource } from "../../lib/features/types";
	import { sourceIcon } from "./sourceIcon";

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

<nav class="rail" aria-label="Library sources">
	<ul>
		{#each sections as section (section.source)}
			<li class="item" class:active={section.source === selected}>
				<IconButton
					size={32}
					icon={sourceIcon(section.source)}
					label={section.label}
					filled={section.source === selected}
					aria-pressed={section.source === selected}
					onclick={() => onSelect(section.source)}
				/>
				<span class="count" aria-hidden="true">{section.count}</span>
				<span class="flyout" aria-hidden="true">
					<span class="flyout-name">{section.label}</span>
					<span class="flyout-count">{section.count}</span>
				</span>
			</li>
		{/each}
	</ul>
</nav>

<style lang="scss">
	/* Sticks below the page padding, inside the content. The column behind it is the panel. */
	.rail {
		position: sticky;
		top: 76px;
		z-index: 10;
		box-sizing: border-box;
		width: 100%;
		padding: 12px 4px;
	}

	ul {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.item {
		position: relative;
		display: grid;
		place-items: center;
		width: 48px;
		height: 48px;
		border-radius: var(--radius-md);

		&.active {
			background: var(--bg-elevated);

			&::before {
				content: "";
				position: absolute;
				left: -6px;
				top: 10px;
				bottom: 10px;
				width: 3px;
				border-radius: 2px;
				background: var(--accent-primary);
			}
		}

	}

	/* Sits in the tile corner, outside the 18px glyph, so it never covers the icon. */
	.count {
		position: absolute;
		z-index: 1;
		top: -4px;
		right: -4px;
		min-width: 18px;
		padding: 0 5px;
		border-radius: var(--radius-pill);
		background: var(--bg-surface);
		color: var(--text-muted);
		font-size: 0.6875rem;
		font-weight: 600;
		line-height: 18px;
		text-align: center;
		pointer-events: none;
	}

	/* The name slides out to the right of the icon. It shows at once on focus, no delay. */
	.flyout {
		position: absolute;
		left: calc(100% + 8px);
		top: 50%;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 4px 10px;
		border-radius: var(--radius-pill);
		background: var(--bg-elevated);
		color: var(--text-main);
		font-size: 0.875rem;
		white-space: nowrap;
		box-shadow: 0 4px 14px rgb(0 0 0 / 0.25);
		opacity: 0;
		transform: translate(-6px, -50%);
		pointer-events: none;
		transition:
			opacity 120ms var(--ease-out),
			transform 120ms var(--ease-out);
	}

	.flyout-count {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 600;
	}

	.item:hover .flyout,
	.item:focus-within .flyout {
		opacity: 1;
		transform: translate(0, -50%);
	}

	@media (prefers-reduced-motion: reduce) {
		.flyout {
			transform: translate(0, -50%);
			transition: opacity 120ms linear;
		}
	}

	/* Touch screens have no hover; the aria-label still names each source. */
	@media (max-width: 899px) {
		.flyout {
			display: none;
		}
	}
</style>
