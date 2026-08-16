import { getCategoryColor, LEVEL_COLORS, type LogLevel, type LogEntry } from "@shared";

const isDevelopment = import.meta.env.DEV;

function log(level: LogLevel, category: string, message: string, data?: unknown) {
	if (level === "debug" && !isDevelopment) {
		return;
	}

	const catColor = getCategoryColor(category);
	const lvlColor = LEVEL_COLORS[level];
	const time = new Date().toLocaleTimeString();

	const catStyle = `background: ${catColor.cssBg}; color: ${catColor.hex}; border: 1px solid ${catColor.cssBorder}; border-radius: 4px; padding: 1px 6px; font-weight: 700; font-family: monospace; font-size: 0.9em;`;
	const lvlStyle = `background: ${lvlColor.bg}; color: ${lvlColor.hex}; border-radius: 4px; padding: 1px 4px; font-weight: 800; font-family: monospace; font-size: 0.85em;`;
	const timeStyle = `color: rgba(255, 255, 255, 0.4); font-size: 0.85em; font-family: monospace;`;
	const msgStyle = `color: inherit; font-weight: 500;`;

	const prefix = `%c${time}%c %c${level.toUpperCase().padEnd(5)}%c %c[${category}]%c %c${message}`;

	const styles = [timeStyle, "", lvlStyle, "", catStyle, "", msgStyle];

	switch (level) {
		case "error":
			if (data !== undefined) {
				console.error(prefix, ...styles, data);
			} else {
				console.error(prefix, ...styles);
			}
			break;
		case "warn":
			if (data !== undefined) {
				console.warn(prefix, ...styles, data);
			} else {
				console.warn(prefix, ...styles);
			}
			break;
		default:
			if (data !== undefined) {
				console.log(prefix, ...styles, data);
			} else {
				console.log(prefix, ...styles);
			}
			break;
	}
}

export function createLogger(category: string) {
	return {
		debug: (message: string, data?: unknown) => log("debug", category, message, data),
		info: (message: string, data?: unknown) => log("info", category, message, data),
		warn: (message: string, data?: unknown) => log("warn", category, message, data),
		error: (message: string, data?: unknown) => log("error", category, message, data)
	};
}

export const logger = {
	debug: (category: string, message: string, data?: unknown) =>
		log("debug", category, message, data),
	info: (category: string, message: string, data?: unknown) => log("info", category, message, data),
	warn: (category: string, message: string, data?: unknown) => log("warn", category, message, data),
	error: (category: string, message: string, data?: unknown) =>
		log("error", category, message, data)
};
