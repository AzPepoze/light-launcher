<script lang="ts">
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";
	import Modal from "@components/shared/Modal.svelte";

	export let show: boolean;
	export let selectedCount: number;
	export let onClose: () => void;
	export let onConfirm: () => void;
</script>

<Modal {show} title="Remove Games" {onClose}>
	<div class="confirm-modal-content">
		<div class="warning-icon">
			<Icon name="warning" size={48} />
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
		<Button variant="secondary" onclick={onClose}>Cancel</Button>
		<Button variant="danger" onclick={onConfirm}>Remove Games</Button>
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
			color: #ef4444;
		}
	}

	.modal-footer-actions {
		display: flex;
		gap: 12px;
		width: 100%;

		:global(.button) {
			flex: 1;
		}
	}
</style>
