<script lang="ts">
	export let label: string = "Value";
	export let value: string = "2";
	export let min: number = 2;
	export let max: number = 100;
	export let onChange: ((val: string) => void) | null = null;

	function handleChange(e: Event) {
		const input = e.target as HTMLInputElement;
		let numVal = parseInt(input.value);

		// If empty, set to min
		if (isNaN(numVal)) {
			numVal = min;
		}

		// Clamp to range, but allow -1
		if (numVal !== -1) {
			numVal = Math.max(min, Math.min(max, numVal));
		}

		const newVal = numVal.toString();

		value = newVal;
		onChange?.(newVal);
	}

	function handleBlur(e: Event) {
		const input = e.target as HTMLInputElement;
		if (input.value === "") {
			value = min.toString();
			onChange?.(value);
		}
	}
</script>

<div class="numeric-input-container">
	<label for="numeric-input">{label}</label>
	<div class="input-wrapper">
		<input
			id="numeric-input"
			type="number"
			class="input"
			value={parseInt(value) || min}
			on:change={handleChange}
			on:blur={handleBlur}
		/>
		<span class="range-hint">
			{#if value === "-1"}
				No Limit
			{:else}
				{min}-{max}
			{/if}
		</span>
	</div>
</div>

<style lang="scss">
	.numeric-input-container {
		display: flex;
		flex-direction: column;
		gap: 8px;

		label {
			font-size: 0.75rem;
			font-weight: 600;
			color: var(--text-dim);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}
	}

	.input-wrapper {
		display: flex;
		align-items: center;
		gap: 8px;

		.input {
			max-width: 140px;
			padding: 8px 12px;
			background: var(--glass-surface);
			border: 1px solid var(--glass-border);
			border-radius: 6px;
			color: var(--text-main);
			font-size: 0.95rem;
			font-weight: 600;

			&:focus {
				outline: none;
				border-color: var(--accent-primary);
				box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
			}

			/* Remove number input spinner */
			&::-webkit-inner-spin-button,
			&::-webkit-outer-spin-button {
				-webkit-appearance: none;
				margin: 0;
			}
		}

		.range-hint {
			font-size: 0.7rem;
			color: var(--text-dim);
			opacity: 0.7;
		}
	}
</style>
