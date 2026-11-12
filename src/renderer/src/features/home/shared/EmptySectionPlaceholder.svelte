<script lang="ts">
	export let icon: string = "folder_open";
	export let title: string = "No games found";
	export let subtitle: string = "";
</script>

<div class="folder-empty-placeholder">
	<span class="material-icons placeholder-icon">{icon}</span>
	<span class="placeholder-text">{title}</span>
	{#if subtitle}
		<span class="placeholder-subtext">{subtitle}</span>
	{/if}
	<slot />
</div>

<style lang="scss">
	.folder-empty-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		background: rgba(255, 255, 255, 0.015);
		border: 1px dashed rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-lg);
		margin: 8px 12px 24px 12px;
		text-align: center;
		gap: 8px;

		.placeholder-icon {
			font-size: 36px;
			color: var(--text-dim);
			opacity: 0.4;
			margin-bottom: 4px;
		}

		.placeholder-text {
			font-size: 0.95rem;
			font-weight: 700;
			color: var(--text-muted);
		}

		.placeholder-subtext {
			font-size: 0.8rem;
			color: var(--text-dim);
			max-width: 480px;
			line-height: 1.4;
		}
	}
</style>
