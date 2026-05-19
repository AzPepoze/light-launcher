<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import {
		DetectLosslessDll,
		GetSystemToolsStatus,
		GetUtilsStatus,
		InstallLsfg,
		UninstallLsfg,
	} from "@bindings/light-launcher/internal/app/app";
	import * as core from "@bindings/light-launcher/internal/types/models";
	import { notifications } from "@stores/notificationStore";
	import { Events } from "@wailsio/runtime";
	import { onDestroy, onMount } from "svelte";

	import lsfgPng from "@icons/lsfg.png";
	import SystemDependencies from "@components/utils/SystemDependencies.svelte";
	import LsfgCard from "@components/utils/LsfgCard.svelte";

	let status: core.UtilsStatus = { isLsfgInstalled: false, lsfgVersion: "" };
	let systemStatus: core.SystemToolsStatus & { hasLosslessDll: boolean } = {
		hasGamescope: false,
		hasMangoHud: false,
		hasGameMode: false,
		hasVulkanInfo: false,
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
	<PageHeader title="Utilities & Extras" icon="handyman" />

	<SystemDependencies {systemStatus} />

	<div class="utils-grid">
		<LsfgCard
			{status}
			{isInstalling}
			{isUninstalling}
			{progressMessage}
			{progressPercent}
			handleInstall={handleInstall}
			handleUninstall={handleUninstall}
			{lsfgPng}
		/>
	</div>
</div>

<style lang="scss">
	.utils-container {
		padding: 40px 48px;
		height: 100%;
		display: flex;
		flex-direction: column;
		overflow-y: auto;

		&::-webkit-scrollbar {
			width: 8px;
		}
		&::-webkit-scrollbar-track {
			background: transparent;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 102, 171, 0.2);
			border-radius: 10px;
		}
	}

	.utils-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 400px), 520px));
		gap: 24px;
		width: 100%;
		margin-top: 8px;
	}
</style>
