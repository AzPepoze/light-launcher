<script lang="ts">
	export let name: string;
	export let currentPrefixName: string = "";
	export let onSelectPrefix: (name: string) => void;
	export let onRemovePrefix: (name: string) => void;
	export let subtitle: string = "";
	export let runningGames: string[] = [];

	$: isActive = currentPrefixName === name;
	$: hasRunning = runningGames.length > 0;
</script>

<div
	class="prefix-item"
	class:active={isActive}
	class:has-running={hasRunning}
	role="option"
	aria-selected={isActive}
>
	<button
		type="button"
		class="prefix-item-btn"
		on:click={() => onSelectPrefix(name)}
	>
		<span class="folder-icon">
			<span class="material-icons">
				{isActive ? "folder_open" : "folder"}
			</span>
		</span>
		<span class="info">
			<span class="name">{name}</span>
			{#if subtitle}
				<span class="subtitle">{subtitle}</span>
			{/if}
		</span>
	</button>
	{#if hasRunning}
		<span class="running-indicator" title="{runningGames.length} running">
			<span class="pulse-dot"></span>
			<span class="running-count">{runningGames.length}</span>
		</span>
		<div class="running-tooltip" role="tooltip">
			<div class="tooltip-title">Running — {runningGames.length}</div>
			<ul class="tooltip-list">
				{#each runningGames as g}
					<li>{g}</li>
				{/each}
			</ul>
		</div>
	{/if}
	{#if name !== "Default"}
		<button
			type="button"
			class="remove-btn"
			title="Delete {name}"
			aria-label="Delete {name}"
			on:click|stopPropagation={() => onRemovePrefix(name)}
		>
			<span class="material-icons">delete_outline</span>
		</button>
	{/if}
</div>

<style lang="scss">
	.prefix-item {
		display: flex;
		align-items: center;
		border-radius: var(--radius-md);
		transition: background var(--transition-fast);

		&:hover {
			background: var(--bg-elevated);

			.remove-btn {
				opacity: 1;
			}
		}

		&.active {
			background: var(--accent-primary);
			color: var(--bg-base);

			.folder-icon .material-icons {
				color: var(--bg-base);
			}

			.remove-btn {
				color: var(--bg-base);
				opacity: 0.85;

				&:hover {
					background: rgba(0, 0, 0, 0.12);
					color: var(--bg-base);
				}
			}
		}
	}

	.prefix-item-btn {
		display: flex;
		align-items: center;
		gap: 10px;
		flex: 1;
		min-width: 0;
		padding: 11px 12px;
		background: transparent;
		border: none;
		color: inherit;
		cursor: pointer;
		text-align: left;
		font-weight: inherit;

		.folder-icon .material-icons {
			font-size: 20px;
			color: var(--accent-primary);
		}

		.info {
			flex: 1;
			min-width: 0;
			display: flex;
			flex-direction: column;
			gap: 2px;
		}

		.name {
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
			font-size: 0.85rem;
			font-weight: 700;
			letter-spacing: 0.3px;
		}

		.subtitle {
			font-size: 0.7rem;
			font-weight: 600;
			color: var(--text-muted);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	}

	.prefix-item.active .subtitle {
		color: rgba(0, 0, 0, 0.6);
	}

	.prefix-item {
		position: relative;
	}

	.running-indicator {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 4px 8px;
		border-radius: var(--radius-pill);
		background: rgba(52, 199, 89, 0.12);
		border: 1px solid rgba(52, 199, 89, 0.35);
		flex-shrink: 0;
		margin-right: 4px;
	}

	.pulse-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--success);
		box-shadow: 0 0 8px rgba(52, 199, 89, 0.7);
		animation: pulse 1.6s ease-in-out infinite;
		flex-shrink: 0;
	}

	.running-count {
		font-size: 0.7rem;
		font-weight: 800;
		color: var(--success);
		line-height: 1;
	}

	.prefix-item.active .running-indicator {
		background: rgba(0, 0, 0, 0.12);
		border-color: rgba(0, 0, 0, 0.2);
		.running-count { color: var(--bg-base); }
		.pulse-dot {
			background: var(--bg-base);
			box-shadow: 0 0 6px rgba(0, 0, 0, 0.3);
		}
	}

	.running-tooltip {
		position: absolute;
		right: 8px;
		top: calc(100% + 6px);
		background: var(--bg-elevated);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		padding: 10px 12px;
		min-width: 160px;
		max-width: 240px;
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.45);
		z-index: 20;
		opacity: 0;
		visibility: hidden;
		transform: translateY(-4px);
		transition: opacity var(--transition-fast), transform var(--transition-fast), visibility var(--transition-fast);
		pointer-events: none;
	}

	.prefix-item.has-running:hover .running-tooltip {
		opacity: 1;
		visibility: visible;
		transform: translateY(0);
	}

	.tooltip-title {
		font-size: 0.7rem;
		font-weight: 800;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: var(--text-muted);
		margin-bottom: 6px;
	}

	.tooltip-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.tooltip-list li {
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--text-main);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		padding: 4px 6px;
		background: var(--bg-surface);
		border-radius: var(--radius-sm);
		border: 1px solid var(--glass-border);
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; transform: scale(1); }
		50% { opacity: 0.6; transform: scale(0.85); }
	}

	.remove-btn {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		margin-right: 4px;
		padding: 0;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		opacity: 0;
		transition:
			opacity var(--transition-fast),
			background var(--transition-fast),
			color var(--transition-fast);

		.material-icons {
			font-size: 18px;
		}

		&:hover {
			background: rgba(255, 59, 48, 0.15);
			color: var(--danger);
			opacity: 1;
		}
	}

	.prefix-item.active .remove-btn,
	.prefix-item:hover .remove-btn {
		opacity: 0.7;
	}
</style>
