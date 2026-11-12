<script lang="ts">
	import type { Snippet } from "svelte";

	let {
		checked = false,
		disabled = false,
		onChange = () => {},
		children
	}: {
		checked?: boolean;
		disabled?: boolean;
		onChange?: (checked: boolean) => void;
		children: Snippet;
	} = $props();
</script>

<label class="checkbox" class:disabled>
	<input
		type="checkbox"
		{checked}
		{disabled}
		onchange={(event) => onChange(event.currentTarget.checked)}
	/>
	{@render children()}
</label>

<style lang="scss">
	.checkbox {
		display: flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;

		&.disabled {
			cursor: not-allowed;
			opacity: 0.6;
		}

		input {
			width: 18px;
			height: 18px;
			accent-color: var(--accent-primary);
		}
	}
</style>
