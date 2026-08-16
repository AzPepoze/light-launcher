<script lang="ts">
	import { PickFolder, PickFile } from "@lib/api";

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
				console.error("Error executing custom browse handler:", err);
			}
			return;
		}
		try {
			const selected = type === "folder" ? await PickFolder() : await PickFile();
			if (selected) {
				value = selected;
			}
		} catch (err) {
			console.error(`Error picking ${type}:`, err);
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
		<button class="btn browse-btn" type="button" on:click={handleBrowse}>
			<span class="material-icons btn-icon">
				{type === "folder" ? "folder_open" : "insert_drive_file"}
			</span>
			Browse
		</button>
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
		align-items: center;
		gap: 12px;
		width: 100%;
	}

	.path-input {
		flex: 1;
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

	.browse-btn {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 12px 24px;
		border-radius: var(--radius-pill);
		font-weight: 700;
		font-size: 0.9rem;
		cursor: pointer;
		transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
		border: 2px solid var(--glass-border);
		background: var(--bg-surface);
		color: var(--text-main);
		white-space: nowrap;

		&:hover {
			background: var(--bg-elevated);
			border-color: var(--accent-secondary);
			transform: scale(1.05);
		}

		&:active {
			transform: scale(0.95);
		}

		.btn-icon {
			font-size: 18px;
		}
	}
</style>
