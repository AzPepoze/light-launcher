<script lang="ts">
	import type { Snippet } from "svelte";
	import Icon from "./Icon.svelte";

	let {
		title,
		subtitle = "",
		icon = "",
		isCustomIcon = false,
		children
	}: {
		title: string;
		subtitle?: string;
		icon?: string;
		isCustomIcon?: boolean;
		/** Buttons and other actions shown on the right. */
		children?: Snippet;
	} = $props();
</script>

<div class="page-header">
	<div class="header-content">
		{#if icon && isCustomIcon}
			<img src={icon} class="header-custom-icon" alt="" />
		{:else if icon}
			<span class="header-icon"><Icon name={icon} size={32} /></span>
		{/if}
		<div class="header-text">
			<h1>{title}</h1>
			{#if subtitle}
				<p class="subtitle">{subtitle}</p>
			{/if}
		</div>
	</div>
	<div class="header-actions">
		{@render children?.()}
	</div>
</div>

<style lang="scss">
	.page-header {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		justify-content: space-between;
		gap: 16px;
		margin-bottom: 28px;
		flex-shrink: 0;
	}

	.header-content {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.header-icon {
		display: inline-flex;
		color: var(--accent-primary);
		opacity: 0.8;
	}

	.header-custom-icon {
		width: 32px;
		height: 32px;
		object-fit: contain;
		filter: drop-shadow(0 0 4px var(--accent-glow));
	}

	.header-text {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;

		h1 {
			font-size: 1.75rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 2px;
			margin: 0;
			line-height: 1.1;
		}
	}

	@media (max-width: 900px) {
		.header-text h1 {
			font-size: 1.35rem;
			letter-spacing: 1.5px;
		}
	}

	.subtitle {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-muted);
		margin: 0;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-shrink: 0;
	}
</style>
