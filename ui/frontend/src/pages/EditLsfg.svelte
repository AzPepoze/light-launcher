<script lang="ts">
	import Modal from "@components/shared/Modal.svelte";
	import LsfgConfigForm from "@components/editlsfg/LsfgConfigForm.svelte";
	import {
		GetLsfgProfileForGame,
		GetInitialGamePath,
		SaveLsfgProfile,
		DisableLsfgProfile,
		CloseWindow,
		PickFileCustom,
	} from "@bindings/light-launcher/internal/app/app";
	import * as core from "@bindings/light-launcher/internal/types/models";
	import { onMount } from "svelte";
	import { loadLsfgResources, createLaunchOptions } from "@lib/formService";
	import { notifications } from "@stores/notificationStore";

	export let gamePath = "";

	let options: core.LaunchOptions = createLaunchOptions();

	let loading = true;
	let error = "";
	let gpuList: string[] = [];
	let saving = false;
	let disabling = false;

	onMount(async () => {
		try {
			const currentGamePath = gamePath || (await GetInitialGamePath());

			if (!currentGamePath) {
				error = "No game path provided";
				loading = false;
				return;
			}

			options.GamePath = currentGamePath;
			options.Name = currentGamePath.split(/[/\\]/).pop()?.replace(/\.exe$/i, "") || "Game";

			// Load profile data
			const data = await GetLsfgProfileForGame(options.Name, currentGamePath);
			if (data) {
				if (data.name) options.Name = data.name;
				options.Extras.Lsfg.Multiplier = String(data.multiplier || 2);
				options.Extras.Lsfg.PerfMode = data.performanceMode || false;
				options.Extras.Lsfg.Gpu = data.gpu || "";
				options.Extras.Lsfg.FlowScale = String(data.flowScale || 0.8);
				options.Extras.Lsfg.Pacing = data.pacing || "none";
				options.Extras.Lsfg.DllPath = data.dllPath || "";
				options.Extras.Lsfg.AllowFp16 = data.allowFp16 || false;
			}

			// Auto-detect DLL and load GPUs
			const { gpus, dll } = await loadLsfgResources();
			if (dll && !options.Extras.Lsfg.DllPath) {
				options.Extras.Lsfg.DllPath = dll;
				console.log("Auto-detected DLL:", dll);
			}
			if (gpus && gpus.length > 0) {
				gpuList = gpus;
			}

			loading = false;
		} catch (err) {
			error = String(err);
			loading = false;
		}
	});

	async function handleApply() {
		saving = true;

		try {
			await SaveLsfgProfile(
				options.Name,
				options.GamePath,
				parseInt(options.Extras.Lsfg.Multiplier) || 2,
				options.Extras.Lsfg.PerfMode,
				options.Extras.Lsfg.DllPath,
				options.Extras.Lsfg.Gpu,
				options.Extras.Lsfg.FlowScale,
				options.Extras.Lsfg.Pacing,
				options.Extras.Lsfg.AllowFp16,
			);
			notifications.success("Configuration saved successfully!");
		} catch (err) {
			notifications.error(`Failed to save configuration: ${err}`);
		} finally {
			saving = false;
		}
	}

	async function handleDisable() {
		disabling = true;
		try {
			await DisableLsfgProfile(options.Name, options.GamePath);
			notifications.success("LSFG profile disabled in global config.");
		} catch (err) {
			notifications.error(`Failed to disable profile: ${err}`);
		} finally {
			disabling = false;
		}
	}

	async function handleBrowseDll() {
		try {
			const path = await PickFileCustom("Select Lossless.dll", [
				{ DisplayName: "Lossless.dll", Pattern: "Lossless.dll" },
			]);
			if (path) options.Extras.Lsfg.DllPath = path;
		} catch (err) {
			console.error(err);
		}
	}

	function handleClose() {
		CloseWindow();
	}
</script>

{#if loading}
	<Modal
		show={true}
		title="LSFG-VK Configuration"
		onClose={handleClose}
		fullscreen={true}
		showDone={false}
	>
		<div class="loading-container">
			<p>Loading LSFG configuration...</p>
		</div>
	</Modal>
{:else if error}
	<Modal
		show={true}
		title="LSFG-VK Configuration"
		onClose={handleClose}
		fullscreen={true}
		showDone={false}
	>
		<div class="error-container">
			<p>Error loading profile: {error}</p>
		</div>
		<div slot="footer">
			<button class="btn secondary" on:click={handleClose}
				>Close</button
			>
		</div>
	</Modal>
{:else}
	<Modal
		show={true}
		title="LSFG-VK Configuration"
		onClose={handleClose}
		fullscreen={true}
		showDone={false}
	>
		<div class="modal-content">
			<div class="profile-header">
				<h2 class="profile-title">{options.Name}</h2>
				<p class="game-path">{options.GamePath}</p>
			</div>

			<LsfgConfigForm
				{options}
				{gpuList}
				onDllBrowse={handleBrowseDll}
			/>
		</div>

		<div slot="footer" class="actions">
			<button class="btn secondary" on:click={handleClose}
				>Close</button
			>
			<button 
				class="btn danger" 
				on:click={handleDisable}
				disabled={disabling || saving}
			>
				{disabling ? "Disabling..." : "Disable Profile"}
			</button>
			<button
				class="btn primary"
				on:click={handleApply}
				disabled={saving || disabling}
			>
				{saving ? "Saving..." : "Apply"}
			</button>
		</div>
	</Modal>
{/if}

<style lang="scss">
	.loading-container,
	.error-container {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		text-align: center;

		p {
			color: var(--text-muted);
			font-size: 1.1rem;
			font-weight: 700;
		}
	}

	.error-container p {
		color: var(--danger);
	}

	.modal-content {
		margin: 0 auto;
	}

	.profile-header {
		margin-bottom: 32px;
		padding-bottom: 20px;
		border-bottom: 2px solid rgba(255, 255, 255, 0.05);

		.profile-title {
			margin: 0 0 8px 0;
			font-size: 1.85rem;
			font-weight: 800;
			color: var(--text-main);
			text-transform: uppercase;
			letter-spacing: 1px;
		}

		.game-path {
			margin: 0;
			color: var(--text-muted);
			font-size: 0.9rem;
			word-break: break-all;
			font-family: monospace;
			font-weight: 600;
		}
	}

	.actions {
		display: flex;
		gap: 12px;
	}

	.btn {
		padding: 12px 24px;
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-pill);
		font-weight: 800;
		cursor: pointer;
		transition: transform var(--transition-spring), background var(--transition-fast), border-color var(--transition-fast), color var(--transition-fast);
		font-size: 0.9rem;
		background: var(--bg-surface);
		color: var(--text-main);

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}

		&.primary {
			background: var(--accent-primary);
			border: none;
			color: #ffffff;
			box-shadow: 0 4px 12px var(--accent-glow);

			&:hover:not(:disabled) {
				background: var(--accent-hover);
				transform: scale(1.05);
			}

			&:active:not(:disabled) {
				transform: scale(0.95);
			}
		}

		&.secondary {
			background: var(--bg-surface);
			color: var(--text-muted);

			&:hover:not(:disabled) {
				background: var(--bg-elevated);
				color: var(--text-main);
				border-color: var(--accent-secondary);
				transform: scale(1.05);
			}

			&:active:not(:disabled) {
				transform: scale(0.95);
			}
		}

		&.danger {
			background: var(--danger);
			border: none;
			color: #ffffff;

			&:hover:not(:disabled) {
				background: #ff2e63;
				transform: scale(1.05);
			}

			&:active:not(:disabled) {
				transform: scale(0.95);
			}
		}
	}
</style>
