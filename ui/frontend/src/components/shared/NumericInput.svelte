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
		gap: 10px;

		label {
			font-size: 0.85rem;
			font-weight: 800;
			color: var(--accent-primary);
			text-transform: uppercase;
			letter-spacing: 1px;
		}
	}

	.input-wrapper {
		display: flex;
		align-items: center;
		gap: 12px;

		.input {
			max-width: 140px;
			padding: 10px 16px;
			background: var(--bg-input);
			border: 2px solid transparent;
			border-radius: var(--radius-md);
			color: var(--text-main);
			font-size: 0.95rem;
			font-weight: 700;
			outline: none;
			transition: border-color var(--transition-fast), transform var(--transition-fast);

			&:focus {
				border-color: var(--accent-primary);
				transform: scale(1.005);
			}

			/* Remove number input spinner */
			&::-webkit-inner-spin-button,
			&::-webkit-outer-spin-button {
				-webkit-appearance: none;
				margin: 0;
			}
		}

		.range-hint {
			font-size: 0.75rem;
			font-weight: 700;
			text-transform: uppercase;
			color: var(--text-muted);
		}
	}
</style>
