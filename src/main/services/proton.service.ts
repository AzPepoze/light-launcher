import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import http from "http";
import https from "https";
import { exec } from "child_process";
import { promisify } from "util";
import type { GitHubRelease, ProtonTool, ProtonVariant } from "../../shared/types/prefix.types";

const execAsync = promisify(exec);

const KNOWN_VARIANTS: ProtonVariant[] = [
	{
		ID: "ge-proton",
		Name: "GE-Proton (GloriousEggroll)",
		Description:
			"The most popular custom Proton build. Includes many game fixes and codec patches.",
		RepoOwner: "GloriousEggroll",
		RepoName: "proton-ge-custom"
	},
	{
		ID: "proton-cachyos",
		Name: "Proton-CachyOS",
		Description: "Optimized for performance with CachyOS patches and schedulers.",
		RepoOwner: "CachyOS",
		RepoName: "proton-cachyos"
	},
	{
		ID: "kron4ek",
		Name: "Proton-Kron4ek",
		Description: "Vanilla builds and TKG builds. Often smaller and faster updates.",
		RepoOwner: "Kron4ek",
		RepoName: "Proton-Builds"
	},
	{
		ID: "luxtorpeda",
		Name: "Luxtorpeda (Native Tools)",
		Description: "Runs Windows games using native Linux engines (e.g. GZDoom, ScummVM).",
		RepoOwner: "luxtorpeda-dev",
		RepoName: "luxtorpeda"
	}
];

export class ProtonService {
	static getProtonVariants(): ProtonVariant[] {
		return KNOWN_VARIANTS;
	}

	static async scanProtonVersions(): Promise<ProtonTool[]> {
		const tools: ProtonTool[] = [];
		const seenPaths = new Set<string>();
		const home = os.homedir();

		const steamCommonPaths = [
			path.join(home, ".steam/root/steamapps/common"),
			path.join(home, ".local/share/Steam/steamapps/common"),
			path.join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam/steamapps/common")
		];

		const customToolsPaths = [
			path.join(home, ".steam/root/compatibilitytools.d"),
			path.join(home, ".local/share/Steam/compatibilitytools.d"),
			path.join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam/compatibilitytools.d")
		];

		// 1. Scan Steam official Proton installations
		for (const basePath of steamCommonPaths) {
			if (!fsSync.existsSync(basePath)) continue;
			try {
				const entries = await fs.readdir(basePath, { withFileTypes: true });
				for (const entry of entries) {
					if (entry.isDirectory() && entry.name.toLowerCase().startsWith("proton")) {
						const protonPath = path.join(basePath, entry.name);
						const realP = fsSync.existsSync(protonPath)
							? fsSync.realpathSync(protonPath)
							: protonPath;
						if (seenPaths.has(realP)) continue;

						const protonExe = path.join(protonPath, "proton");
						if (fsSync.existsSync(protonExe)) {
							seenPaths.add(realP);
							tools.push({
								Name: entry.name,
								Path: protonPath,
								IsSteam: true,
								DisplayName: `Steam: ${entry.name}`
							});
						}
					}
				}
			} catch {}
		}

		// 2. Scan custom compatibility tools
		for (const basePath of customToolsPaths) {
			if (!fsSync.existsSync(basePath)) continue;
			try {
				const entries = await fs.readdir(basePath, { withFileTypes: true });
				for (const entry of entries) {
					if (entry.isDirectory()) {
						const protonPath = path.join(basePath, entry.name);
						const realP = fsSync.existsSync(protonPath)
							? fsSync.realpathSync(protonPath)
							: protonPath;
						if (seenPaths.has(realP)) continue;

						const protonExe = path.join(protonPath, "proton");
						if (fsSync.existsSync(protonExe)) {
							seenPaths.add(realP);
							tools.push({
								Name: entry.name,
								Path: protonPath,
								IsSteam: false,
								DisplayName: `Custom: ${entry.name}`
							});
						}
					}
				}
			} catch {}
		}

		return tools;
	}

	static findProtonMatch(savedPath: string, protonVersions: ProtonTool[]): ProtonTool | null {
		if (!savedPath) return null;

		// 1. Match by DisplayName
		for (const tool of protonVersions) {
			if (tool.DisplayName === savedPath) return tool;
		}

		// 2. Match by exact Path
		for (const tool of protonVersions) {
			if (tool.Path === savedPath) return tool;
		}

		// 3. Match by directory name
		const dirName = path.basename(path.normalize(savedPath));
		const isSteam = savedPath.includes("steamapps/common");

		for (const tool of protonVersions) {
			if (tool.Name === dirName && tool.IsSteam === isSteam) {
				return tool;
			}
		}

		for (const tool of protonVersions) {
			if (tool.Name === dirName) {
				return tool;
			}
		}

		return null;
	}

