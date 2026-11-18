<script lang="ts">
	import { onDestroy, onMount } from "svelte";
	import IconButton from "./IconButton.svelte";
	import {
		clampPan,
		clampZoom,
		fitScale,
		minimapToPan,
		minimapViewRect,
		zoomAround,
		type Point,
		type Size
	} from "./imageViewerMath";

	type ViewerImage = { url: string; alt: string };

	let {
		images,
		index,
		onIndexChange,
		onClose
	}: {
		images: ViewerImage[];
		index: number;
		onIndexChange: (index: number) => void;
		onClose: () => void;
	} = $props();

	/** The largest zoom, as a multiple of the image's natural size. */
	const MAX_SCALE = 4;
	const ZOOM_BUTTON_FACTOR = 1.25;
	/** Wheel zoom per pixel of wheel movement. */
	const WHEEL_ZOOM_RATE = 0.0015;
	const FIT_EPSILON = 1e-6;

	const current = $derived(images[index]);
	const returnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;

	let rootEl = $state<HTMLElement>();
	let stageEl = $state<HTMLElement>();
	let minimapEl = $state<HTMLElement>();
	let stripEl = $state<HTMLElement>();
	let stageBox = $state({ width: 0, height: 0 });
	let minimapBox = $state({ width: 0, height: 0 });
	let imageSize = $state<Size | null>(null);
	/** Null means the fit scale. A number is an explicit zoom. */
	let zoom = $state<number | null>(null);
	let pan = $state<Point>({ x: 0, y: 0 });
	let dragging = $state(false);
	let minimapDragging = $state(false);

	let dragStart: Point = { x: 0, y: 0 };
	let dragOrigin: Point = { x: 0, y: 0 };

	const stageSize = $derived<Size>({ width: stageBox.width, height: stageBox.height });
	const minimapSize = $derived<Size>({ width: minimapBox.width, height: minimapBox.height });
	const fitValue = $derived(
		imageSize && stageSize.width > 0
			? fitScale(imageSize.width, imageSize.height, stageSize.width, stageSize.height)
			: 1
	);
	const maxScale = $derived(Math.max(MAX_SCALE, fitValue));
	const scale = $derived(clampZoom(zoom ?? fitValue, fitValue, maxScale));
	const pictureSize = $derived<Size>({
		width: (imageSize?.width ?? 0) * scale,
		height: (imageSize?.height ?? 0) * scale
	});
	const shownPan = $derived(
		clampPan(pan, pictureSize.width, pictureSize.height, stageSize.width, stageSize.height)
	);
	const canPan = $derived(
		pictureSize.width > stageSize.width || pictureSize.height > stageSize.height
	);
	const minimapRect = $derived(
		imageSize && minimapSize.width > 0
			? minimapViewRect(shownPan, scale, stageSize, imageSize, minimapSize)
			: null
	);
	const percent = $derived(Math.round(scale * 100));

	/** Moves to another image. Each image opens at fit. */
	function show(next: number) {
		if (next < 0 || next >= images.length || next === index) return;
		zoom = null;
		pan = { x: 0, y: 0 };
		imageSize = null;
		onIndexChange(next);
	}

	function resetToFit() {
		zoom = null;
		pan = { x: 0, y: 0 };
	}

	/** Zooms to a scale, keeping the image point at the anchor still. The anchor is from the stage centre. */
	function zoomTo(nextScale: number, anchor: Point) {
		if (!imageSize) return;
		const next = clampZoom(nextScale, fitValue, maxScale);
		if (next <= fitValue + FIT_EPSILON) {
			resetToFit();
			return;
		}
		const nextPan = zoomAround(anchor, shownPan, scale, next);
		pan = clampPan(
			nextPan,
			imageSize.width * next,
			imageSize.height * next,
			stageSize.width,
			stageSize.height
		);
		zoom = next;
	}

	function zoomBy(factor: number) {
		zoomTo(scale * factor, { x: 0, y: 0 });
	}

	/** A pointer position from the stage centre. */
	function stagePoint(event: MouseEvent): Point {
		const rect = stageEl?.getBoundingClientRect();
		if (!rect) return { x: 0, y: 0 };
		return {
			x: event.clientX - rect.left - rect.width / 2,
			y: event.clientY - rect.top - rect.height / 2
		};
	}

	function handleWheel(event: WheelEvent) {
		event.preventDefault();
		zoomTo(scale * Math.exp(-event.deltaY * WHEEL_ZOOM_RATE), stagePoint(event));
	}

	function handleDoubleClick(event: MouseEvent) {
		if (zoom === null) zoomTo(2, stagePoint(event));
		else resetToFit();
	}

	function handlePointerDown(event: PointerEvent) {
		if (!canPan || event.button !== 0) return;
		stageEl?.setPointerCapture(event.pointerId);
		dragging = true;
		dragStart = { x: event.clientX, y: event.clientY };
		dragOrigin = shownPan;
	}

	function handlePointerMove(event: PointerEvent) {
		if (!dragging) return;
		pan = clampPan(
			{
				x: dragOrigin.x + event.clientX - dragStart.x,
				y: dragOrigin.y + event.clientY - dragStart.y
			},
			pictureSize.width,
			pictureSize.height,
			stageSize.width,
			stageSize.height
		);
	}

	function endDrag() {
		dragging = false;
	}

	/** Moves the view to the point under the pointer on the minimap. */
	function panFromMinimap(event: PointerEvent) {
		if (!imageSize || !minimapEl) return;
		const rect = minimapEl.getBoundingClientRect();
		const point = { x: event.clientX - rect.left, y: event.clientY - rect.top };
		pan = minimapToPan(point, minimapSize, imageSize, scale, stageSize);
	}

	function handleMinimapDown(event: PointerEvent) {
		minimapEl?.setPointerCapture(event.pointerId);
		minimapDragging = true;
		panFromMinimap(event);
	}

	function handleMinimapMove(event: PointerEvent) {
		if (minimapDragging) panFromMinimap(event);
	}

	function endMinimapDrag() {
		minimapDragging = false;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Escape") onClose();
		else if (event.key === "ArrowLeft") show(index - 1);
		else if (event.key === "ArrowRight") show(index + 1);
		else if (event.key === "+" || event.key === "=") zoomBy(ZOOM_BUTTON_FACTOR);
		else if (event.key === "-") zoomBy(1 / ZOOM_BUTTON_FACTOR);
		else if (event.key === "0") resetToFit();
		else return;
		// The viewer owns these keys while it is open, so the page behind it does not also act on them.
		event.stopImmediatePropagation();
		event.preventDefault();
	}

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) onClose();
	}

	function loadedSize(event: Event) {
		const picture = event.currentTarget as HTMLImageElement;
		imageSize = { width: picture.naturalWidth, height: picture.naturalHeight };
	}

	// Keep the active thumbnail in view when the image changes.
	$effect(() => {
		const thumb = stripEl?.children[index];
		const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
		thumb?.scrollIntoView({
			block: "nearest",
			inline: "center",
			behavior: reduceMotion ? "auto" : "smooth"
		});
	});

	onMount(() => {
		rootEl?.focus();
		// A non-passive listener so the wheel can be stopped from scrolling the page.
		stageEl?.addEventListener("wheel", handleWheel, { passive: false });
		// Capture phase runs before the page's own key handlers, so Escape closes only the viewer.
		window.addEventListener("keydown", handleKeydown, true);
		return () => {
			stageEl?.removeEventListener("wheel", handleWheel);
			window.removeEventListener("keydown", handleKeydown, true);
		};
	});

	onDestroy(() => {
		returnFocus?.focus();
	});
