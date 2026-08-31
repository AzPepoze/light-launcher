<script lang="ts">
	import { fade } from "svelte/transition";
	import { tick } from "svelte";

	export let visible: boolean = false;
	export let x: number = 0;
	export let y: number = 0;
	/** 'cursor' = fixed at x/y with backdrop, 'anchor' = absolute anchored to trigger */
	export let mode: "cursor" | "anchor" = "cursor";
	export let onClose: () => void = () => {};

	let menuEl: HTMLDivElement;
	let backdropEl: HTMLDivElement;

	// internal render state to allow close animation before DOM removal
	let isRendered = false;
	let isOpen = false;
	let isClosing = false;
	let openTimer: ReturnType<typeof setTimeout>;
	let closeTimer: ReturnType<typeof setTimeout>;

	$: {
		if (visible) {
			clearTimeout(closeTimer);
			clearTimeout(openTimer);
			if (!isRendered) isRendered = true;
			isClosing = false;
			// next tick trigger CSS transition (like Dropdown does with 10ms)
			openTimer = setTimeout(() => {
				isOpen = true;
				clampPosition();
			}, 10);
		} else if (isRendered) {
			clearTimeout(openTimer);
			isOpen = false;
			isClosing = true;
			// keep DOM for exit animation, then unmount
			closeTimer = setTimeout(() => {
				isRendered = false;
				isClosing = false;
			}, 170);
		}
	}

	// Clamp to viewport for cursor mode after render
	async function clampPosition() {
		await tick();
		if (!menuEl || mode !== "cursor") return;
		const rect = menuEl.getBoundingClientRect();
		const vw = window.innerWidth;
		const vh = window.innerHeight;
		let nx = x;
		let ny = y;
		if (nx + rect.width > vw - 10) nx = vw - rect.width - 10;
		if (ny + rect.height > vh - 10) ny = vh - rect.height - 10;
		if (nx < 10) nx = 10;
		if (ny < 10) ny = 10;
		if (nx !== x) x = nx;
		if (ny !== y) y = ny;
	}

	function handleBackdropClick() {
		onClose();
	}

	function handleKeydown(e: KeyboardEvent) {
		if ((visible || isRendered) && e.key === "Escape") {
			onClose();
		}
	}

	function handleAnchorOutsideClick(e: MouseEvent) {
		if (mode !== "anchor" || !isRendered || isClosing) return;
		const target = e.target as HTMLElement;
		if (menuEl && !menuEl.contains(target) && !target.closest(".menu-trigger")) {
			onClose();
		}
	}

	// transform origin based on mode
	$: transformOrigin = mode === "anchor" ? "top right" : "top left";
</script>

<svelte:window on:keydown={handleKeydown} on:click={handleAnchorOutsideClick} />

