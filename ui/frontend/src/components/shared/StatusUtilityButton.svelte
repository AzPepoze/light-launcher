<script lang="ts">
	import { fade, fly } from "svelte/transition";

	export let icon: string;
	export let title: string;
	export let subtitle: string;
	export let isPulsing = false;
	export let showSuccess = false;
	export let btnClass = "";
	export let onclick: () => void = () => {};
</script>

<button
	class="util-btn {btnClass}"
	class:pulsing={isPulsing}
	on:click={onclick}
>
	<div class="content-left">
		<span class="material-icons util-icon">{icon}</span>
		<div class="btn-stack">
			<span>{title}</span>
			<small>{subtitle}</small>
		</div>
	</div>
	{#if showSuccess}
		<div
			class="check-indicator"
			in:fly={{ x: 20, duration: 400 }}
			out:fade={{ duration: 200 }}
		>
			<span class="material-icons" style="font-size: 14px;">check</span
			>
		</div>
	{/if}
</button>

<style lang="scss">
	.util-btn {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 14px 20px;
		min-height: 72px;
		background: var(--bg-base);
		border: 2px solid rgba(255, 255, 255, 0.03);
		border-radius: var(--radius-md);
		color: var(--text-main);
		cursor: pointer;
		transition: transform var(--transition-spring), border-color var(--transition-fast), background var(--transition-fast), box-shadow var(--transition-fast);
		position: relative;
		overflow: hidden;

		.content-left {
			display: flex;
			align-items: center;
			gap: 14px;
			flex: 1;
			min-width: 0;
		}

		.util-icon {
			font-size: 24px;
			z-index: 2;
			transition: transform var(--transition-spring);
		}

		.btn-stack {
			display: flex;
			flex-direction: column;
			align-items: flex-start;
			z-index: 2;
		}

		span {
			font-size: 0.9rem;
			font-weight: 800;
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		small {
			font-size: 0.75rem;
			font-weight: 600;
			color: var(--text-muted);
			display: block;
			margin-top: 2px;
		}

		&:hover {
			transform: scale(1.03);
			.util-icon {
				transform: scale(1.1);
			}
		}

		&.cleanup {
			background: rgba(255, 102, 171, 0.05);
			border-color: rgba(255, 102, 171, 0.2);

			.util-icon, span {
				color: var(--accent-primary);
			}
			small {
				color: rgba(255, 102, 171, 0.8);
			}

			&:hover {
				background: rgba(255, 102, 171, 0.1);
				border-color: var(--accent-primary);
				box-shadow: 0 4px 15px rgba(255, 102, 171, 0.2);
			}
			&.pulsing::after {
				background: linear-gradient(
					90deg,
					transparent,
					rgba(255, 102, 171, 0.4),
					transparent
				);
			}
		}

		&.cache {
			background: rgba(0, 240, 255, 0.05);
			border-color: rgba(0, 240, 255, 0.2);

			.util-icon, span {
				color: var(--accent-secondary);
			}
			small {
				color: rgba(0, 240, 255, 0.8);
			}

			&:hover {
				background: rgba(0, 240, 255, 0.1);
				border-color: var(--accent-secondary);
				box-shadow: 0 4px 15px rgba(0, 240, 255, 0.2);
			}
			&.pulsing::after {
				background: linear-gradient(
					90deg,
					transparent,
					rgba(0, 240, 255, 0.4),
					transparent
				);
			}
		}

		&.drop-caches {
			background: rgba(177, 151, 252, 0.05);
			border-color: rgba(177, 151, 252, 0.2);

			.util-icon, span {
				color: #b197fc;
			}
			small {
				color: rgba(177, 151, 252, 0.8);
			}

			&:hover {
				background: rgba(177, 151, 252, 0.1);
				border-color: #b197fc;
				box-shadow: 0 4px 15px rgba(177, 151, 252, 0.2);
			}
			&.pulsing::after {
				background: linear-gradient(
					90deg,
					transparent,
					rgba(177, 151, 252, 0.4),
					transparent
				);
			}
		}

		&.clear-swap {
			background: rgba(16, 185, 129, 0.05);
			border-color: rgba(16, 185, 129, 0.2);

			.util-icon, span {
				color: #10b981;
			}
			small {
				color: rgba(16, 185, 129, 0.8);
			}

			&:hover {
				background: rgba(16, 185, 129, 0.1);
				border-color: #10b981;
				box-shadow: 0 4px 15px rgba(16, 185, 129, 0.2);
			}
			&.pulsing::after {
				background: linear-gradient(
					90deg,
					transparent,
					rgba(16, 185, 129, 0.4),
					transparent
				);
			}
		}

		&.pulsing {
			pointer-events: none;

			&::after {
				content: "";
				position: absolute;
				top: 0;
				left: -150%;
				width: 150%;
				height: 100%;
				z-index: 1;
				animation: pulse-scan 1.5s cubic-bezier(0.4, 0, 0.2, 1);
			}
		}
	}

	.check-indicator {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		background: var(--accent-primary);
		border-radius: 50%;
		color: #ffffff;
		padding: 3px;
		z-index: 2;

		.material-icons {
			width: 100%;
			height: 100%;
		}
	}

	@keyframes pulse-scan {
		0% {
			transform: translateX(0);
		}
		100% {
			transform: translateX(250%);
		}
	}
</style>
