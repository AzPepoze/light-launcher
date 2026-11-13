<script lang="ts">
	import type { HTMLButtonAttributes } from "svelte/elements";
	import Icon from "./Icon.svelte";

	type Props = Omit<HTMLButtonAttributes, "type" | "onclick"> & {
		icon: string;
		/** Used as the accessible name and the tooltip, unless title or aria-label is passed. */
		label: string;
		size?: 32 | 36;
		tone?: "default" | "danger" | "accent";
		filled?: boolean;
		disabled?: boolean;
		onclick?: (event: MouseEvent) => void;
		class?: string;
	};

	let {
		icon,
		label,
		size = 36,
		tone = "default",
		filled = false,
		disabled = false,
		onclick,
		class: className = "",
		...rest
	}: Props = $props();
</script>

<button
	class="icon-button {tone} {className}"
	style:width="{size}px"
	style:height="{size}px"
	title={label}
	aria-label={label}
	{...rest}
	type="button"
	{disabled}
	{onclick}
>
	<Icon name={icon} size={size === 32 ? 18 : 20} {filled} />
</button>

<style lang="scss">
	.icon-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		padding: 0;
		border: none;
		border-radius: var(--radius-pill);
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		transition:
			background var(--transition-fast),
			color var(--transition-fast);

		&:hover:not(:disabled) {
			background: var(--bg-elevated);
			color: var(--text-main);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}

		&:disabled {
			opacity: 0.4;
			cursor: not-allowed;
		}

		&.accent {
			color: var(--accent-primary);
		}

		&.danger:hover:not(:disabled) {
			color: var(--danger);
		}
	}
</style>
