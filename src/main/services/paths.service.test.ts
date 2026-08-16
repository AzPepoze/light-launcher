import { describe, it, expect, afterEach } from "vitest";
import path from "path";
import os from "os";
import { PathsService } from "./paths.service";

describe("PathsService", () => {
	const originalXdg = process.env.XDG_CONFIG_HOME;

	afterEach(() => {
		process.env.XDG_CONFIG_HOME = originalXdg;
	});

	it("resolves base directory using XDG_CONFIG_HOME when present", () => {
		process.env.XDG_CONFIG_HOME = "/custom/config";
		expect(PathsService.getBaseDirectory()).toBe("/custom/config/light-launcher");
	});

	it("falls back to ~/.config/light-launcher when XDG_CONFIG_HOME is empty", () => {
		delete process.env.XDG_CONFIG_HOME;
		expect(PathsService.getBaseDirectory()).toBe(path.join(os.homedir(), ".config/light-launcher"));
	});

	it("computes config, prefix, and logs directories accurately", () => {
		const base = PathsService.getBaseDirectory();
		expect(PathsService.getConfigDirectory()).toBe(path.join(base, "games"));
		expect(PathsService.getPrefixBaseDirectory()).toBe(path.join(base, "prefixes"));
		expect(PathsService.getLogsDirectory()).toBe(path.join(base, "logs"));
	});

	it("computes executable config path by id or name fallback", () => {
		const configDir = PathsService.getConfigDirectory();
		expect(PathsService.getExecutableConfigPath("MyGame", "id123")).toBe(
			path.join(configDir, "id123")
		);
		expect(PathsService.getExecutableConfigPath("MyGame", "")).toBe(path.join(configDir, "MyGame"));
	});

	it("resolves preload and renderer paths", () => {
		expect(PathsService.getPreloadPath()).toContain("preload");
		expect(PathsService.getRendererPath()).toContain("renderer");
	});

	it("expands tilde paths correctly", () => {
		const home = os.homedir();
		expect(PathsService.expandPath("~")).toBe(home);
		expect(PathsService.expandPath("~/games/steam")).toBe(path.join(home, "games/steam"));
		expect(PathsService.expandPath("/absolute/path")).toBe("/absolute/path");
		expect(PathsService.expandPath("")).toBe("");
	});
});
