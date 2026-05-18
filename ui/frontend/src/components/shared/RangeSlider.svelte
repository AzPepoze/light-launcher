<script lang="ts">

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
			<span class="material-icons warning-icon">warning</span>
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
		font-size: 0.95rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--text-main);
	}

	.input-wrapper {
		display: flex;
		align-items: center;
		background: var(--bg-input);
		border-radius: var(--radius-pill);
		padding: 6px 14px;
		border: 2px solid transparent;
		transition: border-color var(--transition-fast);

		&:focus-within {
			border-color: var(--accent-primary);
		}

		.manual-input {
			background: transparent;
			border: none;
			color: var(--text-main);
			width: 60px;
			text-align: right;
			font-weight: 800;
			font-size: 0.95rem;
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
			font-size: 0.75rem;
			font-weight: 800;
			text-transform: uppercase;
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
		background: var(--bg-input);
		border-radius: var(--radius-pill);
		overflow: visible; // Allow thumb to overflow
	}

	.warning-zone {
		position: absolute;
		height: 100%;
		background: rgba(255, 74, 122, 0.15); // Red/pink tint
		border-radius: 0 var(--radius-pill) var(--radius-pill) 0;
		pointer-events: none;
	}

	.fill {
		position: absolute;
		height: 100%;
		left: 0;
		top: 0;
		background: var(--accent-primary);
		box-shadow: 0 0 10px var(--accent-primary);
		border-radius: var(--radius-pill);
		transition: background var(--transition-fast);

		&.warning {
			background: var(--danger); // Red
			box-shadow: 0 0 10px var(--danger);
		}
	}

	.thumb {
		position: absolute;
		top: 50%;
		width: 20px;
		height: 20px;
		background: #ffffff;
		border-radius: 50%;
		transform: translate(-50%, -50%);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
		transition: transform 0.1s;
		pointer-events: none; // Let wrapper handle events

		&.active {
			transform: translate(-50%, -50%) scale(1.2);
			box-shadow: 0 0 12px var(--accent-primary);
		}
	}

	.snap-point {
		position: absolute;
		top: 50%;
		width: 2px;
		height: 4px;
		background: rgba(255, 255, 255, 0.1);
		transform: translate(-50%, -50%);
		pointer-events: none;
	}

	.warning-text {
		margin-top: 8px;
		font-size: 0.8rem;
		font-weight: 700;
		color: var(--danger);
		text-align: right;
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 6px;

		.warning-icon {
			font-size: 16px;
			color: var(--danger);
		}
	}
</style>
