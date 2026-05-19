<script lang="ts">
	import { notifications } from "@stores/notificationStore";
	import SelectionView from "./SelectionView.svelte";
	import ConfigView from "./ConfigView.svelte";
	import ReviewView from "./ReviewView.svelte";
	import Modal from "@components/shared/Modal.svelte";
	import Dropdown from "@components/shared/Dropdown.svelte";
	import { onMount } from "svelte";
	import { fade } from "svelte/transition";
	import { AddGameModalState } from "@lib/AddGameModalState.svelte";

	let { show = false, onClose, onRefresh } = $props<{
		show?: boolean;
		onClose: () => void;
		onRefresh: () => void;
	}>();

	const state = new AddGameModalState();

	onMount(() => {
		state.initialize();
	});

	$effect(() => {
		if (!show) {
			state.resetState();
		}
	});

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === "Escape" && show) onClose();
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<Modal
	{show}
	title={state.addMode === "select"
		? "Add Game"
		: state.addMode === "folder-config"
			? "Search Configuration"
			: "Found Executables"}
	onClose={() => onClose()}
	showDone={false}
	contentClass={state.addMode === "select" ? "selection-modal-style" : ""}
>
	<div class="add-container">
		{#if state.addMode === "select"}
			<div class="prefix-selection-quick" transition:fade>
				<label for="quick-prefix">Target Prefix</label>
				<Dropdown
					options={state.prefixes}
					bind:value={state.selectedPrefix}
					placeholder="Select WINE Prefix"
				/>
			</div>
			<SelectionView
				onAddFile={() => state.handleAddFile(onRefresh, onClose)}
				onAddFolder={() => state.handleAddFolder()}
			/>
		{:else}
			{#if state.addMode === "folder-config"}
				<ConfigView
					selectedFolder={state.selectedFolder}
					bind:searchDepth={state.searchDepth}
					bind:excludeNames={state.excludeNames}
					prefixes={state.prefixes}
					bind:selectedPrefix={state.selectedPrefix}
				/>
			{:else}
				<ReviewView
					foundExecutables={state.foundExecutables}
					discardedExecutables={state.discardedExecutables}
					onToggleDiscard={(path) => state.toggleDiscard(path)}
				/>
			{/if}
		{/if}
	</div>

	<div slot="footer" class="modal-footer-wizard">
		{#if state.addMode === "select"}
			<p class="selection-footer-text">
				Select how you want to add games to your library
			</p>
		{:else}
			{#if state.addMode === "folder-config"}
				<button
					class="secondary-btn"
					onclick={() => (state.addMode = "select")}>Back</button
				>
				<button
					class="primary-btn"
					onclick={() => state.startFolderScan()}
					disabled={state.isSearching}
				>
					{#if state.isSearching}
						<div class="spinner small"></div>
						Scanning...
					{:else}
						Start Search
					{/if}
				</button>
			{:else}
				{#if state.addMode === "folder-review"}
					<button
						class="secondary-btn"
						onclick={() => (state.addMode = "folder-config")}>Back</button
					>
					<button class="primary-btn" onclick={() => state.confirmAddFolder(onRefresh, onClose)}>
						Add {state.foundExecutables.length - state.discardedExecutables.size} Games
					</button>
				{/if}
			{/if}
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
