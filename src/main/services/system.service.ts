import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import { exec, execSync } from "child_process";
import { promisify } from "util";
import { LoggerService } from "./logger.service";
import type { SystemInfo, SystemToolsStatus, SystemUsage } from "../../shared/types/system.types";

const execAsync = promisify(exec);

let lastTotal = 0;
let lastIdle = 0;
let cachedGpuDevices: { name: string; pci?: string; rank: number }[] | null = null;

export class SystemService {
	static isCommandAvailable(commandName: string): boolean {
		try {
			execSync(`which ${commandName}`, { stdio: "ignore" });
			return true;
		} catch {
			return false;
		}
	}

	static getSystemToolsStatus(): SystemToolsStatus {
		return {
			hasGamescope: this.isCommandAvailable("gamescope"),
			hasMangoHud: this.isCommandAvailable("mangohud"),
			hasGameMode: this.isCommandAvailable("gamemoderun"),
			hasVulkanInfo: this.isCommandAvailable("vulkaninfo")
		};
	}

	private static async parseMemInfo(): Promise<{ memTotal: number; memAvailable: number }> {
		let memTotal = 0;
		let memAvailable = 0;
		try {
			const memContent = await fs.readFile("/proc/meminfo", "utf-8");
			for (const line of memContent.split("\n")) {
				if (line.startsWith("MemTotal:")) {
					memTotal = parseInt(line.replace(/\D/g, ""), 10);
				} else if (line.startsWith("MemAvailable:")) {
					memAvailable = parseInt(line.replace(/\D/g, ""), 10);
				}
				if (memTotal > 0 && memAvailable > 0) break;
			}
		} catch {}
		return { memTotal, memAvailable };
	}

	static async getSystemUsage(): Promise<SystemUsage> {
		const usage: SystemUsage = { cpu: 0, ram: 0, ramUsedGb: 0, ramTotalGb: 0, gpus: [] };

		try {
			const statContent = await fs.readFile("/proc/stat", "utf-8");
			const lines = statContent.split("\n");
			if (lines.length > 0 && lines[0].startsWith("cpu ")) {
				const fields = lines[0].trim().split(/\s+/).slice(1).map(Number);
				const total = fields.reduce((acc, v) => acc + v, 0);
				const idle = fields[3] || 0;

				if (lastTotal > 0) {
					const totalDelta = total - lastTotal;
					const idleDelta = idle - lastIdle;
					if (totalDelta > 0) {
						usage.cpu = Math.min(100, Math.max(0, (100 * (totalDelta - idleDelta)) / totalDelta));
					}
				}
				lastTotal = total;
				lastIdle = idle;
			}
		} catch {}

		const { memTotal, memAvailable } = await this.parseMemInfo();
		if (memTotal > 0) {
			const used = memTotal - memAvailable;
			usage.ramUsedGb = Math.round((used / 1024 / 1024) * 10) / 10;
			usage.ramTotalGb = Math.round(memTotal / 1024 / 1024);
			usage.ram = Math.min(100, Math.max(0, Math.round((used / memTotal) * 100)));
		}

		usage.gpus = await this.getGpuUsages();
		return usage;
	}

	static detectGpuDevices(forceRefresh = false): { name: string; pci?: string; rank: number }[] {
		if (cachedGpuDevices && !forceRefresh) {
			return cachedGpuDevices;
		}

		const gpus: { name: string; pci?: string; rank: number }[] = [];

		try {
			const output = execSync("lspci | grep -i 'vga\\|3d\\|display'", { encoding: "utf-8" });
			const lines = output.trim().split("\n").filter(Boolean);
			for (const line of lines) {
				const pciMatch = line.match(/^([0-9a-fA-F:.]+)\s+/);
				const pci = pciMatch ? pciMatch[1] : undefined;
				const colonIdx = line.indexOf(": ");
				const name = colonIdx !== -1 ? line.substring(colonIdx + 2).trim() : line;
				if (!name) continue;
				const lower = line.toLowerCase();
				const rank = lower.includes("vga") ? 0 : lower.includes("3d") ? 1 : 2;
				gpus.push({ name, pci, rank });
			}
			gpus.sort((a, b) => a.rank - b.rank);
		} catch {}

		if (gpus.length === 0) {
			try {
				if (fsSync.existsSync("/sys/class/drm")) {
					const entries = fsSync.readdirSync("/sys/class/drm");
					for (const entry of entries) {
						if (entry.startsWith("card") && !entry.includes("-")) {
							const devicePath = path.join("/sys/class/drm", entry, "device");
							if (fsSync.existsSync(devicePath)) {
								let name = "";
								for (const file of ["product_name", "product"]) {
									try {
										const value = fsSync.readFileSync(path.join(devicePath, file), "utf-8").trim();
										if (value) {
											name = value;
											break;
										}
									} catch {}
								}
								gpus.push({ name: name || entry, rank: 0 });
							}
						}
					}
				}
			} catch {}
		}

		cachedGpuDevices = gpus;
		return gpus;
	}

