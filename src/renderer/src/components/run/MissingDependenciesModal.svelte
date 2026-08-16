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
			margin-bottom: 20px;
			display: flex;
			justify-content: center;
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
		button {
			flex: 1;
			padding: 14px 28px;
			font-weight: 800;
			border-radius: var(--radius-pill);
			cursor: pointer;
			transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
		}

		.btn.primary {
			background: var(--accent-primary);
			color: #ffffff;
			border: none;
			box-shadow: 0 4px 12px var(--accent-glow);
			&:hover {
				background: var(--accent-hover);
				transform: scale(1.05);
			}
			&:active {
				transform: scale(0.95);
			}
		}

		.btn.secondary {
			background: var(--bg-surface);
			color: var(--text-muted);
			border: 2px solid rgba(255, 255, 255, 0.05);
			&:hover {
				background: var(--bg-elevated);
				color: var(--text-main);
				border-color: var(--accent-secondary);
				transform: scale(1.05);
			}
			&:active {
				transform: scale(0.95);
			}
		}
	}
</style>
