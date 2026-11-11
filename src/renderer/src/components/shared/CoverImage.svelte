<script lang="ts">
	import type { CoverSource } from "../../lib/features/coverCache";
	import { useCover } from "../../lib/features/useCover.svelte";

	/** Shows the whole image centered over a blurred copy of itself. */
	let { game, framed = false }: { game: CoverSource; framed?: boolean } = $props();
	const cover = useCover(() => game);
</script>

{#if cover.url}
	{#if framed}
		<img class="backdrop" src={cover.url} alt="" aria-hidden="true" loading="lazy" />
	{/if}
	<img class:framed src={cover.url} alt="" loading="lazy" />
{/if}

<style lang="scss">
	img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
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
