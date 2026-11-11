import type { ScannedFolderGroup } from "@shared";

export const UNLIMITED_DEPTH = -1;
const MIN_DEPTH = 1;
const MAX_DEPTH = 999;

export type DepthResult = { ok: true; depth: number } | { ok: false; message: string };

/** Reads the scan depth typed by the user: -1 (no limit) or a whole number from 1 to 999. */
export function validateDepth(text: string): DepthResult {
	const trimmed = text.trim();
	const depth = Number(trimmed);
	if (trimmed === "" || !Number.isInteger(depth)) {
		return { ok: false, message: "Enter a whole number." };
	}
	if (depth === UNLIMITED_DEPTH || (depth >= MIN_DEPTH && depth <= MAX_DEPTH)) {
		return { ok: true, depth };
	}
	return { ok: false, message: "Use -1 for no limit, or a number from 1 to 999." };
}

/** The message for a depth that cannot be saved, or an empty string when it can. */
export function depthError(text: string): string {
	const result = validateDepth(text);
	return result.ok === false ? result.message : "";
}

/** Splits comma-separated ignore names, drops blanks and repeats (case-insensitive). */
export function normalizeExcludeNames(text: string): string[] {
	const seen = new Set<string>();
	const names: string[] = [];
	for (const raw of text.split(",")) {
		const name = raw.trim();
		const key = name.toLowerCase();
		if (name === "" || seen.has(key)) continue;
		seen.add(key);
		names.push(name);
	}
	return names;
}

export function excludeNamesText(names: string[]): string {
	return names.join(", ");
}

/** A key that treats the same folder with another case or a trailing slash as one folder. */
export function folderKey(path: string): string {
	return path.trim().replace(/[\\/]+$/, "").toLowerCase();
}

/** Returns the message for a folder that cannot be added, or null when it can. */
export function folderInputError(path: string, watched: string[]): string | null {
	const key = folderKey(path);
	if (key === "") return "Enter a folder path.";
	if (watched.some((folder) => folderKey(folder) === key)) {
		return "This folder is already watched.";
	}
	return null;
}

/** Counts the games a scan found under one watched folder, or null when it was not scanned. */
export function gamesInFolder(groups: ScannedFolderGroup[], folder: string): number | null {
	const key = folderKey(folder);
	const group = groups.find((item) => folderKey(item.folderPath) === key);
	return group ? group.games.length : null;
}

export function gamesFoundText(count: number): string {
	return `Found ${count} ${count === 1 ? "game" : "games"}`;
}
