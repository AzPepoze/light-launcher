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
		background: var(--bg-surface);
		border: 2px solid rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		transform: translateY(calc(100% - 78px));
		transition: transform 0.6s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.3s ease;
		z-index: 100;
		padding: 0 24px 24px 24px;
		box-shadow: 0 10px 50px rgba(0, 0, 0, 0.4);
		overflow: hidden;
		margin-right: 24px;
		width: -webkit-fill-available;
		box-sizing: border-box;

		&.expanded {
			transform: translateY(0);
			box-shadow: 0 10px 60px rgba(255, 102, 171, 0.1);
		}
	}

	.toggle-btn {
		width: 100%;
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		background: var(--bg-base);
		border: 2px solid rgba(255, 255, 255, 0.03);
		border-radius: var(--radius-pill);
		margin: 12px 0;
		transition: transform var(--transition-spring), border-color var(--transition-fast), background var(--transition-fast);

		.trigger-text {
			font-size: 0.85rem;
			font-weight: 800;
			color: var(--text-muted);
			letter-spacing: 1.5px;
			text-transform: uppercase;
		}

		&:hover {
			background: var(--bg-elevated);
			border-color: var(--accent-primary);

			.trigger-text {
				color: var(--text-main);
			}
		}

		&:active {
			transform: scale(0.98);
		}
	}

	.drawer-content {
		padding-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}


</style>
