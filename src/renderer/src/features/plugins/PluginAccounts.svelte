<script lang="ts">
	import { onMount } from "svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { AuthState, PluginInfo } from "../../lib/features/types";
	import { LEGAL_NOTICE } from "./pluginCopy";

	let { plugin }: { plugin: PluginInfo } = $props();

	let statuses = $state<Record<string, AuthState>>({});
	let credentials = $state<Record<string, string>>({});
	let busySource = $state("");
	let message = $state("");

	const STATUS_LABELS: Record<AuthState["status"], string> = {
		unconfigured: "Not signed in",
		ok: "Signed in",
		expired: "Sign-in expired",
		invalid: "Sign-in rejected",
		error: "Check failed",
	};

	onMount(() => {
		for (const source of plugin.sources) refresh(source);
	});

	async function refresh(source: string) {
		try {
			statuses[source] = await pluginsApi.authStatus(plugin.id, source);
		} catch (err) {
			statuses[source] = { status: "error", message: errorMessage(err) };
		}
	}

	async function run(source: string, action: () => Promise<unknown>) {
		busySource = source;
		message = "";
		try {
			await action();
			await refresh(source);
		} catch (err) {
			message = errorMessage(err);
		} finally {
			busySource = "";
		}
	}

	function saveCredential(source: string) {
		const value = credentials[source]?.trim();
		if (!value) return;
		run(source, () => pluginsApi.call(plugin.id, "source.setCredential", { source, value }));
		credentials[source] = "";
	}

	function clearCredential(source: string) {
		run(source, () => pluginsApi.call(plugin.id, "source.clearCredential", { source }));
	}
</script>

<section aria-labelledby="accounts-title">
	<h3 id="accounts-title">Accounts</h3>
	<p class="legal">{LEGAL_NOTICE}</p>

	<ul class="sources">
		{#each plugin.sources as source (source)}
			{@const auth = statuses[source]}
			<li>
				<div class="head">
					<span class="name">{source}</span>
					<span class="status">{auth ? STATUS_LABELS[auth.status] : "Checking..."}</span>
				</div>
				{#if auth?.user}
					<p class="muted">Signed in as {auth.user}</p>
				{/if}
				<div class="row">
					<label class="sr-only" for="credential-{source}">Credential for {source}</label>
					<input
						id="credential-{source}"
						class="input"
						type="password"
						autocomplete="off"
						placeholder="Paste a session token"
						bind:value={credentials[source]}
					/>
					<button
						class="btn sm primary"
						disabled={busySource === source || !credentials[source]?.trim()}
						onclick={() => saveCredential(source)}
					>
						Save
					</button>
					<button class="btn sm" disabled={busySource === source} onclick={() => clearCredential(source)}>
						Clear
					</button>
				</div>
			</li>
		{/each}
	</ul>

	{#if message}
		<p class="error" role="alert">{message}</p>
	{/if}
</section>

<style lang="scss">
	section {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	h3 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-muted);
	}

	p {
		margin: 0;
	}

	.legal,
	.muted {
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.sources {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 10px;

		li {
			display: flex;
			flex-direction: column;
			gap: 8px;
			padding: 12px;
			border-radius: var(--radius-md);
			background: var(--bg-surface);
		}
	}

	.head {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		font-weight: 700;
		font-size: 0.9375rem;
	}

	.status {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-muted);
	}

	.row {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;

		.input {
			flex: 1 1 200px;
			min-width: 0;
		}
	}

	.error {
		font-size: 0.9375rem;
		color: var(--danger);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
</style>
