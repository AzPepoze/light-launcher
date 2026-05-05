<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import {
		GetSystemInfo,
		GetSystemUsage,
		CleanupProcesses,
		GetShaderCacheSize,
		ClearShaderCache,
		DropCaches,
		ClearSwap,
	} from "@bindings/light-launcher/internal/app/app";
	import * as core from "@bindings/light-launcher/internal/types/models";

	import StatusUtilityButton from "@components/shared/StatusUtilityButton.svelte";
	import SystemResources from "@components/shared/SystemResources.svelte";
	import CleanupActions from "@components/shared/CleanupActions.svelte";

	let isExpanded = false;
	let isCleaning = false;
	let isClearingCache = false;
	let isDroppingCaches = false;
	let isClearingSwap = false;
	let showCleanupSuccess = false;
	let showCacheSuccess = false;
	let showDropSuccess = false;
	let showSwapSuccess = false;
	let sysInfo: core.SystemInfo = {
		os: "",
		kernel: "",
		cpu: "",
		gpu: "",
		ram: "",
		driver: "",
	};
	let sysUsage: core.SystemUsage = { cpu: "0%", ram: "0%", gpu: "0%" };
	let shaderCacheSize = "...";
	let usageInterval;

	async function fetchData() {
		try {
			const [info, usage, cache] = await Promise.all([
				GetSystemInfo(),
				GetSystemUsage(),
				GetShaderCacheSize(),
			]);
			sysInfo = info;
			sysUsage = usage;
			shaderCacheSize = cache;
		} catch (err) {
			console.error("Failed to fetch status drawer data:", err);
		}
	}

	onMount(() => {
		fetchData();
		usageInterval = setInterval(async () => {
			try {
				sysUsage = await GetSystemUsage();
			} catch (e) {}
		}, 3000);
	});

	onDestroy(() => {
		if (usageInterval) clearInterval(usageInterval);
	});

	async function handleCleanup() {
		if (isCleaning) return;
		isCleaning = true;
		showCleanupSuccess = false;
		try {
			await CleanupProcesses();
			await fetchData();
			// Faster pop
			setTimeout(() => {
				showCleanupSuccess = true;
				// Longer visibility
				setTimeout(() => {
					showCleanupSuccess = false;
				}, 2000);
			}, 100);
		} catch (err) {
			console.error(`Cleanup failed: ${err}`);
		} finally {
			setTimeout(() => {
				isCleaning = false;
			}, 1500);
		}
	}

	async function handleClearCache() {
		if (isClearingCache) return;
		isClearingCache = true;
		showCacheSuccess = false;
		try {
			await ClearShaderCache();
			const newCache = await GetShaderCacheSize();
			shaderCacheSize = newCache;
			setTimeout(() => {
				showCacheSuccess = true;
				setTimeout(() => {
					showCacheSuccess = false;
				}, 2000);
			}, 100);
		} catch (err) {
			console.error(`Failed to clear cache: ${err}`);
		} finally {
			setTimeout(() => {
				isClearingCache = false;
			}, 1500);
		}
	}

	async function handleDropCaches() {
		if (isDroppingCaches) return;
		isDroppingCaches = true;
		showDropSuccess = false;
		try {
			await DropCaches();
			await fetchData();
			setTimeout(() => {
				showDropSuccess = true;
				setTimeout(() => {
					showDropSuccess = false;
				}, 2000);
			}, 100);
		} catch (err) {
			console.error(`Failed to drop caches: ${err}`);
		} finally {
			setTimeout(() => {
				isDroppingCaches = false;
			}, 1500);
		}
	}

	async function handleClearSwap() {
		if (isClearingSwap) return;
		isClearingSwap = true;
		showSwapSuccess = false;
		try {
			await ClearSwap();
			await fetchData();
			setTimeout(() => {
				showSwapSuccess = true;
				setTimeout(() => {
					showSwapSuccess = false;
				}, 2000);
			}, 100);
		} catch (err) {
			console.error(`Failed to clear swap: ${err}`);
		} finally {
			setTimeout(() => {
				isClearingSwap = false;
			}, 1500);
		}
	}
</script>

<div class="status-drawer-wrapper" class:expanded={isExpanded}>
	<button class="toggle-btn" on:click={() => (isExpanded = !isExpanded)}>
		<span class="trigger-text"
			>{isExpanded
				? "CLOSE DRAWER"
				: "SYSTEM STATUS & UTILITIES"}</span
		>
	</button>

	<div class="drawer-content">
		<SystemResources {sysInfo} {sysUsage} />

		<div class="divider"></div>

		<CleanupActions
			{isCleaning}
			{showCleanupSuccess}
			handleCleanup={handleCleanup}
			{isClearingCache}
			{showCacheSuccess}
			handleClearCache={handleClearCache}
			{shaderCacheSize}
			{isDroppingCaches}
			{showDropSuccess}
			handleDropCaches={handleDropCaches}
			{isClearingSwap}
			{showSwapSuccess}
			handleClearSwap={handleClearSwap}
		/>
	</div>
</div>

<style lang="scss">
	.status-drawer-wrapper {
		position: fixed;
		bottom: 20px;
		background: var(--glass-bg);
		border: 1px solid var(--glass-border);
		border-radius: 20px;
		transform: translateY(calc(100% - 55px));
		transition: all 0.5s cubic-bezier(0.23, 1, 0.32, 1);
		z-index: 100;
		padding: 0 20px 20px 20px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
		overflow: hidden;
		margin-right: 20px;
		width: -webkit-fill-available;

		&.expanded {
			transform: translateY(0);
		}
	}

	.toggle-btn {
		width: 100%;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		background: var(--glass-surface);
		border: 1px solid var(--glass-border);
		border-radius: 12px;
		margin: 10px 0;
		transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);

		.trigger-text {
			font-size: 0.8rem;
			font-weight: 800;
			color: var(--text-dim);
			letter-spacing: 1.5px;
			text-transform: uppercase;
		}

		&:hover {
			background: var(--glass-border);
			border-color: var(--glass-border-bright);

			.trigger-text {
				color: var(--text-main);
			}
		}

		&:active {
			background: var(--glass-hover);
			transform: scale(0.995);
		}
	}

	.drawer-content {
		padding-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}


</style>
