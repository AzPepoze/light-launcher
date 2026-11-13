<script lang="ts">
	import { marqueeTextShift } from "../../features/home/state/marqueeGeometry";

	let {
		text,
		title,
		class: className = "",
		viewportClass = ""
	}: {
		text: string;
		/** Tooltip with the full text, for hosts that showed one. */
		title?: string;
		/** Extra class on the sliding label, for host typography. */
		class?: string;
		/** Extra class on the clipping viewport, such as a flex basis. */
		viewportClass?: string;
	} = $props();

	let viewportEl = $state<HTMLElement>();
	let textEl = $state<HTMLElement>();
	/** How far the label slides to show its end; 0 when it fits. */
	let shift = $state(0);

	function measure() {
		if (!viewportEl || !textEl) return;
		shift = marqueeTextShift(viewportEl.clientWidth, textEl.scrollWidth);
	}

	// Re-measures when the text, its font or the available width changes.
	$effect(() => {
		void text;
		const viewport = viewportEl;
		const label = textEl;
		if (!viewport || !label) return;
		measure();
		const observer = new ResizeObserver(measure);
		observer.observe(viewport);
		observer.observe(label);
		return () => observer.disconnect();
	});
</script>

<span class="marquee-viewport {viewportClass}" bind:this={viewportEl}>
	<span
		class="marquee-text {className}"
		class:scrolling={shift > 0}
		style="--marquee-shift: -{shift}px;"
		title={title}
		bind:this={textEl}>{text}</span
	>
</span>

<style lang="scss">
	.marquee-viewport {
		display: block;
		min-width: 0;
		overflow: hidden;
	}

	.marquee-text {
		display: inline-block;
		white-space: nowrap;
		transform: translateX(0);
	}

	.marquee-text.scrolling {
		animation: marquee-slide 6s ease-in-out infinite alternate;
	}

	// Holds at both ends, then slides across the overflow.
	@keyframes marquee-slide {
		0%,
		20% {
			transform: translateX(0);
		}
		80%,
		100% {
			transform: translateX(var(--marquee-shift));
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.marquee-text.scrolling {
			animation: none;
		}
	}
</style>
