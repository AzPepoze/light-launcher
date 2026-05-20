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

export class StatusDrawerState {
	isExpanded = $state(false);
	isCleaning = $state(false);
	isClearingCache = $state(false);
	isDroppingCaches = $state(false);
	isClearingSwap = $state(false);
	showCleanupSuccess = $state(false);
	showCacheSuccess = $state(false);
	showDropSuccess = $state(false);
	showSwapSuccess = $state(false);
	sysInfo = $state<core.SystemInfo>({
		os: "",
		kernel: "",
		cpu: "",
		gpu: "",
		ram: "",
		driver: "",
	});
	sysUsage = $state<core.SystemUsage>({ cpu: "0%", ram: "0%", gpu: "0%" });
	shaderCacheSize = $state("...");
	usageInterval: any = null;

	async fetchData() {
		try {
			const [info, usage, cache] = await Promise.all([
				GetSystemInfo(),
				GetSystemUsage(),
				GetShaderCacheSize(),
			]);
			this.sysInfo = info;
			this.sysUsage = usage;
			this.shaderCacheSize = cache;
		} catch (err) {
			console.error("Failed to fetch status drawer data:", err);
		}
	}

	initialize() {
		this.fetchData();
		this.usageInterval = setInterval(async () => {
			try {
				this.sysUsage = await GetSystemUsage();
			} catch (e) {}
		}, 3000);
	}

	destroy() {
		if (this.usageInterval) clearInterval(this.usageInterval);
	}

	private triggerSuccess(stateKey: "showCleanupSuccess" | "showCacheSuccess" | "showDropSuccess" | "showSwapSuccess") {
		setTimeout(() => {
			this[stateKey] = true;
			setTimeout(() => {
				this[stateKey] = false;
			}, 2000);
		}, 100);
	}

	async handleCleanup() {
		if (this.isCleaning) return;
		this.isCleaning = true;
		this.showCleanupSuccess = false;
		try {
			await CleanupProcesses();
			await this.fetchData();
			this.triggerSuccess("showCleanupSuccess");
		} catch (err) {
			console.error(`Cleanup failed: ${err}`);
		} finally {
			setTimeout(() => {
				this.isCleaning = false;
			}, 1500);
		}
	}

	async handleClearCache() {
		if (this.isClearingCache) return;
		this.isClearingCache = true;
		this.showCacheSuccess = false;
		try {
			await ClearShaderCache();
			const newCache = await GetShaderCacheSize();
			this.shaderCacheSize = newCache;
			this.triggerSuccess("showCacheSuccess");
		} catch (err) {
			console.error(`Failed to clear cache: ${err}`);
		} finally {
			setTimeout(() => {
				this.isClearingCache = false;
			}, 1500);
		}
	}

	async handleDropCaches() {
		if (this.isDroppingCaches) return;
		this.isDroppingCaches = true;
		this.showDropSuccess = false;
		try {
			await DropCaches();
			await this.fetchData();
			this.triggerSuccess("showDropSuccess");
		} catch (err) {
			console.error(`Failed to drop caches: ${err}`);
		} finally {
			setTimeout(() => {
				this.isDroppingCaches = false;
			}, 1500);
		}
	}

	async handleClearSwap() {
		if (this.isClearingSwap) return;
		this.isClearingSwap = true;
		this.showSwapSuccess = false;
		try {
			await ClearSwap();
			await this.fetchData();
			this.triggerSuccess("showSwapSuccess");
		} catch (err) {
			console.error(`Failed to clear swap: ${err}`);
		} finally {
			setTimeout(() => {
				this.isClearingSwap = false;
			}, 1500);
		}
	}
}
