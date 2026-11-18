<script lang="ts">
	import { PickFolder, PickFile } from "@lib/api";
	import { createLogger } from "@lib/logger";
	import Button from "./Button.svelte";

	const log = createLogger("BrowseInput");

	export let value = "";
	export let placeholder = "Select path...";
	export let type: "folder" | "file" = "folder";
	export let label = "";
	export let id = "";
	export let focusBorderColor: string = "var(--accent-primary)";
	export let browseHandler: (() => Promise<void>) | null = null;

	async function handleBrowse() {
		if (browseHandler) {
			try {
				await browseHandler();
			} catch (err) {
				log.error("Error executing custom browse handler", err);
			}
			return;
		}
		try {
			const selected = type === "folder" ? await PickFolder() : await PickFile();
			if (selected) {
				value = selected;
			}
		} catch (err) {
			log.error(`Error picking ${type}`, err);
		}
	}
</script>

<div class="browse-input-container">
	{#if label}
		<label class="field-label" for={id}>{label}</label>
	{/if}
	<div class="input-row">
		<input
			{id}
			type="text"
			class="input path-input"
			{placeholder}
			bind:value
			style="--focus-border-color: {focusBorderColor}"
		/>
		<Button
			class="browse-btn"
			icon={type === "folder" ? "folder_open" : "insert_drive_file"}
			onclick={handleBrowse}
		>
			Browse
		</Button>
	</div>
</div>

<style lang="scss">
	.browse-input-container {
		display: flex;
		flex-direction: column;
		gap: 8px;
		width: 100%;
	}

	.field-label {
		display: block;
		font-size: 0.75rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 1px;
		color: var(--accent-primary);
		margin-bottom: 2px;
	}

	.input-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 12px;
		width: 100%;
	}

	.path-input {
		flex: 1 1 160px;
		min-width: 0;
		padding: 12px 16px;
		background: var(--bg-input);
		border: 2px solid var(--glass-border);
		border-radius: var(--radius-md);
		color: var(--text-main);
		font-size: 0.95rem;
		font-weight: 600;
		outline: none;
		transition: border-color var(--transition-fast), transform var(--transition-fast);

		&:focus {
			border-color: var(--focus-border-color, var(--accent-primary));
			transform: scale(1.005);
		}
	}

	.input-row :global(.browse-btn) {
		flex-shrink: 0;
	}
</style>
