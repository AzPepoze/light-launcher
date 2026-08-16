<script lang="ts">
	import { fade, scale } from "svelte/transition";
	import { backOut } from "svelte/easing";

	export let show = false;
	export let title = "Settings";
	export let fullscreen = false;
	export let onClose: () => void = () => {};
	export let showDone = true;
	export let contentClass = "";

	function close() {
		onClose();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Escape") {
			close();
		}
	}
</script>

<svelte:window on:keydown={handleKeydown} />

{#if show}
	<div
		class="modal-backdrop"
		on:click={close}
		on:keydown={(e) => e.key === "Escape" && close()}
		transition:fade={{ duration: 250 }}
		role="presentation"
	>
		<div
			class="modal-content {contentClass}"
			class:fullscreen
			on:click|stopPropagation
			on:keydown|stopPropagation={handleKeydown}
			transition:scale={{ duration: 400, start: 0.85, easing: backOut }}
			role="dialog"
			tabindex="0"
			aria-modal="true"
		>
			<div class="modal-header">
				<h3>{title}</h3>
				<button
					class="close-btn"
					on:click={close}
					aria-label="Close modal"
				>
					<span class="material-icons">close</span>
				</button>
			</div>
			<div class="modal-body">
				<slot></slot>
			</div>
			{#if $$slots.footer || showDone}
				<div class="modal-footer">
					<slot name="footer">
						<button class="btn primary" on:click={close}
							>Done</button
						>
					</slot>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style lang="scss">
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(8, 8, 14, 0.7);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal-content {
		width: 90%;
		max-width: 600px;
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
		display: flex;
		flex-direction: column;
		max-height: 90vh; /* Default max height */

		&.fullscreen {
			width: 100%;
			height: 100%;
			max-width: none;
			max-height: none;
			border-radius: 0;
			border: none;
			background: var(--bg-base);
		}
	}

	.modal-header {
		padding: 24px 32px;
		border-bottom: 2px solid rgba(255, 255, 255, 0.05);
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-shrink: 0;

		h3 {
			margin: 0;
			font-size: 1.5rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 1px;
		}

		.close-btn {
			background: var(--bg-base);
			border: 2px solid rgba(255, 255, 255, 0.05);
			color: var(--text-muted);
			width: 36px;
			height: 36px;
			border-radius: var(--radius-pill);
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);

			&:hover {
				color: var(--text-main);
				background: var(--bg-elevated);
				border-color: var(--accent-secondary);
				transform: scale(1.1);
			}

			&:active {
				transform: scale(0.9);
			}
		}
	}

	.modal-body {
		padding: 32px;
		overflow-y: auto;
		flex: 1; /* Take remaining space */
		display: flex;
		flex-direction: column;

		&::-webkit-scrollbar {
			width: 8px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 102, 171, 0.2);
			border-radius: 10px;
		}
	}

	.modal-footer {
		padding: 24px 32px;
		border-top: 2px solid rgba(255, 255, 255, 0.05);
		display: flex;
		justify-content: flex-end;
		flex-shrink: 0;
	}
</style>
