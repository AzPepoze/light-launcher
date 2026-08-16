import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import { spawn } from "child_process";
import { PathsService } from "./paths.service";
import { ConfigService } from "./config.service";
import { ProtonService } from "./proton.service";
import type { LaunchOptions } from "../../shared/types/config.types";
import type { PrefixConfigWithProton } from "../../shared/types/prefix.types";

export class PrefixService {
	static async getPrefixBaseDir(): Promise<string> {
		const settings = await ConfigService.loadAppSettings();
		return PathsService.getPrefixBaseDirectory(settings?.CustomPrefixDir);
	}

	static async listPrefixes(): Promise<string[]> {
		const baseDir = await this.getPrefixBaseDir();
		if (!fsSync.existsSync(baseDir)) {
			return ["Default"];
		}

		try {
			const entries = await fs.readdir(baseDir, { withFileTypes: true });
			const prefixes = entries.filter((e) => e.isDirectory()).map((e) => e.name);
			if (!prefixes.includes("Default")) {
				prefixes.unshift("Default");
			}
			return prefixes;
		} catch {
			return ["Default"];
		}
	}

	static async createPrefix(name: string): Promise<void> {
		if (!name || name.trim() === "") {
			throw new Error("Prefix name cannot be empty");
		}
		const baseDir = await this.getPrefixBaseDir();
		const prefixPath = path.join(baseDir, name.trim());
		await fs.mkdir(prefixPath, { recursive: true });
	}

	static async removePrefix(name: string): Promise<void> {
		if (name === "Default") {
			throw new Error("Cannot delete Default prefix");
		}
		const baseDir = await this.getPrefixBaseDir();
		const prefixPath = path.join(baseDir, name);
		if (fsSync.existsSync(prefixPath)) {
			await fs.rm(prefixPath, { recursive: true, force: true });
		}
	}

	static async savePrefixConfig(prefixName: string, options: LaunchOptions): Promise<void> {
		await ConfigService.savePrefixConfig(prefixName, options);
	}

	static async loadPrefixConfig(prefixName: string): Promise<LaunchOptions> {
		return ConfigService.loadPrefixConfig(prefixName);
	}

	static async loadPrefixConfigWithProton(prefixName: string): Promise<PrefixConfigWithProton> {
		const cfg = await ConfigService.loadPrefixConfig(prefixName);
		const result: PrefixConfigWithProton = {
			config: cfg,
			protonDisplayName: "",
			protonName: "",
			protonPath: cfg.ProtonPath || "",
			protonIsSteam: false
		};

		if (cfg.ProtonPath) {
			const protonTools = await ProtonService.scanProtonVersions();
			const match = ProtonService.findProtonMatch(cfg.ProtonPath, protonTools);
			if (match) {
				result.protonDisplayName = match.DisplayName;
				result.protonName = match.Name;
				result.protonPath = match.Path;
				result.protonIsSteam = match.IsSteam;
			} else {
				result.protonDisplayName = cfg.ProtonPath;
			}
		}

		return result;
	}

	static async runPrefixTool(
		prefixPath: string,
		toolName: string,
		protonPath: string
	): Promise<void> {
		if (!toolName) {
			throw new Error("Tool name cannot be empty");
		}

		const resolvedPrefix = PathsService.expandPath(prefixPath);
		if (!fsSync.existsSync(resolvedPrefix)) {
			await fs.mkdir(resolvedPrefix, { recursive: true });
		}

		const env: NodeJS.ProcessEnv = {
			...process.env,
			WINEPREFIX: resolvedPrefix
		};

		if (protonPath) {
			const expandedProton = PathsService.expandPath(protonPath);
			env.PROTONPATH = expandedProton;
			env.UMU_PROTON_PATTERN = path.basename(expandedProton);
		}

		const hasSetsid = fsSync.existsSync("/usr/bin/setsid");
		const cmd = hasSetsid ? "/usr/bin/setsid" : "umu-run";
		const args = hasSetsid ? ["umu-run", toolName] : [toolName];

		try {
			const child = spawn(cmd, args, {
				env,
				detached: true,
				stdio: "ignore"
			});
			child.unref();
		} catch (err: any) {
			throw new Error(`Failed to launch prefix tool "${toolName}": ${err?.message || err}`);
		}
	}
}
