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
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		padding: 12px 14px;
		display: flex;
		align-items: center;
		gap: 12px;
		cursor: pointer;
		transition: transform var(--transition-spring), border-color var(--transition-fast), background var(--transition-fast);
		color: var(--text-main);
		text-align: left;
		position: relative;
		overflow: hidden;

		&:hover {
			background: var(--bg-elevated);
			border-color: var(--accent-primary);
			transform: scale(1.03);
		}

		&:active {
			transform: scale(0.97);
		}

		&.loading {
			opacity: 0.7;
			cursor: wait;
			pointer-events: none;
		}

		.icon {
			font-size: 1.5rem;
			display: flex;
			align-items: center;
			justify-content: center;
			width: 28px;
			height: 28px;
			color: var(--accent-primary);
			flex-shrink: 0;

			.svg-icon {
				width: 100%;
				height: 100%;
			}
		}

		.text {
			min-width: 0;
			flex: 1;
		}

		h3 {
			font-size: 0.9rem;
			margin: 0;
			font-weight: 800;
			text-transform: uppercase;
			letter-spacing: 0.5px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
		p {
			font-size: 0.75rem;
			margin: 2px 0 0 0;
			color: var(--text-muted);
			font-weight: 600;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
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
