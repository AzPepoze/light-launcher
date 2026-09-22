import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import { app, BrowserWindow, dialog, nativeImage, shell } from "electron";
import { execFile } from "child_process";
import { promisify } from "util";
import { IconCacheMaxEntries, IconPngSize } from "../../shared/constants";
import type { IconCacheEntry } from "../../shared/types/system.types";

const execFileAsync = promisify(execFile);

let initialLauncherPath = "";
let initialGamePath = "";
let shouldEditLsfg = false;

const iconMemoryCache = new Map<string, IconCacheEntry>();

export class AppService {
	static setInitialArgs(launcherPath: string, gamePath: string, editLsfg: boolean) {
		initialLauncherPath = launcherPath;
		initialGamePath = gamePath;
		shouldEditLsfg = editLsfg;
	}

	static getInitialLauncherPath(): string {
		return initialLauncherPath || process.env.LIGHT_LAUNCHER_LAUNCHER_PATH || "";
	}

	static getInitialGamePath(): string {
		return initialGamePath || process.env.LIGHT_LAUNCHER_GAME_PATH || "";
	}

	static getShouldEditLsfg(): boolean {
		return shouldEditLsfg || process.env.LIGHT_LAUNCHER_EDIT_LSFG === "1";
	}

	static isDir(targetPath: string): boolean {
		try {
			const stat = fsSync.statSync(targetPath);
			return stat.isDirectory();
		} catch {
			return false;
		}
	}

