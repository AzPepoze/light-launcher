<script lang="ts">
	import { fade, scale } from "svelte/transition";

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
		transition:fade={{ duration: 200 }}
		role="presentation"
	>
		<div
			class="modal-content glass {contentClass}"
			class:fullscreen
			on:click|stopPropagation
			on:keydown|stopPropagation={handleKeydown}
			transition:scale={{ duration: 200, start: 0.95 }}
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
		background: rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.modal-content {
		width: 90%;
		max-width: 600px;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border);
		border-radius: 16px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
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
			background: var(--glass-bg); /* Match app bg */
		}
	}

	.modal-header {
		padding: 20px 32px;
		border-bottom: 1px solid var(--glass-border);
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-shrink: 0;

		h3 {
			margin: 0;
			font-size: 1.5rem;
			color: var(--text-main);
		}

		.close-btn {
			background: var(--glass-surface);
			border: 1px solid var(--glass-border);
			color: var(--text-dim);
			width: 32px;
			height: 32px;
			border-radius: 10px;
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			transition: all 0.2s;

			&:hover {
				color: var(--text-main);
				background: var(--glass-border-bright);
				transform: scale(1.1);
			}
		}
	}

	.modal-body {
		padding: 32px;
		overflow-y: auto;
		flex: 1; /* Take remaining space */
		display: flex;
		flex-direction: column;
	}

	.modal-footer {
		padding: 20px 32px;
		border-top: 1px solid var(--glass-border);
		display: flex;
		justify-content: flex-end;
		flex-shrink: 0;
	}
</style>
