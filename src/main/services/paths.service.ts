import path from "path";
import os from "os";
import { AppName, ConfigDirName, PrefixesDirName, LogsDirName } from "../../shared/constants";

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
