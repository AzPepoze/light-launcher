<script lang="ts">
	import { fade } from "svelte/transition";
	import { tick } from "svelte";
	import { fadeParams, menuPanel } from "./motion";

	export let visible: boolean = false;
	export let x: number = 0;
	export let y: number = 0;
	/** 'cursor' = fixed at x/y with backdrop, 'anchor' = absolute anchored to trigger */
	export let mode: "cursor" | "anchor" = "cursor";
	/** Which side of the trigger an anchored panel opens on. */
	export let placement: "bottom" | "top" = "bottom";
	export let onClose: () => void = () => {};

	let menuEl: HTMLDivElement;
	let backdropEl: HTMLDivElement;

	// The panel stays in the DOM until its exit transition ends.
	let isRendered = false;
	$: isRendered = visible;
	$: if (visible) clampPosition();

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
		if (visible && e.key === "Escape") {
			onClose();
		}
	}

	function handleAnchorOutsideClick(e: MouseEvent) {
		if (mode !== "anchor" || !visible) return;
		const target = e.target as HTMLElement;
		if (menuEl && !menuEl.contains(target) && !target.closest(".menu-trigger")) {
			onClose();
		}
	}

	// transform origin based on mode
	$: transformOrigin =
		mode === "anchor" ? (placement === "top" ? "bottom left" : "top right") : "top left";
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
			in:fade={fadeParams("enter")}
			out:fade={fadeParams("exit")}
		>
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				bind:this={menuEl}
				class="app-menu-panel glass"
				style="left: {x}px; top: {y}px; transform-origin: {transformOrigin};"
				on:click|stopPropagation
				role="menu"
				in:menuPanel={{ placement }}
				out:menuPanel|global={{ placement }}
			>
				<slot />
			</div>
		</div>
	{:else}
		<div
			bind:this={menuEl}
			class="app-menu-panel anchor glass"
			class:anchor-top={placement === "top"}
			style="transform-origin: {transformOrigin};"
			role="menu"
			in:menuPanel={{ placement }}
			out:menuPanel={{ placement }}
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

		&.anchor {
			position: absolute;
			right: 0;
			top: calc(100% + 8px);
			min-width: 180px;
			max-height: var(--app-menu-max-height, none);
			overflow-y: auto;
			z-index: 100;
			box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);

			&.anchor-top {
				top: auto;
				right: auto;
				bottom: 0;
				left: 0;
				width: var(--app-menu-anchor-width, auto);
				min-width: 0;
			}
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

	@keyframes menuDividerIn {
		from { opacity: 0; }
		to { opacity: 1; }
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
		.app-menu-panel :global(.menu-item),
		.app-menu-panel :global(.menu-divider) {
			animation: none !important;
			opacity: 1 !important;
		}
	}
</style>
