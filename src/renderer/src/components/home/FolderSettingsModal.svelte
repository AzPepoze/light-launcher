<script lang="ts">
	import { notifications } from "@stores/notificationStore";
	import Modal from "@components/shared/Modal.svelte";
	import FolderConfiguration from "@components/shared/FolderConfiguration.svelte";
	import { GetScanFolderConfig, UpdateScanFolderConfig } from "@bindings/light-launcher/internal/app/app";

	export let show = false;
	export let folderPath = "";
	export let onClose: () => void;
	export let onSave: () => void;

	let searchDepth = "2";
	let excludeNames = "";
	let isLoading = false;

	$: if (show && folderPath) {
		loadFolderConfig();
	}

	async function loadFolderConfig() {
		isLoading = true;
		try {
			const config = await GetScanFolderConfig(folderPath);
			if (config) {
				searchDepth = config.Depth.toString();
				excludeNames = config.ExcludeNames ? config.ExcludeNames.join(", ") : "";
			}
		} catch (error) {
			console.error("Failed to load folder config:", error);
			notifications.add(`Failed to load folder config: ${error}`, "error");
		} finally {
			isLoading = false;
		}
	}

	async function handleSave() {
		try {
			const depth = parseInt(searchDepth);
			const excludes = excludeNames
				.split(",")
				.map((item) => item.trim())
				.filter((item) => item.length > 0);

			await UpdateScanFolderConfig(folderPath, depth, excludes);
			notifications.add("Folder settings saved successfully", "success");
			onSave();
			onClose();
		} catch (error) {
			console.error("Failed to save folder config:", error);
			notifications.add(`Failed to save folder config: ${error}`, "error");
		}
	}
</script>

<Modal
	{show}
	title="Folder Settings"
	onClose={onClose}
	showDone={false}
>
	{#if isLoading}
		<div class="loading-state">
			<div class="spinner"></div>
			<span>Loading settings...</span>
		</div>
	{:else}
		<FolderConfiguration
			selectedFolder={folderPath}
			bind:searchDepth
			bind:excludeNames
			showFolderInput={false}
			showPrefixInput={false}
		/>
	{/if}

	<div slot="footer" class="modal-footer-buttons">
		<button class="btn secondary" on:click={onClose}>Cancel</button>
		<button class="btn primary" on:click={handleSave} disabled={isLoading}>Save Settings</button>
	</div>
</Modal>

<style lang="scss">
	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px;
		gap: 16px;
		color: var(--text-dim);

		.spinner {
			width: 24px;
			height: 24px;
			border: 2px solid var(--glass-border);
			border-top-color: var(--accent-primary);
			border-radius: 50%;
			animation: spin 0.8s linear infinite;
		}
	}

	.modal-footer-buttons {
		display: flex;
		gap: 12px;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
