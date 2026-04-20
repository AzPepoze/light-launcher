<script lang="ts">
	import Modal from "@components/shared/Modal.svelte";

	export let show: boolean;
	export let missingTools: string[] = [];
	export let onClose: () => void;
	export let onConfirm: () => void;
</script>

<Modal {show} title="Missing Dependencies" {onClose}>
	<div class="warning-modal-content">
		<div class="warning-icon">
			<span
				class="material-icons"
				style="font-size: 48px; color: #ef4444;">warning</span
			>
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
			<button class="btn secondary" on:click={onClose}>Cancel</button>
			<button class="btn primary" on:click={onConfirm}
				>Launch Anyway</button
			>
		</div>
	</div>
</Modal>

<style lang="scss">
	.warning-modal-content {
		text-align: center;
		.warning-icon {
			font-size: 3rem;
			margin-bottom: 16px;
			display: flex;
			justify-content: center;
		}
		p {
			color: var(--text-main);
			line-height: 1.5;
		}
		.missing-list {
			margin: 16px 0;
			display: flex;
			flex-wrap: wrap;
			justify-content: center;
			gap: 12px;
			.tool-tag {
				background: rgba(239, 68, 68, 0.1);
				color: #ef4444;
				padding: 6px 16px;
				border-radius: 20px;
				font-size: 0.9rem;
				border: 1px solid rgba(239, 68, 68, 0.2);
				font-weight: bold;
			}
		}
		.question {
			margin-top: 24px;
			font-weight: 600;
			color: var(--accent-secondary);
		}
	}
	.modal-actions {
		display: flex;
		gap: 12px;
		margin-top: 32px;
		button {
			flex: 1;
			padding: 12px;
			font-weight: 600;
			border-radius: 10px;
			cursor: pointer;
			transition: all 0.2s;
		}

		.btn.primary {
			background: var(--accent-primary);
			color: var(--glass-bg);
			border: none;
			&:hover {
				filter: brightness(1.2);
			}
		}

		.btn.secondary {
			background: var(--glass-surface);
			color: var(--text-muted);
			border: 1px solid var(--glass-border);
			&:hover {
				background: var(--glass-border);
				color: var(--text-main);
			}
		}
	}
</style>
