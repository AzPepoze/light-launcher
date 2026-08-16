import { app, BrowserWindow, shell } from "electron";
import path from "path";
import fs from "fs";
import { registerIpcHandlers } from "./ipc";
import { ConfigService } from "./services/config.service";
import { AppService } from "./services/app.service";

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

	const preloadPath = path.join(__dirname, "../preload/index.js");

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
			preload: preloadPath,
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
	const devServerUrl = process.env.VITE_DEV_SERVER_URL;
	if (devServerUrl) {
		await mainWindow.loadURL(devServerUrl);
	} else {
		const rendererDist = path.join(__dirname, "../renderer/index.html");
		if (fs.existsSync(rendererDist)) {
			await mainWindow.loadFile(rendererDist);
		} else {
			// Fallback path in development
			const fallbackDist = path.join(__dirname, "../../renderer/dist/index.html");
			if (fs.existsSync(fallbackDist)) {
				await mainWindow.loadFile(fallbackDist);
			} else {
				console.error("Could not find renderer index.html at", rendererDist);
			}
		}
	}

	mainWindow.on("closed", () => {
		mainWindow = null;
	});

	return mainWindow;
}

// App lifecycle
app.whenReady().then(async () => {
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
