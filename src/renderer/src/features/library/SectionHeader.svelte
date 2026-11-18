<script lang="ts">
	import type { Snippet } from "svelte";
	import { gameCount } from "./librarySections";

	let {
		label,
		count,
		children
	}: {
		label: string;
		count: number;
		/** The section's actions, such as Sync. */
		children?: Snippet;
	} = $props();
</script>

<header class="header">
	<h2>{label}</h2>
	<p class="count">{gameCount(count)}</p>
	<div class="actions">
		{@render children?.()}
	</div>
</header>

<style lang="scss">
	.header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 4px 12px;
		margin-bottom: 12px;
	}

	h2 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.count {
		margin: 0;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin-left: auto;
	}
</style>
