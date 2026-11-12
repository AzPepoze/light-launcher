<script lang="ts">
	import type { PermissionKind, PermissionLines, PluginPreview } from "../../lib/features/types";
	import Checkbox from "@components/shared/Checkbox.svelte";
	import Notice from "./Notice.svelte";
	import PluginIdentity from "./PluginIdentity.svelte";
	import {
		LEGAL_NOTICE,
		REPOSITORY_CHANGED_WARNING,
		UNSIGNED_WARNING,
		sitesTitle
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
		/** Sentence above the permission rows. */
		hint: string;
		lines: PermissionLines;
		/** Shown when there are no rows. */
		emptyText: string;
		confirmLabel: string;
		busy: boolean;
		onConfirm: () => void;
		onCancel: () => void;
	} = $props();

	/** Rows past this one appear at once, so a long list does not keep the user waiting. */
	const STAGGER_LIMIT = 8;
	const KIND_ICONS: Record<PermissionKind, string> = {
		site: "language",
		files: "folder",
		"sign-in": "key",
		browser: "open_in_browser",
		program: "terminal",
		other: "extension"
	};

	const uid = $props.id();
	const sitesId = `${uid}-sites`;
	let accepted = $state(false);
	let sitesOpen = $state(false);
	const hasSites = $derived(lines.sites.length > 0);
	/** The sites row is the first row, so permission rows start one place later. */
	const siteOffset = $derived(hasSites ? 1 : 0);
	const isEmpty = $derived(!hasSites && lines.other.length === 0);
</script>

<article class="consent" aria-label="Plugin details">
	<header class="head">
		<PluginIdentity id={preview.id} name={preview.name} large />
		<div>
			<h3>{preview.name}</h3>
			<p class="version">v{preview.version}</p>
		</div>
	</header>

	{#if !preview.signed}
		<Notice icon="warning">
			<p>{UNSIGNED_WARNING}</p>
		</Notice>
	{/if}

	{#if preview.repositoryChanged}
		<Notice icon="report">
			<p>
				{REPOSITORY_CHANGED_WARNING}
				{#if preview.installedRepository}
					Installed from {preview.installedRepository}.
				{/if}
			</p>
		</Notice>
	{/if}

	<section class="ledger" aria-label={hint}>
		<p class="hint">{hint}</p>
		{#if isEmpty}
			<p class="empty">{emptyText}</p>
		{:else}
			<ul class="rows">
				{#if hasSites}
					<li class="row sites-row staggered" style:--row-index={0}>
						<button
							class="sites-toggle"
							type="button"
							aria-expanded={sitesOpen}
							aria-controls={sitesId}
							onclick={() => (sitesOpen = !sitesOpen)}
						>
							<span class="material-icons icon" aria-hidden="true">language</span>
							<span class="sites-title">{sitesTitle(lines.sites.length)}</span>
							<span class="material-icons chevron" class:turned={sitesOpen} aria-hidden="true">
								expand_more
							</span>
						</button>
						<div id={sitesId} class="sites-body" class:open={sitesOpen} inert={!sitesOpen}>
							<div class="clip">
								<ul class="sites">
									{#each lines.sites as site (site)}
										<li>{site}</li>
									{/each}
								</ul>
							</div>
						</div>
					</li>
				{/if}

				{#each lines.other as row, index (row.text)}
					{@const position = index + siteOffset}
					<li
						class="row"
						class:sensitive={row.sensitive}
						class:staggered={position < STAGGER_LIMIT}
						style:--row-index={position}
					>
						<span class="material-icons icon" aria-hidden="true">{KIND_ICONS[row.kind]}</span>
						<span class="text">{row.text}</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<p class="legal">{LEGAL_NOTICE}</p>

	<Checkbox checked={accepted} onChange={(value) => (accepted = value)}>
		I trust this plugin and accept these permissions.
	</Checkbox>

	<div class="actions">
		<button class="btn sm primary install" disabled={!accepted || busy} onclick={onConfirm}>
			{confirmLabel}
		</button>
		<button class="btn sm secondary" disabled={busy} onclick={onCancel}>Cancel</button>
	</div>
</article>

<style lang="scss">
	.consent {
		margin-top: 18px;
		display: flex;
		flex-direction: column;
		gap: 16px;
		align-items: stretch;
		font-size: 1rem;
		line-height: 1.5;

		p {
			margin: 0;
		}
	}

	.head {
		display: flex;
		align-items: center;
		gap: 14px;

		h3 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 600;
		}

		.version {
			font-size: 0.875rem;
			color: var(--text-muted);
		}
	}

	.ledger {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.hint {
		font-size: 1.05rem;
		font-weight: 600;
	}

	.empty,
	.legal {
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.rows {
		list-style: none;
		margin: 0;
		padding: 0;
		border-top: 1px solid var(--glass-border);
	}

	.row {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 12px 14px 12px 12px;
		border-bottom: 1px solid var(--glass-border);
		border-left: 3px solid var(--glass-border-bright);

		&.sensitive {
			border-left-color: var(--danger);
			background: color-mix(in srgb, var(--danger) 8%, transparent);

			.icon {
				color: var(--danger);
			}
		}

		&.staggered {
			animation: row-in var(--motion-base) var(--ease-out) both;
			animation-delay: calc(var(--row-index) * var(--motion-stagger));
		}
	}

	.icon {
		color: var(--text-muted);
		font-size: 1.25rem;
		margin-top: 1px;
	}

	.text {
		min-width: 0;
	}

	.sites-row {
		display: block;
		padding: 0;
	}

	.sites-toggle {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px 12px 12px;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--text-main);
		font: inherit;
		text-align: left;
		cursor: pointer;

		&:focus-visible {
			outline: 2px solid var(--accent-primary);
			outline-offset: 2px;
		}

		.icon {
			margin-top: 0;
		}
	}

	.sites-title {
		flex: 1;
		min-width: 0;
	}

	.chevron {
		color: var(--text-muted);
		transition: transform var(--motion-quick) var(--ease-out);

		&.turned {
			transform: rotate(180deg);
		}
	}

	.sites-body {
		display: grid;
		grid-template-rows: 0fr;
		transition: grid-template-rows var(--motion-base) var(--ease-out);

		&.open {
			grid-template-rows: 1fr;
		}
	}

	.clip {
		min-height: 0;
		overflow: hidden;
	}

	.sites {
		list-style: none;
		margin: 0;
		padding: 0 14px 12px 49px;
		display: flex;
		flex-direction: column;
		gap: 6px;
		color: var(--text-muted);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.install {
		transition:
			background-color var(--motion-quick) var(--ease-out),
			opacity var(--motion-quick) var(--ease-out),
			transform var(--motion-quick) var(--ease-out);
	}

	@keyframes row-in {
		from {
			opacity: 0;
			transform: translateY(6px);
		}
	}
</style>
