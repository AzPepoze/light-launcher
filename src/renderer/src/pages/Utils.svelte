<script lang="ts">
	import PageHeader from "@components/shared/PageHeader.svelte";
	import {
		DetectLosslessDll,
		GetSystemToolsStatus,
		GetUtilsStatus,
		InstallLsfg,
		UninstallLsfg,
		onEvent,
	} from "@lib/api";
	import * as core from "@shared";
	import { notifications } from "@stores/notificationStore";
	import { onDestroy, onMount } from "svelte";

	import lsfgPng from "@icons/lsfg.png";
	import SystemDependencies from "@components/utils/SystemDependencies.svelte";
	import UtilityCard from "@components/utils/UtilityCard.svelte";

	let status = $state<core.UtilsStatus>({ isLsfgInstalled: false, lsfgVersion: "" });
	let systemStatus = $state<core.SystemToolsStatus & { hasLosslessDll: boolean }>({
		hasGamescope: false,
		hasMangoHud: false,
		hasGameMode: false,
		hasVulkanInfo: false,
		hasLosslessDll: false,
	});
	let isInstalling = $state(false);
	let isUninstalling = $state(false);
	let progressMessage = $state("");
	let progressPercent = $state(0);
	let unsubProgress: (() => void) | null = null;

	const utilities = $derived([
		{
			id: "lsfg",
			name: "LSFG-VK",
			isInstalled: status.isLsfgInstalled,
			isInstalling,
			isUninstalling,
			progressMessage,
			progressPercent,
			description: [
				"Lossless Scaling is a Windows-exclusive program featuring various algorithms for scaling and interpolating programs.",
				"<strong>lsfg-vk</strong> is a Vulkan layer that hooks into Vulkan applications and generates additional frames using Lossless Scaling's frame generation algorithm."
			],
			note: "Note: Requires Lossless Scaling downloaded on Steam.",
			icon: lsfgPng,
			onInstall: handleInstall,
			onUninstall: handleUninstall,
		}
	]);

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
		unsubProgress = onEvent("lsfg-install-progress", (data: any) => {
			progressMessage = data.message;
			progressPercent = data.percent;
		});
	});

	onDestroy(() => {
		if (unsubProgress) unsubProgress();
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
		{#each utilities as util (util.id)}
			<UtilityCard {...util} />
		{/each}
	</div>
</div>

<style lang="scss">
	.utils-container {
		display: flex;
		flex-direction: column;
		width: 100%;
	}

	.utils-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 400px), 520px));
		gap: 24px;
		width: 100%;
		margin-top: 8px;
	}
</style>
