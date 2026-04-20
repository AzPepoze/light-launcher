<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import {
		GetUtilsStatus,
		InstallLsfg,
		UninstallLsfg,
		GetSystemToolsStatus,
		DetectLosslessDll,
	} from "@bindings/light-launcher-wails/backend/app";
	import { notifications } from "@stores/notificationStore";
	import { Events } from "@wailsio/runtime";
	import * as core from "@bindings/light-launcher/pkg/core/models";

	import lsfgPng from "@icons/lsfg.png";

	let status: core.UtilsStatus = { isLsfgInstalled: false, lsfgVersion: "" };
	let systemStatus: core.SystemToolsStatus & { hasLosslessDll: boolean } = {
		hasGamescope: false,
		hasMangoHud: false,
		hasGameMode: false,
		hasLosslessDll: false,
	};
	let isInstalling = false;
	let isUninstalling = false;
	let progressMessage = "";
	let progressPercent = 0;

	async function loadStatus() {
		const [utilStatus, sysTools, dllPath] = await Promise.all([
			GetUtilsStatus(),
			GetSystemToolsStatus(),
			DetectLosslessDll(),
		]);

		status = utilStatus;
		systemStatus = { ...sysTools, hasLosslessDll: !!dllPath };
	}

	onMount(() => {
		loadStatus();
		Events.On("lsfg-install-progress", (data: any) => {
			progressMessage = data.message;
			progressPercent = data.percent;
		});
	});

	onDestroy(() => {
		Events.Off("lsfg-install-progress");
	});

	async function handleInstall() {
		isInstalling = true;
		progressMessage = "Starting installation...";
		progressPercent = 0;
		try {
			await InstallLsfg();
			await loadStatus();
			notifications.add("LSFG-VK installed successfully!", "success");
		} catch (err) {
			notifications.add(`Installation failed: ${err}`, "error");
		} finally {
			isInstalling = false;
			progressMessage = "";
			progressPercent = 0;
		}
	}

	async function handleUninstall() {
		isUninstalling = true;
		try {
			await UninstallLsfg();
			await loadStatus();
			notifications.add("LSFG-VK removed successfully.", "info");
		} catch (err) {
			notifications.add(`Removal failed: ${err}`, "error");
		} finally {
			isUninstalling = false;
		}
	}
</script>

