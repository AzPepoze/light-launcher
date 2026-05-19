<script lang="ts">
	export let onLaunch: () => void;
</script>

<button class="btn primary launch-btn" on:click={onLaunch}>
	<span class="material-icons run-icon">rocket_launch</span>
	<span>LAUNCH GAME</span>
</button>

<style lang="scss">
	.launch-btn {
		width: 100%;
		height: 60px;
		padding: 0 32px;
		border-radius: var(--radius-pill);
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		background: var(--accent-primary) !important;
		color: var(--bg-base) !important;
		border: none;
		box-shadow: 0 4px 14px var(--accent-glow);
		cursor: pointer;
		transition: background var(--transition-fast), box-shadow var(--transition-fast) !important;
		transform: none !important;
		font-weight: 800;
		font-size: 1.15rem;
		text-transform: uppercase;
		letter-spacing: 2px;

		span {
			display: block;
		}

		&:hover {
			background: var(--accent-hover) !important;
			box-shadow: 0 8px 24px var(--accent-glow);
			transform: none !important;

			.run-icon {
				transform: rotate(-15deg);
			}
		}

		&:active {
			transform: none !important;
		}

		.run-icon {
			font-size: 28px;
			transition: transform var(--transition-spring);
		}
	}
</style>
