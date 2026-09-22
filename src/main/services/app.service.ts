import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import { app, BrowserWindow, dialog, shell } from "electron";
import { execFile } from "child_process";
import { promisify } from "util";
import { IconCacheMaxEntries, IconPngSize } from "../../shared/constants";
import { IconCacheService } from "./iconCache.service";
import { LoggerService } from "./logger.service";
import type { IconCacheEntry } from "../../shared/types/system.types";

const execFileAsync = promisify(execFile);

let initialLauncherPath = "";
let initialGamePath = "";
let shouldEditLsfg = false;

const iconMemoryCache = new Map<string, IconCacheEntry>();

interface ToolCandidate<T extends string> {
	tool: T;
	probe: string[];
}

function createToolResolver<T extends string>(candidates: readonly ToolCandidate<T>[]) {
	let resolved: T | null | undefined;
	return {
		async get(): Promise<T | null> {
			if (resolved !== undefined) return resolved;
			for (const { tool, probe } of candidates) {
				try {
					await execFileAsync(tool, probe);
					resolved = tool;
					return tool;
				} catch {}
			}
			resolved = null;
			return null;
		},
		reset(): void {
			resolved = undefined;
		}
	};
}

const extractorResolver = createToolResolver([
	{ tool: "wrestool", probe: ["--version"] },
	{ tool: "icoextract", probe: ["--help"] }
] as const);

const converterResolver = createToolResolver([
	{ tool: "magick", probe: ["-version"] },
	{ tool: "convert", probe: ["-version"] }
] as const);

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
		const name = path.basename(executablePath);

		const cached = iconMemoryCache.get(executablePath);
		if (cached && cached.mtimeMs === stat.mtimeMs && cached.size === stat.size) {
			LoggerService.debug("Icon", `hit memory ${name}`);
			return cached.icon;
		}

		const diskIcon = await IconCacheService.get(executablePath, stat);
		if (diskIcon !== null) {
			if (diskIcon) AppService.rememberIcon(executablePath, stat, diskIcon);
			LoggerService.debug("Icon", diskIcon ? `hit disk ${name}` : `hit disk (no icon) ${name}`);
			return diskIcon;
		}

		LoggerService.debug("Icon", `miss ${name}`);
		const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), "light-launcher-icon-"));
		try {
			const icoPath = await AppService.extractIco(executablePath, tempDir);
			if (!icoPath) {
				AppService.rememberIcon(executablePath, stat, "");
				await IconCacheService.markMiss(executablePath, stat);
				LoggerService.debug("Icon", `no icon ${name}`);
				return "";
			}
			const icon = await AppService.buildAndCacheIcon(executablePath, stat, icoPath, tempDir);
			LoggerService.debug("Icon", `extracted ${name} (${icon.length} bytes)`);
			return icon;
		} finally {
			try {
				await fs.rm(tempDir, { recursive: true, force: true });
			} catch {}
		}
	}

	private static async buildAndCacheIcon(
		executablePath: string,
		stat: fsSync.Stats,
		icoPath: string,
		tempDir: string
	): Promise<string> {
		const ico = await fs.readFile(icoPath);
		const png = await extractIconPng(ico, icoPath, tempDir);
		const bytes = png ?? ico;
		const icon = `data:${png ? "image/png" : "image/x-icon"};base64,${bytes.toString("base64")}`;
		AppService.rememberIcon(executablePath, stat, icon);
		await IconCacheService.put(executablePath, stat, bytes, png ? "png" : "ico");
		return icon;
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
		const extractor = await extractorResolver.get();
		if (!extractor) {
			return "";
		}

		try {
			if (extractor === "wrestool") {
				await execFileAsync("wrestool", ["-x", `--output=${tempDir}`, executablePath]);
				const files = await fs.readdir(tempDir);
				const icoFile = files.find((f) => f.toLowerCase().endsWith(".ico"));
				if (icoFile) {
					return path.join(tempDir, icoFile);
				}
			} else {
				const outIco = path.join(tempDir, "icon.ico");
				await execFileAsync("icoextract", [executablePath, outIco]);
				if (fsSync.existsSync(outIco)) {
					return outIco;
				}
			}
		} catch {
			// Tool failed at runtime; re-probe on the next call.
			extractorResolver.reset();
		}
		return "";
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

interface IcoFrame {
	index: number;
	width: number;
	height: number;
	size: number;
	offset: number;
	png: boolean;
}

function parseIcoFrames(ico: Buffer): IcoFrame[] {
	if (ico.length < 6 || ico.readUInt16LE(0) !== 0 || ico.readUInt16LE(2) !== 1) return [];
	const count = ico.readUInt16LE(4);
	const pngSignature = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
	const frames: IcoFrame[] = [];

	for (let i = 0; i < count; i++) {
		const entry = 6 + i * 16;
		if (entry + 16 > ico.length) break;
		const size = ico.readUInt32LE(entry + 8);
		const offset = ico.readUInt32LE(entry + 12);
		if (size < 8 || offset + size > ico.length) continue;
		frames.push({
			index: i,
			width: ico[entry] || 256,
			height: ico[entry + 1] || 256,
			size,
			offset,
			png: ico.subarray(offset, offset + 8).equals(pngSignature)
		});
	}
	return frames;
}

/** Largest ICO frame as PNG; BMP frames are converted, else null to fall back to the raw ICO. */
async function extractIconPng(ico: Buffer, icoPath: string, tempDir: string): Promise<Buffer | null> {
	const frames = parseIcoFrames(ico);
	if (!frames.length) return null;

	const largest = frames.reduce((a, b) => (b.width * b.height > a.width * a.height ? b : a));
	if (largest.png) {
		return Buffer.from(ico.subarray(largest.offset, largest.offset + largest.size));
	}

	const converter = await converterResolver.get();
	if (!converter) return null;
	const outPng = path.join(tempDir, "icon.png");
	try {
		await execFileAsync(converter, [
			`${icoPath}[${largest.index}]`,
			"-resize",
			`${IconPngSize}x${IconPngSize}>`,
			`PNG32:${outPng}`
		]);
		const png = await fs.readFile(outPng);
		return png.length ? png : null;
	} catch {
		converterResolver.reset();
		return null;
	}
}
