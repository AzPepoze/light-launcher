<script lang="ts">
	import type { Snippet } from "svelte";
	import type { HTMLInputAttributes } from "svelte/elements";

	let {
		checked = false,
		disabled = false,
		onChange = () => {},
		children,
		...rest
	}: {
		checked?: boolean;
		disabled?: boolean;
		onChange?: (checked: boolean) => void;
		children: Snippet;
	} & Omit<HTMLInputAttributes, "type" | "checked" | "disabled"> = $props();
</script>

<label class="switch">
	<input
		{...rest}
		type="checkbox"
		role="switch"
		{checked}
		{disabled}
		onchange={(event) => onChange(event.currentTarget.checked)}
	/>
	<span class="track" aria-hidden="true"></span>
	{@render children()}
</label>

<style lang="scss">
	.switch {
		display: flex;
		align-items: center;
		gap: 10px;
		cursor: pointer;
		font-size: 0.9375rem;
		font-weight: 600;

		input {
			position: absolute;
			opacity: 0;
			pointer-events: none;
		}

		.track {
			position: relative;
			width: 38px;
			height: 22px;
			border-radius: 11px;
			background: var(--bg-input);
			border: 2px solid var(--glass-border);
			transition: background var(--motion-quick) var(--ease-out);
			flex-shrink: 0;

			&::after {
				content: "";
				position: absolute;
				top: 2px;
				left: 2px;
				width: 14px;
				height: 14px;
				border-radius: 50%;
				background: var(--text-muted);
				transition: transform var(--motion-quick) var(--ease-out);
			}
		}

		input:checked + .track {
			background: var(--accent-primary);
			border-color: var(--accent-primary);

			&::after {
				transform: translateX(16px);
				background: var(--bg-elevated);
			}
		}

		input:focus-visible + .track {
			box-shadow: 0 0 0 4px var(--accent-glow);
		}
	}
</style>
