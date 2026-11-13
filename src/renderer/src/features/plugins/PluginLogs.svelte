<script lang="ts">
	import { onMount } from "svelte";
	import Button from "@components/shared/Button.svelte";
	import EmptyState from "@components/shared/EmptyState.svelte";
	import LoadingState from "@components/shared/LoadingState.svelte";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginLogLine } from "../../lib/features/types";

	let { pluginId }: { pluginId: string } = $props();

	let lines = $state<PluginLogLine[]>([]);
	let loading = $state(true);
	let failed = $state(false);

	onMount(() => {
		load();
	});

	async function load() {
		loading = true;
		failed = false;
		try {
			lines = await pluginsApi.logs(pluginId);
		} catch {
			failed = true;
		} finally {
			loading = false;
		}
	}
</script>

<section aria-labelledby="logs-title">
	<div class="head">
		<h3 id="logs-title">Logs</h3>
		<Button size="sm" disabled={loading} onclick={load}>Refresh</Button>
	</div>

	{#if loading}
		<LoadingState label="Loading logs..." />
	{:else if failed}
		<EmptyState icon="error" title="Logs could not be loaded">
			<Button size="sm" onclick={load}>Retry</Button>
		</EmptyState>
	{:else if lines.length === 0}
		<p class="muted">No log lines yet.</p>
	{:else}
		<ul class="lines" aria-label="Plugin log">
			{#each lines as line, index (index)}
				<li class={line.level}>
					<span class="time">{line.time}</span>
					<span class="level">{line.level}</span>
					<span class="message">{line.message}</span>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style lang="scss">
	section {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	h3 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-muted);
	}

	.muted {
		margin: 0;
		font-size: 0.9375rem;
		color: var(--text-muted);
	}

	.lines {
		list-style: none;
		margin: 0;
		padding: 10px;
		max-height: 220px;
		overflow: auto;
		border-radius: var(--radius-md);
		background: var(--bg-input);
		font-family: monospace;
		font-size: 0.875rem;
		display: flex;
		flex-direction: column;
		gap: 4px;

		li {
			display: flex;
			gap: 8px;
			word-break: break-word;
		}

		.time {
			color: var(--text-muted);
			flex-shrink: 0;
		}

		.level {
			flex-shrink: 0;
			width: 44px;
			text-transform: uppercase;
			color: var(--text-muted);
		}

		.warn .level {
			color: var(--color-cleanup);
		}

		.error .level {
			color: var(--danger);
		}
	}
</style>
