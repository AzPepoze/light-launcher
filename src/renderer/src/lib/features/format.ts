const UNITS = ["B", "KB", "MB", "GB", "TB"];

export function bytes(value: number): string {
	if (!Number.isFinite(value) || value <= 0) return "0 B";
	const unit = Math.min(Math.floor(Math.log(value) / Math.log(1024)), UNITS.length - 1);
	const scaled = value / 1024 ** unit;
	const digits = unit === 0 || scaled >= 10 ? 0 : 1;
	return `${scaled.toFixed(digits)} ${UNITS[unit]}`;
}

export function speed(bytesPerSecond: number): string {
	return `${bytes(bytesPerSecond)}/s`;
}

/* Always one decimal for KB and above; a missing value renders as a dash. */
export function formatBytes(value: number | null | undefined): string {
	if (value == null) return "--";
	if (value < 1024) return `${value} B`;
	if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
	if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`;
	return `${(value / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

export function percent(done: number, total: number): number {
	if (total <= 0) return 0;
	return Math.min(100, Math.max(0, Math.round((done / total) * 100)));
}

export function eta(seconds: number | null): string {
	if (seconds === null || !Number.isFinite(seconds) || seconds < 0) return "--";
	if (seconds < 60) return `${Math.round(seconds)}s`;
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes}m ${String(Math.round(seconds % 60)).padStart(2, "0")}s`;
	const hours = Math.floor(minutes / 60);
	return `${hours}h ${String(minutes % 60).padStart(2, "0")}m`;
}

export function playtime(seconds: number): string {
	if (seconds < 60) return "0 min";
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes} min`;
	const hours = Math.floor(minutes / 60);
	return `${hours}h ${minutes % 60}m`;
}

export function ago(iso: string | null, now: number = Date.now()): string {
	if (!iso) return "never";
	const time = Date.parse(iso);
	if (Number.isNaN(time)) return "never";
	const minutes = Math.floor((now - time) / 60_000);
	if (minutes < 1) return "just now";
	if (minutes < 60) return `${minutes}m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours}h ago`;
	return `${Math.floor(hours / 24)}d ago`;
}
