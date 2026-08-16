import { app, BrowserWindow, shell } from "electron";
import path from "path";
import fs from "fs";
import { registerIpcHandlers } from "./ipc";
import { ConfigService } from "./services/config.service";
import { AppService } from "./services/app.service";
import { PathsService } from "./services/paths.service";
import { MigrationService } from "./services/migration.service";
import { LoggerService } from "./services/logger.service";

let mainWindow: BrowserWindow | null = null;

function initPlatformFlags(): void {
	// 1. Check for custom flags file (e.g. ~/.config/light-launcher-flags.conf)
	const flagsFile = PathsService.getFlagsFilePath();
	let hasOzoneInFlagsFile = false;

	if (fs.existsSync(flagsFile)) {
		try {
			const content = fs.readFileSync(flagsFile, "utf-8");
			const lines = content.split("\n");
			for (const line of lines) {
				const trimmed = line.trim();
				if (!trimmed || trimmed.startsWith("#")) continue;

				if (trimmed.startsWith("--ozone-platform")) {
					hasOzoneInFlagsFile = true;
				}

				if (trimmed.startsWith("--")) {
					const withoutDash = trimmed.slice(2);
					const eqIdx = withoutDash.indexOf("=");
					if (eqIdx !== -1) {
						app.commandLine.appendSwitch(withoutDash.slice(0, eqIdx), withoutDash.slice(eqIdx + 1));
					} else {
						app.commandLine.appendSwitch(withoutDash);
					}
				}
			}
			LoggerService.info("Platform", `Applied custom flags from ${flagsFile}`);
		} catch (err) {
			LoggerService.error("Platform", `Failed to read flags file: ${err}`);
		}
	}

	// 2. If ozone platform was not explicitly set via flags file, determine from settings / CLI
	if (!hasOzoneInFlagsFile) {
		let nativeWayland = false;
		const settingsPath = PathsService.getAppSettingsPath();
		if (fs.existsSync(settingsPath)) {
			try {
				const raw = fs.readFileSync(settingsPath, "utf-8");
				const parsed = JSON.parse(raw);
				nativeWayland = Boolean(parsed.NativeWayland);
			} catch (err) {
				LoggerService.error("Platform", `Failed to read settings for Ozone configuration: ${err}`);
			}
		}

		const isWaylandForced = process.argv.includes("--wayland");
		const isX11Forced = process.argv.includes("--x11");

		const platform = (nativeWayland || isWaylandForced) && !isX11Forced ? "wayland" : "x11";

		app.commandLine.appendSwitch("ozone-platform", platform);
		if (platform === "wayland") {
			app.commandLine.appendSwitch("enable-features", "UseOzonePlatform,WaylandWindowDecorations");
		}

		LoggerService.info("Platform", `Ozone platform initialized: ${platform}`, {
			nativeWayland,
			isWaylandForced,
			isX11Forced
		});
	}
}

initPlatformFlags();

function parseCliArgs(): void {
	const args = process.argv.slice(app.isPackaged ? 1 : 2);
	let launcherPath = "";
	let gamePath = "";
	let editLsfg = false;

	for (const arg of args) {
		if (arg === "--edit-lsfg") {
			editLsfg = true;
		} else if (!arg.startsWith("--") && fs.existsSync(arg)) {
			launcherPath = path.resolve(arg);
			gamePath = path.resolve(arg);
		}
	}

	AppService.setInitialArgs(launcherPath, gamePath, editLsfg);
}

async function createWindow(): Promise<BrowserWindow> {
	const settings = await ConfigService.loadAppSettings();
	const isTransparent = settings.TransparentMode ?? true;

	mainWindow = new BrowserWindow({
		title: "LightLauncher",
		width: 1024,
		height: 768,
		minWidth: 800,
		minHeight: 600,
		transparent: isTransparent,
		backgroundColor: isTransparent ? "#00000000" : "#18181b",
		frame: true,
		autoHideMenuBar: true,
		webPreferences: {
			preload: PathsService.getPreloadPath(),
			contextIsolation: true,
			nodeIntegration: false,
			sandbox: false
		}
	});

	// Handle external link clicks securely
	mainWindow.webContents.setWindowOpenHandler((details) => {
		shell.openExternal(details.url);
		return { action: "deny" };
	});

	// Load renderer
	if (!app.isPackaged) {
		const devUrl = process.env.VITE_DEV_SERVER_URL || "http://localhost:9245";
		await mainWindow.loadURL(devUrl);
	} else {
		await mainWindow.loadFile(PathsService.getRendererPath());
	}

	mainWindow.on("closed", () => {
		mainWindow = null;
	});

	return mainWindow;
}

// App lifecycle
app.whenReady().then(async () => {
	await MigrationService.migrateIfNeeded();
	parseCliArgs();
	registerIpcHandlers();
	await createWindow();

	app.on("activate", () => {
		if (BrowserWindow.getAllWindows().length === 0) {
			createWindow();
		}
	});
});

app.on("window-all-closed", () => {
	if (process.platform !== "darwin") {
		app.quit();
	}
});
