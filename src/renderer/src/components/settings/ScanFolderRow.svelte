<script lang="ts">
	import { UpdateScanFolderConfig } from "@lib/api";
	import {
		depthError,
		excludeNamesText,
		gamesFoundText,
		normalizeExcludeNames,
		validateDepth
	} from "@lib/features/scanFolders";
	import { createLogger } from "@lib/logger";

	const log = createLogger("ScanFolderRow");

	let {
		folder,
		depth,
		excludeNames,
		gamesFound,
		rescanNote,
		rescanning,
		onRescan,
		onRemove,
		onSaved
	} = $props<{
		folder: string;
		depth: number;
		excludeNames: string[];
		gamesFound: number | null;
		rescanNote: string;
		rescanning: boolean;
		onRescan: () => void;
		onRemove: () => void;
		onSaved: () => Promise<void>;
	}>();

	// null means the user has not edited the field, so the saved value is shown.
	let depthDraft = $state<string | null>(null);
	let excludeDraft = $state<string | null>(null);
	let saving = $state(false);
	let saveError = $state("");

	const depthText = $derived(depthDraft ?? String(depth));
	const depthCheck = $derived(validateDepth(depthText));
	const depthMessage = $derived(depthError(depthText));
	const excludeText = $derived(excludeDraft ?? excludeNamesText(excludeNames));
	const dirty = $derived(depthDraft !== null || excludeDraft !== null);
	const countText = $derived(gamesFound === null ? "Not scanned yet" : gamesFoundText(gamesFound));

	async function save() {
		if (!depthCheck.ok) return;
		saving = true;
		saveError = "";
		try {
			const names = normalizeExcludeNames(excludeText);
			await UpdateScanFolderConfig(folder, depthCheck.depth, names);
			depthDraft = null;
			excludeDraft = null;
			await onSaved();
		} catch (err) {
			log.error("Failed to save folder settings", err);
			saveError = "Could not save these settings. Try again.";
		} finally {
			saving = false;
		}
	}
</script>

<li class="folder-row">
	<div class="folder-head">
		<span class="material-icons folder-icon" aria-hidden="true">folder</span>
		<div class="folder-text">
			<span class="folder-path">{folder}</span>
			<span class="games-count">{countText}</span>
		</div>
		<button class="btn secondary" onclick={onRescan} disabled={rescanning}>
			{rescanning ? "Scanning..." : "Rescan"}
		</button>
		<button class="btn danger" onclick={onRemove}>Remove</button>
	</div>

	<p class="rescan-note" role="status">{rescanNote}</p>

	<div class="folder-fields">
		<div class="field">
			<label for="depth-{folder}">Scan depth</label>
			<input
				id="depth-{folder}"
				type="number"
				step="1"
				value={depthText}
				oninput={(event) => (depthDraft = event.currentTarget.value)}
				aria-invalid={!depthCheck.ok}
				aria-describedby="depth-hint-{folder}"
			/>
			{#if depthMessage}
				<p class="error" role="alert">{depthMessage}</p>
			{/if}
			<p class="hint" id="depth-hint-{folder}">
				How many folder levels to search. Use -1 for no limit.
			</p>
		</div>

		<div class="field">
			<label for="ignore-{folder}">Ignore names</label>
			<input
				id="ignore-{folder}"
				type="text"
				value={excludeText}
				placeholder="e.g. Unity.*, *64, redist"
				oninput={(event) => (excludeDraft = event.currentTarget.value)}
			/>
			<p class="hint">Comma-separated regex patterns or wildcards.</p>
		</div>
	</div>

	{#if saveError}
		<p class="error" role="alert">{saveError}</p>
	{/if}

	<div class="folder-actions">
		<button class="btn primary" onclick={save} disabled={!dirty || saving || !depthCheck.ok}>
			{saving ? "Saving..." : "Save settings"}
		</button>
	</div>
</li>

<style lang="scss">
	.folder-row {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 16px;
		background: var(--bg-base);
		border: 1px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-md);
		list-style: none;
	}

	.folder-head {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
	}

	.folder-icon {
		color: var(--accent-secondary);
		font-size: 22px;
	}

	.folder-text {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 1;
		min-width: 0;
	}

	.folder-path {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-main);
		overflow-wrap: anywhere;
	}

	.games-count {
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.rescan-note {
		margin: 0;
		min-height: 0;
		font-size: 0.875rem;
		color: var(--text-muted);
	}

	.folder-fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 16px;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;

		label {
			font-size: 0.875rem;
			font-weight: 600;
			color: var(--text-main);
		}

		input {
			font: inherit;
			font-size: 1rem;
			color: var(--text-main);
			background: var(--bg-input);
			border: 1px solid rgba(255, 255, 255, 0.1);
			border-radius: var(--radius-md);
			padding: 8px 12px;

			&[aria-invalid="true"] {
				border-color: var(--danger);
			}

			&:focus-visible {
				outline: 2px solid var(--accent-primary);
				outline-offset: 2px;
			}
		}

		.hint {
			margin: 0;
			font-size: 0.875rem;
			color: var(--text-muted);
		}
	}

	.error {
		margin: 0;
		font-size: 0.875rem;
		color: var(--danger);
	}

	.folder-actions {
		display: flex;
		justify-content: flex-end;
	}

	.btn:focus-visible {
		outline: 2px solid var(--accent-primary);
		outline-offset: 2px;
	}
</style>
