<script lang="ts">
	import type { Snippet } from "svelte";

	let {
		id,
		label,
		value,
		min,
		max,
		step = 1,
		invalid = false,
		describedBy,
		oninput,
		children
	}: {
		id: string;
		label: string;
		value: string;
		min?: number;
		max?: number;
		step?: number | "any";
		invalid?: boolean;
		describedBy?: string;
		oninput: (value: string) => void;
		children?: Snippet;
	} = $props();

	function handleInput(event: Event) {
		oninput((event.currentTarget as HTMLInputElement).value);
	}
</script>

<div class="number-field">
	<label for={id}>{label}</label>
	<input
		{id}
		type="number"
		class="input"
		{min}
		{max}
		{step}
		{value}
		aria-invalid={invalid}
		aria-describedby={describedBy}
		oninput={handleInput}
	/>
	{@render children?.()}
</div>

<style lang="scss">
	.number-field {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;

		label {
			font-size: 0.875rem;
			font-weight: 600;
			color: var(--text-main);
		}

		.input {
			width: 100%;

			&[aria-invalid="true"] {
				border-color: var(--danger);
			}
		}
	}
</style>
