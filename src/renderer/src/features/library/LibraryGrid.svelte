<script lang="ts">
	import { onMount } from "svelte";
	import type { Game } from "../../lib/features/types";
	import GameTile from "./GameTile.svelte";
	import type { LibraryState } from "./libraryState.svelte";
	import { revealIndex } from "./libraryMotion";
	import { gridLayout, gridWindow, indexesIn, PAGE_SIZE } from "./libraryWindow";
	import type { TileShape } from "./sourceTileShape";

	let {
		library,
		shape,
		revealed,
		onOpen
	}: {
		library: LibraryState;
		shape: TileShape;
		/** True once the first reveal has run. Later renders do not animate tiles in. */
		revealed: boolean;
		onOpen: (game: Game) => void;
	} = $props();

	let gridEl = $state<HTMLElement>();
	let width = $state(0);
	let viewportHeight = $state(0);
	/** How far the visible area has moved past the top of the grid. */
	let scrolledPast = $state(0);
	let frame = 0;

	const layout = $derived(gridLayout(width, shape));
	const total = $derived(library.total ?? PAGE_SIZE);
	const range = $derived(
		gridWindow({ total, layout, scrollTop: scrolledPast, viewportHeight })
	);
	const indexes = $derived(indexesIn(range));

	// Keeps the pages behind the visible rows loaded. Runs again when the query resets the pages.
	$effect(() => {
		library.requestRange(range.start, range.end);
	});

	/** The nearest ancestor that scrolls. Null when the page itself scrolls. */
	function findScrollParent(element: HTMLElement): HTMLElement | null {
		for (let node = element.parentElement; node; node = node.parentElement) {
			const overflow = getComputedStyle(node).overflowY;
			if (overflow === "auto" || overflow === "scroll") return node;
		}
		return null;
	}

	function visibleTop(scroller: HTMLElement | null): number {
		return scroller ? scroller.getBoundingClientRect().top : 0;
	}

	function visibleHeight(scroller: HTMLElement | null): number {
		return scroller ? scroller.clientHeight : window.innerHeight;
	}

	/** Reads the position once per frame. Scroll events only ask for a frame. */
	function measure(scroller: HTMLElement | null) {
		frame = 0;
		if (!gridEl) return;
		scrolledPast = visibleTop(scroller) - gridEl.getBoundingClientRect().top;
		viewportHeight = visibleHeight(scroller);
		width = gridEl.clientWidth;
	}

	function scheduleMeasure(scroller: HTMLElement | null) {
		if (frame === 0) frame = requestAnimationFrame(() => measure(scroller));
	}

	/** A new section starts at the top of its grid, not wherever the last one was scrolled to. */
	function showGridTop(scroller: HTMLElement | null) {
		if (!gridEl || !scroller) return;
		const gap = visibleTop(scroller) - gridEl.getBoundingClientRect().top;
		if (gap > 0) scroller.scrollTop -= gap;
	}

	onMount(() => {
		if (!gridEl) return;
		const scroller = findScrollParent(gridEl);
		const target: EventTarget = scroller ?? window;
		const onScroll = () => scheduleMeasure(scroller);
		const resizeObserver = new ResizeObserver(onScroll);

		showGridTop(scroller);
		resizeObserver.observe(gridEl);
		if (scroller) resizeObserver.observe(scroller);
		target.addEventListener("scroll", onScroll, { passive: true });
		measure(scroller);

		return () => {
			resizeObserver.disconnect();
			target.removeEventListener("scroll", onScroll);
			cancelAnimationFrame(frame);
			frame = 0;
		};
	});
</script>

<div class="virtual" bind:this={gridEl}>
	<div style:height="{range.topSpacer}px" aria-hidden="true"></div>
	<ul
		class="grid"
		style:grid-template-columns="repeat({layout.columns}, minmax(0, 1fr))"
		style:grid-auto-rows="{layout.rowHeight}px"
	>
		{#each indexes as index (index)}
			{@const game = library.gameAt(index)}
			<li>
				{#if game}
					<GameTile
						{game}
						{shape}
						revealIndex={revealIndex(index, revealed)}
						{onOpen}
					/>
				{:else}
					<span class="placeholder" aria-hidden="true">
						<span class="skeleton cover" style:aspect-ratio={shape.ratio}></span>
						<span class="skeleton line"></span>
					</span>
				{/if}
			</li>
		{/each}
	</ul>
	<div style:height="{range.bottomSpacer}px" aria-hidden="true"></div>
</div>

<style lang="scss">
	.grid {
		display: grid;
		gap: 0 14px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* Each row is exactly rowHeight tall, so the spacers line up with the rows. */
	li {
		min-width: 0;
	}

	.placeholder {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.cover {
		display: block;
		width: 100%;
		border-radius: var(--radius-md);
	}

	.line {
		display: block;
		width: 70%;
		height: 14px;
		border-radius: var(--radius-sm);
	}
</style>
