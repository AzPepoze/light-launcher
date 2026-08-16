<script lang="ts">
	import { fade } from "svelte/transition";
	import NumericInput from "@components/shared/NumericInput.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import BrowseInput from "@components/shared/BrowseInput.svelte";

	export let selectedFolder: string;
	export let searchDepth: string;
	export let excludeNames: string;
	export let prefixes: string[] = [];
	export let selectedPrefix: string = "";
	export let showFolderInput = true;
	export let showPrefixInput = true;
</script>

<div class="config-view" transition:fade={{ duration: 200 }}>
	{#if showFolderInput}
		<BrowseInput
			bind:value={selectedFolder}
			type="folder"
			label="Scanning Target"
			placeholder="Target folder path..."
		/>
	{:else}
		<div class="folder-info">
			<span class="info-label">Folder Path</span>
			<span class="info-value">{selectedFolder}</span>
		</div>
	{/if}

	<div class="settings-grid">
		{#if showPrefixInput}
			<div class="setting-item">
				<label class="item-label" for="prefix-select"
					>Target Prefix</label
				>
				<Dropdown
					options={prefixes}
					bind:value={selectedPrefix}
					placeholder="Select WINE Prefix"
				/>
				<p class="setting-hint">
					Games will be assigned to this prefix.
				</p>
			</div>
		{/if}

		<div class="setting-item">
			<NumericInput
				label="Search Depth"
				bind:value={searchDepth}
				min={1}
				max={999}
			/>
			<p class="setting-hint">
				Use <b>-1</b> for recursive search without limit.
			</p>
		</div>

		<div class="setting-item">
			<div class="exclude-input">
				<label for="exclude-names-settings">Exclusion Rules</label>
				<input
					id="exclude-names-settings"
					type="text"
					bind:value={excludeNames}
					placeholder="e.g. Unity.*, *64, redist"
				/>
			</div>
			<p class="setting-hint">
				Comma-separated regex patterns or wildcards.
			</p>
		</div>
	</div>
</div>

<style lang="scss">
	.config-view {
		display: flex;
		flex-direction: column;
		gap: 20px;

		.settings-grid {
			display: grid;
			grid-template-columns: 1fr;
			gap: 16px;
		}

		.setting-item {
			display: flex;
			flex-direction: column;
			gap: 6px;

			.item-label {
				font-size: 0.75rem;
				font-weight: 600;
				color: var(--text-dim);
				text-transform: uppercase;
				letter-spacing: 0.5px;
			}
		}

		.setting-hint {
			margin: 0;
			font-size: 0.75rem;
			color: var(--text-dim);
			opacity: 0.6;
		}
	}

	.folder-info {
		display: flex;
		flex-direction: column;
		gap: 6px;
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid var(--glass-border);
		border-radius: 8px;
		padding: 12px 16px;

		.info-label {
			font-size: 0.7rem;
			font-weight: 800;
			color: var(--text-dim);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		.info-value {
			font-size: 0.9rem;
			color: var(--text-main);
			word-break: break-all;
			font-weight: 500;
		}
	}

	.exclude-input {
		display: flex;
		flex-direction: column;
		gap: 8px;

		label {
			font-size: 0.75rem;
			font-weight: 600;
			color: var(--text-dim);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		input {
			background: var(--glass-bg);
			border: 1px solid var(--glass-border);
			border-radius: 8px;
			padding: 12px 16px;
			color: var(--text-main);
			font-size: 1rem;
			width: 100%;

			&:focus {
				outline: none;
				border-color: var(--accent-primary);
				box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1);
			}
		}
	}
</style>
