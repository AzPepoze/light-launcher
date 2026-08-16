<script lang="ts">
	import { fade } from "svelte/transition";
	import { commandPaletteState as state } from "./CommandPaletteState.svelte";

	export let show = false;
	export let onClose: () => void = () => {};

	$: if (show) {
		state.onShowChange(true);
	} else {
		state.onShowChange(false);
	}

	// Tell state how to close the UI layer
	$: state.onCloseCallback = onClose;

	$: if (state.searchQuery !== undefined) {
		state.filterItems();
	}
</script>

{#if show}
	<div
		class="palette-backdrop"
		role="button"
		tabindex="-1"
		on:click={() => state.close()}
		on:keydown={(e) => {
			if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') {
				e.stopPropagation();
				state.close();
			}
		}}
		transition:fade={{ duration: 100 }}
	>
		<div
			class="command-palette"
			role="dialog"
			tabindex="-1"
			on:click|stopPropagation
			on:keydown|stopPropagation={(e) => state.handleKeyDown(e)}
		>
			<div class="search-header">
				<span class="material-icons search-icon">search</span>
				<input
					bind:this={state.inputElement}
					bind:value={state.searchQuery}
					type="text"
					placeholder="Search games, pages, and actions..."
					spellcheck="false"
					autocomplete="off"
				/>
				<button class="kbd-hint" on:click={() => state.close()}>ESC</button>
			</div>

			<div class="results-list" bind:this={state.resultsContainer}>
				{#each state.filteredItems as item, i}
					<div
						class="result-item"
						class:active={i === state.selectedIndex}
						on:click={() => state.executeItem(item)}
						on:mouseenter={() => (state.selectedIndex = i)}
						role="button"
						tabindex="0"
						on:keydown={(e) => e.key === "Enter" && state.executeItem(item)}
					>
						{#if item.type === "game" && state.gameIcons[item.game.path || item.game.config?.LauncherPath]}
							<img src={state.gameIcons[item.game.path || item.game.config?.LauncherPath]} class="item-img-icon" alt="" />
						{:else if item.isCustomIcon}
							<img src={item.icon} class="item-img-icon" alt="" />
						{:else}
							<span class="material-icons item-icon">{item.icon}</span>
						{/if}
						<span class="item-name">{item.name}</span>
						{#if item.type === "page"}
							<span class="type-badge page">Navigation</span>
						{:else}
							<span class="type-badge game">Game</span>
						{/if}
					</div>
				{:else}
					<div class="no-results">
						<span class="material-icons">search_off</span>
						<p>No results found for "{state.searchQuery}"</p>
					</div>
				{/each}
			</div>

			<div class="palette-footer">
				<span class="tip"><span class="kbd">↑↓</span> Navigate</span>
				<span class="tip"><span class="kbd">Enter</span> Select</span>
				<span class="tip"><span class="kbd">Esc</span> Close</span>
			</div>
		</div>
	</div>
{/if}

<style lang="scss">
	.palette-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.7);
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding-top: 15vh;
		z-index: 9999;
		border: none;
		outline: none;
		cursor: default;
	}

	.command-palette {
		width: 90%;
		max-width: 640px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-lg);
		box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6);
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.search-header {
		display: flex;
		align-items: center;
		padding: 16px 20px;
		border-bottom: 2px solid rgba(255, 255, 255, 0.05);
		gap: 12px;

		.search-icon {
			color: var(--text-dim);
			font-size: 24px;
		}

		input {
			flex: 1;
			background: transparent;
			border: none;
			outline: none;
			color: var(--text-main);
			font-size: 1.1rem;
			font-weight: 600;

			&::placeholder {
				color: var(--text-dim);
			}
		}

		.kbd-hint {
			font-size: 0.75rem;
			font-weight: 800;
			color: var(--text-dim);
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			padding: 4px 8px;
			border-radius: var(--radius-sm);
			cursor: pointer;
			transition: background var(--transition-fast), color var(--transition-fast);

			&:hover {
				background: rgba(255, 255, 255, 0.1);
				color: var(--text-main);
			}
		}
	}

	.results-list {
		position: relative;
		max-height: 360px;
		overflow-y: auto;
		padding: 8px;
		display: flex;
		flex-direction: column;
		gap: 4px;

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 255, 255, 0.1);
			border-radius: var(--radius-pill);
		}
	}

	.result-item {
		display: flex;
		align-items: center;
		padding: 12px 16px;
		border-radius: var(--radius-md);
		gap: 14px;
		cursor: pointer;
		outline: none;
		transition: background var(--transition-fast), transform var(--transition-fast);

		&:hover, &.active {
			background: rgba(255, 255, 255, 0.06);
			transform: translateX(4px);
		}

		.item-icon {
			color: var(--text-dim);
			font-size: 20px;
		}

		.item-img-icon {
			width: 20px;
			height: 20px;
			object-fit: contain;
			border-radius: var(--radius-sm);
		}

		&:hover .item-icon, &.active .item-icon {
			color: var(--text-main);
		}

		.item-name {
			flex: 1;
			font-size: 0.95rem;
			font-weight: 700;
			color: var(--text-muted);
			text-align: left;
		}

		&:hover .item-name, &.active .item-name {
			color: var(--text-main);
		}

		.type-badge {
			font-size: 0.7rem;
			font-weight: 800;
			text-transform: uppercase;
			padding: 2px 6px;
			border-radius: var(--radius-sm);
			letter-spacing: 0.5px;

			&.page {
				background: rgba(255, 255, 255, 0.05);
				color: var(--text-muted);
			}

			&.game {
				background: rgba(16, 185, 129, 0.1);
				color: var(--success);
			}
		}
	}

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px 20px;
		color: var(--text-dim);
		gap: 8px;

		span {
			font-size: 32px;
		}

		p {
			margin: 0;
			font-size: 0.9rem;
			font-style: italic;
		}
	}

	.palette-footer {
		display: flex;
		align-items: center;
		padding: 12px 20px;
		background: var(--bg-surface);
		border-top: 2px solid rgba(255, 255, 255, 0.05);
		gap: 16px;

		.tip {
			font-size: 0.75rem;
			font-weight: 600;
			color: var(--text-dim);
			display: flex;
			align-items: center;
			gap: 4px;

			.kbd {
				background: rgba(255, 255, 255, 0.03);
				border: 1px solid rgba(255, 255, 255, 0.08);
				padding: 2px 5px;
				border-radius: var(--radius-sm);
				font-family: monospace;
			}
		}
	}
</style>
