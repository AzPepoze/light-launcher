<script lang="ts">
	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let isSelectionMode: boolean = false;
	export let isSelected: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};
	export let onSelect: (game: any) => void = () => {};

	function handleLaunch() {
		if (isSelectionMode) {
			onSelect(game);
			return;
		}
		onLaunch(game);
	}

	function handleConfigure() {
		onConfigure(game);
	}
</script>

<div
	class="game-card"
	class:running={isRunning}
	class:selection-mode={isSelectionMode}
	class:selected={isSelected}
>
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="game-icon-container"
		on:click={handleLaunch}
		role="button"
		tabindex="0"
	>
		{#if isSelectionMode}
			<div class="selection-overlay">
				<div class="checkbox">
					{#if isSelected}
						<span class="material-icons" style="font-size: 16px;"
							>check</span
						>
					{/if}
				</div>
			</div>
		{/if}

		{#if isRunning}
			<div class="running-indicator">
				<span class="pulse"></span>
				<span class="indicator-text">RUNNING</span>
			</div>
		{/if}

		<div class="rainbow-glow"></div>

		<div class="icon-wrapper">
			{#if icon}
				<img src={icon} alt={game.name} class="game-icon" />
			{:else}
				<span class="material-icons system-icon">rocket_launch</span>
			{/if}
		</div>

		<div class="play-overlay">
			<span class="material-icons launch-icon-large">play_arrow</span>
		</div>
	</div>

	<div class="game-footer">
		<span class="game-name" title={game.name}>{game.name}</span>
		<button
			class="config-btn"
			title="Configure"
			on:click|stopPropagation={handleConfigure}
		>
			<span class="material-icons">settings</span>
		</button>
	</div>
</div>

<style lang="scss">
	.game-card {
		display: flex;
		flex-direction: column;
		gap: 8px;
		transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
		position: relative;
		width: 100%;
		max-width: 160px;
		margin: 4px;

		&.selected {
			.game-icon-container {
				border-color: var(--accent-primary);
				box-shadow: 0 0 20px
					rgba(var(--accent-primary-rgb, 255, 255, 255), 0.3);
			}

			.checkbox {
				background: var(--accent-primary) !important;
				border-color: var(--accent-primary) !important;
				color: #000;
			}
		}

		&:hover {
			transform: scale(1.05);

			.rainbow-glow {
				opacity: 0.8;
				animation: rainbow-glow-animation 3s linear infinite;
			}

			.game-icon-container {
				border-color: var(--glass-border-bright);
				box-shadow: var(--glass-shadow);

				.play-overlay {
					opacity: 1;
				}

				img.game-icon {
					transform: scale(1.1);
				}
			}

			.game-footer .game-name {
				color: var(--text-main);
				text-shadow: 0 0 10px var(--glass-surface);
			}

			.config-btn {
				opacity: 1;
				visibility: visible;
				transform: translateX(0);
			}

			.game-footer .game-name {
				white-space: normal;
				overflow: visible;
				word-break: break-all;
			}
		}

		&.selection-mode:hover {
			.play-overlay {
				opacity: 0 !important;
			}
		}

		&.running {
			.rainbow-glow {
				opacity: 1;
				animation: rainbow-glow-animation 2s linear infinite;
			}

			.game-icon-container {
				border-color: var(--success, #22c55e);
				box-shadow: 0 0 20px rgba(34, 197, 94, 0.3);
			}
		}
	}

	.game-icon-container {
		aspect-ratio: 1;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
		cursor: pointer;
		transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
		z-index: 1;
	}

	.icon-wrapper {
		position: absolute;
		inset: 2px;
		background: var(--glass-bg);
		border-radius: calc(var(--radius-lg) - 2px);
		z-index: 2;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.rainbow-glow {
		position: absolute;
		inset: -2px;
		background: conic-gradient(
			from 0deg,
			#ff0000,
			#ffff00,
			#00ff00,
			#00ffff,
			#0000ff,
			#ff00ff,
			#ff0000
		);
		opacity: 0;
		transition: opacity 0.4s;
		z-index: 1;
		filter: blur(3px) brightness(1.5);
		border-radius: calc(var(--radius-lg) + 4px);
	}

	.game-icon {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.6s cubic-bezier(0.23, 1, 0.32, 1);
	}

	.system-icon {
		font-size: 64px;
		color: var(--text-dim);
		opacity: 0.4;
	}

	.play-overlay {
		position: absolute;
		inset: 0;
		background: transparent;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.3s;
		z-index: 3;

		.launch-icon-large {
			font-size: 60px;
			color: #fff;
			filter: drop-shadow(0 0 10px rgba(0, 0, 0, 0.8))
				drop-shadow(0 0 20px rgba(0, 0, 0, 0.4));
			transform: scale(0.8);
			transition: transform 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
		}
	}

	.selection-overlay {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: flex-start;
		justify-content: flex-end;
		padding: 12px;
		z-index: 5;
		background: var(--glass-low-alpha);
		border-radius: calc(var(--radius-lg) - 2px);

		.checkbox {
			width: 24px;
			height: 24px;
			border: 2px solid var(--glass-border-bright);
			border-radius: var(--radius-sm);
			background: var(--glass-surface);
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.2s;
			color: transparent;
		}
	}

	.game-card:hover .launch-icon-large {
		transform: scale(1);
	}

	.running-indicator {
		position: absolute;
		top: 12px;
		right: 12px;
		background: var(--success, #22c55e);
		color: #fff;
		padding: 4px 10px;
		border-radius: var(--radius-xl);
		font-size: 0.65rem;
		font-weight: 900;
		display: flex;
		align-items: center;
		gap: 6px;
		letter-spacing: 0.5px;
		z-index: 10;
		box-shadow: 0 4px 10px rgba(0, 0, 0, 0.3);

		.pulse {
			width: 6px;
			height: 6px;
			background: #fff;
			border-radius: 50%;
			display: inline-block;
			animation: blink 1s infinite;
		}
	}

	.game-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0 4px;

		.game-name {
			font-size: 0.85rem;
			font-weight: 700;
			color: var(--text-main);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			flex: 1;
			transition: all 0.3s;
			letter-spacing: -0.2px;
			line-height: 1.2;
		}

		.config-btn {
			background: var(--glass-surface);
			border: 1px solid var(--glass-border);
			padding: 6px;
			border-radius: var(--radius-md);
			cursor: pointer;
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.3s;
			opacity: 0.6;
			visibility: visible;
			transform: translateX(0);
			color: var(--text-main);
			margin-left: 6px;

			&:hover {
				background: var(--glass-border-bright);
				transform: rotate(45deg) !important;
				border-color: var(--glass-border-bright);
			}

			.material-icons {
				font-size: 16px;
			}
		}
	}

	@keyframes rainbow-glow-animation {
		0% {
			filter: blur(3px) hue-rotate(0deg);
		}
		50% {
			filter: blur(3px) hue-rotate(180deg);
			transform: scale(1.02);
		}
		100% {
			filter: blur(3px) hue-rotate(360deg);
		}
	}

	@keyframes blink {
		0%,
		100% {
			opacity: 1;
			transform: scale(1);
		}
		50% {
			opacity: 0.5;
			transform: scale(0.8);
		}
	}
</style>
