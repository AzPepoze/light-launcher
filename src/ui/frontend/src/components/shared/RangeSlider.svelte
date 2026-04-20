<script lang="ts">
	import warningIcon from "@icons/warning.svg";

	export let value: number = 4;
	export let min: number = 0.1;
	export let max: number = 16;
	export let inputMin: number | null = null; // Optional separate min for input
	export let inputMax: number | null = null; // Optional separate max for input
	export let snapValues: number[] = [4, 8, 12, 16, 24, 32, 64];
	export let unit: string = "GB"; // "GB", "x", or any custom unit
	export let label: string = ""; // Custom label, if empty uses "value unit / max unit"
	export let showWarning: boolean = true; // Only show warning for RAM-like sliders
	export let onChange: (value: number) => void = () => {};

	const warningPercentage = 0.75; // 75% of max is warning zone

	let isDragging = false;
	let trackElement: HTMLElement;
	let percentage = 0;

	$: {
		percentage = ((value - min) / (max - min)) * 100;
	}

	function handleStart(e: MouseEvent | TouchEvent) {
		isDragging = true;
		updateValue(e);
		window.addEventListener("mousemove", handleMove);
		window.addEventListener("touchmove", handleMove);
		window.addEventListener("mouseup", handleEnd);
		window.addEventListener("touchend", handleEnd);
	}

	function handleEnd() {
		isDragging = false;
		window.removeEventListener("mousemove", handleMove);
		window.removeEventListener("touchmove", handleMove);
		window.removeEventListener("mouseup", handleEnd);
		window.removeEventListener("touchend", handleEnd);
	}

	function handleMove(e: MouseEvent | TouchEvent) {
		if (!isDragging) return;
		updateValue(e);
	}

	function updateValue(e: MouseEvent | TouchEvent) {
		const rect = trackElement.getBoundingClientRect();
		const clientX = "touches" in e ? e.touches[0].clientX : e.clientX;
		let x = clientX - rect.left;
		x = Math.max(0, Math.min(x, rect.width));

		let rawValue = min + (x / rect.width) * (max - min);

		let closest = rawValue;
		closest = Math.round(rawValue * 10) / 10;

		// Snap to integers and specific values if very close
		for (const snap of snapValues) {
			if (snap > max || snap < min) continue;
			const diff = Math.abs(rawValue - snap);
			if (diff < 0.5) {
				closest = snap;
			}
		}

		if (closest < min) closest = min;

		value = closest;
		onChange(value);
	}

	$: isWarning = showWarning && value > max * warningPercentage;
	$: displayLabel = label || `${value} ${unit} / ${max} ${unit}`;
</script>

<div class="range-container">
	<div class="range-header">
		<div class="range-label">
			<span>{displayLabel}</span>
		</div>
		<div class="input-wrapper">
			<input
				type="number"
				step="0.1"
				min={inputMin ?? min}
				max={inputMax ?? max}
				bind:value
				on:input={() => onChange(value)}
				class="manual-input"
			/>
			<span class="unit">{unit}</span>
		</div>
	</div>

	<div
		class="track-wrapper"
		bind:this={trackElement}
		on:mousedown={handleStart}
		on:touchstart|nonpassive={handleStart}
		role="slider"
		tabindex="0"
		aria-valuemin="0"
		aria-valuemax="100"
		aria-valuenow={value}
	>
		<div class="track">
			<!-- Warning Zone Background -->
			{#if showWarning}
				<div
					class="warning-zone"
					style="width: {(1 - warningPercentage) *
						100}%; left: {warningPercentage * 100}%"
				></div>
			{/if}

			<!-- Fill -->
			<div
				class="fill {isWarning ? 'warning' : ''}"
				style="width: {percentage}%"
			></div>

			<!-- Thumb -->
			<div
				class="thumb {isDragging ? 'active' : ''}"
				style="left: {percentage}%"
			></div>

			<!-- Snap Points -->
			{#each snapValues as point}
				{#if point <= max && point >= min}
					<div
						class="snap-point"
						style="left: {((point - min) / (max - min)) *
							100}%"
					></div>
				{/if}
			{/each}
		</div>
	</div>

	{#if isWarning}
		<div class="warning-text">
			<img src={warningIcon} alt="warning" class="svg-icon" />
			High value may affect system stability
		</div>
	{/if}
</div>

<style lang="scss">
	.range-container {
		width: 100%;
		padding: 8px 0;
		user-select: none;
	}

	.range-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 12px;
	}

	.range-label {
		display: flex;
		flex-direction: column;
		font-size: 1.25rem;
		font-weight: bold;
		color: var(--text-main);
	}

	.input-wrapper {
		display: flex;
		align-items: center;
		background: var(--glass-low-alpha);
		border-radius: 8px;
		padding: 4px 10px;
		border: 1px solid var(--glass-border);

		&:focus-within {
			border-color: var(--accent-primary);
		}

		.manual-input {
			background: transparent;
			border: none;
			color: var(--text-main);
			width: 60px;
			text-align: right;
			font-weight: bold;
			font-size: 1rem;
			outline: none;

			/* Remove arrows */
			&::-webkit-inner-spin-button,
			&::-webkit-outer-spin-button {
				-webkit-appearance: none;
				margin: 0;
			}
		}

		.unit {
			margin-left: 4px;
			font-size: 0.8rem;
			color: var(--text-muted);
		}
	}

	.track-wrapper {
		height: 32px;
		display: flex;
		align-items: center;
		cursor: pointer;
		touch-action: none;
	}

	.track {
		position: relative;
		width: 100%;
		height: 8px;
		background: var(--glass-border);
		border-radius: 4px;
		overflow: visible; // Allow thumb to overflow
	}

	.warning-zone {
		position: absolute;
		height: 100%;
		background: rgba(239, 68, 68, 0.2); // Red tint
		border-radius: 0 4px 4px 0;
		pointer-events: none;
	}

	.fill {
		position: absolute;
		height: 100%;
		left: 0;
		top: 0;
		background: var(--accent-primary);
		border-radius: 4px;
		transition: background 0.2s;

		&.warning {
			background: #ef4444; // Red
		}
	}

	.thumb {
		position: absolute;
		top: 50%;
		width: 20px;
		height: 20px;
		background: white;
		border-radius: 50%;
		transform: translate(-50%, -50%);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
		transition: transform 0.1s;
		pointer-events: none; // Let wrapper handle events

		&.active {
			transform: translate(-50%, -50%) scale(1.2);
		}
	}

	.snap-point {
		position: absolute;
		top: 50%;
		width: 2px;
		height: 4px;
		background: var(--glass-border-bright);
		transform: translate(-50%, -50%);
		pointer-events: none;
	}

	.warning-text {
		margin-top: 8px;
		font-size: 0.75rem;
		color: #ef4444;
		text-align: right;
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 6px;

		.svg-icon {
			width: 14px;
			height: 14px;
			filter: invert(36%) sepia(84%) saturate(4644%) hue-rotate(345deg)
				brightness(98%) contrast(93%); // #ef4444
		}
	}
</style>
