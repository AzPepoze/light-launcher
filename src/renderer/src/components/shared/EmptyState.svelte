<script lang="ts">
	import type { Snippet } from "svelte";
	import Icon from "./Icon.svelte";

	let {
		icon,
		title,
		hint,
		children
	}: {
		icon: string;
		title: string;
		hint?: string;
		/** Optional action, such as a button. */
		children?: Snippet;
	} = $props();
</script>

<div class="empty-state">
	<Icon name={icon} size={40} class="empty-icon" />
	<p class="title">{title}</p>
	{#if hint}<p class="hint">{hint}</p>{/if}
	{#if children}<div class="action">{@render children()}</div>{/if}
</div>

<style lang="scss">
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 48px 24px;
		text-align: center;
	}

	.empty-state :global(.empty-icon) {
		color: var(--text-dim);
		margin-bottom: 8px;
	}

	.title {
		margin: 0;
		color: var(--text-main);
		font-size: 1.05rem;
		font-weight: 700;
	}

	.hint {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.action {
		margin-top: 12px;
	}
</style>
