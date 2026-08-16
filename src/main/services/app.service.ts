import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import os from "os";
import { app, dialog, shell } from "electron";
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
		if (!fsSync.existsSync(executablePath)) {
			return "";
		}

		const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), "light-launcher-icon-"));
		try {
			// Try wrestool
			try {
				await execAsync(`wrestool -x --output="${tempDir}" "${executablePath}"`);
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
				const outIco = path.join(tempDir, "icon.ico");
				await execAsync(`icoextract "${executablePath}" "${outIco}"`);
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

	static async runSystemPicker(
		title: string,
		isFolder: boolean,
		filters?: { displayName: string; pattern: string }[]
	): Promise<string | null> {
		// Zenity
		try {
			execSync("which zenity", { stdio: "ignore" });
			const args = ["--file-selection", `--title="${title}"`];
			if (isFolder) args.push("--directory");
			if (filters && filters.length > 0) {
				for (const f of filters) {
					args.push(`--file-filter="${f.displayName}|${f.pattern.replace(/;/g, " ")}"`);
				}
			}
			const { stdout } = await execAsync(`zenity ${args.join(" ")}`);
			return stdout.trim() || null;
		} catch {}

		// Kdialog
		try {
			execSync("which kdialog", { stdio: "ignore" });
			if (isFolder) {
				const { stdout } = await execAsync(`kdialog --getexistingdirectory . --title "${title}"`);
				return stdout.trim() || null;
			} else {
				let filterStr = "";
				if (filters && filters.length > 0) {
					filterStr = filters
						.map((f) => `${f.displayName} (${f.pattern.replace(/;/g, " ")})`)
						.join(";;");
				}
				const { stdout } = await execAsync(
					`kdialog --getopenfilename . "${filterStr}" --title "${title}"`
				);
				return stdout.trim() || null;
			}
		} catch {}

		return null;
	}

	static async pickFile(): Promise<string> {
		const sysResult = await this.runSystemPicker("Select Game Executable", false, [
			{ displayName: "Executables (*.exe)", pattern: "*.exe" },
			{ displayName: "All Files", pattern: "*.*" }
		]);
		if (sysResult !== null) return sysResult;

		const result = await dialog.showOpenDialog({
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
		const sysResult = await this.runSystemPicker("Select Directory", true);
		if (sysResult !== null) return sysResult;

		const result = await dialog.showOpenDialog({
			title: "Select Directory",
			properties: ["openDirectory"]
		});
		return result.canceled || result.filePaths.length === 0 ? "" : result.filePaths[0];
	}

	static async pickFileCustom(
		title: string,
		filters: { displayName: string; pattern: string }[]
	): Promise<string> {
		const sysResult = await this.runSystemPicker(title, false, filters);
		if (sysResult !== null) return sysResult;

		const electronFilters = (filters || []).map((f) => ({
			name: f.displayName,
			extensions: f.pattern.split(";").map((p) => p.replace(/^\*\./, ""))
		}));

		const result = await dialog.showOpenDialog({
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
