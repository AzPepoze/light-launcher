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

	static getPrefixBaseDirectory(): string {
		return path.join(this.getBaseDirectory(), PrefixesDirName);
	}

	static getLogsDirectory(): string {
		return path.join(this.getBaseDirectory(), LogsDirName);
	}

	static getAppSettingsPath(): string {
		return path.join(this.getBaseDirectory(), "settings.json");
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

	static getPrefixConfigPath(prefixName: string): string {
		return path.join(this.getPrefixBaseDirectory(), prefixName, "light-launcher.json");
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

	static expandPath(targetPath: string): string {
		if (!targetPath) return "";
		if (targetPath === "~") {
			return os.homedir();
		}
		if (targetPath.startsWith("~/")) {
			return path.join(os.homedir(), targetPath.slice(2));
		}
		return targetPath;
	}
}
