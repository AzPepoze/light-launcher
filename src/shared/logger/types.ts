export type LogLevel = "debug" | "info" | "warn" | "error";

export interface LogColor {
	name: string;
	ansi: string;
	hex: string;
	cssBg: string;
	cssBorder: string;
}

export interface LogEntry {
	timestamp: string;
	level: LogLevel;
	category: string;
	message: string;
	args?: unknown[];
}
