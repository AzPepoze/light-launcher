<script lang="ts">
	import type { Snippet } from "svelte";
	import Icon from "./Icon.svelte";

	let {
		title,
		help,
		icon,
		children,
		actions
	}: {
		title: string;
		help?: string;
		icon?: string;
		children?: Snippet;
		/** Buttons or icon buttons shown at the right of the header. */
		actions?: Snippet;
	} = $props();
</script>

<section class="section">
	<header class="section-header">
		<div class="heading">
			{#if icon}
				<span class="section-icon"><Icon name={icon} size={22} /></span>
			{/if}
			<div class="heading-text">
				<h2>{title}</h2>
				{#if help}<p class="help">{help}</p>{/if}
			</div>
		</div>
		{#if actions}<div class="actions">{@render actions()}</div>{/if}
	</header>
	<div class="section-body">
		{#if children}{@render children()}{/if}
	</div>
</section>

<style lang="scss">
	.section {
		display: flex;
		flex-direction: column;
		gap: 20px;
		padding: 28px 32px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding-bottom: 16px;
		border-bottom: 2px solid rgba(255, 255, 255, 0.05);
	}

	.heading {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 0;
	}

	.section-icon {
		color: var(--accent-primary);
	}

	.heading-text {
		min-width: 0;

		h2 {
			margin: 0;
			color: var(--text-main);
			font-size: 1.1rem;
			font-weight: 700;
		}
	}

	.help {
		margin: 4px 0 0;
		color: var(--text-muted);
		font-size: 0.85rem;
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-shrink: 0;
	}

	.section-body {
		display: flex;
		flex-direction: column;
	}
</style>
