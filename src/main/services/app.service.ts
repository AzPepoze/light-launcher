import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import { app, BrowserWindow, dialog, shell } from "electron";
import { exec, execSync } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

let initialLauncherPath = "";
let initialGamePath = "";
let shouldEditLsfg = false;

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
		if (!executablePath || !fsSync.existsSync(executablePath)) {
			return "";
		}

		const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), "light-launcher-icon-"));
		try {
			// Try wrestool
			try {
				const { execFile } = require("child_process");
				const execFileAsync = promisify(execFile);
				await execFileAsync("wrestool", ["-x", `--output=${tempDir}`, executablePath]);
				const files = await fs.readdir(tempDir);
				const icoFile = files.find((f) => f.toLowerCase().endsWith(".ico"));
				if (icoFile) {
					const data = await fs.readFile(path.join(tempDir, icoFile));
					if (data.length > 0) {
						return `data:image/x-icon;base64,${data.toString("base64")}`;
					}
				}
			} catch {}

			// Try icoextract
			try {
				const { execFile } = require("child_process");
				const execFileAsync = promisify(execFile);
				const outIco = path.join(tempDir, "icon.ico");
				await execFileAsync("icoextract", [executablePath, outIco]);
				if (fsSync.existsSync(outIco)) {
					const data = await fs.readFile(outIco);
					if (data.length > 0) {
						return `data:image/x-icon;base64,${data.toString("base64")}`;
					}
				}
			} catch {}
		} finally {
			try {
				await fs.rm(tempDir, { recursive: true, force: true });
			} catch {}
		}

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

	static closeWindow(): void {
		app.quit();
	}

	static restartApp(): void {
		app.relaunch();
		app.exit(0);
	}
}
