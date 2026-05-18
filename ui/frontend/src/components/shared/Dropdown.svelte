<script lang="ts">
	import { fade } from "svelte/transition";

	export let options: string[] = [];
	export let value: string = "";
	export let placeholder: string = "Select an option";
	export let disabled: boolean = false;
	export let onChange: (value: string) => void = () => {};

	let isOpen = false;

	function toggle() {
		if (!disabled) isOpen = !isOpen;
	}

	function select(option: string) {
		value = option;
		isOpen = false;
		onChange(option);
	}

	function handleOutsideClick(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (isOpen && !target.closest(".custom-dropdown")) {
			isOpen = false;
		}
	}
</script>

<svelte:window on:click={handleOutsideClick} />

<div class="custom-dropdown" class:disabled>
	<button
		class="dropdown-trigger"
		class:open={isOpen}
		on:click={toggle}
		type="button"
	>
		<span class="text">{value || placeholder}</span>
		<span class="material-icons arrow">expand_more</span>
	</button>

	{#if isOpen}
		<div class="dropdown-menu" transition:fade={{ duration: 80 }}>
			{#each options as option}
				<button
					class="dropdown-item"
					class:selected={option === value}
					on:click={() => select(option)}
					type="button"
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
			opacity: 0.5;
			pointer-events: none;
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
		transition: background var(--transition-fast), color var(--transition-fast);
		margin-bottom: 2px;

		&:last-child {
			margin-bottom: 0;
		}

		&:hover {
			background: var(--bg-elevated);
			color: var(--text-main);
		}

		&.selected {
			background: var(--accent-primary);
			color: var(--bg-base);
			font-weight: 800;
		}
	}
</style>
