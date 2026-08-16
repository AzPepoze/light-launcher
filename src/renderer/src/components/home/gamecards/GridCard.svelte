<script lang="ts">
	import CardRunningIndicator from "./shared/CardRunningIndicator.svelte";
	import CardSelectionCheckbox from "./shared/CardSelectionCheckbox.svelte";

	export let game: any;
	export let icon: string = "";
	export let isRunning: boolean = false;
	export let isSelectionMode: boolean = false;
	export let isSelected: boolean = false;
	export let onLaunch: (game: any) => void = () => {};
	export let onConfigure: (game: any) => void = () => {};
	export let onSelect: (game: any, shiftKey: boolean) => void = () => {};

	function handleLaunch(event?: MouseEvent) {
		if (isSelectionMode) {
			onSelect(game, event ? event.shiftKey : false);
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
			<CardSelectionCheckbox {isSelected} />
		{/if}

		{#if isRunning}
			<CardRunningIndicator />
		{/if}

		<div class="rainbow-glow"></div>

		<div class="icon-wrapper">
			{#if icon}
				<img src={icon} alt={game.name} class="game-icon" loading="lazy" />
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
		gap: 12px;
		transition: transform var(--transition-spring), filter var(--transition-fast);
		position: relative;
		width: 100%;
		max-width: 200px;
		margin: 6px;

		&.selected {
			.game-icon-container {
				border-color: var(--accent-primary);
				box-shadow: 0 0 24px var(--accent-glow);
			}
		}

		&:hover {
			transform: scale(1.08) translateY(-6px);

			.rainbow-glow {
				opacity: 0.8;
				animation: rainbow-glow-animation 3s linear infinite;
			}

			.game-icon-container {
				border-color: var(--accent-primary);
				box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);

				.play-overlay {
					opacity: 1;
				}

				img.game-icon {
					transform: scale(1.1);
				}
			}

			.game-footer .game-name {
				color: var(--text-main);
				white-space: normal;
				overflow: visible;
				word-break: break-all;
			}

			.config-btn {
				opacity: 1;
				visibility: visible;
				transform: scale(1);
			}
		}

		&:active {
			transform: scale(0.97);
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
				border-color: var(--success, #44ffaa);
				box-shadow: 0 0 24px rgba(68, 255, 170, 0.35);
			}
		}
	}

	.game-icon-container {
		aspect-ratio: 1;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		position: relative;
		cursor: pointer;
		transition: transform var(--transition-spring), border-color var(--transition-fast), box-shadow var(--transition-fast);
		z-index: 1;
	}

	.icon-wrapper {
		position: absolute;
		inset: 2px;
		background: var(--bg-base);
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
		background: var(--accent-primary);
		opacity: 0;
		transition: opacity 0.4s;
		z-index: 1;
		filter: blur(4px) brightness(1.2);
		border-radius: calc(var(--radius-lg) + 4px);
	}

	.game-icon {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.6s var(--ease-spring);
	}

	.system-icon {
		font-size: 64px;
		color: var(--text-muted);
		opacity: 0.3;
	}

	.play-overlay {
		position: absolute;
		inset: 0;
		background: rgba(var(--bg-base-rgb), 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity var(--transition-fast);
		z-index: 3;
		border-radius: calc(var(--radius-lg) - 2px);

		.launch-icon-large {
			font-size: 64px;
			color: #fff;
			filter: drop-shadow(0 0 10px rgba(0, 0, 0, 0.8))
				drop-shadow(0 0 20px var(--accent-glow));
			transform: scale(0.8);
			transition: transform var(--transition-spring);
		}
	}

	.game-card:hover .launch-icon-large {
		transform: scale(1);
	}

	.game-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0 4px;

		.game-name {
			font-size: 0.95rem;
			font-weight: 800;
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			flex: 1;
			transition: color var(--transition-fast);
			letter-spacing: -0.2px;
			line-height: 1.2;
		}

		.config-btn {
			background: var(--bg-surface);
			border: 2px solid rgba(255, 255, 255, 0.05);
			padding: 8px;
			border-radius: var(--radius-pill);
			cursor: pointer;
			display: flex;
			align-items: center;
			justify-content: center;
			transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast);
			color: var(--text-main);
			margin-left: 6px;

			&:hover {
				background: var(--bg-elevated);
				transform: rotate(90deg) scale(1.1);
				border-color: var(--accent-secondary);
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
</style>