	static async getExeIcon(executablePath: string): Promise<string> {
		const stat = AppService.readExeStat(executablePath);
		if (!stat) {
			return "";
		}

		const cached = iconMemoryCache.get(executablePath);
		if (cached && cached.mtimeMs === stat.mtimeMs && cached.size === stat.size) {
			return cached.icon;
		}

		const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), "light-launcher-icon-"));
		try {
			const icoPath = await AppService.extractIco(executablePath, tempDir);
			if (!icoPath) {
				return "";
			}
			const icon = AppService.icoToPngDataUrl(icoPath) ?? (await AppService.rawIcoDataUrl(icoPath));
			if (icon) {
				AppService.rememberIcon(executablePath, stat, icon);
			}
			return icon;
		} finally {
			try {
				await fs.rm(tempDir, { recursive: true, force: true });
			} catch {}
		}
	}

	private static readExeStat(executablePath: string): fsSync.Stats | null {
		try {
			const stat = fsSync.statSync(executablePath);
			return stat.isFile() ? stat : null;
		} catch {
			return null;
		}
	}

	private static async extractIco(executablePath: string, tempDir: string): Promise<string> {
		try {
			await execFileAsync("wrestool", ["-x", `--output=${tempDir}`, executablePath]);
			const files = await fs.readdir(tempDir);
			const icoFile = files.find((f) => f.toLowerCase().endsWith(".ico"));
			if (icoFile) {
				return path.join(tempDir, icoFile);
			}
		} catch {}

		try {
			const outIco = path.join(tempDir, "icon.ico");
			await execFileAsync("icoextract", [executablePath, outIco]);
			if (fsSync.existsSync(outIco)) {
				return outIco;
			}
		} catch {}
		return "";
	}

	// Small PNG instead of the full multi-image ICO, so scrolling ships kilobytes.
	private static icoToPngDataUrl(icoPath: string): string | null {
		try {
			const image = nativeImage.createFromBuffer(fsSync.readFileSync(icoPath));
			if (image.isEmpty()) {
				return null;
			}
			const { width, height } = image.getSize();
			const sized =
				Math.max(width, height) > IconPngSize
					? image.resize({ width: IconPngSize, height: IconPngSize })
					: image;
			const png = sized.toPNG();
			if (png.length === 0) {
				return null;
			}
			return `data:image/png;base64,${png.toString("base64")}`;
		} catch {
			return null;
		}
	}

	private static rememberIcon(executablePath: string, stat: fsSync.Stats, icon: string): void {
		if (iconMemoryCache.size >= IconCacheMaxEntries) {
			const oldest = iconMemoryCache.keys().next();
			if (!oldest.done) {
				iconMemoryCache.delete(oldest.value);
			}
		}
		iconMemoryCache.set(executablePath, { mtimeMs: stat.mtimeMs, size: stat.size, icon });
	}

	private static async rawIcoDataUrl(icoPath: string): Promise<string> {
		try {
			const data = await fs.readFile(icoPath);
			if (data.length > 0) {
				return `data:image/x-icon;base64,${data.toString("base64")}`;
			}
		} catch {}
		return "";
	}

	static async getTotalRam(): Promise<number> {
		try {
			const memInfo = await fs.readFile("/proc/meminfo", "utf-8");
			for (const line of memInfo.split("\n")) {
				if (line.startsWith("MemTotal:")) {
					const memKb = parseInt(line.replace(/\D/g, ""), 10);
					return Math.round(memKb / 1024 / 1024);
				}
			}
		} catch {}
		return Math.round(os.totalmem() / 1024 / 1024 / 1024);
	}

	static async getImageBase64(imagePath: string): Promise<string> {
		if (!fsSync.existsSync(imagePath)) {
			return "";
		}

		try {
			const data = await fs.readFile(imagePath);
			if (data.length === 0) return "";

			const ext = path.extname(imagePath).toLowerCase();
			let mimeType = "image/png";
			switch (ext) {
				case ".jpg":
				case ".jpeg":
					mimeType = "image/jpeg";
					break;
				case ".svg":
					mimeType = "image/svg+xml";
					break;
				case ".webp":
					mimeType = "image/webp";
					break;
				case ".ico":
					mimeType = "image/x-icon";
					break;
			}
			return `data:${mimeType};base64,${data.toString("base64")}`;
		} catch {
			return "";
		}
	}

	static async pickFile(): Promise<string> {
		const focusedWindow = BrowserWindow.getFocusedWindow() || undefined;
		const result = await dialog.showOpenDialog(focusedWindow!, {
			title: "Select Game Executable",
			properties: ["openFile"],
			filters: [
				{ name: "Executables (*.exe)", extensions: ["exe"] },
				{ name: "All Files", extensions: ["*"] }
			]
		});
		return result.canceled || result.filePaths.length === 0 ? "" : result.filePaths[0];
	}

	static async pickFolder(): Promise<string> {
		const focusedWindow = BrowserWindow.getFocusedWindow() || undefined;
		const result = await dialog.showOpenDialog(focusedWindow!, {
			title: "Select Directory",
			properties: ["openDirectory"]
		});
		return result.canceled || result.filePaths.length === 0 ? "" : result.filePaths[0];
	}

	static async pickFileCustom(
		title: string,
		filters: { displayName: string; pattern: string }[]
	): Promise<string> {
		const electronFilters = (filters || []).map((f) => ({
			name: f.displayName,
			extensions: f.pattern.split(";").map((p) => p.replace(/^\*\./, ""))
		}));

		const focusedWindow = BrowserWindow.getFocusedWindow() || undefined;
		const result = await dialog.showOpenDialog(focusedWindow!, {
			title,
			properties: ["openFile"],
			filters: electronFilters
		});
		return result.canceled || result.filePaths.length === 0 ? "" : result.filePaths[0];
	}

	static async openExternal(url: string): Promise<void> {
		if (url) {
			await shell.openExternal(url);
		}
	}

	static async openFileLocation(targetPath: string): Promise<void> {
		if (!targetPath) {
			throw new Error("No path provided");
		}

		const stat = await fs.stat(targetPath).catch(() => null);
		if (!stat) {
			throw new Error(`Path does not exist: ${targetPath}`);
		}

		if (stat.isDirectory()) {
			await shell.openPath(targetPath);
		} else {
			shell.showItemInFolder(targetPath);
		}
	}

	static closeWindow(): void {
		app.quit();
	}

	static restartApp(): void {
		app.relaunch();
		app.exit(0);
	}
}
