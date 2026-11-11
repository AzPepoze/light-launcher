<script lang="ts">
	import type { PluginPreview } from "../../lib/features/types";
	import { LEGAL_NOTICE, UNSIGNED_WARNING } from "./pluginCopy";

	let {
		preview,
		hint,
		lines,
		emptyText,
		confirmLabel,
		busy,
		onConfirm,
		onCancel,
	}: {
		preview: PluginPreview;
		/** Sentence above the permission lines. */
		hint: string;
		lines: string[];
		/** Shown when there are no lines. */
		emptyText: string;
		confirmLabel: string;
		busy: boolean;
		onConfirm: () => void;
		onCancel: () => void;
	} = $props();

	let accepted = $state(false);
</script>

<article class="consent" aria-label="Plugin details">
	<h3>{preview.name} <span class="muted">v{preview.version}</span></h3>

	{#if !preview.signed}
		<p class="notice" role="note">{UNSIGNED_WARNING}</p>
	{/if}

	<p class="label">{hint}</p>
	{#if lines.length}
		<ul class="permissions">
			{#each lines as line (line)}
				<li>{line}</li>
			{/each}
		</ul>
	{:else}
		<p class="muted">{emptyText}</p>
	{/if}

	<p class="legal">{LEGAL_NOTICE}</p>

	<label class="consent-check">
		<input type="checkbox" bind:checked={accepted} />
		<span>I trust this plugin and accept these permissions.</span>
	</label>

	<div class="actions">
		<button class="btn sm primary" disabled={!accepted || busy} onclick={onConfirm}>
			{confirmLabel}
		</button>
		<button class="btn sm" disabled={busy} onclick={onCancel}>Cancel</button>
	</div>
</article>

<style lang="scss">
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

	.consent-check {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.85rem;
		cursor: pointer;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
</style>
