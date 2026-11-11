<script lang="ts">
	import { useScreenshot } from "../../lib/features/useMedia.svelte";

	let {
		gameId,
		index,
		local,
		remote
	}: { gameId: number; index: number; local: string | null; remote: string } = $props();
	const shot = useScreenshot(() => ({ gameId, index, local, remote }));
</script>

{#if shot.url}
	<img src={shot.url} alt="" loading="lazy" />
{/if}

<style lang="scss">
	img {
		display: block;
		width: 100%;
		aspect-ratio: 16 / 9;
		object-fit: cover;
		border-radius: var(--radius-sm);
	}
</style>
