<script lang="ts">
	import type { Snippet } from "svelte";
	import NumberStepper from "./NumberStepper.svelte";

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

	// Arrow buttons report a number, so it is sent on as text like typed input.
	function handleStep(next: number) {
		oninput(String(next));
	}
</script>

<div class="number-field">
	<label for={id}>{label}</label>
	<NumberStepper
		{id}
		{label}
		{value}
		{min}
		{max}
		{step}
		{invalid}
		{describedBy}
		{oninput}
		onChange={handleStep}
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
	}
</style>