	static async getProtonReleases(variantID: string): Promise<GitHubRelease[]> {
		const variant = KNOWN_VARIANTS.find((v) => v.ID === variantID);
		if (!variant) {
			throw new Error(`Unknown variant: ${variantID}`);
		}

		const url = `https://api.github.com/repos/${variant.RepoOwner}/${variant.RepoName}/releases?per_page=50`;

		return new Promise((resolve, reject) => {
			const req = https.get(
				url,
				{
					headers: {
						"User-Agent": "LightLauncher-App",
						Accept: "application/vnd.github.v3+json"
					},
					timeout: 15000
				},
				(res) => {
					if (res.statusCode === 403 || res.statusCode === 429) {
						reject(
							new Error(
								"GitHub API rate limit exceeded. Please wait a few minutes before trying again."
							)
						);
						return;
					}
					if (res.statusCode !== 200) {
						reject(new Error(`GitHub API returned HTTP ${res.statusCode}`));
						return;
					}

					let body = "";
					res.on("data", (chunk) => (body += chunk));
					res.on("end", () => {
						try {
							const releases = JSON.parse(body) as GitHubRelease[];
							resolve(releases);
						} catch (e) {
							reject(new Error("Failed to parse GitHub releases response"));
						}
					});
				}
			);
			req.on("timeout", () => {
				req.destroy();
				reject(new Error("Request timed out connecting to GitHub"));
			});
			req.on("error", (err) => {
				reject(new Error(`Network error: ${err.message}`));
			});
		});
	}

	static async installProtonVersion(
		url: string,
		versionTag: string,
		onProgress: (percent: number, message: string) => void
	): Promise<void> {
		const home = os.homedir();
		let targetBase = path.join(home, ".steam/root/compatibilitytools.d");
		if (!fsSync.existsSync(path.join(home, ".steam/root"))) {
			targetBase = path.join(home, ".local/share/Steam/compatibilitytools.d");
		}

		await fs.mkdir(targetBase, { recursive: true });
		onProgress(0, "Downloading...");

		const ext = url.endsWith(".tar.zst") ? ".tar.zst" : ".tar.gz";
		const tempFile = path.join(os.tmpdir(), `proton-install-${Date.now()}${ext}`);

		try {
			await new Promise<void>((resolve, reject) => {
				const getter = url.startsWith("https") ? https : http;
				const req = getter.get(
					url,
					{
						headers: { "User-Agent": "LightLauncher-App" },
						timeout: 30000
					},
					(res) => {
						if (res.statusCode === 302 || res.statusCode === 301) {
							if (res.headers.location) {
								return this.installProtonVersion(res.headers.location, versionTag, onProgress)
									.then(resolve)
									.catch(reject);
							}
						}

						if (res.statusCode !== 200) {
							reject(new Error(`Download failed with status ${res.statusCode}`));
							return;
						}

						const total = parseInt(res.headers["content-length"] || "0", 10);
						let current = 0;

						const fileStream = fsSync.createWriteStream(tempFile);
						fileStream.on("error", reject);

						res.on("data", (chunk) => {
							current += chunk.length;
							fileStream.write(chunk);
							if (total > 0) {
								const percent = Math.round((current / total) * 50);
								onProgress(
									percent,
									`Downloading... ${(current / 1024 / 1024).toFixed(1)} MB / ${(total / 1024 / 1024).toFixed(1)} MB`
								);
							}
						});

						res.on("end", () => {
							fileStream.end();
							resolve();
						});

						res.on("error", reject);
					}
				);
				req.on("timeout", () => {
					req.destroy();
					reject(new Error("Download request timed out"));
				});
				req.on("error", reject);
			});

			onProgress(50, "Extracting Proton package...");

			let extractCmd = `tar -xf "${tempFile}" -C "${targetBase}"`;
			if (tempFile.endsWith(".tar.zst")) {
				extractCmd = `tar --use-compress-program=unzstd -xf "${tempFile}" -C "${targetBase}"`;
			}

			await execAsync(extractCmd);
			onProgress(100, "Installation Complete!");
		} catch (error: any) {
			throw new Error(`Failed to install Proton ${versionTag}: ${error?.message || error}`);
		} finally {
			try {
				if (fsSync.existsSync(tempFile)) {
					await fs.unlink(tempFile);
				}
			} catch {}
		}
	}
}
