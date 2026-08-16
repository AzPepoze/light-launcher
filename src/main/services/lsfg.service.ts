import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import https from "https";
import { exec } from "child_process";
import { promisify } from "util";
import { ConfigService } from "./config.service";
import { PathsService } from "./paths.service";
import { SystemService } from "./system.service";
import type { LsfgProfileData } from "../../shared/types/lsfg.types";
import type { UtilsStatus } from "../../shared/types/system.types";

const execAsync = promisify(exec);

export class LsfgService {
	static getManifestPath(): string {
		return "/usr/share/vulkan/implicit_layer.d/lsfg-vk.json";
	}

	static getLibraryPath(): string {
		return "/usr/lib/liblsfg-vk-layer.so";
	}

	static getConfigPath(): string {
		return path.join(os.homedir(), ".config/lsfg-vk/lsfg-vk.toml");
	}

	static isInstalled(): boolean {
		return fsSync.existsSync(this.getManifestPath()) || fsSync.existsSync(this.getLibraryPath());
	}

	static getUtilsStatus(): UtilsStatus {
		const installed = this.isInstalled();
		return {
			isLsfgInstalled: installed,
			lsfgVersion: installed ? "Installed" : ""
		};
	}

	static detectLosslessDll(): string {
		const home = os.homedir();
		const candidates = [
			path.join(home, ".steam/root/steamapps/common/Lossless Scaling/Lossless.dll"),
			path.join(home, ".local/share/Steam/steamapps/common/Lossless Scaling/Lossless.dll"),
			path.join(
				home,
				".var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps/common/Lossless Scaling/Lossless.dll"
			)
		];

		for (const candidate of candidates) {
			if (fsSync.existsSync(candidate)) {
				return candidate;
			}
		}
		return "";
	}

	static async getLsfgProfileForGame(
		name: string,
		gamePath: string
	): Promise<LsfgProfileData | null> {
		const cfg = await ConfigService.loadGameConfig(gamePath);
		const id = cfg ? cfg.ID : "";
		const gameLsfgPath = PathsService.getGameLsfgConfigPath(name, id);

		let dllPath = "";
		let allowFp16 = false;

		const globalConfigPath = this.getConfigPath();
		if (fsSync.existsSync(globalConfigPath)) {
			try {
				const content = await fs.readFile(globalConfigPath, "utf-8");
				const dllMatch = content.match(/dll\s*=\s*"([^"]+)"/i);
				if (dllMatch) dllPath = dllMatch[1];
				const fp16Match = content.match(/allow_fp16\s*=\s*(true|false)/i);
				if (fp16Match) allowFp16 = fp16Match[1].toLowerCase() === "true";
			} catch {}
		}

		if (fsSync.existsSync(gameLsfgPath)) {
			try {
				const content = await fs.readFile(gameLsfgPath, "utf-8");
				const multMatch = content.match(/multiplier\s*=\s*(\d+)/i);
				const perfMatch = content.match(/performance_mode\s*=\s*(true|false)/i);
				const gpuMatch = content.match(/gpu\s*=\s*"([^"]+)"/i);
				const flowMatch = content.match(/flow_scale\s*=\s*([\d.]+)/i);
				const pacingMatch = content.match(/pacing\s*=\s*"([^"]+)"/i);