	static async getGpuUsages(
		devices?: { name: string; pci?: string; rank: number }[]
	): Promise<number[]> {
		const gpus = devices ?? this.detectGpuDevices();
		if (gpus.length === 0) {
			return [];
		}

		const nvidiaMap: Record<string, number> = {};
		if (this.isCommandAvailable("nvidia-smi")) {
			try {
				const { stdout } = await execAsync(
					"nvidia-smi --query-gpu=pci.bus_id,utilization.gpu --format=csv,noheader,nounits"
				);
				for (const line of stdout.trim().split("\n")) {
					const parts = line.split(",").map((s) => s.trim());
					if (parts.length >= 2) {
						const util = parseFloat(parts[1]);
						if (!isNaN(util)) nvidiaMap[parts[0].toLowerCase()] = util;
					}
				}
			} catch {}
		}

		const usages: number[] = [];
		const drmPath = "/sys/class/drm";
		let drmEntries: string[] = [];
		try {
			if (fsSync.existsSync(drmPath)) {
				drmEntries = (await fs.readdir(drmPath)).filter(
					(e) => e.startsWith("card") && !e.includes("-")
				);
			}
		} catch {}

		for (const gpu of gpus) {
			let foundUsage: number | null = null;

			if (gpu.pci && Object.keys(nvidiaMap).length > 0) {
				const lowerPci = gpu.pci.toLowerCase();
				for (const [bus, util] of Object.entries(nvidiaMap)) {
					if (bus.includes(lowerPci)) {
						foundUsage = util;
						break;
					}
				}
			}

			if (foundUsage === null && drmEntries.length > 0) {
				for (const entry of drmEntries) {
					const cardPath = path.join(drmPath, entry);
					const devPath = path.join(cardPath, "device");
					let matches = false;
					if (gpu.pci && fsSync.existsSync(devPath)) {
						try {
							const target = await fs.readlink(devPath);
							if (target.includes(gpu.pci)) matches = true;
						} catch {}
					} else if (!gpu.pci) {
						matches = true;
					}

					if (matches) {
						const busyFile = path.join(cardPath, "device/gpu_busy_percent");
						if (fsSync.existsSync(busyFile)) {
							try {
								const val = parseInt((await fs.readFile(busyFile, "utf-8")).trim(), 10);
								if (!isNaN(val)) {
									foundUsage = Math.min(100, Math.max(0, val));
									break;
								}
							} catch {}
						}

						const actFile = path.join(cardPath, "gt_act_freq_mhz");
						const maxFile = path.join(cardPath, "gt_max_freq_mhz");
						const minFile = path.join(cardPath, "gt_min_freq_mhz");
						if (fsSync.existsSync(actFile) && fsSync.existsSync(maxFile)) {
							try {
								const act = parseInt((await fs.readFile(actFile, "utf-8")).trim(), 10);
								const max = parseInt((await fs.readFile(maxFile, "utf-8")).trim(), 10);
								const min = fsSync.existsSync(minFile)
									? parseInt((await fs.readFile(minFile, "utf-8")).trim(), 10)
									: 0;
								if (max > min && act > min) {
									foundUsage = Math.round(((act - min) / (max - min)) * 100);
								} else {
									foundUsage = 0;
								}
								break;
							} catch {}
						}
					}
				}
			}

			usages.push(foundUsage ?? 0);
		}

		return usages;
	}