{#if isRendered}
	{#if mode === "cursor"}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			bind:this={backdropEl}
			class="app-menu-backdrop"
			on:click={handleBackdropClick}
			on:contextmenu|preventDefault={handleBackdropClick}
			transition:fade={{ duration: 100 }}
		>
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				bind:this={menuEl}
				class="app-menu-panel glass"
				class:open={isOpen && !isClosing}
				class:closing={isClosing}
				style="left: {x}px; top: {y}px; transform-origin: {transformOrigin};"
				on:click|stopPropagation
				role="menu"
			>
				<slot />
			</div>
		</div>
	{:else}
		<div
			bind:this={menuEl}
			class="app-menu-panel anchor glass"
			class:open={isOpen && !isClosing}
			class:closing={isClosing}
			style="transform-origin: {transformOrigin};"
			role="menu"
		>
			<slot />
		</div>
	{/if}
{/if}

<style lang="scss">
	.app-menu-backdrop {
		position: fixed;
		inset: 0;
		z-index: 9999;
		background: transparent;
	}

	.app-menu-panel {
		position: absolute;
		min-width: 210px;
		background: var(--bg-surface);
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-md);
		padding: 6px;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
		display: flex;
		flex-direction: column;
		gap: 2px;
		z-index: 10000;

		// container animation : closed state
		opacity: 0;
		transform: translateY(-8px) scale(0.96);
		transition:
			opacity 180ms cubic-bezier(0.215, 0.61, 0.355, 1),
			transform 180ms cubic-bezier(0.175, 0.885, 0.32, 1.275);
		pointer-events: none;

		&.open {
			opacity: 1;
			transform: translateY(0) scale(1);
			pointer-events: auto;
		}

		&.closing {
			opacity: 0;
			transform: translateY(-6px) scale(0.97);
			pointer-events: none;
			transition:
				opacity 130ms cubic-bezier(0.4, 0, 1, 1),
				transform 130ms cubic-bezier(0.4, 0, 1, 1);
		}

		&.anchor {
			position: absolute;
			right: 0;
			top: calc(100% + 8px);
			min-width: 180px;
			z-index: 100;
			box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
		}
	}

	/* Staggered entrance for menu items */
	.app-menu-panel :global(.menu-item) {
		opacity: 0;
		animation: menuItemIn 260ms cubic-bezier(0.34, 1.56, 0.64, 1) both;
	}

	.app-menu-panel :global(.menu-item:nth-child(1)) { animation-delay: 0ms; }
	.app-menu-panel :global(.menu-item:nth-child(2)) { animation-delay: 22ms; }
	.app-menu-panel :global(.menu-item:nth-child(3)) { animation-delay: 44ms; }
	.app-menu-panel :global(.menu-item:nth-child(4)) { animation-delay: 66ms; }
	.app-menu-panel :global(.menu-item:nth-child(5)) { animation-delay: 88ms; }
	.app-menu-panel :global(.menu-item:nth-child(6)) { animation-delay: 110ms; }
	.app-menu-panel :global(.menu-item:nth-child(7)) { animation-delay: 132ms; }
	.app-menu-panel :global(.menu-item:nth-child(8)) { animation-delay: 154ms; }
	.app-menu-panel :global(.menu-item:nth-child(9)) { animation-delay: 154ms; }
	.app-menu-panel :global(.menu-item:nth-child(10)) { animation-delay: 154ms; }

	/* Divider stagger */
	.app-menu-panel :global(.menu-divider) {
		opacity: 0;
		animation: menuDividerIn 160ms ease both;
	}
	.app-menu-panel :global(.menu-divider:nth-child(1)) { animation-delay: 0ms; }
	.app-menu-panel :global(.menu-divider:nth-child(2)) { animation-delay: 22ms; }
	.app-menu-panel :global(.menu-divider:nth-child(3)) { animation-delay: 44ms; }
	.app-menu-panel :global(.menu-divider:nth-child(4)) { animation-delay: 66ms; }
	.app-menu-panel :global(.menu-divider:nth-child(5)) { animation-delay: 88ms; }
	.app-menu-panel :global(.menu-divider:nth-child(6)) { animation-delay: 110ms; }

	/* Close: items fade/scale out quickly, reversed stagger for polish */
	.app-menu-panel.closing :global(.menu-item) {
		animation: menuItemOut 110ms cubic-bezier(0.4, 0, 1, 1) both;
	}
	.app-menu-panel.closing :global(.menu-item:nth-child(1)) { animation-delay: 56ms; }
	.app-menu-panel.closing :global(.menu-item:nth-child(2)) { animation-delay: 42ms; }
	.app-menu-panel.closing :global(.menu-item:nth-child(3)) { animation-delay: 28ms; }
	.app-menu-panel.closing :global(.menu-item:nth-child(4)) { animation-delay: 14ms; }
	.app-menu-panel.closing :global(.menu-item:nth-child(5)) { animation-delay: 0ms; }
	.app-menu-panel.closing :global(.menu-item:nth-child(6)) { animation-delay: 0ms; }
	.app-menu-panel.closing :global(.menu-item:nth-child(7)) { animation-delay: 0ms; }
	.app-menu-panel.closing :global(.menu-item:nth-child(8)) { animation-delay: 0ms; }

	.app-menu-panel.closing :global(.menu-divider) {
		animation: menuDividerOut 90ms ease both;
	}

	@keyframes menuItemIn {
		from {
			opacity: 0;
			transform: scale(0.92) translateY(-4px);
		}
		to {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}

	@keyframes menuItemOut {
		from {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
		to {
			opacity: 0;
			transform: scale(0.94) translateY(-4px);
		}
	}

	@keyframes menuDividerIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	@keyframes menuDividerOut {
		from { opacity: 1; }
		to { opacity: 0; }
	}

	/* Shared menu-item styles */
	.app-menu-panel :global(.menu-item) {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 14px;
		background: none;
		border: none;
		border-radius: var(--radius-sm);
		color: var(--text-main);
		font-size: 0.88rem;
		font-weight: 700;
		cursor: pointer;
		text-align: left;
		transition: background var(--transition-fast), color var(--transition-fast), transform var(--transition-fast);
		width: 100%;
		box-sizing: border-box;
		font-family: inherit;

		&:hover {
			background: var(--accent-glow);
			color: var(--accent-primary);
		}

		&:active {
			transform: scale(0.97);
		}

		:global(.material-icons) {
			font-size: 18px;
		}
	}

	// FIX: danger styles must be separate :global selectors (svelte nesting with & breaks :global)
	.app-menu-panel :global(.menu-item.danger:hover),
	.app-menu-panel :global(.menu-item.danger-soft:hover) {
		background: rgba(255, 68, 68, 0.1);
		color: var(--danger, #ff4444);
	}

	.app-menu-panel :global(.menu-divider) {
		height: 1px;
		background: rgba(255, 255, 255, 0.06);
		margin: 4px 6px;
		border: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.app-menu-panel,
		.app-menu-panel.open,
		.app-menu-panel.closing {
			transition: none !important;
		}
		.app-menu-panel :global(.menu-item),
		.app-menu-panel :global(.menu-divider) {
			animation: none !important;
			opacity: 1 !important;
		}
	}
</style>
