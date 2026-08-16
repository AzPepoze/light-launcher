import fsSync from "fs";
import os from "os";
import path from "path";
import { AppName, ConfigDirName, LogsDirName, PrefixesDirName } from "../../shared/constants";

export class PathsService {
	static getBaseDirectory(): string {
		const xdgConfigHome = process.env.XDG_CONFIG_HOME;
		if (xdgConfigHome) {
			return path.join(xdgConfigHome, AppName);
		}
		return path.join(os.homedir(), ".config", AppName);
	}

	static getConfigDirectory(): string {
		return path.join(this.getBaseDirectory(), ConfigDirName);
	}

	static getPrefixBaseDirectory(customDir?: string): string {
		if (customDir && customDir.trim() !== "") {
			return this.expandPath(customDir.trim());
		}
		return path.join(this.getBaseDirectory(), PrefixesDirName);
	}

	static getLogsDirectory(): string {
		return path.join(this.getBaseDirectory(), LogsDirName);
	}

	static getAppSettingsPath(): string {
		return path.join(this.getBaseDirectory(), "settings.json");
	}

	static getFlagsFilePath(): string {
		const xdgConfigHome = process.env.XDG_CONFIG_HOME;
		const base = xdgConfigHome || path.join(os.homedir(), ".config");
		return path.join(base, `${AppName}-flags.conf`);
	}

	static getExecutableConfigPath(name: string, id: string): string {
		if (id) {
			return path.join(this.getConfigDirectory(), id);
		}
		return path.join(this.getConfigDirectory(), name);
	}

	static getGameConfigFilePath(name: string, id: string): string {
		return path.join(this.getExecutableConfigPath(name, id), "config.json");
	}

	static getGameLsfgConfigPath(name: string, id: string): string {
		return path.join(this.getExecutableConfigPath(name, id), "lsfg_vk.toml");
	}

	static getPrefixConfigPath(prefixName: string, customDir?: string): string {
		return path.join(this.getPrefixBaseDirectory(customDir), prefixName, "light-launcher.json");
	}

	static getPreloadPath(): string {
		const candidates = [
			path.resolve(__dirname, "../../preload/index.js"),
			path.resolve(__dirname, "../preload/index.js")
		];
		return candidates.find((p) => fsSync.existsSync(p)) || candidates[0];
	}

	static getRendererPath(): string {
		const candidates = [
			path.resolve(__dirname, "../../renderer/index.html"),
			path.resolve(__dirname, "../../../src/renderer/dist/index.html"),
			path.resolve(__dirname, "../renderer/index.html")
		];
		return candidates.find((p) => fsSync.existsSync(p)) || candidates[0];
	}

	static safeExists(targetPath: string): boolean {
		if (!targetPath) return false;
		try {
			return fsSync.existsSync(this.expandPath(targetPath));
		} catch {
			return false;
		}
	}

	static expandPath(targetPath: string): string {
		if (!targetPath || typeof targetPath !== "string") return "";
		const trimmed = targetPath.trim();
		if (trimmed === "") return "";
		if (trimmed === "~") {
			return os.homedir();
		}
		if (trimmed.startsWith("~/")) {
			return path.join(os.homedir(), trimmed.slice(2));
		}
		return path.normalize(trimmed);
	}
}