	static getListGpus(): string[] {
		return this.detectGpuDevices().map((g) => g.name);
	}

	static async getSystemInfo(): Promise<SystemInfo> {
		const info: SystemInfo = {
			os: "Unknown",
			kernel: "Unknown",
			cpu: "Unknown",
			gpu: "Unknown",
			ram: "Unknown",
			driver: "Unknown"
		};

		try {
			if (fsSync.existsSync("/etc/os-release")) {
				const content = await fs.readFile("/etc/os-release", "utf-8");
				for (const line of content.split("\n")) {
					if (line.startsWith("PRETTY_NAME=")) {
						info.os = line.replace("PRETTY_NAME=", "").replace(/"/g, "").trim();
						break;
					}
				}
			}
		} catch {}

		try {
			const { stdout } = await execAsync("uname -r");
			info.kernel = stdout.trim();
		} catch {}

		try {
			const { stdout } = await execAsync("lscpu | grep 'Model name' | cut -d':' -f2");
			info.cpu = stdout.trim();
		} catch {}

		const gpus = this.getListGpus();
		if (gpus.length > 0) {
			info.gpu = gpus[0];
			info.gpus = gpus;
		}

		const { memTotal } = await this.parseMemInfo();
		if (memTotal > 0) {
			info.ram = `${Math.round(memTotal / 1024 / 1024)} GB`;
		}

		try {
			const { stdout } = await execAsync(
				"vulkaninfo --summary | grep -m 1 'driverVersion' | awk '{print $3}'"
			);
			info.driver = stdout.trim() || "Unknown";
		} catch {}

		return info;
	}

	static async dropCaches(): Promise<void> {
		await execAsync("sync");
		await execAsync("pkexec sysctl -w vm.drop_caches=3");
	}

	static async clearSwap(): Promise<void> {
		try {
			const swapContent = await fs.readFile("/proc/swaps", "utf-8");
			const lines = swapContent.trim().split("\n").slice(1);
			const swaps = lines.map((l) => l.trim().split(/\s+/)[0]).filter(Boolean);
			if (swaps.length === 0) return;

			const swaponCmds = swaps.map((s) => `swapon ${s}`).join(" ; ");
			const cmdStr = `swapoff -a ; ${swaponCmds}`;
			await execAsync(`pkexec sh -c "${cmdStr}"`);
		} catch (err) {
			LoggerService.error("System", `Error clearing swap: ${err}`);
		}
	}

	static async cleanupProcesses(): Promise<void> {
		const commands = [
			"umu-run",
			"pressure-vessel",
			"gamescopereaper",
			"steam-runtime-launcher-service",
			"srt-bwrap",
			"reaper"
		];
		for (const command of commands) {
			try {
				await execAsync(`pkill -f ${command}`);
			} catch {}
		}
	}

	static getShaderCachePaths(): string[] {
		const home = os.homedir();
		return [
			path.join(home, ".cache/mesa_shader_cache"),
			path.join(home, ".cache/nvidia/GLCache"),
			path.join(home, ".nv/GLCache"),
			path.join(home, ".cache/AMD/VkCache"),
			path.join(home, ".cache/radv_builtin_shaders")
		];
	}

	static async getShaderCacheSize(): Promise<string> {
		let totalBytes = 0;
		for (const p of this.getShaderCachePaths()) {
			if (fsSync.existsSync(p)) {
				try {
					const { stdout } = await execAsync(`du -sb "${p}"`);
					const size = parseInt(stdout.trim().split(/\s+/)[0], 10);
					if (!isNaN(size)) totalBytes += size;
				} catch {}
			}
		}
		if (totalBytes === 0) return "0 MB";
		return `${(totalBytes / 1024 / 1024).toFixed(1)} MB`;
	}

	static async clearShaderCache(): Promise<void> {
		for (const p of this.getShaderCachePaths()) {
			if (fsSync.existsSync(p)) {
				try {
					await fs.rm(p, { recursive: true, force: true });
				} catch {}
			}
		}
	}
}
