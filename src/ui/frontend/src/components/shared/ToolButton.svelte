<script lang="ts">
	export let icon: string;
	export let title: string;
	export let subtitle: string;
	export let loading: boolean = false;
	export let onClick: () => void;
</script>

<button class="tool-card" class:loading on:click={onClick} disabled={loading}>
	{#if loading}
		<div class="spinner"></div>
	{:else}
		<div class="icon">
			{#if icon.includes("/") || icon.includes(".svg") || icon.startsWith("data:")}
				<img src={icon} alt={title} class="svg-icon" />
			{:else}
				<span class="material-icons">{icon}</span>
			{/if}
		</div>
	{/if}
	<div class="text">
		<h3>{title}</h3>
		<p>{loading ? "Opening..." : subtitle}</p>
	</div>
</button>

<style lang="scss">
	.tool-card {
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 16px;
		padding: 16px;
		display: flex;
		align-items: center;
		gap: 16px;
		cursor: pointer;
		transition: all 0.2s;
		color: var(--text-main);
		text-align: left;
		position: relative;
		overflow: hidden;

		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--accent-primary);
			transform: translateY(-2px);
		}

		&.loading {
			opacity: 0.7;
			cursor: wait;
			pointer-events: none;
		}

		.icon {
			font-size: 1.75rem;
			display: flex;
			align-items: center;
			justify-content: center;
			width: 32px;
			height: 32px;
			color: var(--accent-primary);

			.svg-icon {
				width: 100%;
				height: 100%;
				opacity: 0.8;
			}
		}
		h3 {
			font-size: 1rem;
			margin: 0;
			font-weight: 700;
		}
		p {
			font-size: 0.75rem;
			margin: 2px 0 0 0;
			color: var(--text-dim);
		}
	}

	.spinner {
		width: 28px;
		height: 28px;
		border: 3px solid var(--glass-border);
		border-radius: 50%;
		border-top-color: var(--accent-primary);
		animation: spin 1s ease-in-out infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
