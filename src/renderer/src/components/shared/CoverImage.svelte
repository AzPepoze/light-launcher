<script lang="ts">
	import type { CoverSource } from "../../lib/features/mediaCache";
	import { useCover } from "../../lib/features/useMedia.svelte";

	/** Shows the whole image centered over a blurred copy of itself. A skeleton holds the space until it loads. */
	let { game, framed = false }: { game: CoverSource; framed?: boolean } = $props();
	const cover = useCover(() => game);
	let shown = $state(false);
	const loading = $derived(cover.pending || (cover.url !== null && !shown));

	function markShown() {
		shown = true;
	}
</script>

<span class="frame">
	{#if cover.url}
		{#if framed}
			<img class="backdrop" class:shown src={cover.url} alt="" aria-hidden="true" loading="lazy" />
		{/if}
		<img class:framed class:shown src={cover.url} alt="" loading="lazy" onload={markShown} onerror={markShown} />
	{/if}
	{#if loading}
		<span class="skeleton"></span>
	{/if}
</span>

<style lang="scss">
	.frame {
		display: block;
		position: relative;
		width: 100%;
		height: 100%;
	}

	.skeleton {
		position: absolute;
		inset: 0;
		display: block;
	}

	img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
		opacity: 0;
		transition: opacity var(--motion-base) var(--ease-out);

		&.shown {
			opacity: 1;
		}
	}

	.framed {
		position: relative;
		object-fit: contain;
	}

	.backdrop {
		position: absolute;
		inset: 0;
		object-fit: cover;
		filter: blur(18px) brightness(0.45);
		transform: scale(1.15);
	}
</style>
