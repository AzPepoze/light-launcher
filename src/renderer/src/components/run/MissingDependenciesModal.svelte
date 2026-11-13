<script lang="ts">
	import Modal from "@components/shared/Modal.svelte";
	import Button from "@components/shared/Button.svelte";
	import Icon from "@components/shared/Icon.svelte";

	export let show: boolean;
	export let missingTools: string[] = [];
	export let onClose: () => void;
	export let onConfirm: () => void;
</script>

<Modal {show} title="Missing Dependencies" {onClose}>
	<div class="warning-modal-content">
		<div class="warning-icon">
			<Icon name="warning" size={48} />
		</div>
		<p>
			The following requested features are not installed on your
			system:
		</p>
		<div class="missing-list">
			{#each missingTools as tool}
				<span class="tool-tag">{tool}</span>
			{/each}
		</div>
		<p class="question">
			Do you want to launch the game without these features?
		</p>
		<div class="modal-actions">
			<Button onclick={onClose}>Cancel</Button>
			<Button variant="primary" onclick={onConfirm}>Launch Anyway</Button>
		</div>
	</div>
</Modal>

<style lang="scss">
	.warning-modal-content {
		text-align: center;
		.warning-icon {
			margin-bottom: 20px;
			display: flex;
			justify-content: center;
			color: #ef4444;
		}
		p {
			color: var(--text-main);
			line-height: 1.6;
			font-size: 1rem;
		}
		.missing-list {
			margin: 20px 0;
			display: flex;
			flex-wrap: wrap;
			justify-content: center;
			gap: 12px;
			.tool-tag {
				background: var(--danger);
				color: #ffffff;
				padding: 8px 18px;
				border-radius: var(--radius-pill);
				font-size: 0.9rem;
				font-weight: 800;
				box-shadow: 0 4px 10px rgba(255, 74, 122, 0.2);
			}
		}
		.question {
			margin-top: 28px;
			font-weight: 800;
			color: var(--accent-secondary);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}
	}
	.modal-actions {
		display: flex;
		gap: 16px;
		margin-top: 36px;

		:global(.button) {
			flex: 1;
		}
	}
</style>
