<script lang="ts">
	import type { LibrarySyncProgress } from "../../lib/features/types";
	import { syncMessage } from "./librarySync";

	let {
		label,
		count,
		isPlugin,
		syncing,
		progress,
		onSync
	}: {
		label: string;
		count: number;
		/** Plugin sections can be synced from their source. */
		isPlugin: boolean;
		syncing: boolean;
		progress: LibrarySyncProgress | undefined;
		onSync: () => void;
	} = $props();

	const countText = $derived(count === 1 ? "1 game" : `${count} games`);
</script>

<header class="header">
	<div class="title">
		<h2>{label}</h2>
		<p class="count">{countText}</p>
	</div>
	{#if isPlugin}
		<button class="btn sm" disabled={syncing} onclick={onSync}>
			{syncing ? "Syncing..." : "Sync now"}
		</button>
	{/if}
</header>
<p class="sync" class:failed={progress?.phase === "error"} role="status">
	{#if isPlugin && progress}{syncMessage(progress)}{/if}
</p>

<style lang="scss">
	.header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 4px;
	}

	.title {
		display: flex;
		align-items: baseline;
		gap: 12px;
		min-width: 0;
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