				return {
					name: name || path.parse(path.basename(gamePath)).name,
					multiplier: multMatch ? parseInt(multMatch[1], 10) : 2,
					performanceMode: perfMatch ? perfMatch[1].toLowerCase() === "true" : false,
					gpu: gpuMatch ? gpuMatch[1] : "",
					flowScale: flowMatch ? parseFloat(flowMatch[1]) : 1.0,
					pacing: pacingMatch ? pacingMatch[1] : "smooth",
					dllPath,
					allowFp16
				};
			} catch {}
		}

		return null;
	}

	static async saveLsfgProfile(
		profileName: string,
		gamePath: string,
		multiplier: number,
		performanceMode: boolean,
		dllPath: string,
		gpu: string,
		flowScale: string,
		pacing: string,
		allowFp16: boolean
	): Promise<void> {
		const cfg = await ConfigService.loadGameConfig(gamePath);
		const id = cfg ? cfg.ID : "";
		const gameLsfgPath = PathsService.getGameLsfgConfigPath(profileName, id);

		if (!gpu) {
			const gpus = SystemService.getListGpus();
			if (gpus.length > 0) gpu = gpus[0];
		}

		await fs.mkdir(path.dirname(gameLsfgPath), { recursive: true });

		const tomlContent = `[profile]
name = "${profileName}"
multiplier = ${multiplier || 2}
performance_mode = ${performanceMode ? "true" : "false"}
gpu = "${gpu || ""}"
flow_scale = ${parseFloat(flowScale) || 1.0}
pacing = "${pacing || "smooth"}"
allow_fp16 = ${allowFp16 ? "true" : "false"}
`;

		await fs.writeFile(gameLsfgPath, tomlContent, "utf-8");

		// Sync global config with dllPath & allowFp16
		const globalConfigPath = this.getConfigPath();
		await fs.mkdir(path.dirname(globalConfigPath), { recursive: true });

		let globalContent = `[global]
dll = "${dllPath || ""}"
allow_fp16 = ${allowFp16 ? "true" : "false"}
`;
		await fs.writeFile(globalConfigPath, globalContent, "utf-8");
	}

	static async disableLsfgProfile(profileName: string, gamePath: string): Promise<void> {
		const cfg = await ConfigService.loadGameConfig(gamePath);
		const id = cfg ? cfg.ID : "";
		const gameLsfgPath = PathsService.getGameLsfgConfigPath(profileName, id);

		if (fsSync.existsSync(gameLsfgPath)) {
			let content = await fs.readFile(gameLsfgPath, "utf-8");
			if (!content.includes("disable = true")) {
				content += "\ndisable = true\n";
				await fs.writeFile(gameLsfgPath, content, "utf-8");
			}
		}
	}

	static async removeProfile(mainExecutablePath: string): Promise<void> {
		const cfg = await ConfigService.loadGameConfig(mainExecutablePath);
		if (cfg) {
			const gameLsfgPath = PathsService.getGameLsfgConfigPath(cfg.Name, cfg.ID);
			if (fsSync.existsSync(gameLsfgPath)) {
				await fs.rm(gameLsfgPath, { force: true });
			}
		}
	}

	static async installLsfg(onProgress: (percent: number, message: string) => void): Promise<void> {
		onProgress(0, "Fetching release info from GitHub...");

		const releasesUrl = "https://api.github.com/repos/Pumboo/lsfg-vk/releases";

		const releaseData: any = await new Promise((resolve, reject) => {
			https
				.get(releasesUrl, { headers: { "User-Agent": "LightLauncher-App" } }, (res) => {
					if (res.statusCode !== 200)
						return reject(new Error(`GitHub API error: ${res.statusCode}`));
					let body = "";
					res.on("data", (chunk) => (body += chunk));
					res.on("end", () => resolve(JSON.parse(body)));
				})
				.on("error", reject);
		});

		if (!releaseData || releaseData.length === 0) {
			throw new Error("No releases found for lsfg-vk");
		}

		let downloadURL = "";
		let assetName = "";
		for (const release of releaseData) {
			for (const asset of release.assets || []) {
				const name = asset.name.toLowerCase();
				if (
					(name.includes("x86_64") && name.endsWith(".tar.zst")) ||
					(name.includes("linux") && name.endsWith(".tar.xz"))
				) {
					downloadURL = asset.browser_download_url;
					assetName = asset.name;
					break;
				}
			}
			if (downloadURL) break;
		}

		if (!downloadURL) {
			throw new Error("No compatible Linux package found for lsfg-vk");
		}

		onProgress(10, `Downloading ${assetName}...`);
		const ext = assetName.endsWith(".tar.zst") ? ".tar.zst" : ".tar.xz";
		const tempFile = path.join(os.tmpdir(), `lsfg-vk-dl${ext}`);

		await new Promise<void>((resolve, reject) => {
			https
				.get(downloadURL, { headers: { "User-Agent": "LightLauncher-App" } }, (res) => {
					if (res.statusCode === 302 || res.statusCode === 301) {
						if (res.headers.location) {
							return https.get(
								res.headers.location,
								{ headers: { "User-Agent": "LightLauncher-App" } },
								(res2) => {
									const fileStream = fsSync.createWriteStream(tempFile);
									res2.pipe(fileStream);
									fileStream.on("finish", () => resolve());
									fileStream.on("error", reject);
								}
							);
						}
					}
					const fileStream = fsSync.createWriteStream(tempFile);
					res.pipe(fileStream);
					fileStream.on("finish", () => resolve());
					fileStream.on("error", reject);
				})
				.on("error", reject);
		});

		onProgress(85, "Installing to system directories (requires sudo)...");

		let extractCmd = `tar -xf "${tempFile}" -C /usr`;
		if (tempFile.endsWith(".tar.zst")) {
			extractCmd = `tar --use-compress-program=unzstd -xf "${tempFile}" -C /usr`;
		}

		await execAsync(`pkexec sh -c "${extractCmd}"`);
		try {
			await fs.unlink(tempFile);
		} catch {}

		onProgress(100, "Installation complete!");
	}

	static async uninstallLsfg(): Promise<void> {
		const manifest = this.getManifestPath();
		const lib = this.getLibraryPath();
		await execAsync(`pkexec rm -f "${manifest}" "${lib}"`);
	}
}
