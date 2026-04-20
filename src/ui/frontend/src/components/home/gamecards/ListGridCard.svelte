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
	class="list-card"
	class:running={isRunning}
	class:selection-mode={isSelectionMode}
	class:selected={isSelected}
	on:click={handleLaunch}
	role="button"
	tabindex="0"
	on:keydown={(e) => e.key === "Enter" && handleLaunch()}
>
	{#if isSelectionMode}
		<div class="selection-checkbox">
			<div class="checkbox" class:checked={isSelected}>
				{#if isSelected}
					<span class="material-icons" style="font-size: 16px;"
						>check</span
					>
				{/if}
			</div>
		</div>
	{/if}

	<div class="icon-section">
		{#if icon}
			<img src={icon} alt={game.name} class="game-icon" />
		{:else}
			<div class="fallback-wrapper">
				<span
					class="material-icons"
					style="font-size: 32px; color: var(--text-dim); opacity: 0.5;"
					>rocket_launch</span
				>
			</div>
		{/if}

		{#if isRunning}
			<div class="running-indicator-small">
				<span class="pulse"></span>
			</div>
		{/if}
	</div>

	<div class="content-section">
		<div class="info">
			<span class="game-name">{game.name}</span>
			<span class="game-path"
				>{game.path || game.config.LauncherPath}</span
			>
		</div>

		<div class="actions">
			<button class="action-btn play" title="Play Now">
				<span class="material-icons">play_arrow</span>
			</button>
			<button
				class="action-btn config"
				title="Configure"
				on:click|stopPropagation={handleConfigure}
			>
				<span class="material-icons">settings</span>
			</button>
		</div>
	</div>
</div>

<style lang="scss">
	.list-card {
		display: flex;
		align-items: center;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 16px;
		padding: 12px 20px;
		gap: 20px;
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.23, 1, 0.32, 1);
		max-width: 100%;

		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--glass-border-bright);
			transform: translateX(8px);

			.game-icon {
				transform: scale(1.1);
			}

			.play .material-icons {
				transform: scale(1.2);
			}
		}

		&.running {
			border-color: var(--success, #22c55e);
			background: linear-gradient(
				90deg,
				rgba(34, 197, 94, 0.1) 0%,
				rgba(34, 197, 94, 0.02) 100%
			);
		}

		&.selected {
			border-color: var(--accent-primary);
			background: rgba(var(--accent-primary-rgb, 255, 255, 255), 0.05);

			.checkbox {
				background: var(--accent-primary) !important;
				border-color: var(--accent-primary) !important;
				color: #000;
			}
		}
	}

	.selection-checkbox {
		flex-shrink: 0;

		.checkbox {
			width: 24px;
			height: 24px;
			border: 2px solid var(--glass-border);
			border-radius: 6px;
			background: var(--glass-bg);
			display: flex;
			align-items: center;
			justify-content: center;
			transition: all 0.2s;
			color: transparent;
		}
	}

	.icon-section {
		height: 64px;
		aspect-ratio: 1/1;
		border-radius: 12px;
		overflow: hidden;
		position: relative;
		flex-shrink: 0;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border);

		.game-icon {
			width: 100%;
			height: 100%;
			object-fit: cover;
			transition: transform 0.4s;
		}

		.fallback-wrapper {
			width: 100%;
			height: 100%;
			display: flex;
			align-items: center;
			justify-content: center;
			opacity: 1;
		}

		.running-indicator-small {
			position: absolute;
			top: 4px;
			right: 4px;
			width: 8px;
			height: 8px;
			background: var(--success, #22c55e);
			border-radius: 50%;
			box-shadow: 0 0 8px var(--success, #22c55e);

			.pulse {
				position: absolute;
				inset: 0;
				background: inherit;
				border-radius: inherit;
				animation: ping 1.5s cubic-bezier(0, 0, 0.2, 1) infinite;
			}
		}
	}

	.content-section {
		flex: 1;
		display: flex;
		justify-content: space-between;
		align-items: center;
		min-width: 0;
	}

	.info {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;

		.game-name {
			font-weight: 700;
			color: var(--text-main);
			font-size: 1.1rem;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.game-path {
			font-size: 0.8rem;
			color: var(--text-dim);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			max-width: 100%;
		}
	}

	.actions {
		display: flex;
		gap: 8px;
	}
	.action-btn {
		.material-icons {
			font-size: 20px;
			transition: transform 0.2s;
			display: flex;
			align-items: center;
			justify-content: center;
		}

		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--glass-border-bright);
		}

		&.play:hover {
			background: var(--success, #22c55e);
			border-color: transparent;
		}
	}

	@keyframes ping {
		75%,
		100% {
			transform: scale(2.5);
			opacity: 0;
		}
	}
</style>