<div class="utils-container">
	<h1 class="page-title">Utilities & Extras</h1>

	<div class="section-container glass">
		<h3>System Dependencies</h3>
		<div class="system-status-grid">
			<div class="status-item" class:ok={systemStatus.hasGamescope}>
				<span class="dot"></span>
				<span class="label">Gamescope</span>
				<span class="value"
					>{systemStatus.hasGamescope
						? "Available"
						: "Not Found"}</span
				>
			</div>
			<div class="status-item" class:ok={systemStatus.hasMangoHud}>
				<span class="dot"></span>
				<span class="label">MangoHud</span>
				<span class="value"
					>{systemStatus.hasMangoHud
						? "Available"
						: "Not Found"}</span
				>
			</div>
			<div class="status-item" class:ok={systemStatus.hasGameMode}>
				<span class="dot"></span>
				<span class="label">GameMode</span>
				<span class="value"
					>{systemStatus.hasGameMode
						? "Available"
						: "Not Found"}</span
				>
			</div>
			<div class="status-item" class:ok={systemStatus.hasLosslessDll}>
				<span class="dot"></span>
				<span class="label">Lossless.dll</span>
				<span class="value"
					>{systemStatus.hasLosslessDll
						? "Found"
						: "Not Found"}</span
				>
			</div>
		</div>
		{#if !systemStatus.hasGamescope || !systemStatus.hasMangoHud || !systemStatus.hasGameMode || !systemStatus.hasLosslessDll}
			<p class="warning-text">
				Some features may not work until you install these tools.
				LSFG-VK requires Lossless Scaling installed via Steam.
			</p>
		{/if}
	</div>

	<div class="utils-grid">
		<!-- LSFG-VK Card -->
		<div class="util-card glass">
			<div class="util-header">
				<div class="icon-bg">
					<img src={lsfgPng} alt="lsfg" class="lsfg-logo" />
				</div>
				<div class="title-area">
					<h3>LSFG-VK</h3>
					<span
						class="badge"
						class:installed={status.isLsfgInstalled}
					>
						{status.isLsfgInstalled
							? "Installed"
							: "Not Installed"}
					</span>
				</div>
			</div>

			<div class="description">
				<p>
					Lossless Scaling is a Windows-exclusive program
					featuring various algorithms for scaling and
					interpolating programs.
				</p>
				<p>
					<strong>lsfg-vk</strong> is a Vulkan layer that hooks into
					Vulkan applications and generates additional frames using
					Lossless Scaling's frame generation algorithm.
				</p>
				<p class="note">
					Note: Requires Lossless Scaling downloaded on Steam.
				</p>
			</div>

			<div class="action-area">
				{#if status.isLsfgInstalled}
					<button
						class="btn danger"
						on:click={handleUninstall}
						disabled={isUninstalling}
					>
						{isUninstalling
							? "Removing..."
							: "Remove Utility"}
					</button>
				{:else}
					<div class="install-controls">
						<button
							class="btn primary"
							on:click={handleInstall}
							disabled={isInstalling}
						>
							{isInstalling
								? "Installing..."
								: "Install LSFG-VK"}
						</button>
						{#if isInstalling}
							<div class="install-progress-area">
								<div class="progress-header">
									<span class="msg"
										>{progressMessage}</span
									>
									<span class="pct"
										>{progressPercent}%</span
									>
								</div>
								<div class="progress-bar-container">
									<div
										class="progress-fill"
										style="width: {progressPercent}%"
									></div>
								</div>
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>

<style lang="scss">
	.utils-container {
		padding: 32px;
		height: 100%;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
	}

	.page-title {
		font-size: 2rem;
		font-weight: 800;
		color: var(--text-main);
		margin: 0 0 32px 0;
	}

	.section-container {
		padding: 20px 24px;
		border-radius: 16px;
		border: 1px solid var(--glass-border);
		background: var(--glass-surface);
		margin-bottom: 32px;

		h3 {
			margin: 0 0 16px 0;
			font-size: 1rem;
			color: var(--text-dim);
			text-transform: uppercase;
			letter-spacing: 1px;
		}
	}

	.system-status-grid {
		display: flex;
		gap: 24px;
		flex-wrap: wrap;
	}

	.status-item {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 0.9rem;
		padding: 8px 16px;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border);
		border-radius: 10px;
		color: var(--text-muted);

		.dot {
			width: 8px;
			height: 8px;
			border-radius: 50%;
			background: #ef4444;
			box-shadow: 0 0 8px #ef4444;
		}
		.label {
			font-weight: 600;
			color: var(--text-main);
		}
		.value {
			font-size: 0.8rem;
			opacity: 0.8;
		}

		&.ok {
			.dot {
				background: #10b981;
				box-shadow: 0 0 8px #10b981;
			}
		}
	}

	.warning-text {
		font-size: 0.8rem;
		color: var(--accent-secondary);
		margin: 12px 0 0 0;
		font-style: italic;
	}

	.utils-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
		gap: 24px;
	}

	.util-card {
		padding: 24px;
		border-radius: 20px;
		border: 1px solid var(--glass-border);
		background: var(--glass-surface);
		display: flex;
		flex-direction: column;
		gap: 20px;
		transition: all 0.2s;

		&:hover {
			background: var(--glass-border-bright);
			border-color: var(--glass-border-bright);
		}
	}

	.util-header {
		display: flex;
		align-items: center;
		gap: 20px;

		.icon-bg {
			width: 56px;
			height: 56px;
			background: var(--glass-bg);
			border: 1px solid var(--glass-border);
			border-radius: 14px;
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 1.75rem;

			.lsfg-logo {
				width: 32px;
				height: 32px;
				color: var(--accent-primary);
				opacity: 0.9;
				object-fit: contain;
			}
		}

		.title-area {
			display: flex;
			flex-direction: column;
			gap: 4px;

			h3 {
				margin: 0;
				font-size: 1.25rem;
				font-weight: 700;
			}
		}
	}

	.badge {
		display: inline-block;
		padding: 2px 10px;
		border-radius: 20px;
		font-size: 0.7rem;
		font-weight: 700;
		text-transform: uppercase;
		background: var(--glass-high-alpha);
		color: var(--text-muted);

		&.installed {
			background: rgba(16, 185, 129, 0.2);
			color: #10b981;
		}
	}

	.description {
		font-size: 0.9rem;
		color: var(--text-muted);
		line-height: 1.5;
		margin: 0;
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 12px;

		p {
			margin: 0;
		}
		strong {
			color: var(--text-main);
		}
		.note {
			font-size: 0.8rem;
			color: var(--accent-secondary);
			font-style: italic;
			padding-top: 8px;
			border-top: 1px solid var(--glass-border);
		}
	}

	.install-controls {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.install-progress-area {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 16px;
		background: var(--glass-bg);
		border-radius: 12px;
		border: 1px solid var(--glass-border);
	}

	.progress-header {
		display: flex;
		justify-content: space-between;
		font-size: 0.85rem;
		font-weight: 600;

		.msg {
			color: var(--text-muted);
		}
		.pct {
			color: var(--accent-primary);
		}
	}

	.progress-bar-container {
		height: 8px;
		background: var(--glass-low-alpha);
		border-radius: 10px;
		overflow: hidden;
		border: 1px solid var(--glass-border);
	}

	.progress-fill {
		height: 100%;
		background: var(--accent-primary);
		transition: width 0.3s ease;
		box-shadow: 0 0 10px var(--glass-border-bright);
	}

	.action-area {
		margin-top: 8px;

		button {
			width: 100%;
			padding: 12px;
			font-weight: 600;
			letter-spacing: 0.5px;
		}
	}
</style>
