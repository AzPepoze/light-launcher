<script lang="ts">
	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let active: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};

	function handleLaunch() {
		onLaunch(game);
	}

	function handleConfigure() {
		onConfigure(game);
	}
</script>

<div class="perspective-card" class:active class:running={isRunning}>
	<div
		class="card-inner"
		on:click={handleLaunch}
		role="button"
		tabindex="0"
		on:keydown={(e) => e.key === "Enter" && handleLaunch()}
	>
		<div class="glow"></div>

		<div class="image-container">
			{#if icon}
				<img src={icon} alt={game.name} class="game-icon" />
			{:else}
				<div class="fallback">
					<span class="material-icons">rocket_launch</span>
				</div>
			{/if}
		</div>

		<div class="info-overlay">
			<div class="name-row">
				<span class="name">{game.name}</span>
				<button
					class="config-btn-small"
					title="Configure"
					on:click|stopPropagation={handleConfigure}
				>
					<span class="material-icons" style="font-size: 14px;"
						>settings</span
					>
				</button>
			</div>
			{#if isRunning}
				<div class="status">RUNNING</div>
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	.perspective-card {
		width: 220px;
		height: 310px;
		cursor: pointer;
		transition: transform var(--transition-spring), filter var(--transition-fast);
		position: relative;
		perspective: 1000px;
		flex-shrink: 0;
		transform: rotateY(-15deg);

		&:hover {
			transform: scale(1.1) translateY(-10px) rotateY(0deg);
			z-index: 10;

			.card-inner {
				box-shadow:
					0 30px 60px rgba(0, 0, 0, 0.8),
					0 0 24px var(--accent-glow);
				border-color: var(--accent-primary);
			}

			.glow {
				opacity: 0.8;
			}

			.info-overlay {
				opacity: 1;
				transform: translateY(0);
			}
		}

		&.running {
			.card-inner {
				border-color: var(--success, #44ffaa);
				box-shadow: 0 0 30px rgba(68, 255, 170, 0.35);
			}
			.status {
				color: var(--success, #44ffaa);
			}
		}
	}

	.card-inner {
		width: 100%;
		height: 100%;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		overflow: hidden;
		position: relative;
		transition: transform var(--transition-spring), border-color var(--transition-fast), box-shadow var(--transition-fast);
		box-shadow: 0 10px 20px rgba(0, 0, 0, 0.3);
	}

	.glow {
		position: absolute;
		inset: 0;
		background: radial-gradient(
			circle at 50% 0%,
			var(--accent-glow),
			transparent
		);
		opacity: 0.2;
		transition: opacity 0.4s;
		pointer-events: none;
	}

	.image-container {
		width: 100%;
		height: 100%;

		.game-icon {
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		.fallback {
			width: 100%;
			height: 100%;
			display: flex;
			align-items: center;
			justify-content: center;
			background: var(--bg-base);

			.material-icons {
				font-size: 48px;
				opacity: 0.15;
				color: var(--text-muted);
			}
		}
	}

	.info-overlay {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		padding: 20px;
		background: linear-gradient(transparent, var(--bg-base));
		display: flex;
		flex-direction: column;
		gap: 4px;
		opacity: 0.8;
		transform: translateY(5px);
		transition: all 0.4s;

		.name {
			font-weight: 800;
			color: var(--text-main);
			font-size: 1.05rem;
			text-shadow: 0 2px 4px rgba(0, 0, 0, 0.4);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.name-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			gap: 8px;
		}

		.config-btn-small {
			background: var(--bg-base);
			border: 2px solid rgba(255, 255, 255, 0.05);
			color: var(--text-main);
			padding: 4px;
			border-radius: var(--radius-pill);
			cursor: pointer;
			display: flex;
			transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast);

			&:hover {
				background: var(--bg-elevated);
				border-color: var(--accent-secondary);
				transform: rotate(90deg) scale(1.1);
			}
		}

		.status {
			font-size: 0.7rem;
			font-weight: 900;
			letter-spacing: 1px;
		}
	}
</style>
