<script lang="ts">
	export let compact: boolean = false;
	export let label: string = "RUNNING";
</script>

{#if compact}
	<div class="running-indicator-small">
		<span class="pulse"></span>
	</div>
{:else}
	<div class="running-indicator">
		<span class="pulse"></span>
		<span class="indicator-text">{label}</span>
	</div>
{/if}

<style lang="scss">
	.running-indicator {
		position: absolute;
		top: 10px;
		left: 10px;
		display: flex;
		align-items: center;
		gap: 6px;
		background: rgba(10, 10, 15, 0.85);
		backdrop-filter: blur(8px);
		border: 1px solid rgba(0, 230, 118, 0.4);
		padding: 4px 8px;
		border-radius: var(--radius-full);
		z-index: 10;
		box-shadow: 0 0 12px rgba(0, 230, 118, 0.3);

		.indicator-text {
			font-size: 0.65rem;
			font-weight: 800;
			color: var(--success);
			letter-spacing: 0.8px;
		}
	}

	.running-indicator-small {
		position: absolute;
		bottom: -2px;
		right: -2px;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 12px;
		height: 12px;
		background: var(--bg-surface);
		border-radius: 50%;
		z-index: 10;
	}

	.pulse {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--success);
		box-shadow: 0 0 8px var(--success);
		animation: pulse-animation 1.5s infinite;
	}

	@keyframes pulse-animation {
		0% {
			transform: scale(0.95);
			box-shadow: 0 0 0 0 rgba(0, 230, 118, 0.7);
		}
		70% {
			transform: scale(1);
			box-shadow: 0 0 0 6px rgba(0, 230, 118, 0);
		}
		100% {
			transform: scale(0.95);
			box-shadow: 0 0 0 0 rgba(0, 230, 118, 0);
		}
	}
</style>
