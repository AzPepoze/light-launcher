import { app, BrowserWindow, shell } from "electron";
import path from "path";
import fs from "fs";
import { registerIpcHandlers } from "./ipc";
import { ConfigService } from "./services/config.service";
import { AppService } from "./services/app.service";
import { PathsService } from "./services/paths.service";
import { MigrationService } from "./services/migration.service";

let mainWindow: BrowserWindow | null = null;

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
