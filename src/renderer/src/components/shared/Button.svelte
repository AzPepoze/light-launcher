<script lang="ts">
	import type { Snippet } from "svelte";
	import type { HTMLButtonAttributes } from "svelte/elements";
	import Icon from "./Icon.svelte";

	type Props = Omit<HTMLButtonAttributes, "type" | "onclick"> & {
		variant?: "primary" | "secondary" | "danger" | "text";
		size?: "sm" | "md";
		icon?: string;
		loading?: boolean;
		disabled?: boolean;
		type?: "button" | "submit" | "reset";
		onclick?: (event: MouseEvent) => void;
		children?: Snippet;
	};

	let {
		variant = "secondary",
		size = "md",
		icon,
		loading = false,
		disabled = false,
		type = "button",
		onclick,
		class: className = "",
		children,
		...rest
	}: Props = $props();

	const isDisabled = $derived(disabled || loading);
	const iconSize = $derived(size === "sm" ? 16 : 18);
</script>

<button
	class="button {variant} {size} {className}"
	class:loading
	{type}
	disabled={isDisabled}
	aria-busy={loading}
	{onclick}
	{...rest}
>
	{#if loading}
		<span class="spinner" aria-hidden="true"></span>
	{:else if icon}
		<Icon name={icon} size={iconSize} />
	{/if}
	{#if children}{@render children()}{/if}
</button>

<style lang="scss">
	.button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 12px 26px;
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-pill);
		background: var(--bg-elevated);
		color: var(--text-main);
		font: inherit;
		font-size: 0.95rem;
		font-weight: 800;
		letter-spacing: 0.5px;
		white-space: nowrap;
		cursor: pointer;
		transition:
			background var(--transition-fast),
			border-color var(--transition-fast),
			color var(--transition-fast),
			opacity var(--transition-fast);

		&.sm {
			padding: 8px 18px;
			font-size: 0.8rem;
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}

		&:disabled {
			opacity: 0.55;
			cursor: not-allowed;
		}

		&.loading {
			cursor: wait;
		}

		&.primary {
			background: var(--accent-primary);
			border-color: var(--accent-primary);
			color: var(--bg-base);

			&:hover:not(:disabled) {
				background: var(--accent-hover);
				border-color: var(--accent-hover);
			}
		}

		&.secondary {
			background: var(--bg-surface);
			color: var(--text-muted);

			&:hover:not(:disabled) {
				border-color: var(--accent-secondary);
				color: var(--text-main);
			}
		}

		&.danger {
			background: var(--danger);
			border-color: var(--danger);
			color: white;

			&:hover:not(:disabled) {
				filter: brightness(1.15);
			}
		}

		&.text {
			background: transparent;
			border-color: transparent;
			color: var(--text-muted);

			&:hover:not(:disabled) {
				background: var(--bg-elevated);
				color: var(--text-main);
			}
		}
	}

	.spinner {
		width: 16px;
		height: 16px;
		border: 2px solid currentColor;
		border-right-color: transparent;
		border-radius: 50%;
		opacity: 0.8;
		animation: ll-spin 0.8s linear infinite;
	}

	/* No rotation for people who ask for less motion: a dimmed static ring. */
	@media (prefers-reduced-motion: reduce) {
		.spinner {
			animation: none;
			opacity: 0.5;
		}
	}
</style>
