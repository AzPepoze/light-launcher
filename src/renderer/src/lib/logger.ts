/**
 * Shared logger utility for frontend
 * Provides consistent logging with prefixes and timestamps
 */

type LogLevel = "debug" | "info" | "warn" | "error";

interface LogEntry {
	timestamp: string;
	level: LogLevel;
	module: string;
	message: string;
	data?: unknown;
}

const isDevelopment = import.meta.env.DEV;

/**
 * Format a log entry with timestamp and module
 */
function formatLog(level: LogLevel, module: string, message: string): string {
	const timestamp = new Date().toLocaleTimeString();
	return `[${timestamp}] [${module}] ${message}`;
}

/**
 * Internal function to log with different levels
 */
function log(level: LogLevel, module: string, message: string, data?: unknown) {
	const formatted = formatLog(level, module, message);

	switch (level) {
		case "debug":
			if (isDevelopment) {
				console.debug(formatted, data);
			}
			break;
		case "info":
			console.log(formatted, data);
			break;
		case "warn":
			console.warn(formatted, data);
			break;
		case "error":
			console.error(formatted, data);
			break;
	}
}

/**
 * Create a logger instance for a specific module
 */
export function createLogger(module: string) {
	return {
		debug: (message: string, data?: unknown) =>
			log("debug", module, message, data),
		info: (message: string, data?: unknown) =>
			log("info", module, message, data),
		warn: (message: string, data?: unknown) =>
			log("warn", module, message, data),
		error: (message: string, data?: unknown) =>
			log("error", module, message, data),
	};
}

/**
 * Global logger for generic logging
 */
export const logger = {
	debug: (module: string, message: string, data?: unknown) =>
		log("debug", module, message, data),
	info: (module: string, message: string, data?: unknown) =>
		log("info", module, message, data),
	warn: (module: string, message: string, data?: unknown) =>
		log("warn", module, message, data),
	error: (module: string, message: string, data?: unknown) =>
		log("error", module, message, data),
};
