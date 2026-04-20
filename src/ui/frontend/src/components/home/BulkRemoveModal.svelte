<script lang="ts">
	import Modal from "../shared/Modal.svelte";

	export let show: boolean;
	export let selectedCount: number;
	export let onClose: () => void;
	export let onConfirm: () => void;
</script>

<Modal {show} title="Remove Games" {onClose}>
	<div class="confirm-modal-content">
		<div class="warning-icon">
			<span
				class="material-icons"
				style="font-size: 48px; color: #ef4444;">warning</span
			>
		</div>
		<p>
			Are you sure you want to remove <strong>{selectedCount}</strong> games
			from the library?
		</p>
		<p class="sub-text">
			This will only remove them from the Quick Launch list, not from
			your disk.
		</p>
	</div>

	<div slot="footer" class="modal-footer-actions">
		<button class="cancel-btn" on:click={onClose}> Cancel </button>
		<button class="confirm-remove-btn" on:click={onConfirm}>
			Remove Games
		</button>
	</div>
</Modal>

<style lang="scss">
	.confirm-modal-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 16px;
		padding: 10px 0;

		p {
			margin: 0;
			font-size: 1.1rem;
			color: rgba(255, 255, 255, 0.9);

			strong {
				color: #ef4444;
			}
		}

		.sub-text {
			font-size: 0.9rem;
			color: rgba(255, 255, 255, 0.4);
		}

		.warning-icon {
			background: rgba(239, 68, 68, 0.1);
			padding: 20px;
			border-radius: 50%;
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}

	.modal-footer-actions {
		display: flex;
		gap: 12px;
		width: 100%;

		button {
			flex: 1;
			padding: 12px;
			border-radius: 12px;
			font-weight: 800;
			cursor: pointer;
			transition: all 0.2s;
		}

		.cancel-btn {
			background: var(--glass-surface);
			color: var(--text-main);
			border: 1px solid var(--glass-border);

			&:hover {
				background: var(--glass-border);
			}
		}

		.confirm-remove-btn {
			background: #ef4444;
			color: #fff;
			border: none;
			box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);

			&:hover {
				filter: brightness(1.2);
				transform: translateY(-2px);
				box-shadow: 0 6px 16px rgba(239, 68, 68, 0.4);
			}
		}
	}
</style>
