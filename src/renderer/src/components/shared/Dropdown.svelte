<script lang="ts">

	export let options: string[] = [];
	export let value: string = "";
	export let placeholder: string = "Select an option";
	export let disabled: boolean = false;
	export let disabledMessage: string = "";
	export let onChange: (value: string) => void = () => {};

	let isRendered = false;
	let isOpen = false;
	let isClosing = false;
	let openTimeout: ReturnType<typeof setTimeout>;
	let closeTimeout: ReturnType<typeof setTimeout>;

	function toggle() {
		if (disabled) return;
		if (isRendered) {
			closeDropdown();
		} else {
			openDropdown();
		}
	}

	function openDropdown() {
		clearTimeout(closeTimeout);
		isClosing = false;
		isRendered = true;
		openTimeout = setTimeout(() => {
			isOpen = true;
		}, 10);
	}

	function closeDropdown() {
		if (!isRendered || isClosing) return;
		clearTimeout(openTimeout);
		isClosing = true;
		closeTimeout = setTimeout(() => {
			isOpen = false;
			isRendered = false;
			isClosing = false;
		}, 150);
	}

	function select(option: string) {
		value = option;
		closeDropdown();
		onChange(option);
	}

	function handleOutsideClick(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (isRendered && !target.closest(".custom-dropdown")) {
			closeDropdown();
		}
	}
</script>

<svelte:window on:click={handleOutsideClick} />

<div class="custom-dropdown" class:disabled>
	<button
		class="dropdown-trigger"
		class:open={isRendered && !isClosing}
		on:click={toggle}
		type="button"
	>
		<span class="text">{value || placeholder}</span>
		<span class="material-icons arrow">expand_more</span>
	</button>

	{#if disabled && disabledMessage}
		<div class="disabled-tooltip">
			<span class="material-icons info-icon">info</span>
			<span>{disabledMessage}</span>
		</div>
	{/if}

	{#if isRendered}
		<div class="dropdown-menu" class:open={isOpen && !isClosing}>
			{#each options as option, i}
				<button
					class="dropdown-item"
					class:selected={option === value}
					on:click={() => select(option)}
					type="button"
					style="animation-delay: {Math.min(i, 8) * 20}ms"
				>
					{option}
				</button>
			{/each}
		</div>
	{/if}
</div>

<style lang="scss">
	.custom-dropdown {
		position: relative;
		width: 100%;

		&.disabled {
			opacity: 0.6;
			cursor: not-allowed;

			.dropdown-trigger {
				pointer-events: auto;
				cursor: not-allowed;
			}

			&:hover .disabled-tooltip {
				opacity: 1;
				visibility: visible;
				transform: translateX(-50%) translateY(0);
			}
		}
	}

	.disabled-tooltip {
		position: absolute;
		bottom: calc(100% + 8px);
		left: 50%;
		transform: translateX(-50%) translateY(4px);
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		color: var(--text-main);
		padding: 6px 12px;
		border-radius: var(--radius-md);
		font-size: 0.75rem;
		font-weight: 700;
		white-space: nowrap;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
		z-index: 110;
		opacity: 0;
		visibility: hidden;
		transition: opacity var(--transition-fast), transform var(--transition-fast);
		display: flex;
		align-items: center;
		gap: 6px;
		pointer-events: none;

		.info-icon {
			font-size: 1rem;
			color: var(--accent-primary);
		}

		&::after {
			content: '';
			position: absolute;
			top: 100%;
			left: 50%;
			transform: translateX(-50%);
			border-width: 6px;
			border-style: solid;
			border-color: var(--bg-surface) transparent transparent transparent;
		}
	}

	.dropdown-trigger {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 20px;
		font-family: inherit;
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--text-main);
		cursor: pointer;
		text-align: left;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-pill);
		transition: border-color var(--transition-fast), transform var(--transition-spring);

		&:hover,
		&.open {
			border-color: var(--accent-primary);
			transform: scale(1.005);
		}
	}

	.arrow {
		font-size: 1.2rem;
		color: var(--text-muted);
		transition: transform var(--transition-spring);
	}

	.open .arrow {
		transform: rotate(180deg);
		color: var(--accent-primary);
	}

	.dropdown-menu {
		position: absolute;
		top: calc(100% + 8px);
		left: 0;
		width: 100%;
		max-height: 220px;
		overflow-y: auto;
		z-index: 100;
		background: var(--bg-surface);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
		padding: 8px;
		box-sizing: border-box;

		opacity: 0;
		transform: translateY(-8px);
		transition: opacity 150ms cubic-bezier(0.25, 1, 0.5, 1), transform 150ms cubic-bezier(0.25, 1, 0.5, 1);
		pointer-events: none;

		&.open {
			opacity: 1;
			transform: translateY(0);
			pointer-events: auto;
		}

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: var(--glass-border-bright);
			border-radius: 10px;
		}
	}

	@keyframes dropdownItemScaleBounce {
		from {
			opacity: 0;
			transform: scale(0.85);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}

	.dropdown-item {
		width: 100%;
		text-align: left;
		padding: 10px 16px;
		background: transparent;
		border: none;
		color: var(--text-muted);
		font-family: inherit;
		font-size: 0.85rem;
		font-weight: 700;
		border-radius: var(--radius-pill);
		cursor: pointer;
		transition: background var(--transition-fast), color var(--transition-fast), transform var(--transition-fast);
		margin-bottom: 2px;
		opacity: 0;
		animation: dropdownItemScaleBounce 280ms cubic-bezier(0.34, 1.56, 0.64, 1) both;

		&:last-child {
			margin-bottom: 0;
		}

		&:hover {
			background: var(--bg-elevated);
			color: var(--text-main);
		}

		&:active {
			transform: scale(0.97);
		}

		&.selected {
			background: var(--accent-primary);
			color: var(--bg-base);
			font-weight: 800;
		}
	}
</style>
