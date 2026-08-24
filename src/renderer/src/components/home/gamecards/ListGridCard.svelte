<script lang="ts">
	import CardRunningIndicator from "./shared/CardRunningIndicator.svelte";
	import CardSelectionCheckbox from "./shared/CardSelectionCheckbox.svelte";
	import { getGamePath } from "@lib/gameUtils";

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
		<CardSelectionCheckbox {isSelected} inline={true} />
	{/if}

	<div class="icon-section">
		{#if icon}
			<img src={icon} alt={game.name} class="game-icon" loading="lazy" />
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
			<CardRunningIndicator compact={true} />
		{/if}
	</div>

	<div class="content-section">
		<div class="info">
			<span class="game-name">{game.name}</span>
			<span class="game-path">{getGamePath(game)}</span>
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
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		padding: 8px 16px;
		gap: 12px;
		cursor: pointer;
		transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), box-shadow var(--transition-fast);
		width: 100%;
		max-width: 100%;
		min-width: 0;
		box-sizing: border-box;
		overflow: visible;

		&:hover {
			border-color: var(--accent-primary);
			box-shadow: 0 4px 20px rgba(0,0,0,0.3);

			.game-icon {
				transform: scale(1.1);
			}

			.play .material-icons {
				transform: scale(1.15);
			}
		}

		&.running {
			border-color: var(--success, #44ffaa);
			background: var(--bg-surface);
			box-shadow: 0 0 16px rgba(68, 255, 170, 0.15);
		}

		&.selected {
			border-color: var(--accent-primary);
			background: var(--bg-elevated);
		}
	}

	.icon-section {
		height: 108px;
		aspect-ratio: 1/1;
		border-radius: var(--radius-lg);
		overflow: hidden;
		position: relative;
		flex-shrink: 0;
		background: var(--bg-base);
		border: none;

		.game-icon {
			width: 100%;
			height: 100%;
			object-fit: cover;
			border-radius: var(--radius-lg);
			transition: transform 0.4s var(--ease-spring);
		}

		.fallback-wrapper {
			width: 100%;
			height: 100%;
			display: flex;
			align-items: center;
			justify-content: center;
			opacity: 1;
			border-radius: var(--radius-lg);
			overflow: hidden;
		}
	}

	.content-section {
		flex: 1;
		display: flex;
		justify-content: space-between;
		align-items: center;
		min-width: 0;
		overflow: visible;
		gap: 12px;
	}

	.info {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
		flex: 1;
		overflow: hidden;

		.game-name {
			font-weight: 800;
			color: var(--text-main);
			font-size: 1.15rem;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			letter-spacing: -0.3px;
			max-width: 100%;
		}

		.game-path {
			font-size: 0.8rem;
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			max-width: 100%;
			min-width: 0;
			word-break: break-all;
		}
	}

	.actions {
		display: flex;
		gap: 10px;
		flex-shrink: 0;
	}

	.action-btn {
		background: var(--bg-elevated);
		border: 2px solid rgba(255, 255, 255, 0.05);
		color: var(--text-main);
		padding: 10px;
		border-radius: var(--radius-pill);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast);

		.material-icons {
			font-size: 20px;
			display: flex;
			align-items: center;
			justify-content: center;
		}

		&:hover {
			background: var(--bg-surface);
			border-color: var(--accent-secondary);
			transform: scale(1.15);
		}

		&:active {
			transform: scale(0.9);
		}

		&.play:hover {
			background: var(--accent-primary);
			border-color: var(--accent-primary);
			color: #ffffff;
			box-shadow: 0 4px 10px var(--accent-glow);
		}
	}
</style>
