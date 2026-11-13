<script lang="ts">
	import AppMenu from "@components/shared/AppMenu.svelte";
	import { launcherCountLabel } from "@lib/features/launcher";
	import type { GameInfo } from "@shared";
	import { tick } from "svelte";
	import {
		findScrollParent,
		isRectVisible,
		LAUNCHER_MENU_WIDTH,
		launcherChoices,
		launcherMenuPosition,
		pickLauncherSide,
		visibleArea,
		type LauncherMenuSide
	} from "./launcherMenu";

	let {
		game,
		onChoose = () => {}
	}: {
		game: GameInfo;
		/** Receives the picked launcher path, or null for automatic. */
		onChoose?: (launcherPath: string | null) => void;
	} = $props();

	/** The box the menu sits beside: the icon for grid cards, the row for list cards. */
	const ANCHOR_SELECTOR = ".game-icon-container, .list-card";

	let isOpen = $state(false);
	let menuX = $state(0);
	let menuY = $state(0);
	let menuWidth = $state(LAUNCHER_MENU_WIDTH);
	let choices = $derived(launcherChoices(game.launchers ?? []));

	let portalEl: HTMLElement | undefined = $state();

	// Plain values: they only matter while the menu is open and never drive the markup.
	let anchorEl: HTMLElement | null = null;
	let scrollArea: Element | null = null;
	let side: LauncherMenuSide = "right";
	let followFrame: number | null = null;

	/** Moves the panel to the body so the card's transform and overflow cannot clip it. */
	function portal(node: HTMLElement) {
		document.body.append(node);
		return { destroy: () => node.remove() };
	}

	function toggleMenu(event: MouseEvent) {
		event.stopPropagation();
		if (!isOpen) openMenu(event.currentTarget as HTMLElement);
		isOpen = !isOpen;
	}

	function openMenu(trigger: HTMLElement) {
		anchorEl = trigger.closest<HTMLElement>(ANCHOR_SELECTOR);
		if (!anchorEl) return;
		scrollArea = findScrollParent(anchorEl);
		refreshPlacement();
		measurePanel();
	}

	/** Reads the rendered panel's real width, then places the menu with it. */
	async function measurePanel() {
		await tick();
		const panel = portalEl?.querySelector<HTMLElement>(".app-menu-panel");
		if (!panel || !isOpen) return;
		menuWidth = panel.offsetWidth;
		refreshPlacement();
	}

	/** Picks the side from the current menu width, then places the menu beside the anchor. */
	function refreshPlacement() {
		if (!anchorEl) return;
		const anchorRect = anchorEl.getBoundingClientRect();
		side = pickLauncherSide(anchorRect.right, window.innerWidth, menuWidth);
		placeMenu(anchorRect);
	}

	function placeMenu(anchorRect: Pick<DOMRect, "left" | "right" | "top">) {
		const position = launcherMenuPosition(anchorRect, side, menuWidth);
		menuX = position.x;
		menuY = position.y;
	}

	/** Keeps the menu on the anchor; closes it once the anchor scrolls out of the view. */
	function followAnchor() {
		followFrame = null;
		if (!anchorEl || !scrollArea) return;
		const anchorRect = anchorEl.getBoundingClientRect();
		if (!isRectVisible(anchorRect, visibleArea(scrollArea))) {
			isOpen = false;
			return;
		}
		placeMenu(anchorRect);
	}

	function scheduleFollow() {
		if (followFrame !== null) return;
		followFrame = requestAnimationFrame(followAnchor);
	}

	function cancelFollow() {
		if (followFrame === null) return;
		cancelAnimationFrame(followFrame);
		followFrame = null;
	}

	function isInsideMenu(target: EventTarget | null): boolean {
		return target instanceof Element && target.closest(".app-menu-panel") !== null;
	}

	// The backdrop covers the grid, so pass wheel input on to the grid's scroll container.
	function handleWheel(event: WheelEvent) {
		if (isInsideMenu(event.target)) return;
		event.preventDefault();
		scrollArea?.scrollBy({ top: event.deltaY });
	}

	$effect(() => {
		if (!isOpen) return;
		window.addEventListener("wheel", handleWheel, { capture: true, passive: false });
		return () => window.removeEventListener("wheel", handleWheel, { capture: true });
	});

	$effect(() => {
		if (!isOpen) return;
		window.addEventListener("scroll", scheduleFollow, { capture: true, passive: true });
		return () => {
			window.removeEventListener("scroll", scheduleFollow, { capture: true });
			cancelFollow();
		};
	});

	function choose(event: MouseEvent, launcherPath: string | null) {
		event.stopPropagation();
		isOpen = false;
		onChoose(launcherPath);
	}
</script>

{#if game.launcherCount > 1}
	<div class="launcher-menu">
		<!-- Stops propagation so the window click that closes AppMenu does not run for this button. -->
		<button
			type="button"
			class="launcher-note"
			class:open={isOpen}
			title="Choose launcher"
			aria-haspopup="menu"
			aria-expanded={isOpen}
			onclick={toggleMenu}
		>
			<span class="material-icons">tune</span>
			<span>{launcherCountLabel(game.launcherCount)}</span>
		</button>
	</div>

	<div class="launcher-portal" bind:this={portalEl} use:portal>
		<AppMenu visible={isOpen} mode="cursor" x={menuX} y={menuY} onClose={() => (isOpen = false)}>
			{#each choices as choice (choice.path ?? "automatic")}
				<button type="button" class="menu-item" onclick={(event) => choose(event, choice.path)}>
					<span class="material-icons">{choice.icon}</span>
					<span class="launcher-text">
						<span class="launcher-name">{choice.label}</span>
						{#if choice.detail}
							<span class="launcher-kind">{choice.detail}</span>
						{/if}
					</span>
				</button>
			{/each}
		</AppMenu>
	</div>
{/if}

<style lang="scss">
	.launcher-menu {
		position: relative;
		display: inline-flex;
	}

	// Cursor-mode panels default to a wider min-width; keep the launcher list at the anchor width.
	.launcher-portal :global(.app-menu-panel) {
		min-width: 180px;
		max-height: calc(100vh - 20px);
		overflow-y: auto;
	}

	.launcher-note {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 8px;
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-sm);
		background: rgba(255, 255, 255, 0.04);
		color: var(--text-muted);
		font-size: 0.7rem;
		font-weight: 700;
		white-space: nowrap;
		cursor: pointer;
		transition:
			color var(--motion-quick),
			border-color var(--motion-quick),
			background var(--motion-quick);

		.material-icons {
			font-size: 14px;
		}

		&:hover,
		&.open {
			color: var(--text-main);
			border-color: var(--accent-primary);
		}
	}

	.launcher-text {
		display: flex;
		flex-direction: column;
		min-width: 0;
		text-align: left;
	}

	.launcher-name {
		font-weight: 700;
		color: var(--text-main);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.launcher-kind {
		font-size: 0.7rem;
		color: var(--text-muted);
	}

	@media (prefers-reduced-motion: reduce) {
		.launcher-note {
			transition: none;
		}
	}
</style>
