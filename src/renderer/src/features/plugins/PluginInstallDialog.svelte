<script lang="ts">
	import Modal from "@components/shared/Modal.svelte";
	import { errorMessage } from "../../lib/features/rpc";
	import { pluginsApi } from "../../lib/features/pluginsApi";
	import type { PluginInfo, PluginPreview } from "../../lib/features/types";
	import { LEGAL_NOTICE, PERMISSION_HINT, UNSIGNED_WARNING } from "./pluginCopy";

	let {
		onInstalled,
		onClose,
	}: { onInstalled: (plugin: PluginInfo) => void; onClose: () => void } = $props();

	let source = $state("");
	let preview = $state<PluginPreview | null>(null);
	let consented = $state(false);
	let busy = $state(false);
	let message = $state("");

	async function check() {
		const link = source.trim();
		if (!link) return;
		busy = true;
		message = "";
		consented = false;
		try {
			preview = await pluginsApi.previewPermissions(link);
		} catch (err) {
			preview = null;
			message = errorMessage(err);
		} finally {
			busy = false;
		}
	}

	async function install() {
		if (!preview || !consented) return;
		busy = true;
		message = "";
		try {
			onInstalled(await pluginsApi.install(preview.source, preview.permissions));
		} catch (err) {
			message = errorMessage(err);
		} finally {
			busy = false;
		}
	}
</script>

<Modal show={true} title="Install plugin" showDone={false} {onClose}>
	<form
		class="row"
		onsubmit={(event) => {
			event.preventDefault();
			check();
		}}
	>
		<label class="field">
			<span>Plugin link or file</span>
			<input class="input" type="text" bind:value={source} placeholder="https://... or a file path" />
		</label>
		<button class="btn sm" type="submit" disabled={busy || !source.trim()}>Check</button>
	</form>

	{#if message}
		<p class="error" role="alert">{message}</p>
	{/if}

	{#if preview}
		<article class="consent" aria-label="Plugin details">
			<h3>{preview.name} <span class="muted">v{preview.version}</span></h3>
			{#if preview.author}<p class="muted">By {preview.author}</p>{/if}
			{#if preview.description}<p>{preview.description}</p>{/if}

			{#if !preview.signed}
				<p class="notice" role="note">{UNSIGNED_WARNING}</p>
			{/if}

			<p class="label">{PERMISSION_HINT}</p>
			{#if preview.permissions.length}
				<ul class="permissions">
					{#each preview.permissions as permission (permission)}
						<li>{permission}</li>
					{/each}
				</ul>
			{:else}
				<p class="muted">No extra permissions.</p>
			{/if}

			<p class="legal">{LEGAL_NOTICE}</p>

			<label class="consent-check">
				<input type="checkbox" bind:checked={consented} />
				<span>I trust this plugin and accept these permissions.</span>
			</label>

			<button class="btn sm primary" disabled={!consented || busy} onclick={install}>
				Install plugin
			</button>
		</article>
	{/if}
</Modal>

<style lang="scss">
	.row {
		display: flex;
		align-items: flex-end;
		gap: 8px;
		margin: 0;

		.field {
			flex: 1;
			min-width: 0;
			display: flex;
			flex-direction: column;
			gap: 6px;
			font-size: 0.85rem;
			color: var(--text-muted);
		}

		.input {
			width: 100%;
		}
	}

	.consent {
		margin-top: 18px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		align-items: flex-start;

		h3 {
			margin: 0;
			font-size: 1.1rem;
		}

		p {
			margin: 0;
		}
	}

	.muted,
	.legal {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.label {
		font-weight: 700;
		font-size: 0.85rem;
	}

	.permissions {
		margin: 0;
		padding-left: 18px;
		font-size: 0.85rem;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.notice {
		padding: 10px 14px;
		border-radius: var(--radius-md);
		border: 1px solid var(--danger);
		font-size: 0.85rem;
	}

	.error {
		margin: 12px 0 0;
		font-size: 0.85rem;
		color: var(--danger);
	}

	.consent-check {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.85rem;
		cursor: pointer;
	}
</style>
