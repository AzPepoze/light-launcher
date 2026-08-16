<script lang="ts">
	export let name: string;
	export let currentPrefixName: string = "";
	export let onSelectPrefix: (name: string) => void;
	export let onRemovePrefix: (name: string) => void;

	$: isActive = currentPrefixName === name;
</script>

<div
	class="prefix-item"
	class:active={isActive}
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
		<span class="name">{name}</span>
	</button>
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

		.name {
			flex: 1;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
			font-size: 0.85rem;
			font-weight: 700;
			letter-spacing: 0.3px;
		}
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
