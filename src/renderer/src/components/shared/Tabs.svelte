<script lang="ts">
	import Badge from "./Badge.svelte";
	import Icon from "./Icon.svelte";
	import { nextTabId, type TabItem } from "./tabs";

	let {
		tabs,
		value,
		onChange,
		"aria-label": ariaLabel,
		showZeroBadges = false
	}: {
		tabs: TabItem[];
		value: string;
		onChange: (id: string) => void;
		"aria-label": string;
		/** Shows a badge with a zero count too. Off by default, since most tabs hide an empty count. */
		showZeroBadges?: boolean;
	} = $props();

	/* Arrow keys, Home and End move between tabs and select the one reached. */
	function handleKeydown(event: KeyboardEvent) {
		const next = nextTabId(
			tabs.map((tab) => tab.id),
			value,
			event.key
		);
		if (next === null) return;
		event.preventDefault();
		onChange(next);
		const strip = event.currentTarget as HTMLElement;
		strip.querySelector<HTMLElement>(`[data-id="${CSS.escape(next)}"]`)?.focus();
	}
</script>

<div class="tabs" role="tablist" aria-label={ariaLabel} onkeydown={handleKeydown}>
	{#each tabs as tab (tab.id)}
		{@const selected = tab.id === value}
		<button
			type="button"
			role="tab"
			class="tab"
			class:selected
			aria-selected={selected}
			tabindex={selected ? 0 : -1}
			data-id={tab.id}
			onclick={() => onChange(tab.id)}
		>
			{#if tab.icon}<Icon name={tab.icon} size={18} />{/if}
			<span>{tab.label}</span>
			{#if tab.badge !== undefined && (tab.badge > 0 || showZeroBadges)}
				<Badge tone={tab.badge > 0 ? "accent" : "neutral"}>{tab.badge}</Badge>
			{/if}
		</button>
	{/each}
</div>

<style lang="scss">
	.tabs {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.tab {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 14px;
		border: 1px solid transparent;
		border-radius: var(--radius-pill);
		background: transparent;
		color: var(--text-muted);
		font: inherit;
		font-size: 0.9rem;
		font-weight: 700;
		cursor: pointer;
		transition:
			color var(--transition-fast),
			background var(--transition-fast);

		&:hover {
			color: var(--text-main);
		}

		&.selected {
			color: var(--accent-primary);
			border-color: color-mix(in srgb, var(--accent-primary) 45%, transparent);
			background: color-mix(in srgb, var(--accent-primary) 12%, transparent);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}
</style>
