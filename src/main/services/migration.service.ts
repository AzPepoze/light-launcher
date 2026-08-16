import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import { PathsService } from "./paths.service";
import { LoggerService } from "./logger.service";

export class MigrationService {
	static getLegacyDirectory(): string {
		return path.join(os.homedir(), "LightLauncher");
	}

	private static async moveOrCopy(src: string, dest: string): Promise<void> {
		try {
			await fs.rename(src, dest);
		} catch {
			await fs.cp(src, dest, { recursive: true });
			await fs.rm(src, { recursive: true, force: true }).catch(() => {});
		}
	}

	static async migrateIfNeeded(): Promise<boolean> {
		const legacyDir = this.getLegacyDirectory();
		if (!fsSync.existsSync(legacyDir)) {
			return false;
		}

		const targetBaseDir = PathsService.getBaseDirectory();
		await fs.mkdir(targetBaseDir, { recursive: true });

		try {
			// 1. Migrate settings.json if target doesn't have it
			const legacySettings = path.join(legacyDir, "settings.json");
			const targetSettings = PathsService.getAppSettingsPath();
			if (fsSync.existsSync(legacySettings)) {
				if (!fsSync.existsSync(targetSettings)) {
					await this.moveOrCopy(legacySettings, targetSettings);
				} else {
					await fs.rm(legacySettings).catch(() => {});
				}
			}

			// 2. Migrate executable configs
			const targetConfigDir = PathsService.getConfigDirectory();
			await fs.mkdir(targetConfigDir, { recursive: true });

			const legacyConfigDirs = [
				path.join(legacyDir, "config", "executables"),
				path.join(legacyDir, "games")
			];

			for (const oldDir of legacyConfigDirs) {
				if (fsSync.existsSync(oldDir)) {
					const entries = await fs.readdir(oldDir, { withFileTypes: true });
					for (const entry of entries) {
						const src = path.join(oldDir, entry.name);
						const dest = path.join(targetConfigDir, entry.name);
						if (!fsSync.existsSync(dest)) {
							await this.moveOrCopy(src, dest);
						}
					}
					await fs.rm(oldDir, { recursive: true, force: true }).catch(() => {});
				}
			}

			// 3. Migrate prefixes (instant directory moves)
			const legacyPrefixesDir = path.join(legacyDir, "prefixes");
			const targetPrefixesDir = PathsService.getPrefixBaseDirectory();
			if (fsSync.existsSync(legacyPrefixesDir)) {
				await fs.mkdir(targetPrefixesDir, { recursive: true });
				const entries = await fs.readdir(legacyPrefixesDir, { withFileTypes: true });
				for (const entry of entries) {
					const src = path.join(legacyPrefixesDir, entry.name);
					const dest = path.join(targetPrefixesDir, entry.name);
					if (!fsSync.existsSync(dest)) {
						await this.moveOrCopy(src, dest);
					}
				}
				await fs.rm(legacyPrefixesDir, { recursive: true, force: true }).catch(() => {});
			}

			// 4. Migrate logs
			const legacyLogsDir = path.join(legacyDir, "logs");
			const targetLogsDir = PathsService.getLogsDirectory();
			if (fsSync.existsSync(legacyLogsDir)) {
				await fs.mkdir(targetLogsDir, { recursive: true });
				const entries = await fs.readdir(legacyLogsDir, { withFileTypes: true });
				for (const entry of entries) {
					const src = path.join(legacyLogsDir, entry.name);
					const dest = path.join(targetLogsDir, entry.name);
					if (!fsSync.existsSync(dest)) {
						await this.moveOrCopy(src, dest);
					}
				}
				await fs.rm(legacyLogsDir, { recursive: true, force: true }).catch(() => {});
			}

			// Clean up remaining empty legacy root directory or debug files
			await fs.rm(legacyDir, { recursive: true, force: true }).catch(() => {});

			await this.sanitizeExistingConfigs();

			return true;
		} catch (error) {
			LoggerService.error("Migration", `Failed to migrate legacy LightLauncher data: ${error}`);
			return false;
		}
	}

	static async sanitizeExistingConfigs(): Promise<void> {
		const configDir = PathsService.getConfigDirectory();
		if (!fsSync.existsSync(configDir)) return;

		try {
			const entries = await fs.readdir(configDir, { withFileTypes: true });
			const basePrefixDir = PathsService.getPrefixBaseDirectory();

			for (const entry of entries) {
				if (!entry.isDirectory()) continue;
				const configPath = path.join(configDir, entry.name, "config.json");
				if (fsSync.existsSync(configPath)) {
					try {
						const raw = await fs.readFile(configPath, "utf-8");
						if (!raw || raw.trim() === "") continue;
						const cfg = JSON.parse(raw);

						let changed = false;
						if (cfg.PrefixPath) {
							if (cfg.PrefixPath.includes("/LightLauncher/prefixes/")) {
								const prefixName = path.basename(cfg.PrefixPath);
								cfg.PrefixPath = path.join(basePrefixDir, prefixName);
								changed = true;
							}
						}

						if (changed) {
							await fs.writeFile(configPath, JSON.stringify(cfg, null, 2), "utf-8");
						}
					} catch {}
				}
			}

			await this.repairPrefixSymlinks();
		} catch {}
	}

	static async repairPrefixSymlinks(): Promise<void> {
		const basePrefixDir = PathsService.getPrefixBaseDirectory();
		if (!fsSync.existsSync(basePrefixDir)) return;

		try {
			const prefixEntries = await fs.readdir(basePrefixDir, { withFileTypes: true });
			for (const pfx of prefixEntries) {
				if (!pfx.isDirectory()) continue;
				const pfxPath = path.join(basePrefixDir, pfx.name);
				await this.walkAndRepairSymlinks(pfxPath, basePrefixDir);
			}
		} catch (err) {
			LoggerService.error("Migration", `Failed to repair prefix symlinks: ${err}`);
		}
	}

	private static async walkAndRepairSymlinks(dir: string, basePrefixDir: string): Promise<void> {
		try {
			const entries = await fs.readdir(dir, { withFileTypes: true });
			for (const entry of entries) {
				const fullPath = path.join(dir, entry.name);
				try {
					if (entry.isSymbolicLink()) {
						const linkTarget = await fs.readlink(fullPath);
						if (linkTarget.includes("/LightLauncher/prefixes/")) {
							const fixedTarget = linkTarget.replace(
								/^.*\/LightLauncher\/prefixes\//,
								`${basePrefixDir}/`
							);
							await fs.unlink(fullPath);
							await fs.symlink(fixedTarget, fullPath);
						}
					} else if (entry.isDirectory()) {
						await this.walkAndRepairSymlinks(fullPath, basePrefixDir);
					}
				} catch {}
			}
		} catch {}
	}
}
