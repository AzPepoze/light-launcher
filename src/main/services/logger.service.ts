import fs from "fs/promises";
import fsSync from "fs";
import path from "path";
import { BrowserWindow } from "electron";
import {
	getCategoryColor,
	LEVEL_COLORS,
	ANSI_RESET,
	ANSI_DIM,
	ANSI_BOLD,
	type LogLevel,
	type LogEntry
} from "../../shared/logger";
import { PathsService } from "./paths.service";

export class LoggerService {
	private static logFilePath: string = "";
	private static writeQueue: Promise<void> = Promise.resolve();

	static getLogFilePath(): string {
		if (!this.logFilePath) {
			this.logFilePath = path.join(PathsService.getLogsDirectory(), "launcher.log");
		}
		return this.logFilePath;
	}

	private static formatTimestamp(d = new Date()): string {
		const pad = (n: number, z = 2) => String(n).padStart(z, "0");
		const hours = pad(d.getHours());
		const mins = pad(d.getMinutes());
		const secs = pad(d.getSeconds());
		const ms = pad(d.getMilliseconds(), 3);
		return `${hours}:${mins}:${secs}.${ms}`;
	}

	private static async appendToFile(text: string): Promise<void> {
		const logFile = this.getLogFilePath();
		const dir = path.dirname(logFile);

		this.writeQueue = this.writeQueue.then(async () => {
			try {
				if (!fsSync.existsSync(dir)) {
					await fs.mkdir(dir, { recursive: true });
				}
				await fs.appendFile(logFile, text + "\n", "utf-8");
			} catch (err) {
				// Don't crash on log write failure
			}
		});
	}

	static log(level: LogLevel, category: string, message: string, ...args: unknown[]): void {
		const timestamp = this.formatTimestamp();
		const catColor = getCategoryColor(category);
		const lvlColor = LEVEL_COLORS[level];

		const entry: LogEntry = {
			timestamp,
			level,
			category,
			message,
			args: args.length > 0 ? args : undefined
		};

		// 1. CLI Terminal Output (with pastel category and level colors)
		const catBadge = `${catColor.ansi}${ANSI_BOLD}[${category}]${ANSI_RESET}`;
		const lvlBadge = `${lvlColor.ansi}${level.toUpperCase().padEnd(5)}${ANSI_RESET}`;
		const timeBadge = `${ANSI_DIM}${timestamp}${ANSI_RESET}`;

		if (args.length > 0) {
			console.log(`${timeBadge} ${lvlBadge} ${catBadge} ${message}`, ...args);
		} else {
			console.log(`${timeBadge} ${lvlBadge} ${catBadge} ${message}`);
		}

		// 2. File Output (clean formatted plain text)
		const extra = args.length > 0 ? " " + JSON.stringify(args) : "";
		const plainLog = `[${timestamp}] [${level.toUpperCase().padEnd(5)}] [${category}] ${message}${extra}`;
		this.appendToFile(plainLog);

		// 3. UI WebContents Relay (if renderer windows are active)
		try {
			for (const win of BrowserWindow.getAllWindows()) {
				if (!win.isDestroyed()) {
					win.webContents.send("event:log", entry);
				}
			}
		} catch {}
	}

	static debug(category: string, message: string, ...args: unknown[]): void {
		this.log("debug", category, message, ...args);
	}

	static info(category: string, message: string, ...args: unknown[]): void {
		this.log("info", category, message, ...args);
	}

	static warn(category: string, message: string, ...args: unknown[]): void {
		this.log("warn", category, message, ...args);
	}

	static error(category: string, message: string, ...args: unknown[]): void {
		this.log("error", category, message, ...args);
	}
}
