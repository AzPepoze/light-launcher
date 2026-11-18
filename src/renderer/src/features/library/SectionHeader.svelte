<script lang="ts">
	import type { LibrarySyncProgress } from "../../lib/features/types";
	import { syncMessage } from "./librarySync";

	let {
		label,
		count,
		isPlugin,
		progress
	}: {
		label: string;
		count: number;
		/** Plugin sections report their sync progress here. */
		isPlugin: boolean;
		progress: LibrarySyncProgress | undefined;
	} = $props();

	const countText = $derived(count === 1 ? "1 game" : `${count} games`);
</script>

<header class="header">
	<h2>{label}</h2>
	<p class="count">{countText}</p>
</header>
<p class="sync" class:failed={progress?.phase === "error"} role="status">
	{#if isPlugin && progress}{syncMessage(progress)}{/if}
</p>

<style lang="scss">
	.header {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 12px;
		margin-bottom: 4px;
	}

	h2 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.count {
		margin: 0;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.sync {
		min-height: 1.5em;
		margin: 0 0 12px;
		font-size: 0.875rem;
		color: var(--text-muted);

		&.failed {
			color: var(--danger);
		}
	}
</style>
