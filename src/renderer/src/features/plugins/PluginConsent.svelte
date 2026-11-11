<script lang="ts">
	import type { PermissionLines, PluginPreview } from "../../lib/features/types";
	import {
		LEGAL_NOTICE,
		REPOSITORY_CHANGED_WARNING,
		SITES_HEADING,
		UNSIGNED_WARNING
	} from "./pluginCopy";

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
		lines: PermissionLines;
		/** Shown when there are no lines. */
		emptyText: string;
		confirmLabel: string;
		busy: boolean;
		onConfirm: () => void;
		onCancel: () => void;
	} = $props();

	let accepted = $state(false);
	const isEmpty = $derived(lines.sites.length === 0 && lines.other.length === 0);
</script>

<article class="consent" aria-label="Plugin details">
	<h3>{preview.name} <span class="muted">v{preview.version}</span></h3>

	{#if !preview.signed}
		<p class="notice" role="note">{UNSIGNED_WARNING}</p>
	{/if}

	{#if preview.repositoryChanged}
		<p class="notice" role="note">
			{REPOSITORY_CHANGED_WARNING}
			{#if preview.installedRepository}
				Installed from {preview.installedRepository}.
			{/if}
		</p>
	{/if}

	<p class="label">{hint}</p>
	{#if isEmpty}
		<p class="muted">{emptyText}</p>
	{:else}
		{#if lines.sites.length}
			<p class="sites-heading">{SITES_HEADING}</p>
			<ul class="sites">
				{#each lines.sites as site (site)}
					<li>{site}</li>
				{/each}
			</ul>
		{/if}
		{#if lines.other.length}
			<ul class="permissions">
				{#each lines.other as line (line)}
					<li>{line}</li>
				{/each}
			</ul>
		{/if}
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

	.sites-heading {
		font-size: 0.85rem;
		font-weight: 700;
	}

	.sites {
		margin: 0;
		padding-left: 18px;
		font-size: 0.85rem;
		columns: 2;
		column-gap: 24px;
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
