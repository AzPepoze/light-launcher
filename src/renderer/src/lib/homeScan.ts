import type { WireHomeScan } from "./features/wire";

export type HomeScan =
	| { state: "running"; done: number; total: number; current: string }
	| { state: "done" };

/** Maps a home.scan payload to the Home view model. Returns null for a malformed payload. */
export function toHomeScan(payload: unknown): HomeScan | null {
	if (typeof payload !== "object" || payload === null) return null;
	const wire = payload as Partial<WireHomeScan>;
	if (wire.state === "done") return { state: "done" };
	if (wire.state !== "running") return null;
	if (typeof wire.done !== "number" || typeof wire.total !== "number") return null;
	const current = typeof wire.current === "string" ? wire.current.trim() : "";
	return { state: "running", done: wire.done, total: wire.total, current };
}
