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
	<button class="dropdown-trigger" class:open={isOpen} on:click={toggle} type="button">
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
		padding: 12px 16px;
		font-family: inherit;
		font-size: 0.9rem;
		color: var(--text-main);
		cursor: pointer;
		text-align: left;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 10px;
		transition: all 0.2s;

		&:hover,
		&.open {
			border-color: var(--accent-secondary);
		}
	}

	.arrow {
		font-size: 0.6rem;
		color: var(--text-dim);
		transition: transform 0.2s;
	}

	.open .arrow {
		transform: rotate(180deg);
	}

	.dropdown-menu {
		position: absolute;
		top: calc(100% + 6px);
		left: 0;
		width: 100%;
		max-height: 220px;
		overflow-y: auto;
		z-index: 100;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border-bright);
		border-radius: 10px;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
		padding: 4px;
	}

	.dropdown-item {
		width: 100%;
		text-align: left;
		padding: 10px 14px;
		background: transparent;
		border: none;
		color: var(--text-muted);
		font-family: inherit;
		font-size: 0.85rem;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.1s;

		&:hover {
			background: rgba(255, 255, 255, 0.05);
			color: var(--text-main);
		}

		&.selected {
			background: var(--accent-primary);
			color: var(--glass-bg);
			font-weight: 600;
		}
	}
</style>
