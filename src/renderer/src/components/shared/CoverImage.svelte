<script lang="ts">
	import type { CoverSource } from "../../lib/features/coverCache";
	import { useCover } from "../../lib/features/useCover.svelte";

	let { game }: { game: CoverSource } = $props();
	const cover = useCover(() => game);
</script>

{#if cover.url}
	<img src={cover.url} alt="" loading="lazy" />
{/if}

<style lang="scss">
	img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
</style>
