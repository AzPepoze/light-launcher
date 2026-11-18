<script lang="ts">
	import Icon from "./Icon.svelte";

	let {
		id,
		value,
		min,
		max,
		step = 1,
		label,
		onChange
	}: {
		id?: string;
		value: number;
		min: number;
		max: number;
		step?: number;
		/** Names the field and the arrow buttons for screen readers. */
		label: string;
		onChange: (value: number) => void;
	} = $props();

	let input = $state<HTMLInputElement>();

	function clamp(next: number): number {
		return Math.min(max, Math.max(min, next));
	}

	function setValue(next: number) {
		if (next !== value) onChange(next);
	}

	function stepBy(direction: 1 | -1) {
		setValue(clamp(value + direction * step));
	}

	// Saves a typed value as soon as it is in range. Out-of-range values wait for commit.
	function handleInput(event: Event) {
		const text = (event.currentTarget as HTMLInputElement).value;
		const typed = Number(text);
		if (text === "" || !Number.isFinite(typed)) return;
		if (typed >= min && typed <= max) setValue(typed);
	}

	// Clamps the typed value and puts it back in the field. Runs on Enter, blur and change.
	function commit() {
		if (!input) return;
		const typed = Number(input.value);
		const next = input.value === "" || !Number.isFinite(typed) ? value : clamp(typed);
		setValue(next);
		input.value = String(next);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Enter") commit();
	}
</script>

<div class="number-stepper">
	<input
		{id}
		bind:this={input}
		type="number"
		class="field"
		aria-label={label}
		{min}
		{max}
		{step}
		value={String(value)}
		oninput={handleInput}
		onchange={commit}
		onblur={commit}
		onkeydown={handleKeydown}
	/>
	<div class="arrows">
		<button
			type="button"
			class="arrow"
			aria-label="Increase {label}"
			title="Increase {label}"
			disabled={value >= max}
			onclick={() => stepBy(1)}
		>
			<Icon name="keyboard_arrow_up" size={22} />
		</button>
		<button
			type="button"
			class="arrow"
			aria-label="Decrease {label}"
			title="Decrease {label}"
			disabled={value <= min}
			onclick={() => stepBy(-1)}
		>
			<Icon name="keyboard_arrow_down" size={22} />
		</button>
	</div>
</div>

<style lang="scss">
	.number-stepper {
		display: flex;
		align-items: stretch;
		width: 100%;
		min-width: 0;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		transition:
			border-color var(--transition-fast),
			transform var(--transition-fast),
			box-shadow var(--transition-fast);

		&:focus-within {
			border-color: var(--accent-primary);
			box-shadow: 0 0 0 4px var(--accent-glow);
			transform: scale(1.01);
		}

		.field {
			flex: 1;
			min-width: 0;
			padding: 10px 16px;
			border: none;
			background: transparent;
			color: var(--text-main);
			font: inherit;
			font-weight: 600;
			outline: none;
			-moz-appearance: textfield;
			appearance: textfield;

			&::-webkit-inner-spin-button,
			&::-webkit-outer-spin-button {
				-webkit-appearance: none;
				margin: 0;
			}
		}

		.arrows {
			display: flex;
			flex-direction: column;
			width: 32px;
			border-left: 1px solid var(--glass-border);
		}

		.arrow {
			display: flex;
			flex: 1;
			align-items: center;
			justify-content: center;
			padding: 0;
			border: none;
			border-radius: 0;
			background: transparent;
			color: var(--text-muted);
			cursor: pointer;
			transition:
				background var(--transition-fast),
				color var(--transition-fast);

			& + .arrow {
				border-top: 1px solid var(--glass-border);
			}

			&:first-child {
				border-top-right-radius: calc(var(--radius-md) - 2px);
			}

			&:last-child {
				border-bottom-right-radius: calc(var(--radius-md) - 2px);
			}

			&:hover:not(:disabled) {
				background: var(--bg-elevated);
				color: var(--text-main);
			}

			&:focus-visible {
				outline: 2px solid var(--accent-primary);
				outline-offset: -2px;
			}

			&:disabled {
				opacity: 0.4;
				cursor: not-allowed;
			}
		}
	}
</style>
