<script lang="ts">
	import { notifications } from "@stores/notificationStore";
	import { flip } from "svelte/animate";
	import { fade, fly } from "svelte/transition";
</script>

<div class="notification-container">
	{#each $notifications as n (n.id)}
		<div
			class="notification-card {n.type}"
			animate:flip={{ duration: 300 }}
			in:fly={{ y: 20, duration: 300 }}
			out:fade={{ duration: 200 }}
		>
			<span class="material-icons mini-icon">
				{#if n.type === "error"}
					error_outline
				{:else if n.type === "success"}
					check_circle_outline
				{:else}
					info_outline
				{/if}
			</span>
			<div class="message">{n.message}</div>
			<button class="close" on:click={() => notifications.remove(n.id)}>
				<span class="material-icons">close</span>
			</button>
		</div>
	{/each}
</div>

<style lang="scss">
	.notification-container {
		position: fixed;
		bottom: 32px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 9999;
		display: flex;
		flex-direction: column-reverse;
		gap: 12px;
		width: 380px;
		max-width: 90%;
		pointer-events: none;
		align-items: center;
	}

	.notification-card {
		pointer-events: auto;
		padding: 14px 20px;
		border-radius: var(--radius-pill);
		display: flex;
		align-items: center;
		gap: 14px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
		border: 2px solid rgba(255, 255, 255, 0.05);
		background: var(--bg-surface);
		color: var(--text-main);
		width: 100%;
		box-sizing: border-box;
		transition: transform var(--transition-spring);

		&:hover {
			transform: scale(1.02);
		}

		.mini-icon {
			font-size: 1.35rem;
			flex-shrink: 0;
		}

		&.error {
			border-color: var(--danger);
			box-shadow: 0 4px 20px rgba(255, 74, 122, 0.15);
			.mini-icon { color: var(--danger); }
		}
		&.success {
			border-color: var(--accent-secondary);
			box-shadow: 0 4px 20px rgba(0, 240, 255, 0.15);
			.mini-icon { color: var(--accent-secondary); }
		}
		&.info {
			border-color: var(--accent-primary);
			box-shadow: 0 4px 20px rgba(255, 102, 171, 0.15);
			.mini-icon { color: var(--accent-primary); }
		}

		.message {
			flex: 1;
			font-size: 0.85rem;
			font-weight: 800;
			text-transform: uppercase;
			letter-spacing: 0.5px;
			line-height: 1.4;
		}

		.close {
			background: var(--bg-base);
			border: 2px solid rgba(255, 255, 255, 0.05);
			color: var(--text-muted);
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			width: 28px;
			height: 28px;
			border-radius: var(--radius-pill);
			transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);

			.material-icons {
				font-size: 16px;
			}

			&:hover {
				color: var(--text-main);
				background: var(--bg-elevated);
				border-color: var(--accent-primary);
				transform: scale(1.1) rotate(90deg);
			}

			&:active {
				transform: scale(0.9);
			}
		}
	}
</style>
