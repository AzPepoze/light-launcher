<script lang="ts">
	import {
		PickFile,
		PickFolder,
		ListPrefixes,
		GetPrefixBaseDir,
	} from "@bindings/light-launcher/internal/app/app";
	import { notifications } from "@stores/notificationStore";
	import { loadExeIcon } from "@lib/iconService";
	import * as service from "@lib/gameService";
	import SelectionView from "./SelectionView.svelte";
	import ConfigView from "./ConfigView.svelte";
	import ReviewView from "./ReviewView.svelte";
	import Modal from "@components/shared/Modal.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import { onMount } from "svelte";
	import { fade } from "svelte/transition";

	export let show: boolean = false;
	export let onClose: () => void;
	export let onRefresh: () => void;

	let addMode: "select" | "folder-config" | "folder-review" = "select";
	let searchDepth = "2";
	let excludeNames = "UnityCrashHandler64, uninstall, redist";
	let selectedFolder = "";
	let foundExecutables: service.ScannedExecutable[] = [];
	let discardedExecutables = new Set<string>();
	let isSearching = false;

	let prefixes: string[] = [];
	let selectedPrefix = "default";
	let prefixBaseDir = "";

	async function loadPrefixes() {
		try {
			const list = await ListPrefixes();
			prefixes = list || [];
			prefixBaseDir = await GetPrefixBaseDir();

			// If "Default" or "default" is in the list, select it
			if (prefixes.includes("Default")) {
				selectedPrefix = "Default";
			} else if (prefixes.includes("default")) {
				selectedPrefix = "default";
			} else if (prefixes.length > 0) {
				selectedPrefix = prefixes[0];
			}
		} catch (err) {
			console.error("Failed to load prefixes:", err);
		}
	}

	onMount(() => {
		loadPrefixes();
	});

	function resetState() {
		addMode = "select";
		selectedFolder = "";
		foundExecutables = [];
		discardedExecutables = new Set();
		searchDepth = "2";
		excludeNames = "UnityCrashHandler64, uninstall, redist";
		isSearching = false;
		loadPrefixes();
	}

	$: if (!show) resetState();

	async function handleAddFile() {
		try {
			const path = await PickFile();
			if (path) {
				const name = path.split("/").pop()?.replace(".exe", "") || "Game";
				await service.registerGame(path, `${prefixBaseDir}/${selectedPrefix}`);
				notifications.add(`Added ${name}`, "success");
				onRefresh();
				onClose();
			}
		} catch (err) {
			notifications.add(`Failed to add game: ${err}`, "error");
		}
	}

	async function handleAddFolder() {
		try {
			const folder = await PickFolder();
			if (folder) {
				selectedFolder = folder;
				addMode = "folder-config";
			}
		} catch (err) {
			notifications.add(`Failed to select folder: ${err}`, "error");
		}
	}

	async function startFolderScan() {
		if (!selectedFolder) return;
		isSearching = true;
		try {
			const depth = parseInt(searchDepth) || 2;
			const excludes = excludeNames.split(",").map((e) => e.trim()).filter(Boolean);

			const results = await service.scanFolderForExecutables(selectedFolder, depth, excludes);
			
			if (results.length > 0) {
				foundExecutables = results;
				addMode = "folder-review";
				discardedExecutables = new Set();
				
				foundExecutables.forEach((item, index) => {
					loadExeIcon(item.path).then((icon) => {
						if (icon) {
							foundExecutables[index].icon = icon;
							foundExecutables = [...foundExecutables];
						}
					});
				});
			} else {
				notifications.add("No executables found in folder", "info");
				addMode = "select";
			}
		} finally {
			isSearching = false;
		}
	}

	function toggleDiscard(path: string) {
		if (discardedExecutables.has(path)) discardedExecutables.delete(path);
		else discardedExecutables.add(path);
		discardedExecutables = discardedExecutables;
	}

	async function confirmAddFolder() {
		const targetPrefixPath = `${prefixBaseDir}/${selectedPrefix}`;
		const addedCount = await service.batchRegisterGames(foundExecutables, discardedExecutables, targetPrefixPath);
		
		if (addedCount > 0) {
			onRefresh();
			onClose();
		} else if (foundExecutables.length > 0) {
			onClose();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Escape" && show) onClose();
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<Modal
	{show}
	title={addMode === "select"
		? "Add Game"
		: addMode === "folder-config"
			? "Search Configuration"
			: "Found Executables"}
	onClose={() => onClose()}
	showDone={false}
	contentClass={addMode === "select" ? "selection-modal-style" : ""}
>
	<div class="add-container">
		{#if addMode === "select"}
			<div class="prefix-selection-quick" transition:fade>
				<label for="quick-prefix">Target Prefix</label>
				<Dropdown
					options={prefixes}
					bind:value={selectedPrefix}
					placeholder="Select WINE Prefix"
				/>
			</div>
			<SelectionView
				onAddFile={handleAddFile}
				onAddFolder={handleAddFolder}
			/>
		{:else if addMode === "folder-config"}
			<ConfigView
				{selectedFolder}
				bind:searchDepth
				bind:excludeNames
				{prefixes}
				bind:selectedPrefix
			/>
		{:else}
			<ReviewView
				{foundExecutables}
				{discardedExecutables}
				onToggleDiscard={toggleDiscard}
			/>
		{/if}
	</div>

	<div slot="footer" class="modal-footer-wizard">
		{#if addMode === "select"}
			<p class="selection-footer-text">
				Select how you want to add games to your library
			</p>
		{:else if addMode === "folder-config"}
			<button
				class="secondary-btn"
				on:click={() => (addMode = "select")}>Back</button
			>
			<button
				class="primary-btn"
				on:click={startFolderScan}
				disabled={isSearching}
			>
				{#if isSearching}
					<div class="spinner small"></div>
					Scanning...
				{:else}
					Start Search
				{/if}
			</button>
		{:else if addMode === "folder-review"}
			<button
				class="secondary-btn"
				on:click={() => (addMode = "folder-config")}>Back</button
			>
			<button class="primary-btn" on:click={confirmAddFolder}>
				Add {foundExecutables.length - discardedExecutables.size} Games
			</button>
		{/if}
	</div>
</Modal>

<style lang="scss">
	:global(.selection-modal-style) {
		max-width: 620px !important;
		background: var(--glass-bg) !important;
		border-radius: 30px !important;
		border: 1px solid var(--glass-border) !important;
		box-shadow: var(--glass-shadow) !important;

		&::before {
			content: "";
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			height: 1px;
			background: linear-gradient(
				90deg,
				transparent,
				var(--glass-border-bright),
				transparent
			);
		}

		:global(.modal-header) {
			padding: 40px 40px 15px !important;
			border-bottom: none !important;

			:global(h3) {
				font-size: 1.6rem !important;
				font-weight: 900 !important;
				letter-spacing: -1px !important;
				text-transform: uppercase !important;
				color: var(--text-main) !important;
				background: none !important;
				-webkit-text-fill-color: initial !important;
			}
		}

		:global(.modal-body) {
			padding: 24px 40px !important;
		}

		:global(.modal-footer) {
			border-top: none !important;
			padding: 0 40px 30px !important;
			justify-content: center !important;
		}
	}

	.selection-footer-text {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-dim);
		opacity: 0.5;
		font-weight: 600;
		text-align: center;
	}

	.prefix-selection-quick {
		margin-bottom: 24px;
		display: flex;
		flex-direction: column;
		gap: 8px;

		label {
			font-size: 0.75rem;
			font-weight: 800;
			color: var(--text-dim);
			text-transform: uppercase;
			letter-spacing: 1px;
		}
	}

	.modal-footer-wizard {
		display: flex;
		justify-content: flex-end;
		align-items: center;
		gap: 12px;
		width: 100%;
	}

	.add-container {
		display: flex;
		flex-direction: column;
	}

	.primary-btn {
		background: var(--accent-primary);
		color: var(--glass-bg);
		border: none;
		border-radius: 12px;
		padding: 12px 28px;
		font-weight: 900;
		text-transform: uppercase;
		letter-spacing: 1px;
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
		display: flex;
		align-items: center;
		gap: 10px;
		box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);

		&:hover {
			transform: scale(1.05) translateY(-2px);
			box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
			opacity: 0.9;
		}

		&:active {
			transform: scale(0.98);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
			transform: none;
			box-shadow: none;
		}
	}

	.secondary-btn {
		background: var(--glass-surface);
		color: var(--text-main);
		border: 1px solid var(--glass-border);
		border-radius: 12px;
		padding: 12px 24px;
		font-weight: 800;
		cursor: pointer;
		transition: all 0.2s;
		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--glass-border-bright);
		}
	}

	.spinner {
		width: 18px;
		height: 18px;
		border: 2px solid var(--glass-border);
		border-top-color: var(--accent-primary);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
