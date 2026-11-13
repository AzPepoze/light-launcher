<script lang="ts" module>
	export type DropdownOption = { value: string; label: string };

	/** Makes options whose label is the same as the value. */
	export function asOptions(values: string[]): DropdownOption[] {
		return values.map((value) => ({ value, label: value }));
	}
</script>

<script lang="ts">
	import { tick } from "svelte";
	import type { HTMLButtonAttributes } from "svelte/elements";
	import Icon from "./Icon.svelte";
	import IconButton from "./IconButton.svelte";
	import { filterDropdownOptions, showsDropdownSearch } from "./dropdownSearch";

	let {
		options = [],
		value = $bindable(""),
		placeholder = "Select an option",
		disabled = false,
		disabledMessage = "",
		searchable = false,
		onChange = () => {},
		...rest
	}: {
		options: DropdownOption[];
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		disabledMessage?: string;
		/** Shows a search box in the menu once there are more than eight options. */
		searchable?: boolean;
		onChange?: (value: string) => void;
	} & Omit<HTMLButtonAttributes, "value"> = $props();

	let isRendered = $state(false);
	let isOpen = $state(false);
	let isClosing = $state(false);
	let searchQuery = $state("");
	let triggerEl = $state<HTMLButtonElement | null>(null);
	let menuEl = $state<HTMLDivElement | null>(null);
	let searchEl = $state<HTMLInputElement | null>(null);
	let openTimeout: ReturnType<typeof setTimeout>;
	let closeTimeout: ReturnType<typeof setTimeout>;

	const selectedLabel = $derived(options.find((option) => option.value === value)?.label ?? placeholder);
	const showSearch = $derived(showsDropdownSearch(options.length, searchable));
	const visibleOptions = $derived(showSearch ? filterDropdownOptions(options, searchQuery) : options);

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
		searchQuery = "";
		openTimeout = setTimeout(() => {
			isOpen = true;
			focusSearch();
		}, 10);
	}

	async function focusSearch() {
		if (!showSearch) return;
		await tick();
		if (isRendered && isOpen && !isClosing) searchEl?.focus({ preventScroll: true });
	}

	function closeDropdown() {
		if (!isRendered || isClosing) return;
		clearTimeout(openTimeout);
		isClosing = true;
		closeTimeout = setTimeout(() => {
			isOpen = false;
			isRendered = false;
			isClosing = false;
			searchQuery = "";
		}, 150);
	}

	function select(option: DropdownOption) {
		value = option.value;
		closeDropdown();
		onChange(option.value);
	}

	function clearSearch() {
		searchQuery = "";
		searchEl?.focus({ preventScroll: true });
	}

	function itemButtons(): HTMLButtonElement[] {
		return Array.from(menuEl?.querySelectorAll<HTMLButtonElement>(".dropdown-item") ?? []);
	}

	function focusItem(index: number) {
		if (index < 0) {
			searchEl?.focus({ preventScroll: true });
			return;
		}
		const items = itemButtons();
		if (items.length === 0) return;
		items[Math.min(index, items.length - 1)]?.focus({ preventScroll: true });
	}

	function closeAndFocusTrigger() {
		closeDropdown();
		triggerEl?.focus({ preventScroll: true });
	}

	function handleSearchKeydown(event: KeyboardEvent) {
		if (event.key === "ArrowDown") {
			event.preventDefault();
			focusItem(0);
		} else if (event.key === "ArrowUp") {
			event.preventDefault();
			focusItem(visibleOptions.length - 1);
		} else if (event.key === "Escape") {
			event.preventDefault();
			event.stopPropagation();
			closeAndFocusTrigger();
		}
	}

	function handleItemKeydown(event: KeyboardEvent, index: number) {
		// Arrow keys only move between options in the new search mode; the plain menu keeps its own order.
		if (!showSearch) return;
		if (event.key === "ArrowDown") {
			event.preventDefault();
			focusItem(index + 1);
		} else if (event.key === "ArrowUp") {
			event.preventDefault();
			focusItem(index - 1);
		} else if (event.key === "Escape") {
			event.preventDefault();
			event.stopPropagation();
			closeAndFocusTrigger();
		}
	}

	function handleOutsideClick(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (isRendered && !target.closest(".custom-dropdown")) {
			closeDropdown();
		}
	}
</script>

<svelte:window onclick={handleOutsideClick} />

<div class="custom-dropdown" class:disabled>
	<button
		{...rest}
		bind:this={triggerEl}
		class="dropdown-trigger"
		class:open={isRendered && !isClosing}
		onclick={toggle}
		type="button"
	>
		<span class="text">{selectedLabel}</span>
		<span class="material-icons arrow">expand_more</span>
	</button>

	{#if disabled && disabledMessage}
		<div class="disabled-tooltip">
			<span class="material-icons info-icon">info</span>
			<span>{disabledMessage}</span>
		</div>
	{/if}

	{#if isRendered}
		<div class="dropdown-menu" class:open={isOpen && !isClosing} bind:this={menuEl}>
			{#if showSearch}
				<div class="dropdown-search">
					<Icon name="search" size={16} class="search-icon" />
					<input
						bind:this={searchEl}
						class="input dropdown-search-input"
						type="search"
						placeholder="Search options"
						aria-label="Search options"
						bind:value={searchQuery}
						onkeydown={handleSearchKeydown}
					/>
					{#if searchQuery}
						<IconButton icon="close" size={32} label="Clear search" onclick={clearSearch} />
					{/if}
				</div>
			{/if}
			{#each visibleOptions as option, i (option.value)}
				<button
					class="dropdown-item"
					class:selected={option.value === value}
					onclick={() => select(option)}
					onkeydown={(event) => handleItemKeydown(event, i)}
					type="button"
					style="animation-delay: {Math.min(i, 8) * 20}ms"
				>
					{option.label}
				</button>
			{/each}
			{#if showSearch && visibleOptions.length === 0}
				<p class="dropdown-empty">No matches</p>
			{/if}
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

	.dropdown-search {
		position: relative;
		margin-bottom: 6px;
	}

	.dropdown-search-input {
		width: 100%;
		padding: 8px 34px;
		font-size: 0.85rem;
		font-weight: 600;
		border-radius: var(--radius-pill);

		&:focus {
			transform: none;
			box-shadow: 0 0 0 2px var(--accent-glow);
		}
	}

	.dropdown-search :global(.search-icon) {
		position: absolute;
		top: 50%;
		left: 10px;
		transform: translateY(-50%);
		color: var(--text-muted);
		pointer-events: none;
	}

	.dropdown-search :global(.icon-button) {
		position: absolute;
		top: 50%;
		right: 2px;
		transform: translateY(-50%);
	}

	.dropdown-empty {
		margin: 0;
		padding: 10px 16px;
		color: var(--text-muted);
		font-size: 0.85rem;
		font-weight: 600;
		text-align: center;
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
