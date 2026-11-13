import type { WireDownloadSettings } from "@lib/features/wire";

export type DownloadNumberKey =
	| "connections"
	| "maxActive"
	| "hostConnections"
	| "retries"
	| "speedLimitKBps";

export interface NumberRange {
	min: number;
	/** Missing means no upper limit. */
	max?: number;
}

/** The limits core/internal/downloads enforces. */
export const DOWNLOAD_NUMBER_RANGES: Record<DownloadNumberKey, NumberRange> = {
	connections: { min: 1, max: 32 },
	maxActive: { min: 1, max: 10 },
	hostConnections: { min: 1, max: 32 },
	retries: { min: 0, max: 10 },
	speedLimitKBps: { min: 0 }
};

/** The values the backend starts with. Mirrors downloads.DefaultSettings. */
export const DEFAULT_DOWNLOAD_SETTINGS: WireDownloadSettings = {
	connections: 8,
	maxActive: 3,
	hostConnections: 8,
	retries: 3,
	speedLimitKBps: 0,
	keepArchives: false
};

export const DOWNLOAD_SAVE_DELAY_MS = 300;

/** Keeps a number in its range and rounds it to a whole number. A value that is not finite becomes the minimum. */
export function clampDownloadNumber(key: DownloadNumberKey, value: number): number {
	const { min, max = Infinity } = DOWNLOAD_NUMBER_RANGES[key];
	if (!Number.isFinite(value)) return min;
	return Math.min(max, Math.max(min, Math.round(value)));
}

/** Clamps every number of the settings, so the page never shows a value the backend would refuse. */
export function clampDownloadSettings(settings: WireDownloadSettings): WireDownloadSettings {
	return {
		...settings,
		connections: clampDownloadNumber("connections", settings.connections),
		maxActive: clampDownloadNumber("maxActive", settings.maxActive),
		hostConnections: clampDownloadNumber("hostConnections", settings.hostConnections),
		retries: clampDownloadNumber("retries", settings.retries),
		speedLimitKBps: clampDownloadNumber("speedLimitKBps", settings.speedLimitKBps)
	};
}

/** Reads the text of a number field. Null means the field is empty or not a number. */
export function parseDownloadNumber(text: string): number | null {
	if (text.trim() === "") return null;
	const value = Number(text);
	return Number.isNaN(value) ? null : value;
}

/** The settings with one number taken from its field text, or null when the text gives no number. */
export function withDownloadNumber(
	settings: WireDownloadSettings,
	key: DownloadNumberKey,
	text: string
): WireDownloadSettings | null {
	const value = parseDownloadNumber(text);
	if (value === null) return null;
	return { ...settings, [key]: clampDownloadNumber(key, value) };
}
