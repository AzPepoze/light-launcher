<script lang="ts">
	import type { Snippet } from "svelte";
	import Icon from "./Icon.svelte";

	let {
		selected = false,
		onToggle,
		icon,
		count = null,
		children
	}: {
		selected?: boolean;
		onToggle: () => void;
		icon?: string;
		count?: number | null;
		children: Snippet;
	} = $props();
</script>

<button type="button" class="choice-row" class:selected aria-pressed={selected} onclick={onToggle}>
	{#if icon}<Icon name={icon} size={16} />{/if}
	<span class="label">{@render children()}</span>
	{#if count !== null}<span class="count">{count}</span>{/if}
</button>

<style lang="scss">
	.choice-row {
		display: flex;
		align-items: center;
		gap: 8px;
		box-sizing: border-box;
		width: 100%;
		min-height: 32px;
		padding: 4px 10px;
		border: none;
		border-radius: var(--radius-md);
		background: transparent;
		color: var(--text-main);
		font: inherit;
		text-align: left;
		cursor: pointer;
		transition: background-color 120ms ease;

		&:hover {
			background: var(--bg-elevated);
		}

		&.selected {
			background: color-mix(in srgb, var(--accent-primary) 16%, transparent);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: -2px;
		}

		@media (prefers-reduced-motion: reduce) {
			transition: none;
		}
	}

	.label {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.count {
		margin-left: auto;
		color: var(--text-muted);
		font-size: 0.8125rem;
	}
</style>
