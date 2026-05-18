<script lang="ts">
	export let title: string;
	export let subtitle: string = "";
	export let icon: string = "";
</script>

<div class="page-header">
	<div class="header-content">
		{#if icon}
			<span class="material-icons header-icon">{icon}</span>
		{/if}
		<div class="header-text">
			<h1>{title}</h1>
			{#if subtitle}
				<p class="subtitle">{subtitle}</p>
			{/if}
		</div>
	</div>
	<div class="header-actions">
		<slot />
	</div>
</div>

<style lang="scss">
	.page-header {
		display: flex;
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
		font-size: 2rem;
		color: var(--accent-primary);
		opacity: 0.8;
	}

	.header-text {
		display: flex;
		flex-direction: column;
		gap: 4px;

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
