<script lang="ts">
	import NumberStepper from "./NumberStepper.svelte";

	export let label: string = "Value";
	export let value: string = "2";
	export let min: number = 2;
	export let max: number = 100;
	export let onChange: ((val: string) => void) | null = null;

	const NO_LIMIT = "-1";

	$: isNoLimit = value === NO_LIMIT;
	// With no limit the stepper must not clamp, so -1 is kept. The clamp below still applies to other values.
	$: stepperMin = isNoLimit ? undefined : min;
	$: stepperMax = isNoLimit ? undefined : max;
	$: stepperValue = parseInt(value) || min;

	function handleStep(next: number) {
		let numVal = next;

		// Clamp to range, but allow -1
		if (numVal !== -1) {
			numVal = Math.max(min, Math.min(max, numVal));
		}

		value = numVal.toString();
		onChange?.(value);
	}
</script>

<div class="numeric-input-container">
	<label for="numeric-input">{label}</label>
	<div class="input-wrapper">
		<div class="stepper-box">
			<NumberStepper
				id="numeric-input"
				{label}
				value={stepperValue}
				min={stepperMin}
				max={stepperMax}
				onChange={handleStep}
			/>
		</div>
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

		.stepper-box {
			width: 140px;
		}

		.range-hint {
			font-size: 0.75rem;
			font-weight: 700;
			text-transform: uppercase;
			color: var(--text-muted);
		}
	}
</style>
