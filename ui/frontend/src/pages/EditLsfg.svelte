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
			<div class="profile-info">
				<div class="form-group profile-name-group">
					<label for="profileName">Profile Name</label>
					<input
						id="profileName"
						type="text"
						class="input profile-input"
						bind:value={options.Name}
						placeholder="Enter profile name..."
					/>
				</div>
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
		}
	}

	.error-container p {
		color: #ef4444;
	}

	.modal-content {
		margin: 0 auto;
	}

	.profile-info {
		margin-bottom: 16px;

		.game-path {
			margin: 0;
			color: var(--text-dim);
			font-size: 0.9rem;
			word-break: break-all;
		}
	}

	.actions {
		display: flex;
		gap: 12px;
	}

	.btn {
		padding: 10px 20px;
		border: 1px solid var(--glass-border);
		border-radius: 10px;
		font-weight: 800;
		cursor: pointer;
		transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
		font-size: 0.9rem;
		background: var(--glass-hover);
		color: var(--text-main);

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}

		&.primary {
			background: var(--accent-primary);
			border: 1px solid var(--accent-primary);
			color: var(--glass-bg);

			&:hover:not(:disabled) {
				filter: brightness(1.2);
			}
		}

		&.secondary {
			background: var(--glass-hover);
			color: var(--text-muted);

			&:hover:not(:disabled) {
				background: var(--glass-border);
				color: var(--text-main);
				border-color: var(--glass-border-bright);
			}
		}
	}
</style>
