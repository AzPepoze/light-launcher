<script lang="ts">
	import { fade } from "svelte/transition";

	export let selectedCount: number;
	export let onBulkRemove: () => void;
	export let onToggleSelectionMode: () => void;
</script>

<div class="selection-actions" in:fade>
	<span class="selection-count">{selectedCount} selected</span>
	<button
		class="bulk-remove-btn"
		on:click={onBulkRemove}
		disabled={selectedCount === 0}
	>
		<span class="material-icons" style="font-size: 18px;"
			>delete</span
		>
		Remove Selected
	</button>
	<button
		class="cancel-selection-btn"
		on:click={onToggleSelectionMode}
	>
		Cancel
	</button>
</div>

<style lang="scss">
	.selection-actions {
		display: flex;
		align-items: center;
		gap: 14px;
		background: var(--bg-surface);
		padding: 6px 18px;
		border-radius: var(--radius-pill);
		border: 2px solid var(--accent-primary);

		.selection-count {
			font-size: 0.9rem;
			font-weight: 800;
			color: var(--accent-primary);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		.bulk-remove-btn {
			background: var(--danger);
			color: #fff;
			border: none;
			padding: 8px 18px;
			border-radius: var(--radius-pill);
			font-size: 0.85rem;
			font-weight: 800;
			cursor: pointer;
			display: flex;
			align-items: center;
			gap: 6px;
			transition: transform var(--transition-spring), filter var(--transition-fast);

			&:hover:not(:disabled) {
				filter: brightness(1.15);
				transform: scale(1.05);
			}

			&:active {
				transform: scale(0.95);
			}

			&:disabled {
				opacity: 0.5;
				cursor: not-allowed;
			}
		}

		.cancel-selection-btn {
			background: none;
			border: none;
			color: var(--text-muted);
			font-size: 0.85rem;
			font-weight: 800;
			cursor: pointer;
			text-transform: uppercase;
			letter-spacing: 0.5px;
			transition: color var(--transition-fast);

			&:hover {
				color: var(--text-main);
			}
		}
	}
</style>