</script>

<div
	class="viewer"
	role="presentation"
	bind:this={rootEl}
	tabindex="-1"
	onclick={handleBackdropClick}
>
	<div class="frame" role="dialog" aria-modal="true" aria-label="Screenshot viewer">
		<div class="top">
			<p class="counter">{index + 1} / {images.length}</p>
			<div class="zoom">
				<IconButton
					icon="remove"
					label="Zoom out"
					size={32}
					disabled={zoom === null}
					onclick={() => zoomBy(1 / ZOOM_BUTTON_FACTOR)}
				/>
				<span class="percent" aria-live="polite">{percent}%</span>
				<IconButton
					icon="add"
					label="Zoom in"
					size={32}
					disabled={scale >= maxScale}
					onclick={() => zoomBy(ZOOM_BUTTON_FACTOR)}
				/>
				<IconButton
					icon="fit_screen"
					label="Fit to screen"
					size={32}
					disabled={zoom === null}
					onclick={resetToFit}
				/>
			</div>
			<IconButton icon="close" label="Close" size={32} onclick={onClose} />
		</div>

		<div class="stage-wrap">
			<div
				class="stage"
				class:grab={canPan && !dragging}
				class:grabbing={dragging}
				bind:this={stageEl}
				bind:clientWidth={stageBox.width}
				bind:clientHeight={stageBox.height}
				role="presentation"
				ondblclick={handleDoubleClick}
				onpointerdown={handlePointerDown}
				onpointermove={handlePointerMove}
				onpointerup={endDrag}
				onpointercancel={endDrag}
			>
				{#if current}
					<img
						class="picture"
						src={current.url}
						alt={current.alt}
						draggable="false"
						onload={loadedSize}
						style:width="{pictureSize.width}px"
						style:height="{pictureSize.height}px"
						style:transform="translate(calc(-50% + {shownPan.x}px), calc(-50% + {shownPan.y}px))"
					/>
				{/if}
			</div>

			<div class="nav prev">
				<IconButton
					icon="chevron_left"
					label="Previous screenshot"
					disabled={index <= 0}
					onclick={() => show(index - 1)}
				/>
			</div>
			<div class="nav next">
				<IconButton
					icon="chevron_right"
					label="Next screenshot"
					disabled={index >= images.length - 1}
					onclick={() => show(index + 1)}
				/>
			</div>

			{#if canPan && imageSize && current}
				<div
					class="minimap"
					role="presentation"
					bind:this={minimapEl}
					bind:clientWidth={minimapBox.width}
					bind:clientHeight={minimapBox.height}
					style:aspect-ratio="{imageSize.width} / {imageSize.height}"
					onpointerdown={handleMinimapDown}
					onpointermove={handleMinimapMove}
					onpointerup={endMinimapDrag}
					onpointercancel={endMinimapDrag}
				>
					<img class="minimap-picture" src={current.url} alt="" draggable="false" />
					{#if minimapRect}
						<div
							class="minimap-view"
							style:left="{minimapRect.x}px"
							style:top="{minimapRect.y}px"
							style:width="{minimapRect.width}px"
							style:height="{minimapRect.height}px"
						></div>
					{/if}
				</div>
			{/if}
		</div>

		<div class="strip" bind:this={stripEl}>
			{#each images as image, position}
				<button
					class="thumb"
					type="button"
					aria-label="Show screenshot {position + 1}"
					aria-current={position === index ? "true" : undefined}
					onclick={() => show(position)}
				>
					<img src={image.url} alt="" loading="lazy" draggable="false" />
				</button>
			{/each}
		</div>
	</div>
</div>

<style lang="scss">
	.viewer {
		position: fixed;
		inset: 0;
		z-index: 1100;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 24px;
		background: color-mix(in srgb, var(--bg-base) 88%, transparent);
		outline: none;
	}

	.frame {
		display: flex;
		flex-direction: column;
		width: 100%;
		height: 100%;
		max-width: 1400px;
		overflow: hidden;
		background: var(--bg-base);
		border: 1px solid var(--glass-border);
		border-radius: var(--radius-sm);
	}

	.top {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-shrink: 0;
		padding: 8px 12px;
		border-bottom: 1px solid var(--glass-border);
	}

	.counter {
		margin: 0 auto 0 0;
		font-size: 0.9rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-main);
	}

	.zoom {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.percent {
		min-width: 52px;
		text-align: center;
		font-size: 0.875rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-muted);
	}

	.stage-wrap {
		position: relative;
		flex: 1;
		min-height: 0;
	}

	.stage {
		position: absolute;
		inset: 0;
		overflow: hidden;
		touch-action: none;
		user-select: none;

		&.grab {
			cursor: grab;
		}

		&.grabbing {
			cursor: grabbing;
		}
	}

	.picture {
		position: absolute;
		left: 50%;
		top: 50%;
		max-width: none;
		display: block;
	}

	.nav {
		position: absolute;
		top: 50%;
		z-index: 2;
		transform: translateY(-50%);

		&.prev {
			left: 12px;
		}

		&.next {
			right: 12px;
		}

		:global(.icon-button) {
			background: color-mix(in srgb, var(--bg-surface) 70%, transparent);
			color: var(--text-main);
		}
	}

	.minimap {
		position: absolute;
		right: 12px;
		bottom: 12px;
		z-index: 2;
		width: 160px;
		overflow: hidden;
		cursor: pointer;
		touch-action: none;
		background: var(--bg-surface);
		border: 1px solid var(--glass-border);
		border-radius: 6px;
	}

	.minimap-picture {
		display: block;
		width: 100%;
		height: 100%;
		opacity: 0.75;
		pointer-events: none;
	}

	.minimap-view {
		position: absolute;
		box-sizing: border-box;
		border: 1.5px solid var(--accent-primary);
		background: rgb(255 255 255 / 0.08);
		pointer-events: none;
	}

	.strip {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
		overflow-x: auto;
		padding: 10px 12px;
		border-top: 1px solid var(--glass-border);
	}

	.thumb {
		flex: 0 0 auto;
		width: 112px;
		aspect-ratio: 16 / 9;
		padding: 0;
		overflow: hidden;
		border: 2px solid transparent;
		border-radius: var(--radius-sm);
		background: var(--bg-elevated);
		cursor: pointer;

		img {
			display: block;
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		&[aria-current="true"] {
			border-color: var(--accent-primary);
		}

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}
	}

	/* Phone: the minimap shrinks; the strip and zoom buttons stay. */
	@media (max-width: 599px) {
		.viewer {
			padding: 0;
		}

		.frame {
			border: none;
			border-radius: 0;
		}

		.minimap {
			width: 96px;
		}
	}
</style>
