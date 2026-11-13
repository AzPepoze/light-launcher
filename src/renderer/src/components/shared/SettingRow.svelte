<script lang="ts">
	import type { Snippet } from "svelte";

	let {
		title,
		help,
		children
	}: {
		title: string;
		help?: string;
		/** The control, such as a switch or a number field. */
		children?: Snippet;
	} = $props();
</script>

<div class="setting-row">
	<div class="text">
		<span class="title">{title}</span>
		{#if help}<span class="help">{help}</span>{/if}
	</div>
	<div class="control">
		{#if children}{@render children()}{/if}
	</div>
</div>

<style lang="scss">
	.setting-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 24px;
		padding: 16px 0;
	}

	.setting-row + :global(.setting-row) {
		border-top: 2px solid rgba(255, 255, 255, 0.05);
	}

	.text {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}

	.title {
		color: var(--text-main);
		font-size: 0.95rem;
		font-weight: 600;
	}

	.help {
		color: var(--text-muted);
		font-size: 0.8rem;
	}

	.control {
		display: flex;
		align-items: center;
		flex-shrink: 0;
	}
</style>
