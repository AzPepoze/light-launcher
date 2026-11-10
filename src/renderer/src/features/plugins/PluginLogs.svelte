<script lang="ts">
	import { onMount } from "svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginLogLine } from "../../lib/features/types";

	let { pluginId }: { pluginId: string } = $props();

	let lines = $state<PluginLogLine[]>([]);
	let loading = $state(true);
	let message = $state("");

	onMount(() => {
		load();
	});

	async function load() {
		loading = true;
		message = "";
		try {
			lines = await pluginsApi.logs(pluginId);
		} catch (err) {
			message = errorMessage(err);
		} finally {
			loading = false;
		}
	}
</script>

<section aria-labelledby="logs-title">
	<div class="head">
		<h3 id="logs-title">Logs</h3>
		<button class="btn sm" disabled={loading} onclick={load}>Refresh</button>
	</div>

	{#if message}
		<p class="error" role="alert">{message}</p>
	{:else if loading}
		<p class="muted">Loading logs...</p>
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
		font-size: 0.8rem;
		font-weight: 800;
		letter-spacing: 1px;
		text-transform: uppercase;
		color: var(--text-muted);
	}

	p {
		margin: 0;
	}

	.muted {
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.error {
		font-size: 0.85rem;
		color: var(--danger);
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
		font-size: 0.75rem;
		display: flex;
		flex-direction: column;
		gap: 4px;

		li {
			display: flex;
			gap: 8px;
			word-break: break-word;
		}

		.time {
			color: var(--text-dim);
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
