import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import { spawn } from "child_process";
import { ConfigService } from "./config.service";
import { PathsService } from "./paths.service";
import { PrefixService } from "./prefix.service";
import { MigrationService } from "./migration.service";
import { LsfgService } from "./lsfg.service";
import { ProtonService } from "./proton.service";
import { SystemService } from "./system.service";
import { LoggerService } from "./logger.service";
import type { LaunchOptions } from "../../shared/types/config.types";

export class RunnerService {
	static findInstanceManager(): string {
		const instanceName = "light-launcher-instance";
		const candidates = [
			path.join(process.resourcesPath || "", instanceName),
			path.join(process.resourcesPath || "", "bin", instanceName),
			path.join(__dirname, "../../bin", instanceName),
			path.join(__dirname, "../../../bin", instanceName),
			path.join(process.cwd(), "bin", instanceName),
			path.join(process.cwd(), instanceName),
			`/usr/bin/${instanceName}`
		];

		for (const candidate of candidates) {
			if (fsSync.existsSync(candidate)) {
				return path.resolve(candidate);
			}
		}

		return "";
	}

	static async runGame(options: LaunchOptions, showLogs: boolean): Promise<void> {
		if (!options.UseGamePath && options.LauncherPath) {
			options.GamePath = options.LauncherPath;
		}

		if (!options.PrefixPath || options.PrefixPath.includes("/LightLauncher/prefixes/")) {
			const prefixName = options.PrefixPath ? path.basename(options.PrefixPath) : "Default";
			const basePrefixDir = await PrefixService.getPrefixBaseDir();
			options.PrefixPath = path.join(basePrefixDir, prefixName);
		} else if (!fsSync.existsSync(options.PrefixPath)) {
			const prefixName = path.basename(options.PrefixPath);
			const basePrefixDir = await PrefixService.getPrefixBaseDir();
			const candidate = path.join(basePrefixDir, prefixName);
			if (fsSync.existsSync(candidate)) {
				options.PrefixPath = candidate;
			}
		}

		if (!fsSync.existsSync(options.PrefixPath)) {
			await fs.mkdir(options.PrefixPath, { recursive: true });
		}

		await MigrationService.repairPrefixSymlinks();

		if (!fsSync.existsSync(options.GamePath)) {
			LoggerService.error("Runner", `Game executable not found at: ${options.GamePath}`);
			throw new Error(`Game executable not found at: ${options.GamePath}`);
		}

		LoggerService.info("Runner", `Launching "${options.Name || path.basename(options.GamePath)}"`, {
			game: options.GamePath,
			prefix: options.PrefixPath,
			proton: options.ProtonPath || "default"
		});

		// Save configuration
		await ConfigService.saveGameConfig(options);

		// Handle LSFG profile
		if (options.Extras?.Lsfg?.Enabled) {
			let gpu = options.Extras.Lsfg.Gpu;
			if (!gpu) {
				const gpus = SystemService.getListGpus();
				if (gpus.length > 0) gpu = gpus[0];
			}
			await LsfgService.saveLsfgProfile(
				options.Name,
				options.GamePath,
				parseInt(options.Extras.Lsfg.Multiplier || "2", 10),
				options.Extras.Lsfg.PerfMode,
				options.Extras.Lsfg.DllPath,
				gpu,
				options.Extras.Lsfg.FlowScale,
				options.Extras.Lsfg.Pacing,
				options.Extras.Lsfg.AllowFp16
			);
		} else {
			await LsfgService.disableLsfgProfile(options.Name, options.GamePath);
		}

		// Check umu-run availability
		const hasUmu = (() => {
			try {
				const { spawnSync } = require("child_process");
				return spawnSync("which", ["umu-run"]).status === 0;
			} catch {
				return false;
			}
		})();

		if (!hasUmu) {
			LoggerService.error("Runner", "umu-run command not found in PATH");
			throw new Error(
				"umu-run is required to launch games via Proton. Please install umu-launcher (e.g., 'yay -S umu-launcher')."
			);
		}

		const instancePath = this.findInstanceManager();
		if (!instancePath) {
			LoggerService.error("Runner", "light-launcher-instance binary not found");
			throw new Error(
				"light-launcher-instance binary not found. Please build or reinstall LightLauncher."
			);
		}

		// Match Proton path if configured or fallback to first available
		if (options.ProtonPath) {
			const protonTools = await ProtonService.scanProtonVersions();
			const match = ProtonService.findProtonMatch(options.ProtonPath, protonTools);
			if (match) {
				options.ProtonPath = match.Path;
			} else if (!fsSync.existsSync(options.ProtonPath) && protonTools.length > 0) {
				LoggerService.warn(
					"Runner",
					`Configured Proton "${options.ProtonPath}" not found. Falling back to "${protonTools[0].DisplayName}".`
				);
				options.ProtonPath = protonTools[0].Path;
			}
		}

		// Build arguments for light-launcher-instance
		const args: string[] = [
			"--game",
			options.GamePath,
			"--launcher",
			options.LauncherPath || options.GamePath,
			"--prefix",
			PathsService.expandPath(options.PrefixPath),
			"--proton-pattern",
			options.ProtonPath ? path.basename(options.ProtonPath) : "",
			"--proton-path",
			options.ProtonPath ? PathsService.expandPath(options.ProtonPath) : ""
		];

		if (!showLogs) {
			args.push("--logs=false");
		}

		// Extra features
		const extras = options.Extras;
		if (extras) {
			if (extras.EnableMangoHud) {
				args.push("--mangohud");
			}
			if (extras.EnableGamemode) {
				args.push("--gamemode");
			}
			if (extras.Memory?.Enabled) {
				args.push("--memory-min");
				if (extras.Memory.Value) {
					args.push("--memory-min-value", extras.Memory.Value);
				}
			}
			if (extras.Gamescope?.Enabled) {
				const gs = extras.Gamescope;
				args.push("--gamescope");
				if (gs.Width) args.push("--gs-w", gs.Width);
				if (gs.Height) args.push("--gs-h", gs.Height);
				if (gs.OutputWidth) args.push("--gs-out-w", gs.OutputWidth);
				if (gs.OutputHeight) args.push("--gs-out-h", gs.OutputHeight);
				if (gs.RefreshRate) args.push("--gs-r", gs.RefreshRate);
				if (gs.FramerateLimit) args.push("--gs-fr-limit", gs.FramerateLimit);
				if (gs.WindowMode) args.push("--gs-window-mode", gs.WindowMode);
				if (gs.Scaler) args.push("--gs-scaler", gs.Scaler);
				if (gs.Filter) args.push("--gs-filter", gs.Filter);
				if (gs.Sharpness) args.push("--gs-sharpness", gs.Sharpness);
				if (gs.HDR) args.push("--gs-hdr");
				if (gs.AdaptiveSync) args.push("--gs-adaptive-sync");
				if (gs.Mangoapp) args.push("--gs-mangoapp");
				if (gs.CustomArgs) args.push("--gs-custom-args", gs.CustomArgs);
			}
			if (extras.Lsfg?.Enabled) {
				const ls = extras.Lsfg;
				args.push("--lsfg");
				if (ls.Multiplier) args.push("--lsfg-multiplier", ls.Multiplier);
				if (ls.PerfMode) args.push("--lsfg-perf");
				if (ls.DllPath) args.push("--lsfg-dll", ls.DllPath);
				if (ls.Gpu) args.push("--lsfg-gpu", ls.Gpu);
				if (ls.FlowScale) args.push("--lsfg-flow", ls.FlowScale);
				if (ls.Pacing) args.push("--lsfg-pacing", ls.Pacing);
				if (ls.AllowFp16) args.push("--lsfg-fp16");
			}
		}

		// Spawn detached instance using setsid so it survives parent process / concurrently termination
		const hasSetsid = fsSync.existsSync("/usr/bin/setsid");
		const spawnCmd = hasSetsid ? "/usr/bin/setsid" : instancePath;
		const spawnArgs = hasSetsid ? [instancePath, ...args] : args;

		LoggerService.info("Runner", `Executing: ${spawnCmd} ${spawnArgs.join(" ")}`);

		const child = spawn(spawnCmd, spawnArgs, {
			detached: true,
			stdio: "ignore"
		});
		child.unref();
	}
}
