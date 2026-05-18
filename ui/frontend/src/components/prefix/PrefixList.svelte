<script lang="ts">
	export let availablePrefixes: string[] = [];
	export let currentPrefixName: string = "";
	export let newPrefixName: string = "";
	export let onSelectPrefix: (name: string) => void;
	export let onCreatePrefix: () => void;
	export let onRemovePrefix: (name: string) => void;
</script>

<div class="sidebar-section glass">
	<div class="section-header">
		<h2>Available Prefixes</h2>
	</div>
	<div class="prefix-list">
		{#each availablePrefixes as name}
			<div class="prefix-item-wrapper" class:active={currentPrefixName === name}>
				<button
					class="prefix-item-btn"
					on:click={() => onSelectPrefix(name)}
				>
					<span class="folder-icon">
						<span class="material-icons">folder</span>
					</span>
					<span class="name">{name}</span>
				</button>
				{#if name !== "Default"}
					<button 
						class="remove-btn" 
						title="Delete Prefix" 
						on:click|stopPropagation={() => onRemovePrefix(name)}
					>
						<span class="material-icons">delete</span>
					</button>
				{/if}
			</div>
		{/each}
		{#if availablePrefixes.length === 0}
			<div class="empty-state">
				No prefixes found in default directory.
			</div>
		{/if}
	</div>
	<div class="add-prefix-area">
		<input
			type="text"
			placeholder="New prefix..."
			bind:value={newPrefixName}
			class="input sm"
			on:keydown={(e) =>
				e.key === "Enter" && onCreatePrefix()}
		/>
		<button class="btn primary sm" on:click={onCreatePrefix}
			>Create</button
		>
	</div>
</div>

<style lang="scss">
	.sidebar-section {
		display: flex;
		flex-direction: column;
		border-radius: var(--radius-lg);
		overflow: hidden;
		border: 2px solid var(--glass-border);
		background: var(--bg-surface);
		height: 100%;
		box-sizing: border-box;

		.section-header {
			padding: 20px;
			border-bottom: 2px solid var(--glass-border);
			h2 {
				font-size: 0.95rem;
				font-weight: 800;
				margin: 0;
				color: var(--text-main);
				text-transform: uppercase;
				letter-spacing: 1px;
			}
		}
	}

	.prefix-list {
		flex: 1;
		overflow-y: auto;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 6px;

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: var(--glass-border-bright);
			border-radius: 10px;
		}
	}

	.prefix-item-wrapper {
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-radius: var(--radius-md);
		transition: transform var(--transition-spring), background var(--transition-fast), color var(--transition-fast);
		padding-right: 8px;

		&:hover {
			background: var(--bg-elevated);
			transform: scale(1.02);
			
			.remove-btn {
				opacity: 1;
			}
		}

		&.active {
			background: var(--accent-primary);
			color: var(--bg-base);
			font-weight: 800;

			.folder-icon .material-icons {
				color: var(--bg-base);
			}
			
			.remove-btn {
				color: var(--bg-base);
				&:hover {
					background: rgba(0, 0, 0, 0.15);
				}
			}
		}
	}

	.prefix-item-btn {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		background: transparent;
		border: none;
		color: inherit;
		cursor: pointer;
		text-align: left;
		flex: 1;
		font-weight: inherit;
		overflow: hidden;

		.folder-icon {
			font-size: 1.1rem;
			display: flex;
			align-items: center;
			justify-content: center;
			width: 18px;
			height: 18px;

			.material-icons {
				font-size: 1.2rem;
				color: var(--accent-primary);
			}
		}

		.name {
			flex: 1;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
			text-transform: uppercase;
			font-size: 0.85rem;
			letter-spacing: 0.5px;
		}
	}

	.remove-btn {
		background: transparent;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: var(--radius-sm);
		opacity: 0;
		transition: opacity var(--transition-fast), background var(--transition-fast), color var(--transition-fast);

		&:hover {
			background: rgba(255, 74, 122, 0.15);
			color: var(--danger);
		}

		.material-icons {
			font-size: 1.1rem;
		}
	}

	.add-prefix-area {
		padding: 20px;
		border-top: 2px solid var(--glass-border);
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: var(--bg-surface);
	}
	
	.empty-state {
		padding: 32px;
		text-align: center;
		color: var(--text-muted);
		font-size: 0.85rem;
		font-weight: 700;
	}
	
	.input.sm {
		padding: 10px 14px;
		font-size: 0.85rem;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		color: var(--text-main);
		font-weight: 600;
		outline: none;
		transition: border-color var(--transition-fast);

		&:focus {
			border-color: var(--accent-primary);
		}
	}
	
	.btn.sm {
		padding: 10px;
		font-size: 0.85rem;
		border-radius: var(--radius-pill);
		font-weight: 800;
		background: var(--accent-primary);
		color: var(--bg-base);
		border: none;
		cursor: pointer;
		box-shadow: 0 4px 10px var(--accent-glow);
		transition: transform var(--transition-spring), background var(--transition-fast);

		&:hover {
			background: var(--accent-hover);
			transform: scale(1.03);
		}

		&:active {
			transform: scale(0.97);
		}
	}
</style>
