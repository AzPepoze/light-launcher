import type { WireHomeScan } from "./features/wire";
import type { GameInfo, ScannedFolderGroup } from "@shared";

export type HomeScan =
	| { state: "running"; done: number; total: number; found: number; current: string }
	| { state: "done"; found: number };

/** Maps a home.scan payload to the Home view model. Returns null for a malformed payload. */
export function toHomeScan(payload: unknown): HomeScan | null {
	if (typeof payload !== "object" || payload === null) return null;
	const wire = payload as Partial<WireHomeScan>;
	const found = typeof wire.found === "number" ? wire.found : 0;
	if (wire.state === "done") return { state: "done", found };
	if (wire.state !== "running") return null;
	if (typeof wire.done !== "number" || typeof wire.total !== "number") return null;
	const current = typeof wire.current === "string" ? wire.current.trim() : "";
	return { state: "running", done: wire.done, total: wire.total, found, current };
}

/** Reads the finished folder carried by a home.scan event. Returns null when there is none or it is malformed. */
export function toScanGroup(payload: unknown): ScannedFolderGroup | null {
	if (typeof payload !== "object" || payload === null) return null;
	const group = (payload as { group?: unknown }).group;
	if (typeof group !== "object" || group === null) return null;
	const wire = group as { folderPath?: unknown; folderName?: unknown; games?: unknown };
	if (typeof wire.folderPath !== "string" || typeof wire.folderName !== "string") return null;
	const games = Array.isArray(wire.games) ? (wire.games as GameInfo[]) : [];
	return { folderPath: wire.folderPath, folderName: wire.folderName, games };
}
