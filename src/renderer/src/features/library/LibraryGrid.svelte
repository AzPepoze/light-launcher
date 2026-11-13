<script lang="ts">
	import { onMount } from "svelte";
	import type { Game, SourceCard } from "../../lib/features/types";
	import GameTile from "./GameTile.svelte";
	import type { LibraryState } from "./libraryState.svelte";
	import type { TransferKind } from "./transferState";
	import { revealIndex } from "./libraryMotion";
	import { gridLayout, gridWindow, indexesIn, PAGE_SIZE } from "./libraryWindow";

	let {
		library,
		ratio,
		minColumnPx,
		card,
		transfers,
		blurUntilHover,
		revealed,
		onOpen
	}: {
		library: LibraryState;
		/** Width divided by height of each cover. */
		ratio: number;
		/** Smallest width of one column, from the card size setting. */
		minColumnPx: number;
		card: SourceCard | null;
		/** The download or install running for each game id. */
		transfers: Map<number, TransferKind>;
		blurUntilHover: boolean;
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

	const layout = $derived(gridLayout(width, ratio, minColumnPx));
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
						{ratio}
						{card}
						transfer={transfers.get(game.id) ?? null}
						{blurUntilHover}
						revealIndex={revealIndex(index, revealed)}
						{onOpen}
					/>
				{:else}
					<span class="skeleton cover" style:aspect-ratio={ratio} aria-hidden="true"></span>
				{/if}
			</li>
		{/each}
	</ul>
	<div style:height="{range.bottomSpacer}px" aria-hidden="true"></div>
</div>

<style lang="scss">
	.grid {
		display: grid;
		/* Covers form a dense wall: 12px between them both ways. */
		gap: 12px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* Each row is exactly rowHeight tall, so the spacers line up with the rows. */
	li {
		min-width: 0;
	}


	.cover {
		display: block;
		width: 100%;
		border-radius: var(--radius-md);
	}

</style>
