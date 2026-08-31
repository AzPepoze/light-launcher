import {
	GetSystemInfo,
	GetSystemUsage,
	CleanupProcesses,
	GetShaderCacheSize,
	ClearShaderCache,
	DropCaches,
	ClearSwap
} from "@lib/api";
import * as core from "@shared";
import { ResourceHistoryTracker } from "./ResourceHistoryTracker.svelte";

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
		driver: ""
	});
	sysUsage = $state<core.SystemUsage>({
		cpu: 0,
		ram: 0,
		ramUsedGb: 0,
		ramTotalGb: 0,
		gpus: []
	});
	shaderCacheSize = $state("...");
	private usageInterval: ReturnType<typeof setInterval> | null = null;
	private tracker = new ResourceHistoryTracker(30);

	get cpuHistory(): number[] {
		return this.tracker.cpu;
	}

	get ramHistory(): number[] {
		return this.tracker.ram;
	}

	get gpuHistoryMap(): Record<number, number[]> {
		return this.tracker.gpuMap;
	}

	pushUsage(usage: core.SystemUsage) {
		this.tracker.push(usage);
	}

	async fetchData() {
		try {
			const [info, usage, cache] = await Promise.all([
				GetSystemInfo(),
				GetSystemUsage(),
				GetShaderCacheSize()
			]);
			this.sysInfo = info;
			this.sysUsage = usage;
			this.pushUsage(usage);
			this.shaderCacheSize = cache;
		} catch (err) {
			console.error("Failed to fetch status drawer data:", err);
		}
	}

	initialize() {
		this.fetchData();
		if (this.usageInterval) clearInterval(this.usageInterval);
		this.usageInterval = setInterval(async () => {
			try {
				const usage = await GetSystemUsage();
				this.sysUsage = usage;
				this.pushUsage(usage);
			} catch {}
		}, 1000);
	}

	destroy() {
		if (this.usageInterval) {
			clearInterval(this.usageInterval);
			this.usageInterval = null;
		}
	}

	private async runAction(
		loadingKey: "isCleaning" | "isClearingCache" | "isDroppingCaches" | "isClearingSwap",
		successKey: "showCleanupSuccess" | "showCacheSuccess" | "showDropSuccess" | "showSwapSuccess",
		action: () => Promise<void>
	) {
		if (this[loadingKey]) return;
		this[loadingKey] = true;
		this[successKey] = false;
		try {
			await action();
			setTimeout(() => {
				this[successKey] = true;
				setTimeout(() => {
					this[successKey] = false;
				}, 2000);
			}, 100);
		} catch (err) {
			console.error(`Action ${loadingKey} failed:`, err);
		} finally {
			setTimeout(() => {
				this[loadingKey] = false;
			}, 1500);
		}
	}

	async handleCleanup() {
		await this.runAction("isCleaning", "showCleanupSuccess", async () => {
			await CleanupProcesses();
			await this.fetchData();
		});
	}

	async handleClearCache() {
		await this.runAction("isClearingCache", "showCacheSuccess", async () => {
			await ClearShaderCache();
			this.shaderCacheSize = await GetShaderCacheSize();
		});
	}

	async handleDropCaches() {
		await this.runAction("isDroppingCaches", "showDropSuccess", async () => {
			await DropCaches();
			await this.fetchData();
		});
	}

	async handleClearSwap() {
		await this.runAction("isClearingSwap", "showSwapSuccess", async () => {
			await ClearSwap();
			await this.fetchData();
		});
	}
}
