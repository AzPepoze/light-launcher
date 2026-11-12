<script lang="ts">
	let {
		value,
		min,
		max,
		step,
		format = (current: number) => String(current),
		onChange
	}: {
		value: number;
		min: number;
		max: number;
		step: number;
		format?: (value: number) => string;
		onChange: (value: number) => void;
	} = $props();

	function handleInput(event: Event) {
		onChange(Number((event.currentTarget as HTMLInputElement).value));
	}
</script>

<div class="slider-row">
	<input
		type="range"
		{min}
		{max}
		{step}
		{value}
		oninput={handleInput}
		class="slider"
	/>
	<span class="readout">{format(value)}</span>
</div>

<style lang="scss">
	.slider-row {
		display: flex;
		align-items: center;
		gap: 16px;

		.slider {
			flex: 1;
			accent-color: var(--accent-primary);
			height: 6px;
			border-radius: var(--radius-pill);
			background: var(--bg-input);
			outline: none;
		}

		.readout {
			font-size: 1rem;
			font-weight: 800;
			color: var(--accent-primary);
			width: 45px;
			text-align: right;
		}
	}
</style>
